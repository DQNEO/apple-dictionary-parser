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

	switch *flagMode {
	case "html":
		selectWords := getWords(*flagWords, *flagWordsFile)
		renderHTML(entries, selectWords)
	case "htmlsplit":
		outDir := flag.Arg(1)
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
			cssBlock := GetExternalCssBlock()
			htmlTitle := "NOAD - " + strings.ToUpper(string(letter))
			f.Write([]byte(GenHtmlHeader(htmlTitle, cssBlock)))
		}
		for _, ent := range entries {
			t := ent.Title[0]
			f, found := files[t]
			if !found {
				f = files[0]
			}
			f.Write(ent.Body)
		}
	case "text":
		selectWords := getWords(*flagWords, *flagWordsFile)
		renderText(entries, selectWords)
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
	var mapWords = make(map[string]bool, len(words))
	for _, w := range words {
		if len(w) > 0 {
			mapWords[strings.ToLower(w)] = true
		}
	}
	htmlTitle := "NOAD HTML as a single  file"
	cssBlock := GetExternalCssBlock()
	fmt.Print(GenHtmlHeader(htmlTitle, cssBlock))
	for _, ent := range entries {
		if len(words) > 0 && !mapWords[strings.ToLower(ent.Title)] {
			continue
		}
		fmt.Println(string(ent.Body))
	}

	fmt.Print(htmlFooter)
}

type E struct {
	Title string
	Syll  string
	IPA   string
	SG    string
	Phr   string
	Phv   string
	Drv   string
	Etym  string
	Note  string
}

// To human readable line
func (e *E) ToOneline() string {
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
	if e.Etym != "" {
		fields = append(fields, "<"+e.Etym+">")
	}
	fields = append(fields, e.Note)
	return strings.Join(fields, " ")
}

func renderText(entries []*RawEntry, words []string) {
	var mapWords = make(map[string]bool, len(words))
	for _, w := range words {
		if len(w) > 0 {
			mapWords[strings.ToLower(w)] = true
		}
	}

	for _, ent := range entries {
		if len(words) > 0 && !mapWords[strings.ToLower(ent.Title)] {
			continue
		}
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
			case "hg x_xh0": // head group
				pe.HG = elm
			case "sg": // meaning
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
		hg := parseHG(pe.Title, pe.HG)
		etym := parseEtym(pe.Title, pe.Etym)
		//hgDump, err := yaml.Marshal(hg)
		et := &E{
			Title: pe.Title,
			Syll:  hg.SYL_TXT,
			IPA:   hg.PRX,
			SG:    S(pe.SG),
			Phr:   S(pe.Phrases),
			Phv:   S(pe.PhVerbs),
			Drv:   S(pe.Derivatives),
			Etym:  etym,
			Note:  S(pe.Note),
		}

		fmt.Println(et.ToOneline())
	}
}

type HG struct {
	HW      string
	SYL_TXT string
	PRX     string
	PR      string
	VG      string
	LG      string
	FG      string
}

func parseHG(title string, eHG *etree.Element) *HG {
	if eHG == nil {
		panic("HG is nil")
	}
	tokens := eHG.Child
	hg := &HG{}
	for _, tok := range tokens {
		elm, ok := tok.(*etree.Element)
		if ok {
			cls := elm.SelectAttr("class")
			className := cls.Value
			switch className {
			case "hw": // This is always present
				hg.HW = dumpElm(elm)
			case "syl_txt":
				hg.SYL_TXT = elm.Text()
			case "prx":
				// len(elm.Child) varies among 2,4,6,8,12
				// ['|', option1-US, option1-IPA, option2-US, option2-IPA, ..., '|]
				var pronunciation []string
				if len(elm.Child) > 2 {
					options := elm.Child[1 : len(elm.Child)-1] // remove enclosing "| |" pair.
					for i := 1; i < len(options); i += 2 {
						elm, ok := options[i].(*etree.Element)
						if !ok {
							panic("Unexpected prx token")
						}
						pronunciation = append(pronunciation, elm.Text())
					}
				} // else , no IPA
				//fmt.Fprintf(os.Stderr, "prx children=%02d, title=%s%c", len(elm.Child), title, 10)
				hg.PRX = strings.Join(pronunciation, ",")
			case "pr":
				hg.PR = dumpElm(elm)
			case "vg": // can be more than one
				hg.VG = dumpElm(elm)
			case "lg":
				hg.LG = dumpElm(elm)
			case "fg":
				hg.FG = dumpElm(elm)
			default:
				panic(fmt.Sprintf("Unexpected clas name in HG:word=%s, clas=%s", title, className))
			}

		} else {
			panic("Unexpected")
		}
	}
	return hg
}

func assert(cnd bool, expect string) {
	if !cnd {
		panic("Assertion failed. Expect " + expect)
	}
}
func parseEtym(title string, e *etree.Element) string {
	if e == nil {
		return ""
	}
	assert(len(e.Child) == 2, "etym children should be 2")

	etym := e.Child[1].(*etree.Element)
	var s string
	for _, child := range etym.Child {
		switch e := child.(type) {
		case *etree.Element:
			s += e.Text() + " "
		case *etree.CharData:
		default:
			panic("unexpected type")
		}
	}

	return s
}

func dumpTokens(tokens []etree.Token) string {
	var ss []string
	for _, tok := range tokens {
		var s string

		switch tk := tok.(type) {
		case *etree.Element:
			//cls := tk.SelectAttr("class")
			//s = cls.Value
			s = dumpElm(tk)
		case *etree.CharData:
			//s = "(" + tk.Data + ")"
		default:
			typ := fmt.Sprintf("<TOK:%T>", tok)
			panic("Unexpected token type:" + typ)
		}
		ss = append(ss, s+" ")
	}
	return fmt.Sprintf("[%d:%s]", len(tokens), strings.Join(ss, ","))
}

func dumpElm(e *etree.Element) string {
	if e == nil {
		return "<nil>"
	}
	var attrs []string
	for _, attr := range e.Attr {
		attrs = append(attrs, attr.Key+"="+attr.Value)
	}
	return fmt.Sprintf("<%s %s>%s %s</%s>\n",
		e.Tag, strings.Join(attrs, " "),
		e.Text(),
		dumpTokens(e.Child),
		e.Tag)
}

func S(elm *etree.Element) string {
	if elm == nil {
		return ""
	}
	elms := elm.FindElements("./")
	var s string
	for _, e := range elms {
		s += e.Text() + " "
	}
	return s
}
