package obs_analyze_utils

import (
	"bufio"
	"encoding/csv"
	"io"
	"os"
	"strconv"
	"sync"

	"github.com/goccy/go-json"
	"github.com/pterm/pterm"
	"github.com/unhealme/lakehouse-admin-tools/internal/obs"
	"github.com/unhealme/lakehouse-admin-tools/utils"
)

var (
	csvFile   io.WriteCloser
	csvWriter *csv.Writer

	jsonFile   io.WriteCloser
	jsonWriter *bufio.Writer
)

var printCsvHeader = sync.OnceFunc(func() {
	if err := csvWriter.Write([]string{
		"ObsPath",
		"RawSize",
		"Size",
		"DirCount",
		"FileCount",
	}); err != nil {
		panic(err)
	}
})

func OpenCsvWriter(file string) (err error) {
	if csvFile, err = os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644); err == nil {
		csvWriter = csv.NewWriter(csvFile)
	}
	return
}

func OpenJsonWriter(file string) (err error) {
	if jsonFile, err = os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644); err == nil {
		jsonWriter = bufio.NewWriter(jsonFile)
	}
	return
}

func CloseOutput() {
	if csvWriter != nil {
		csvWriter.Flush()
		csvFile.Close()
	}
	if jsonWriter != nil {
		jsonWriter.Flush()
		jsonFile.Close()
	}
}

func WriteOutput(stats obs.ObsPathAnalyzed) {
	if csvWriter != nil {
		printCsvHeader()
		if err := csvWriter.Write([]string{
			stats.URI(),
			strconv.FormatInt(stats.Size, 10),
			utils.FormatSize(stats.Size),
			strconv.FormatInt(int64(stats.DirCount), 10),
			strconv.FormatInt(int64(stats.FileCount), 10),
		}); err != nil {
			panic(err)
		}
	}
	if jsonWriter != nil {
		statsJson, _ := json.Marshal(stats)
		if _, err := jsonWriter.Write(append(statsJson, '\n')); err != nil {
			panic(err)
		}
	}
	pterm.Printf(
		"obs://%s/%s: size: %d (%s), objects: %d (%d dirs, %d files)\n",
		stats.Bucket, stats.Key,
		stats.Size, utils.FormatSize(stats.Size),
		stats.DirCount+stats.FileCount,
		stats.DirCount, stats.FileCount,
	)
}
