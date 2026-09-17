package commands

import (
	"bufio"
	"context"
	"iter"
	"os"
	"strings"

	json "github.com/goccy/go-json"
	"github.com/pterm/pterm"
	"github.com/unhealme/lakehouse-admin-tools/internal"
	"github.com/unhealme/lakehouse-admin-tools/pkg/arguments"
	"github.com/unhealme/lakehouse-admin-tools/pkg/utils"
)

const HiveBackupTableVersion = "2026.09.05-0"

func HiveBackupTable(logger *pterm.Logger, args *arguments.HiveBackupTableArgs) {
	logger.Debug("using backup table args.", logger.Args(internal.ToArgs(*args)...))
	c := args.HiveServerClient.SetMaxConnections(args.Concurrency)

	switch args.Format {
	case arguments.HiveBackupTableOutputCsv:
		headers := []string{"Db", "Table", "Ddl", "Desc"}
		if args.NoHeader {
			headers = nil
		}
		if err := utils.OpenCsvWriter(args.OutputFile, headers); err != nil {
			logger.Fatal("unable to open file to write.", logger.Args("file", args.OutputFile, "error", err))
		}
	case arguments.HiveBackupTableOutputJson:
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
	slot := utils.NewSlot(max(args.Concurrency, 1))
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
		for fetchedTables := range slot.MapValue(fetchTables, tableInputs, true) {
			for table := range fetchedTables {
				tables = append(tables, table)
			}
		}
		tables = utils.SliceDedup(tables)
	}

	var prog *pterm.ProgressbarPrinter
	if !args.NoProg && args.OutputFile != "" {
		ctx, done := context.WithCancel(context.Background())
		prog, _ = utils.NewProgressBar(ctx).WithTitle("Backup tables").WithTotal(len(tables)).Start()
		defer done()
	}
	fetchResult := func(t backupHiveTableResult) *backupHiveTableResult {
		var err error
		logArgs := logger.Args("db", t.Db, "table", t.Table)
		if t.Ddl, err = c.ShowCreateTable(t.Db, t.Table); err != nil {
			logger.Warn("unable to get ddl.", logArgs, logger.Args("error", err))
		}
		if t.Desc, err = c.DescribeFormattedTable(t.Db, t.Table); err != nil {
			logger.Warn("unable to describe.", logArgs, logger.Args("error", err))
		}
		if err == nil {
			logger.Debug("successfully backup table.", logArgs)
		}
		if prog != nil {
			prog.Increment()
		}
		return &t
	}
	for result := range slot.MapValue(fetchResult, tables, true) {
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

func (t backupHiveTableResult) SerTable() []string {
	return nil
}
