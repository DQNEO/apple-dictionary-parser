package extracter

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"unsafe"

	"github.com/DQNEO/apple-dictionary-parser/extracter/raw"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

// Logic is borrowed from here: https://gist.github.com/josephg/5e134adf70760ee7e49d?permalink_comment_id=4554558#gistcomment-4554558
func parseBinaryFile(filePath string) [][]byte {
	var entries [][]byte
	data, err := os.ReadFile(filePath)
	check(err)
	r := bytes.NewReader(data)
	_, err = r.Seek(0x40, io.SeekStart)
	check(err)

	// read the limit marker
	var limitMarker = make([]byte, 4, 4)
	_, err = r.Read(limitMarker)
	check(err)
	var limit = *(*int32)(unsafe.Pointer(&limitMarker[0]))
	entireSize := int64(limit) + 0x40

	_, err = r.Seek(0x60, io.SeekStart)
	check(err)

	var entryId int
	var blockIdx int
	for {
		filePos, err := r.Seek(0, io.SeekCurrent)
		check(err)
		if filePos >= entireSize {
			// End of File
			return entries
		}
		blockIdx++
		var blockSizeMarker = make([]byte, 4)
		_, err = r.Read(blockSizeMarker)
		if err == io.EOF {
			fmt.Printf("reached EOF\n")
			return entries
		}
		var blockSize int32 = *(*int32)(unsafe.Pointer(&blockSizeMarker[0]))
		if blockSize == 0 {
			panic("block size is zero")
		}

		var block = make([]byte, blockSize)
		_, err = r.Read(block)
		check(err)

		r, err := zlib.NewReader(bytes.NewReader(block[8:]))
		check(err)
		blockContents, err := io.ReadAll(r)
		check(err)
		chunkPos := 0
		for chunkPos < len(blockContents) {
			entryId++
			chunkSize := *(*int32)(unsafe.Pointer(&blockContents[chunkPos]))
			chunkPos += 4
			entry := blockContents[chunkPos : chunkPos+int(chunkSize)]
			entries = append(entries, entry)
			chunkPos += int(chunkSize)
		}
		r.Close()
	}
}

const titleStartMarker = `d:title="`

func parseEntry(entry []byte) *raw.Entry {
	titleStart := bytes.Index(entry, []byte(titleStartMarker)) + len(titleStartMarker)
	titleLen := bytes.Index(entry[titleStart:], []byte(`"`))
	if titleLen == 0 || titleLen == -1 {
		// irregular entries whose title is empty
		return nil
	}

	title := entry[titleStart : titleStart+titleLen]

	return &raw.Entry{
		Title: string(title),
		Body:  entry,
	}
}

func ParseBinaryFile(filePath string) []*raw.Entry {
	var entries []*raw.Entry
	rawEntries := parseBinaryFile(filePath)
	for _, rawEntry := range rawEntries {
		entry := parseEntry(rawEntry)
		if entry == nil {
			// ignore irregular entries
			continue
		}
		entries = append(entries, entry)
	}
	return entries
}
