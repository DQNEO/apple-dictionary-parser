package main

import (
	"flag"
	"fmt"
	"github.com/DQNEO/apple-dictionary-parser/cache"
	"github.com/DQNEO/apple-dictionary-parser/extracter"
	"github.com/DQNEO/apple-dictionary-parser/extracter/raw"
	"github.com/DQNEO/apple-dictionary-parser/finder"
	"github.com/DQNEO/apple-dictionary-parser/parser"
	"io"
	"os"
	"sort"
	"strings"
)

const version = "v0.0.5"

var flagCacheFilePath = flag.String("cache-file", cache.DEFAULT_PATH, "cache file path")
var flagDictFilePath = flag.String("dict-file", "", "dictionary file path")

func doVersion() {
	fmt.Println("apple-dictionary-parser version " + version)
}

func doFind() {
	dictDir, dictFilePath, defaultCssPath, err := finder.FindDictFile()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Found a dictionary resources.\n---\n")
	fmt.Printf("Directory:\n  '%s'\n", dictDir)
	fmt.Printf("Body file:\n  '%s'\n", dictFilePath)
	fmt.Printf("CSS file:\n  '%s'\n", defaultCssPath)
}

func doDump() {
	var dictFilePath string
	if *flagDictFilePath == "" {
		fmt.Printf("Searching Dictionary file ...\n")
		_, bodyFilePath, _, err := finder.FindDictFile()
		if err != nil {
			panic(err)
		}
		if bodyFilePath == "" {
			panic("File not bodyFilePath")
		}
		fmt.Printf("Dictionary file is bodyFilePath at '%s'\n", bodyFilePath)
		dictFilePath = bodyFilePath
	} else {
		dictFilePath = *flagDictFilePath
	}
	fmt.Printf("Extracting the dictionary file ...\n")
	entries := extracter.ParseBinaryFile(dictFilePath)
	oFile, err := os.Create(*flagCacheFilePath)
	if err != nil {
		panic(err)
	}
	cache.SaveEntries(oFile, entries)
	fmt.Printf("Dictonary raw data is successfully saved to: %s\n", *flagCacheFilePath)
}

func doEtym(args []string) {
	flag := flag.NewFlagSet("debug", flag.ExitOnError)
	var flagWords = flag.String("words", "", "limit words in csv. Only for HTML mode ")
	var flagWordsFile = flag.String("words-file", "", "limit words by the given file. Only for HTML mode ")
	flag.Parse(args)
	outDir := flag.Arg(0)
	if outDir == "" {
		panic("Please specify an output directory")
	}

	entries := cache.LoadFromCacheFile(*flagCacheFilePath)
	selectWords := getSelectWordsMap(*flagWords, *flagWordsFile)
	slice, mp := collectEtymology(entries, selectWords)
	//println(len(slice), len(mp))
	formatEtymologyToYAML(outDir, slice, mp)
	formatEtymologyToHTML(outDir, slice, mp)
	formatEtymologyToJSON(outDir, slice, mp)
}

func doHTML(args []string) {
	flag := flag.NewFlagSet("debug", flag.ExitOnError)
	var flagWords = flag.String("words", "", "limit words in csv. Only for HTML mode ")
	var flagWordsFile = flag.String("words-file", "", "limit words by the given file. Only for HTML mode ")
	flag.Parse(args)
	outFile := flag.Arg(0)
	if outFile == "" {
		panic("Please specify an output filename")
	}

	entries := cache.LoadFromCacheFile(*flagCacheFilePath)
	_, _, defaultCssPath, err := finder.FindDictFile()
	if err != nil {
		panic(err)
	}
	selectWords := getSelectWordsMap(*flagWords, *flagWordsFile)
	oFile, err := os.Create(outFile)
	if err != nil {
		panic(err)
	}
	renderSingleHTML(defaultCssPath, oFile, entries, selectWords)

}

func doHTMLSplit(args []string) {
	flag := flag.NewFlagSet("debug", flag.ExitOnError)
	var flagWords = flag.String("words", "", "limit words in csv. Only for HTML mode ")
	var flagWordsFile = flag.String("words-file", "", "limit words by the given file. Only for HTML mode ")
	flag.Parse(args)
	outDir := flag.Arg(0)
	if outDir == "" {
		panic("Please specify an output directory")
	}

	entries := cache.LoadFromCacheFile(*flagCacheFilePath)
	_, _, defaultCssPath, err := finder.FindDictFile()
	if err != nil {
		panic(err)
	}
	selectWords := getSelectWordsMap(*flagWords, *flagWordsFile)
	renderSplitHTML(defaultCssPath, outDir, entries, selectWords)
}

