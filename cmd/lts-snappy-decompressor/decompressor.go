package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"sync"

	"github.com/klauspost/compress/snappy"
)

func decompress(in io.Reader, out io.Writer) error {
	var seekOnce sync.Once
	header := make([]byte, 12)
	chunkHeader := make([]byte, 4)

	var cBufSize uint32 = 8192
	compressedChunk := make([]byte, cBufSize)

	dBufSize := 32 * 1024
	decompressedBuf := make([]byte, dBufSize)
	for {
		if _, err := io.ReadFull(in, chunkHeader); err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				break
			}
			return err
		}

		chunkLen := binary.BigEndian.Uint32(chunkHeader)
		switch chunkLen {
		case 0x82534E41:
			if _, err := io.ReadFull(in, header); err != nil {
				return err
			}
			continue
		case 0:
			continue
		}

		if chunkLen > cBufSize {
			cBufSize = chunkLen
			compressedChunk = make([]byte, cBufSize)
		}
		if _, err := io.ReadFull(in, compressedChunk[:chunkLen]); err != nil {
			return err
		}

		decompressedChunk, err := snappy.Decode(decompressedBuf, compressedChunk[:chunkLen])
		if err != nil {
			return err
		}
		if newSize := len(decompressedChunk); newSize > dBufSize {
			dBufSize = newSize
			decompressedBuf = decompressedChunk
		}

		seekOnce.Do(func() {
			startIndex := bytes.IndexByte(decompressedChunk, '{')
			if startIndex > 0 && startIndex < 50 {
				decompressedChunk = decompressedChunk[startIndex:]
			}
		})
		if _, err := out.Write(decompressedChunk); err != nil {
			return err
		}
	}
	return nil
}
