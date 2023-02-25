package main

import (
	"fmt"
	"github.com/DQNEO/apple-dictionary-parser/cache"
	"github.com/DQNEO/apple-dictionary-parser/extracter"
	"github.com/DQNEO/apple-dictionary-parser/extracter/raw"
	"github.com/DQNEO/apple-dictionary-parser/finder"
	"github.com/DQNEO/apple-dictionary-parser/parser"
	"github.com/urfave/cli/v2"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
)

const version = "v0.0.5"

func doVersion(cCtx *cli.Context) error {
	fmt.Println("apple-dictionary-parser version " + version)
	return nil
}

func doFind(cCtx *cli.Context) error {
	dictDir, dictFilePath, defaultCssPath, err := finder.FindDictFile()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Found a dictionary resources.\n---\n")
	fmt.Printf("Directory:\n  '%s'\n", dictDir)
	fmt.Printf("Body file:\n  '%s'\n", dictFilePath)
	fmt.Printf("CSS file:\n  '%s'\n", defaultCssPath)
	return nil
}

func doDump(cCtx *cli.Context) error {
	var dictFilePath string
	flagDictFilePath := cCtx.String("dict-file")
	if flagDictFilePath == "" {
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
		dictFilePath = flagDictFilePath
	}
	fmt.Printf("Extracting the dictionary file ...\n")
	entries := extracter.ParseBinaryFile(dictFilePath)
	flagCacheFilePath := cCtx.String("cache-file")
	oFile, err := os.Create(flagCacheFilePath)
	if err != nil {
		panic(err)
		return err
	}
	cache.SaveEntries(oFile, entries)
	fmt.Printf("Dictonary raw data is successfully saved to: %s\n", flagCacheFilePath)
	return nil
}

func doShowIPA(cCtx *cli.Context) error {
	var msg = `
-- Vowels --
Short: ɪ ɛ æ ɑ ə ʊ
Long: i ɔ ɑ u
Diphthong: eɪ aɪ ɔɪ aʊ oʊ ju
           ɪr ɛr ɑr ɔr ʊr

-- Consonants / Semi-Vowels --
p b f v k ɡ t d 
s z θ ð ʃ ʒ 
r l h  
m ŋ n 
j w  
`
	fmt.Print(msg)
	return nil
}

func doCollectIPA(cCtx *cli.Context) error {
	flagCacheFilePath := cCtx.String("cache-file")
	entries := cache.LoadFromCacheFile(flagCacheFilePath)
	type occurrence struct {
		cnt   int
		words []string
	}
	var ipaChars = make(map[int32]*occurrence, 100)
	for _, ent := range entries {
		et := parser.ParseEntry(ent.Title, ent.Body)
		//fmt.Fprintf(os.Stdout, "%s: ", ent.Title)

		for _, char := range et.IPA {
			switch char {
			case 'ˈ', 'ˌ': // possibly stress
				continue
			case '(', ')':
				continue
			case ',', ';', ' ': // delimiter between multiple candidates
				continue
			case '-': // prefix
				continue
			case '%': // Dictionary's bug ;)
				continue
			case '[', ']', ':', 'ō', 'õ', 'ã', 0x303: // foreign words
				continue
			case 0x35c, 0x329: // irregular character that is ignorable
				continue
			default:
				if oc, ok := ipaChars[char]; ok {
					oc.cnt++
					oc.words = append(oc.words, et.Title+":"+et.IPA)
				} else {
					ipaChars[char] = &occurrence{cnt: 1}
				}
				//ipaChars[char]++
				//fmt.Fprintf(os.Stdout, "%c ", char)
			}
		}
		//fmt.Fprint(os.Stdout, "\n")
	}

	for k, oc := range ipaChars {
		if oc.cnt == 1 {
			continue
		}
		fmt.Printf("count=%07d: %03x %c \n", oc.cnt, k, k)
	}
	return nil
}

func doEtym(cCtx *cli.Context) error {
	outDir := cCtx.Args().First()
	if outDir == "" {
		panic("Please specify an output directory")
	}

	entries := LoadFromCacheFile(cCtx)
	selectWords := GetSelectWordsMap(cCtx)
	slice, mp := collectEtymology(entries, selectWords)
	//println(len(slice), len(mp))
	formatEtymologyToYAML(outDir, slice, mp)
	formatEtymologyToHTML(outDir, slice, mp)
	formatEtymologyToJSON(outDir, slice, mp)
	return nil
}