func doText(args []string) {
	flag := flag.NewFlagSet("debug", flag.ExitOnError)
	var flagWords = flag.String("words", "", "limit words in csv. Only for HTML mode ")
	var flagWordsFile = flag.String("words-file", "", "limit words by the given file. Only for HTML mode ")
	flag.Parse(args)
	outFile := flag.Arg(0)
	if outFile == "" {
		panic("Please specify an output filename")
	}

	entries := cache.LoadFromCacheFile(*flagCacheFilePath)
	selectWords := getSelectWordsMap(*flagWords, *flagWordsFile)
	oFile, err := os.Create(outFile)
	if err != nil {
		panic(err)
	}
	renderText(oFile, entries, selectWords)
}

func doPhonetics(args []string) {
	flag := flag.NewFlagSet("phonetics", flag.ExitOnError)
	var flagWords = flag.String("words", "", "limit words in csv. Only for HTML mode ")
	var flagWordsFile = flag.String("words-file", "", "limit words by the given file. Only for HTML mode ")
	flag.Parse(args)

	entries := cache.LoadFromCacheFile(*flagCacheFilePath)
	//parser.EtymDebugWriter = os.Stderr
	selectWords := getSelectWordsMap(*flagWords, *flagWordsFile)
	renderPhonetics(entries, selectWords)
}

func doDebug(args []string) {
	flag := flag.NewFlagSet("debug", flag.ExitOnError)
	var flagWords = flag.String("words", "", "limit words in csv. Only for HTML mode ")
	var flagWordsFile = flag.String("words-file", "", "limit words by the given file. Only for HTML mode ")
	flag.Parse(args)

	entries := cache.LoadFromCacheFile(*flagCacheFilePath)
	//parser.EtymDebugWriter = os.Stderr
	selectWords := getSelectWordsMap(*flagWords, *flagWordsFile)
	renderForDebug(entries, selectWords)
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		panic("Please specify a subcommand.")
	}
	cmd, args := args[0], args[1:]
	switch cmd {
	case "version":
		doVersion()
	case "find":
		doFind()
	case "dump":
		doDump()
	case "etym":
		doEtym(args)
	case "html":
		doHTML(args)
	case "htmlsplit":
		doHTMLSplit(args)
	case "text":
		doText(args)
	case "debug":
		doDebug(args)
	case "phonetics":
		doPhonetics(args)
	default:
		panic("Invalid mode")
	}
}

type SelectWords map[string]bool

func (mp SelectWords) EmptyOrMatch(w string) bool {
	return len(mp) == 0 || mp[strings.ToLower(w)]
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
		return strings.Split(csv, ",")
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

func renderSplitHTML(defaultCssPath string, outDir string, entries []*raw.Entry, selectWords SelectWords) {
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
		f.Write([]byte(GenHtmlHeader(htmlTitle, true, defaultCssPath)))
	}
	for _, ent := range entries {
		if selectWords.EmptyOrMatch(ent.Title) {
			t := ent.Title[0]
			f, found := files[t]
			if !found {
				f = files[0]
			}
			renderEntry(f, ent.Title, ent.Body)
		}
	}
}

