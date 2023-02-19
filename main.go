package main

import (
	"flag"
	"fmt"
	"github.com/DQNEO/apple-dictionary/cache"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/DQNEO/apple-dictionary/extracter/raw"
	"github.com/DQNEO/apple-dictionary/parser"
)

var flagMode = flag.String("mode", "", "output format (html or text)")
var flagWords = flag.String("words", "", "limit words in csv. Only for HTML mode ")
var flagWordsFile = flag.String("words-file", "", "limit words by the given file. Only for HTML mode ")
var flagCacheFilePath = flag.String("cache-file", cache.DEFAULT_PATH, "cache file path")

func main() {
	flag.Parse()
	entries := cache.LoadFromCacheFile(*flagCacheFilePath)
	//println("entries=", len(entries), *flagCacheFilePath)

	switch *flagMode {
	case "debug":
		parser.DebugWriter = os.Stderr
		selectWords := getSelectWordsMap(*flagWords, *flagWordsFile)
		renderForDebug(entries, selectWords)
	case "etym":
		outDir := flag.Arg(0)
		if outDir == "" {
			panic("invalid argument")
		}
		selectWords := getSelectWordsMap(*flagWords, *flagWordsFile)
		slice, mp := collectEtymology(entries, selectWords)
		//println(len(slice), len(mp))
		formatEtymologyToText(outDir, slice, mp)
		formatEtymologyToHTML(outDir, slice, mp)
	case "html":
		selectWords := getSelectWordsMap(*flagWords, *flagWordsFile)
		oFile := os.Stdout
		renderSingleHTML(oFile, entries, selectWords)
	case "htmlsplit":
		outDir := flag.Arg(0)
		if outDir == "" {
			panic("invalid argument")
		}
		selectWords := getSelectWordsMap(*flagWords, *flagWordsFile)
		renderSplitHTML(outDir, entries, selectWords)
	case "text":
		selectWords := getSelectWordsMap(*flagWords, *flagWordsFile)
		renderText(entries, selectWords)
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

func renderSplitHTML(outDir string, entries []*raw.Entry, selectWords SelectWords) {
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
		if len(selectWords) > 0 && !selectWords.HasKey(ent.Title) {
			continue
		}
		t := ent.Title[0]
		f, found := files[t]
		if !found {
			f = files[0]
		}
		renderEntry(f, ent.Title, ent.Body)
	}
}

func renderSingleHTML(w io.Writer, entries []*raw.Entry, selectWords SelectWords) {
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
	if len(e.FFWords) > 0 {
		ffnote := "<" + strings.Join(e.FFWords, ",") + ">"
		fields = append(fields, ffnote)
	}
	fields = append(fields, e.Note)
	return strings.Join(fields, " ")
}

func convEntryToText(ent *raw.Entry) string {
	title := ent.Title
	body := ent.Body
	et := parser.ParseEntry(title, body)
	//return fmt.Sprintf("%#v", et)
	return ToOneline(et)
}

type BackEtymLink struct {
	EngWord     string
	OriginWords []string
}

type EtymMap map[string][]string

// @TODO: make html, json and yaml formatter as well
const e2oFileName = "english2origin"
const o2eFileName = "origin2english"

const EtymStyle = `
.table { width: 100%; ...}
.table thead {
  background: #d3edeb;
}
.table th,.table td {
    border: 1px solid #ccc; padding: 10px;
}
.english {
    width:15em;
}
`

const EtymHTMLTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>NOAD Etymology</title>
<style>
%s
</style>
</head>
<body>
<h1>NOAD Etymology - English to Origins</h1>
<table class="table"> %s </table>
</body>
</html>
`

func formatEtymologyToHTML(outDir string, backEtymLinks []*BackEtymLink, forwardEtymMap EtymMap) {
	fileE2O, err := os.Create(fmt.Sprintf("%s/%s.html", outDir, e2oFileName))
	if err != nil {
		panic(err)
	}
	defer fileE2O.Close()
	var trs []string
	for _, bel := range backEtymLinks {
		if len(bel.OriginWords) == 0 {
			continue
		}
		// inline yaml
		trs = append(trs, fmt.Sprintf("<tr><td class=\"english\" id=\"%s\">%s</td><td>%s</td></tr>\n", bel.EngWord, bel.EngWord, strings.Join(bel.OriginWords, ", ")))
	}
	fmt.Fprintf(fileE2O, EtymHTMLTemplate, EtymStyle, strings.Join(trs, "\n"))
	var uniqFFs []string
	for k, _ := range forwardEtymMap {
		uniqFFs = append(uniqFFs, k)
	}
	sort.Strings(uniqFFs)
	fileO2E, err := os.Create(fmt.Sprintf("%s/%s.html", outDir, o2eFileName))
	fmt.Fprint(fileO2E, "<html>")

	if err != nil {
		panic(err)
	}
	defer fileO2E.Close()

	for _, ff := range uniqFFs {
		v := forwardEtymMap[ff]
		// inline yaml
		fmt.Fprintf(fileO2E, "%s:[%s]\n", ff, strings.Join(v, ","))
	}
}

func formatEtymologyToText(outDir string, backEtymLinks []*BackEtymLink, forwardEtymMap EtymMap) {
	fileE2O, err := os.Create(fmt.Sprintf("%s/%s.yml", outDir, e2oFileName))
	if err != nil {
		panic(err)
	}
	defer fileE2O.Close()
	fmt.Fprint(fileE2O, "---\n")
	for _, bel := range backEtymLinks {
		if len(bel.OriginWords) == 0 {
			continue
		}
		// inline yaml
		fmt.Fprintf(fileE2O, "%s:[%s]\n", bel.EngWord, strings.Join(bel.OriginWords, ","))
	}

	var uniqFFs []string
	for k, _ := range forwardEtymMap {
		uniqFFs = append(uniqFFs, k)
	}
	sort.Strings(uniqFFs)
	fileO2E, err := os.Create(fmt.Sprintf("%s/%s.yml", outDir, o2eFileName))
	fmt.Fprint(fileO2E, "---\n")

	if err != nil {
		panic(err)
	}
	defer fileO2E.Close()

	for _, ff := range uniqFFs {
		v := forwardEtymMap[ff]
		// inline yaml
		fmt.Fprintf(fileO2E, "%s:[%s]\n", ff, strings.Join(v, ","))
	}

}

func collectEtymology(entries []*raw.Entry, selectWords SelectWords) ([]*BackEtymLink, EtymMap) {
	var forwardEtymMap = make(EtymMap, len(entries))
	var allFF []string
	var backEtymLinks []*BackEtymLink
	for _, ent := range entries {
		if len(selectWords) > 0 && !selectWords.HasKey(ent.Title) {
			continue
		}
		e := parser.ParseEntry(ent.Title, ent.Body)
		if len(e.Etym) == 0 {
			continue
		}
		backEtymLinks = append(backEtymLinks, &BackEtymLink{
			EngWord:     ent.Title,
			OriginWords: e.FFWords,
		})
		for _, ff := range e.FFWords {
			forwardEtymMap[ff] = append(forwardEtymMap[ff], ent.Title)
			allFF = append(allFF, ff)
		}
	}

	return backEtymLinks, forwardEtymMap
}

func renderForDebug(entries []*raw.Entry, selectWords SelectWords) {
	var ffMap = make(map[string][]string, len(entries))
	var ffRevMap = make(map[string][]string, len(entries))
	for _, ent := range entries {
		if len(selectWords) > 0 && !selectWords.HasKey(ent.Title) {
			continue
		}
		e := parser.ParseEntry(ent.Title, ent.Body)
		ffMap[ent.Title] = e.FFWords
		for _, ff := range e.FFWords {
			ffRevMap[ff] = append(ffRevMap[ff], ent.Title)
		}
	}
	// Sort
	var ffMapKeys []string
	for k, _ := range ffMap {
		ffMapKeys = append(ffMapKeys, k)
	}
	sort.Strings(ffMapKeys)

	var ffRevMapKeys []string
	for k, _ := range ffRevMap {
		ffRevMapKeys = append(ffRevMapKeys, k)
	}

	sort.Strings(ffRevMapKeys)

	fmt.Printf("--- FF Map----\n")
	for _, k := range ffMapKeys {
		v := ffMap[k]
		fmt.Printf("[%s] %s\n", k, strings.Join(v, ","))
	}
	fmt.Printf("--- FF Rev Map----\n")
	for _, k := range ffRevMapKeys {
		v := ffRevMap[k]
		fmt.Printf("[%s] %s\n", k, strings.Join(v, ","))
	}
}

func renderText(entries []*raw.Entry, selectWords SelectWords) {
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
