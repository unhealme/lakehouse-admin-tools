package utils

import (
	"bufio"
	"encoding/csv"
	"io"
	"os"
	"sync"
)

type CsvSer interface {
	SerCsv() []string
}

type JsonSer interface {
	SerJson() []byte
}

type OutSer interface {
	CsvSer
	JsonSer
}

var (
	csvOutputFile    io.WriteCloser
	csvOutputWriter  *csv.Writer
	csvOutputHeaders []string

	jsonOutputFile   io.WriteCloser
	jsonOutputWriter *bufio.Writer
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

func CloseOutput() {
	if csvOutputWriter != nil {
		csvOutputWriter.Flush()
		if csvOutputFile != os.Stdout {
			csvOutputFile.Close()
		}
	}
	if jsonOutputWriter != nil {
		jsonOutputWriter.Flush()
		if jsonOutputFile != os.Stdout {
			jsonOutputFile.Close()
		}
	}
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

func WriteOutput(v OutSer) {
	if csvOutputWriter != nil {
		WriteCsv(v)
	}
	if jsonOutputWriter != nil {
		WriteJson(v)
	}
}