func createOutFile(path string) (*os.File, error) {
	if path == "" {
		return os.Stdout, nil
	} else {
		return os.Create(path)
	}
}

func doHTML(cCtx *cli.Context) error {
	flagOutFile := cCtx.String("out-file")

	oFile, err := createOutFile(flagOutFile)
	if err != nil {
		return err
	}
	entries := LoadFromCacheFile(cCtx)
	_, _, defaultCssPath, err := finder.FindDictFile()
	if err != nil {
		return err
	}
	selectWords := GetSelectWordsMap(cCtx)
	renderSingleHTML(defaultCssPath, oFile, entries, selectWords)
	return nil
}

func doHTMLSplit(cCtx *cli.Context) error {
	flagOutFile := cCtx.String("out-dir")
	if flagOutFile == "" {
		panic("Please specify an output directory")
	}

	entries := LoadFromCacheFile(cCtx)
	_, _, defaultCssPath, err := finder.FindDictFile()
	if err != nil {
		panic(err)
	}
	selectWords := GetSelectWordsMap(cCtx)
	renderSplitHTML(defaultCssPath, flagOutFile, entries, selectWords)
	return nil
}

func doText(cCtx *cli.Context) error {
	flagOutFile := cCtx.String("out-file")

	oFile, err := createOutFile(flagOutFile)
	if err != nil {
		return err
	}

	entries := LoadFromCacheFile(cCtx)
	selectWords := GetSelectWordsMap(cCtx)

	renderText(oFile, entries, selectWords)
	return nil
}

func doPhonetics(cCtx *cli.Context) error {
	var flagIPA = cCtx.String("ipa")
	var flagIPARegex = cCtx.String("ipa-regex")
	var flagMin = cCtx.Int("min-syl")
	var flagMax = cCtx.Int("max-syl")
	var flagOutFile = cCtx.String("out-file")
	oFile, err := createOutFile(flagOutFile)
	if err != nil {
		panic(err)
	}

	entries := LoadFromCacheFile(cCtx)
	//parser.EtymDebugWriter = os.Stderr
	selectWords := GetSelectWordsMap(cCtx)

	var ipaRegex *regexp.Regexp
	if flagIPARegex != "" {
		ipaRegex = regexp.MustCompile(flagIPARegex)
	}
	opt := &PhoneticsSelector{
		IPA:          flagIPA,
		IPARegex:     ipaRegex,
		MaxSyllables: flagMax,
		MinSyllables: flagMin,
	}
	renderPhonetics(oFile, entries, selectWords, opt)
	return nil
}

func doDebug(cCtx *cli.Context) error {
	entries := LoadFromCacheFile(cCtx)
	selectWords := GetSelectWordsMap(cCtx)

	renderForDebug(entries, selectWords)
	return nil
}

var localWordsFlag = &cli.StringFlag{
	Name:  "words",
	Usage: "limit words in csv",
}
var localWordsFileFlag = &cli.StringFlag{
	Name:      "words-file",
	TakesFile: true,
	Usage:     "limit words by the given file",
}

var localOutFileFlag = &cli.StringFlag{
	Name:    "out-file",
	Aliases: []string{"o"},
	Usage:   "output file",
}

var localOutDirFlag = &cli.StringFlag{
	Name:    "out-dir",
	Aliases: []string{"o"},
	Usage:   "output directory",
}

func LoadFromCacheFile(cCtx *cli.Context) []*raw.Entry {
	flagCacheFilePath := cCtx.String("cache-file")
	entries := cache.LoadFromCacheFile(flagCacheFilePath)
	return entries
}

