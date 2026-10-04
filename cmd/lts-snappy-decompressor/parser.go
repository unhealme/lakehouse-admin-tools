package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/bytedance/sonic"
	parquet "github.com/parquet-go/parquet-go"
)

// Parallel Parser
// TODO: preserve input order
//
// func parser(in io.Reader, out io.Writer) error {
// 	enc, err := zstd.NewWriter(out)
// 	if err != nil {
// 		return err
// 	}
// 	closeEnc := sync.OnceValue(enc.Close)
// 	defer closeEnc()
//
// 	lines := make(chan []byte, 4)
// 	sc := bufio.NewScanner(in)
// 	go func() {
// 		var s []byte
// 		for sc.Scan() {
// 			lines <- append(s, sc.Bytes()...)
// 		}
// 		close(lines)
// 	}()
//
// 	var wg sync.WaitGroup
// 	cancel := make(chan struct{}, 1)
// 	parsed := make(chan string, 4)
// 	for range min(runtime.GOMAXPROCS(0), 4) {
// 		wg.Go(func() {
// 			for {
// 				select {
// 				case line, ok := <-lines:
// 					if !ok {
// 						return
// 					}
// 					parsed <- gjson.GetBytes(line, "message").Str
// 				case <-cancel:
// 					return
// 				}
// 			}
// 		})
// 	}
// 	go func() {
// 		wg.Wait()
// 		close(parsed)
// 	}()
//
// 	for msg := range parsed {
// 		if _, err := enc.Write(append([]byte(msg), '\n')); err != nil {
// 			cancel <- struct{}{}
// 			return err
// 		}
// 	}
// 	if err := sc.Err(); err != nil {
// 		return err
// 	}
// 	return closeEnc()
// }

type Message struct {
	Code          int32  `json:"code" parquet:"code,dict"`
	ContentLength int64  `json:"content_length" parquet:"content_length,delta"`
	EventType     string `json:"event_type" parquet:"event_type,dict"` // varchar(255)
	ProjectId     string `json:"project_id" parquet:"project_id,dict"` // char(32)
	ReadOnly      bool   `json:"read_only" parquet:"read_only,dict"`
	RecordTime    int64  `json:"record_time" parquet:"record_time,delta"`
	RequestId     string `json:"request_id" parquet:"request_id,delta"` // char(32)
	ResourceName  string `json:"resource_name" parquet:"resource_name,delta"`
	ResourceType  string `json:"resource_type" parquet:"resource_type,dict"` // varchar(255)
	ServiceType   string `json:"service_type" parquet:"service_type,dict"`   // varchar(255)
	SourceIp      string `json:"source_ip" parquet:"source_ip,dict"`         // varchar(255)
	Time          int64  `json:"time" parquet:"time,delta"`
	TotalTime     int64  `json:"total_time" parquet:"total_time,delta"`
	TraceId       string `json:"trace_id" parquet:"trace_id,delta"`        // char(36), primary
	TraceName     string `json:"trace_name" parquet:"trace_name,dict"`     // varchar(255)
	TraceRating   string `json:"trace_rating" parquet:"trace_rating,dict"` // varchar(255)
	TraceType     string `json:"trace_type" parquet:"trace_type,dict"`     // varchar(255)
	TrackerName   string `json:"tracker_name" parquet:"tracker_name,dict"`
	User          string `json:"user" parquet:"user,json"`
}

func parser(in io.Reader, out io.Writer) error {
	enc := parquet.NewGenericWriter[Message](out, parquet.Compression(&parquet.Zstd))
	closeEnc := sync.OnceValue(enc.Close)
	defer closeEnc()

	sc := bufio.NewScanner(in)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	var msg Message
	for l := 1; sc.Scan(); l++ {
		node, _ := sonic.Get(sc.Bytes(), "message")
		rawMsg, err := node.StrictString()
		if err == nil {
			sonic.UnmarshalString(rawMsg, &msg)
			if _, err := enc.Write([]Message{msg}); err != nil {
				return err
			}
		} else {
			fmt.Fprintf(os.Stderr, "empty message at line: %d\n", l)
		}
		if l%50000 == 0 {
			enc.Flush()
		}
	}

	if err := sc.Err(); err != nil {
		return err
	}
	return closeEnc()
}
