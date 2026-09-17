package utils

import (
	"bufio"
	"encoding/csv"
	"io"
	"os"
	"sync"

	"github.com/pterm/pterm"
)

type CsvSer interface {
	SerCsv() []string
}

type JsonSer interface {
	SerJson() []byte
}

type TableSer interface {
	SerTable() []string
}

type OutSer interface {
	CsvSer
	JsonSer
	TableSer
}

var (
	csvOutputFile    io.WriteCloser
	csvOutputWriter  *csv.Writer
	csvOutputHeaders []string

	jsonOutputFile   io.WriteCloser
	jsonOutputWriter *bufio.Writer

	tableOutputBuffer pterm.TableData
	tableOutputFile   io.WriteCloser
	tableOutputWriter *pterm.TablePrinter
)

var printCsvHeader = sync.OnceFunc(func() {
	if csvOutputHeaders != nil {
		if err := csvOutputWriter.Write(csvOutputHeaders); err != nil {
			panic(err)
		}
	}
})

func OpenCsvWriter(file string, headers []string) (err error) {
	if file == "" {
		csvOutputFile = os.Stdout
	} else if csvOutputFile, err = os.OpenFile(file,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0o644); err != nil {
		return
	}
	csvOutputWriter = csv.NewWriter(csvOutputFile)
	csvOutputHeaders = headers
	return
}

func OpenJsonWriter(file string) (err error) {
	if file == "" {
		jsonOutputFile = os.Stdout
	} else if jsonOutputFile, err = os.OpenFile(file,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0o644); err != nil {
		return
	}
	jsonOutputWriter = bufio.NewWriter(jsonOutputFile)
	return
}

func OpenTableWriter(file string, headers []string) (err error) {
	if file == "" {
		tableOutputFile = os.Stdout
	} else if tableOutputFile, err = os.OpenFile(file,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0o644); err != nil {
		return
	}
	tableOutputWriter = pterm.DefaultTable.WithHeaderRowSeparator("─").WithWriter(tableOutputFile)
	if len(headers) > 0 {
		tableOutputWriter = tableOutputWriter.WithHasHeader()
		tableOutputBuffer = append(tableOutputBuffer, headers)
	}
	return
}

func CloseOutput() {
	if csvOutputWriter != nil {
		csvOutputWriter.Flush()
		if mustClose(csvOutputFile) {
			csvOutputFile.Close()
		}
	}
	if jsonOutputWriter != nil {
		jsonOutputWriter.Flush()
		if mustClose(jsonOutputFile) {
			jsonOutputFile.Close()
		}
	}
	if tableOutputWriter != nil {
		tableOutputWriter.WithData(tableOutputBuffer).Render()
		if mustClose(tableOutputFile) {
			tableOutputFile.Close()
		}
	}
}

func mustClose(f io.Closer) bool {
	return f != os.Stdout && f != os.Stderr
}

func WriteCsv(v CsvSer) {
	printCsvHeader()
	if err := csvOutputWriter.Write(v.SerCsv()); err != nil {
		panic(err)
	}
}

func WriteJson(v JsonSer) {
	if _, err := jsonOutputWriter.Write(append(v.SerJson(), '\n')); err != nil {
		panic(err)
	}
}

func WriteTable(v TableSer) {
	tableOutputBuffer = append(tableOutputBuffer, v.SerTable())
}

func WriteOutput(v OutSer) {
	if csvOutputWriter != nil {
		WriteCsv(v)
	}
	if jsonOutputWriter != nil {
		WriteJson(v)
	}
	if tableOutputWriter != nil {
		WriteTable(v)
	}
}
