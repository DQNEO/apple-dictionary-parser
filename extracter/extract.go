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
	var chunks [][]byte
	r, err := os.Open(filePath)
	check(err)
	_, err = r.Seek(0x40, io.SeekStart)
	check(err)

	// read the entire binary size
	var limitMarker = make([]byte, 4, 4)
	_, err = r.Read(limitMarker)
	check(err)
	var limitP = (*int32)(unsafe.Pointer(&limitMarker[0]))

	_, err = r.Seek(0x60, io.SeekStart)
	check(err)
	var entryId int
	var blockIdx int
	for {
		pos, err := r.Seek(0, io.SeekCurrent)
		check(err)
		if pos >= int64(*limitP)+0x40 {
			return chunks
		}
		blockIdx++
		var blockSizeBin = make([]byte, 4, 4)
		_, err = r.Read(blockSizeBin)
		if err == io.EOF {
			fmt.Printf("reached EOF\n")
			return chunks
		}
		var sz *int32 = (*int32)(unsafe.Pointer(&blockSizeBin[0]))
		if *sz == 0 {
			panic("block size is zero")
		}

		var body = make([]byte, *sz)
		_, err = r.Read(body)
		check(err)

		btsr := bytes.NewReader(body[8:])
		r, err := zlib.NewReader(btsr)
		check(err)
		buf, err := io.ReadAll(r)
		check(err)
		chunkPos := 0
		for chunkPos < len(buf) {
			entryId++
			var chunkSizeP *int32 = (*int32)(unsafe.Pointer(&buf[chunkPos]))
			chunkPos += 4
			entry := buf[chunkPos : chunkPos+int(*chunkSizeP)]
			chunks = append(chunks, entry)
			chunkPos += int(*chunkSizeP)
		}
		r.Close()
	}
}

func parseChunk(buf []byte) [][]byte {
	var entries [][]byte
	buf = buf[4:]
	for {
		idx := bytes.IndexByte(buf, '\n')
		if idx > -1 {
			entry := buf[0:idx]
			entries = append(entries, entry)
			if idx+5 >= len(buf) {
				return entries
			}
			buf = buf[idx+5:]
		} else {
			return entries
		}
	}
}

const titleStartMarker = `d:title="`

func parseEntry(entry []byte) *raw.Entry {
	titleStart := bytes.Index(entry, []byte(titleStartMarker)) + len(titleStartMarker)
	titleLen := bytes.Index(entry[titleStart:], []byte(`"`))
	if titleLen == 0 || titleLen == -1 {
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
		e := parseEntry(rawEntry)
		if e != nil {
			entries = append(entries, e)
		}
	}
	return entries
}
