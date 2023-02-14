package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/DQNEO/apple-dictionary/parser"
)

const SPLT = "\t"

type RawEntry struct {
	Title string
	Body  []byte
}

func parseDumpFile(path string) []*RawEntry {
	var r []*RawEntry
	contents, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	lines := bytes.Split(contents, []byte{'\n'})
	for _, line := range lines[:] {
		if len(line) == 0 {
			// Possibly end of file
			continue
		}
		ttlBytes, rawBody, found := bytes.Cut(line, []byte(SPLT))
		if !found {
			panic("failed to Cut:" + (string(line)))
		}
		title := string(ttlBytes)
		e := &RawEntry{
			Title: title,
			Body:  rawBody,
		}
		r = append(r, e)
	}
	return r
}

var flagMode = flag.String("mode", "", "output format (html or text)")
var flagWords = flag.String("words", "", "limit words in csv. Only for HTML mode ")
var flagWordsFile = flag.String("words-file", "", "limit words by the given file. Only for HTML mode ")

func main() {
	flag.Parse()
	rawDumpFile := flag.Arg(0)
	entries := parseDumpFile(rawDumpFile)

	parser.DebugWriter = io.Discard
	switch *flagMode {
	case "debug":
		parser.DebugWriter = os.Stderr
		selectWords := getSelectWordsMap(*flagWords, *flagWordsFile)
		renderForDebug(entries, selectWords)
	case "html":
		selectWords := getSelectWordsMap(*flagWords, *flagWordsFile)
		oFile := os.Stdout
		renderSingleHTML(oFile, entries, selectWords)
	case "htmlsplit":
		outDir := flag.Arg(1)
		renderSplitHTML(outDir, entries)
	case "text":
		selectWords := getSelectWordsMap(*flagWords, *flagWordsFile)
		renderText(entries, selectWords)
		fmt.Println("<FFCollection>")
		for i, ff := range parser.FFWords {
			fmt.Printf("[%d] %s\n", i, ff)
		}
	default:
		panic("Invalid mode")
	}
}

type SelectWords map[string]bool

func (mp SelectWords) HasKey(w string) bool {
	return mp[strings.ToLower(w)]
}

func getSelectWordsMap(csv string, file string) SelectWords {
	words := getSelectWords(csv, file)
	var mapWords = make(SelectWords, len(words))
	for _, w := range words {
		if len(w) > 0 {
			mapWords[strings.ToLower(w)] = true
		}
	}
	return mapWords
}

func getSelectWords(csv string, file string) []string {
	if csv != "" && file != "" {
		panic("Please do not specify both words and words-file at the same time")
	}
	if csv != "" {
		return strings.Split(*flagWords, ",")
	}
	if file != "" {
		contents, err := os.ReadFile(file)
		if err != nil {
			panic(err)
		}
		return strings.Split(string(contents), "\n")
	}
	return nil
}

const closingTag = "</d:entry>"

func renderEntry(w io.Writer, title string, body []byte) {
	// trim "</d:entry>"
	bodyWithoutClosingTag := body[:len(body)-len(closingTag)]
	w.Write(bodyWithoutClosingTag)
	fmt.Fprintf(w, "<p class='external-links'>[ <a href='https://www.etymonline.com/word/%s' target='_blank'>etym</a> | <a href='https://www.google.com/search?tbm=isch&q=%s' target='_blank'>image</a> ]</p>\n", title, title)
	fmt.Fprintln(w, closingTag)
}

func renderSplitHTML(outDir string, entries []*RawEntry) {
	var letters = [...]byte{'0', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z'}
	var files = make(map[byte]*os.File) // e.g. "a" -> File("out/a.html")
	for _, letter := range letters {
		fname := fmt.Sprintf("%s/%s.html", outDir, string(letter))
		f, err := os.Create(fname)
		if err != nil {
			panic(err)
		}
		defer func(f *os.File) {
			f.Write([]byte(htmlFooter))
			f.Close()
		}(f)
		files[letter] = f
		htmlTitle := "NOAD - " + strings.ToUpper(string(letter))
		f.Write([]byte(GenHtmlHeader(htmlTitle, true)))
	}
	for _, ent := range entries {
		t := ent.Title[0]
		f, found := files[t]
		if !found {
			f = files[0]
		}
		renderEntry(f, ent.Title, ent.Body)
	}
}

func renderSingleHTML(w io.Writer, entries []*RawEntry, selectWords SelectWords) {
	htmlTitle := "NOAD HTML as a single file"
	fmt.Fprintln(w, GenHtmlHeader(htmlTitle, true))
	for _, ent := range entries {
		if len(selectWords) > 0 && !selectWords.HasKey(ent.Title) {
			continue
		}
		renderEntry(w, ent.Title, ent.Body)
	}
	fmt.Fprintln(w, htmlFooter)
}

// To human readable line
func ToOneline(e *parser.Entry) string {
	var fields []string
	fields = append(fields, "["+e.Title+"]")
	if e.Syll != "" {
		fields = append(fields, e.Syll)
	}
	if e.IPA != "" {
		fields = append(fields, "|"+e.IPA+"|")
	}
	fields = append(fields, "{ "+e.SG+" }")
	fields = append(fields, e.Phr)
	fields = append(fields, e.Phv)
	fields = append(fields, e.Drv)
	for _, ee := range e.Etym {
		s := strings.TrimSpace(string(ee))
		fields = append(fields, " "+s+" ")
	}
	fields = append(fields, e.Note)
	return strings.Join(fields, " ")
}

func convEntryToText(ent *RawEntry) string {
	title := ent.Title
	body := ent.Body
	et := parser.ParseEntry(title, body)
	//return fmt.Sprintf("%#v", et)
	return ToOneline(et)
}

func renderForDebug(entries []*RawEntry, selectWords SelectWords) {
	for _, ent := range entries {
		if len(selectWords) > 0 && !selectWords.HasKey(ent.Title) {
			continue
		}
		e := parser.ParseEntry(ent.Title, ent.Body)
		for _, etym := range e.Etym {
			fmt.Printf("%#v\n", etym)
		}
	}
}

func renderText(entries []*RawEntry, selectWords SelectWords) {
	for _, ent := range entries {
		if len(selectWords) > 0 && !selectWords.HasKey(ent.Title) {
			continue
		}
		et := parser.ParseEntry(ent.Title, ent.Body)
		s := ToOneline(et)
		fmt.Println(s)
	}
}

func assert(cnd bool, expect string) {
	if !cnd {
		panic("Assertion failed. Expect " + expect)
	}
}