func main() {
	app := &cli.App{
		Name:    "apple-dictionary-parser",
		Version: version,
		Usage:   "a tool to parse and analyze MacOS's built-in dictionaries",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "cache-file",
				Value: cache.DEFAULT_PATH,
				Usage: "cache file path",
			},
			&cli.StringFlag{
				Name:  "dict-file",
				Usage: "dictionary file path",
			},
		},
		Commands: []*cli.Command{
			{
				Name:   "version",
				Usage:  "print the version",
				Action: doVersion,
			},
			{
				Name:   "find",
				Usage:  "find a dictionary file",
				Action: doFind,
			},
			{
				Name:   "dump",
				Usage:  "Dump dictionary raw data",
				Action: doDump,
			},
			{
				Name:  "text",
				Usage: "Convert dictionary data into text format",
				Flags: []cli.Flag{
					localWordsFlag,
					localWordsFileFlag,
					localOutFileFlag,
				},
				Action: doText,
			},
			{
				Name:  "html",
				Usage: "Convert dictionary data into html format",
				Flags: []cli.Flag{
					localWordsFlag,
					localWordsFileFlag,
					localOutFileFlag,
				},
				Action: doHTML,
			},
			{
				Name:  "html-split",
				Usage: "Convert dictionary data into html format",
				Flags: []cli.Flag{
					localWordsFlag,
					localWordsFileFlag,
					localOutDirFlag,
				},
				Action: doHTMLSplit,
			},
			{
				Name:  "phonetics",
				Usage: "Convert dictionary data into html format",
				Flags: []cli.Flag{
					localWordsFlag,
					localWordsFileFlag,
					localOutFileFlag,
					&cli.StringFlag{Name: "ipa", Usage: "filter by IPA"},
					&cli.StringFlag{Name: "ipa-regex", Usage: "filter by IPA in regular expression"},
					&cli.IntFlag{Name: "min-syl", Value: 1, Usage: "min number of syllables"},
					&cli.IntFlag{Name: "max-syl", Value: 1000, Usage: "max number of syllables"},
				},
				Action: doPhonetics,
			},
			{
				Name:  "debug",
				Usage: "debug",
				Flags: []cli.Flag{
					localWordsFlag,
					localWordsFileFlag,
					localOutFileFlag,
				},
				Action: doDebug,
			},
			{
				Name:   "collect-ipa",
				Usage:  "collect all IPA occurrences",
				Action: doCollectIPA,
			},
			{
				Name:   "show-ipa",
				Usage:  "show all IPA letters",
				Action: doShowIPA,
			},
			{
				Name:   "etym",
				Usage:  "Output Etymology info",
				Action: doEtym,
				Flags: []cli.Flag{
					localWordsFlag,
					localWordsFileFlag,
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

type SelectWords map[string]bool

func (mp SelectWords) EmptyOrMatch(w string) bool {
	return len(mp) == 0 || mp[strings.ToLower(w)]
}

func GetSelectWordsMap(cCtx *cli.Context) SelectWords {
	csv := cCtx.String("words")
	file := cCtx.String("words-file")
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

type PhoneticsSelector struct {
	IPA          string
	IPARegex     *regexp.Regexp
	MaxSyllables int // default 1000
	MinSyllables int // default 1
}

func (psel *PhoneticsSelector) Match(e *parser.Entry) bool {
	// check syllables number first
	if !(psel.MinSyllables <= e.NumSyll && e.NumSyll <= psel.MaxSyllables) {
		return false
	}

	if e.NumSyll == 0 {
		fmt.Fprintf(os.Stderr, "psel=%#v\n", psel)
		panic("internal error")
	}
	if psel.IPA != "" {
		return strings.Contains(e.IPA, psel.IPA)
	}
	if psel.IPARegex != nil {
		return psel.IPARegex.MatchString(e.IPA)
	}
	return true
}

func renderPhonetics(w io.Writer, entries []*raw.Entry, selectWords SelectWords, psel *PhoneticsSelector) {
	for _, ent := range entries {
		if selectWords.EmptyOrMatch(ent.Title) {
			e := parser.ParseEntry(ent.Title, ent.Body)
			if psel.Match(e) {
				fmt.Fprintf(w, "%20s%  02d%20s%20s\n", e.Syll, e.NumSyll, e.Title, e.IPA)
			}
		}
	}
}

func renderForDebug(entries []*raw.Entry, selectWords SelectWords) {
	for _, ent := range entries {
		if selectWords.EmptyOrMatch(ent.Title) {
			e := parser.ParseEntry(ent.Title, ent.Body)
			fmt.Fprintf(os.Stdout, "%s\t%d\t%s\t%s\n", e.Title, e.NumSyll, e.Syll, e.IPA)
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
