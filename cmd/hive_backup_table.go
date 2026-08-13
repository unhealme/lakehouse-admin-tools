package cmd

import (
	"bufio"
	"iter"
	"os"
	"strings"

	"github.com/goccy/go-json"
	"github.com/pterm/pterm"
	cmd_args "github.com/unhealme/lakehouse-admin-tools/args"
	"github.com/unhealme/lakehouse-admin-tools/internal"
	"github.com/unhealme/lakehouse-admin-tools/utils"
)

const HiveBackupTableVersion = "2026.08.12-0"

func HiveBackupTable(logger *pterm.Logger, args *cmd_args.HiveBackupTableArgs) {
	logger.Debug("using backup table args.", logger.Args(internal.ToArgs(*args)...))
	c := args.HiveServerClient.SetMaxConnections(args.Concurrency)

	switch args.Format {
	case cmd_args.HiveBackupTableOutputCsv:
		headers := []string{"Db", "Table", "Ddl", "Desc"}
		if args.NoHeader {
			headers = nil
		}
		if err := utils.OpenCsvWriter(args.OutputFile, headers); err != nil {
			logger.Fatal("unable to open file to write.", logger.Args("file", args.OutputFile, "error", err))
		}
	case cmd_args.HiveBackupTableOutputJson:
		if err := utils.OpenJsonWriter(args.OutputFile); err != nil {
			logger.Fatal("unable to open file to write.", logger.Args("file", args.OutputFile, "error", err))
		}
	}
	defer utils.CloseOutput()

	tableInputs := args.Tables
	if args.InputFile != "" {
		file, err := os.Open(args.InputFile)
		if err != nil {
			logger.Fatal("unable to read input file.", logger.Args("name", args.InputFile, "error", err))
		}
		scanner := bufio.NewScanner(file)
		line := 1
		for scanner.Scan() {
			tableInputs = append(tableInputs, strings.TrimSpace(scanner.Text()))
			line++
		}
		file.Close()
		if scanner.Err() != nil {
			logger.Fatal("unable to read input file.", logger.Args("file", args.InputFile, "line", line))
		}
	}
	if len(tableInputs) < 1 {
		logger.Fatal("no table input is specified.")
	}
	tableInputs = utils.SliceDedup(tableInputs)

	var tables []backupHiveTableResult
	if args.Fixed {
		tables = make([]backupHiveTableResult, len(tableInputs))
		for _, table := range tableInputs {
			db, tbl, _ := strings.Cut(table, ".")
			tables = append(tables, backupHiveTableResult{Db: db, Table: tbl})
		}
	} else {
		fetchedDb := make(map[string][]string)
		fetchTables := func(table string) iter.Seq[backupHiveTableResult] {
			db, tbl, _ := strings.Cut(table, ".")
			var dbs []string
			if db != "" {
				var fetched bool
				var err error
				dbs, fetched = fetchedDb[db]
				if !fetched {
					if dbs, err = c.ShowDatabasesLike(db); err != nil {
						logger.Fatal("unable to get databases.", logger.Args("db", db, "error", err))
					}
				}
			} else {
				dbs = append(dbs, "")
			}
			return func(yield func(backupHiveTableResult) bool) {
				for _, db := range dbs {
					tbls, err := c.ShowTablesLike(db, tbl)
					if err != nil {
						logger.Warn("unable to get tables.", logger.Args("db", db, "table", tbl, "error", err))
						continue
					}
					for _, tbl := range tbls {
						if !yield(backupHiveTableResult{Db: db, Table: tbl}) {
							return
						}
					}
				}
			}
		}
		for fetchedTables := range utils.ParallelMapOrdered(fetchTables, tableInputs, args.Concurrency) {
			for table := range fetchedTables {
				tables = append(tables, table)
			}
		}
		tables = utils.SliceDedup(tables)
	}

	var prog *pterm.ProgressbarPrinter
	if !args.NoProg && args.OutputFile != "" {
		prog, _ = utils.NewProgressBar().WithTitle("Backup tables").WithTotal(len(tables)).Start()
		defer prog.Stop()
	}
	fetchResult := func(t backupHiveTableResult) *backupHiveTableResult {
		if ddl, err := c.ShowCreateTable(t.Db, t.Table); err != nil {
			logger.Warn("unable to get ddl.", logger.Args("db", t.Db, "table", t.Table))
		} else {
			t.Ddl = ddl
		}
		if desc, err := c.DescribeFormattedTable(t.Db, t.Table); err != nil {
			logger.Warn("unable to describe.", logger.Args("db", t.Db, "table", t.Table))
		} else {
			t.Desc = desc
		}
		if prog != nil {
			prog.Increment()
		}
		return &t
	}
	for result := range utils.ParallelMapOrdered(fetchResult, tables, args.Concurrency) {
		utils.WriteOutput(result)
	}
}

type backupHiveTableResult struct {
	Db    string `json:"db"`
	Table string `json:"table"`
	Ddl   string `json:"ddl"`
	Desc  string `json:"desc"`
}

func (t backupHiveTableResult) SerCsv() []string {
	return []string{t.Db, t.Table, t.Ddl, t.Desc}
}

func (t backupHiveTableResult) SerJson() (v []byte) {
	v, _ = json.Marshal(t)
	return
}
