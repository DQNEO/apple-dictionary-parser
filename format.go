package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/beevik/etree"
)

const SPLT = "\t"

type RawEntry struct {
	Title string
	Body  []byte
}

type ParsedEntry struct {
	Title       string
	HG          *etree.Element
	SG          *etree.Element
	Phrases     *etree.Element
	PhVerbs     *etree.Element
	Derivatives *etree.Element
	Etym        *etree.Element
	Note        *etree.Element
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

	//println(len(entries))
	switch *flagMode {
	case "html":
		words := getWords(*flagWords, *flagWordsFile)
		renderHTML(entries, words)
	case "htmlsplit":
		outDir := flag.Arg(1)
		var letters = [...]byte{'0', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z'}
		var files = make(map[byte]*os.File) // e.g. "a" -> File("out/a.html")
		for _, letter := range letters {
			fname := fmt.Sprintf("%s/%s.html", outDir, letter)
			f, err := os.Create(fname)
			if err != nil {
				panic(err)
			}
			defer f.Close()
			files[letter] = f
			f.Write([]byte(GenHtmlHeader("NOAD - " + string(letter+0x20))))
		}
		for _, ent := range entries {
			t := ent.Title[0]
			f, found := files[t]
			if !found {
				f = files[0]
			}
			f.Write(ent.Body)
		}

		fmt.Print([]byte(htmlFooter))
	case "text":
		renderText(entries)
	default:
		panic("Invalid mode")
	}
}

func getWords(csv string, file string) []string {
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

func renderHTML(entries []*RawEntry, words []string) {
	println("words=", words)
	var mapWords = make(map[string]bool, len(words))
	for _, w := range words {
		if len(w) > 0 {
			mapWords[strings.ToLower(w)] = true
		}
	}
	fmt.Print(GenHtmlHeader("NOAD HTML as one file"))
	for _, ent := range entries {
		if len(words) == 0 || mapWords[strings.ToLower(ent.Title)] {
			fmt.Println(string(ent.Body))
		}
	}

	fmt.Print(htmlFooter)
}

func renderText(entries []*RawEntry) {
	for _, ent := range entries {
		title := ent.Title
		doc := etree.NewDocument()
		err := doc.ReadFromBytes(ent.Body)
		if err != nil {
			panic(err)
		}

		child := doc.Child[0].(*etree.Element)
		numChildren := len(child.Child)
		if numChildren < 2 || 7 < numChildren {
			panic("Unexpected number of children")
		}
		pe := &ParsedEntry{Title: title}
		for _, ch := range child.Child {
			elm := ch.(*etree.Element)
			if elm.Tag != "span" {
				panic("unexpected tag:" + elm.Tag + " --- " + title)
			}
			class := elm.SelectAttr("class").Value
			switch class {
			case "hg x_xh0":
				pe.HG = elm
			case "sg":
				pe.SG = elm
			case "subEntryBlock x_xo0 t_phrases":
				pe.Phrases = elm
			case "subEntryBlock x_xo0 t_phrasalVerbs":
				pe.PhVerbs = elm
			case "subEntryBlock x_xo0 t_derivatives":
				pe.Derivatives = elm
			case "etym x_xo0":
				pe.Etym = elm
			case "note x_xo0":
				pe.Note = elm
			default:
				panic(fmt.Sprintf(`unexpected class: "%s" in entry "%s"`, class, title))
			}

		}
		asText(pe)
	}
}

func asText(pe *ParsedEntry) {
	fmt.Printf("---\n%s\n  %s\n  %s\n  %s\n  %s\n  %s\n  %s\n  %s\n",
		pe.Title, S(pe.HG), S(pe.SG), S(pe.Phrases), S(pe.PhVerbs), S(pe.Derivatives), S(pe.Etym), S(pe.Note))
}

func S(elm *etree.Element) string {
	if elm == nil {
		return ""
	}
	elms := elm.FindElements("./")
	var s string
	for _, e := range elms {
		s += e.Text()
	}
	return s
}