func renderSingleHTML(cssPath string, w io.Writer, entries []*raw.Entry, selectWords SelectWords) {
	htmlTitle := "NOAD HTML as a single file"
	fmt.Fprintln(w, GenHtmlHeader(htmlTitle, true, cssPath))
	for _, ent := range entries {
		if selectWords.EmptyOrMatch(ent.Title) {
			renderEntry(w, ent.Title, ent.Body)
		}
	}
	fmt.Fprint(w, htmlFooter)
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
.table { width: 80%; }

.table th,.table td {
    border: 1px solid #ccc; padding: 10px;
}
.head_word {
    width:15em;
}
`

const EtymHTMLTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
	<style>
	%s
	</style>
    <title>%s</title>
</head>
<body>
<h1>%s</h1>
<table class="table"> %s </table>
</body>
</html>
`

func formatEtymologyToHTML(outDir string, backEtymLinks []*BackEtymLink, forwardEtymMap EtymMap) {
	fileE2O, err := os.Create(fmt.Sprintf("%s/%s.html", outDir, e2oFileName))
	if err != nil {
		panic(err)
	}
	var trs []string
	for _, bel := range backEtymLinks {
		if len(bel.OriginWords) == 0 {
			continue
		}
		trs = append(trs, fmt.Sprintf("<tr><td class=\"head_word\" id=\"%s\">%s</td><td>%s</td></tr>\n", bel.EngWord, bel.EngWord, strings.Join(bel.OriginWords, ", ")))
	}
	title := "NOAD Etymology English to Origin"
	fmt.Fprintf(fileE2O, EtymHTMLTemplate, EtymStyle, title, title, strings.Join(trs, "\n"))
	fileE2O.Close()

	var uniqFFs []string
	for k, _ := range forwardEtymMap {
		uniqFFs = append(uniqFFs, k)
	}
	sort.Strings(uniqFFs)
	fileO2E, err := os.Create(fmt.Sprintf("%s/%s.html", outDir, o2eFileName))

	if err != nil {
		panic(err)
	}
	trs = nil
	for _, ff := range uniqFFs {
		v := forwardEtymMap[ff]
		trs = append(trs, fmt.Sprintf("<tr><td class=\"head_word\" id=\"%s\">%s</td><td>%s</td></tr>\n", ff, ff, strings.Join(v, ", ")))
	}
	title = "NOAD Etymology Origin to English"

	fmt.Fprintf(fileO2E, EtymHTMLTemplate, EtymStyle, title, title, strings.Join(trs, "\n"))
	fileO2E.Close()
}

func formatEtymologyToJSON(outDir string, backEtymLinks []*BackEtymLink, forwardEtymMap EtymMap) {
	fileE2O, err := os.Create(fmt.Sprintf("%s/%s.json", outDir, e2oFileName))
	if err != nil {
		panic(err)
	}
	fmt.Fprint(fileE2O, "{\n")
	for _, bel := range backEtymLinks {
		if len(bel.OriginWords) == 0 {
			continue
		}
		// inline yaml
		fmt.Fprintf(fileE2O, "\"%s\":[\"%s\"],\n", bel.EngWord, strings.Join(bel.OriginWords, `","`))
	}
	fmt.Fprint(fileE2O, "\"__EOF__\":null\n")
	fmt.Fprint(fileE2O, "}\n")
	fileE2O.Close()

	var uniqFFs []string
	for k, _ := range forwardEtymMap {
		uniqFFs = append(uniqFFs, k)
	}
	sort.Strings(uniqFFs)
	fileO2E, err := os.Create(fmt.Sprintf("%s/%s.json", outDir, o2eFileName))

	if err != nil {
		panic(err)
	}

	fmt.Fprint(fileO2E, "{\n")
	for _, ff := range uniqFFs {
		v := forwardEtymMap[ff]
		// inline yaml
		fmt.Fprintf(fileO2E, "\"%s\":[\"%s\"],\n", ff, strings.Join(v, `","`))
	}
	fmt.Fprint(fileO2E, "\"__EOF__\":null\n")
	fmt.Fprint(fileO2E, "}\n")
	fileO2E.Close()
}

func formatEtymologyToYAML(outDir string, backEtymLinks []*BackEtymLink, forwardEtymMap EtymMap) {
	fileE2O, err := os.Create(fmt.Sprintf("%s/%s.yml", outDir, e2oFileName))
	if err != nil {
		panic(err)
	}
	fmt.Fprint(fileE2O, "---\n")
	for _, bel := range backEtymLinks {
		if len(bel.OriginWords) == 0 {
			continue
		}
		// inline yaml
		fmt.Fprintf(fileE2O, "%s:[%s]\n", bel.EngWord, strings.Join(bel.OriginWords, ","))
	}
	fileE2O.Close()

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

	for _, ff := range uniqFFs {
		v := forwardEtymMap[ff]
		// inline yaml
		fmt.Fprintf(fileO2E, "%s:[%s]\n", ff, strings.Join(v, ","))
	}
	fileO2E.Close()
}

func collectEtymology(entries []*raw.Entry, selectWords SelectWords) ([]*BackEtymLink, EtymMap) {
	var forwardEtymMap = make(EtymMap, len(entries))
	var allFF []string
	var backEtymLinks []*BackEtymLink
	for _, ent := range entries {
		if selectWords.EmptyOrMatch(ent.Title) {
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
	}

	return backEtymLinks, forwardEtymMap
}

func renderPhonetics(entries []*raw.Entry, selectWords SelectWords) {
	for _, ent := range entries {
		if selectWords.EmptyOrMatch(ent.Title) {
			e := parser.ParseEntry(ent.Title, ent.Body)
			fmt.Printf("%s\t%d\t%s\t%s\n", e.Title, e.NumSyll, e.Syll, e.IPA)
		}
	}
}

func renderForDebug(entries []*raw.Entry, selectWords SelectWords) {
	for _, ent := range entries {
		if selectWords.EmptyOrMatch(ent.Title) {
			e := parser.ParseEntry(ent.Title, ent.Body)
			fmt.Printf("%s\t%d\t%s\t%s\n", e.Title, e.NumSyll, e.Syll, e.IPA)
		}
	}
}

func renderText(w io.Writer, entries []*raw.Entry, selectWords SelectWords) {
	for _, ent := range entries {
		if selectWords.EmptyOrMatch(ent.Title) {
			et := parser.ParseEntry(ent.Title, ent.Body)
			s := ToOneline(et)
			fmt.Fprintln(w, s)
		}
	}
}

func assert(cnd bool, expect string) {
	if !cnd {
		panic("Assertion failed. Expect " + expect)
	}
}
