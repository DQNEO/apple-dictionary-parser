package parser

import (
	"fmt"
	"github.com/beevik/etree"
	"io"
	"strings"
)

type Entry struct {
	Title   string
	Syll    string
	NumSyll int
	IPA     string
	SG      string
	Phr     string
	Phv     string
	Drv     string
	Etym    Etymology
	Note    string

	FFWords []string
}

type EtymChunk string

type Etymology []EtymChunk

func parseTopLevelElements(title string, body []byte) (*etree.Element, *etree.Element, *etree.Element, *etree.Element, *etree.Element, *etree.Element, *etree.Element) {
	doc := etree.NewDocument()
	err := doc.ReadFromBytes(body)
	if err != nil {
		panic(err)
	}

	child := doc.Child[0].(*etree.Element)
	numChildren := len(child.Child)
	if numChildren < 2 || 7 < numChildren {
		panic("Unexpected number of children")
	}
	var eHG, eSG, ePhrases, ePhVerbs, eDerivatives, eEtym, eNote *etree.Element
	for _, ch := range child.Child {
		elm := ch.(*etree.Element)
		if elm.Tag != "span" {
			panic("unexpected tag:" + elm.Tag + " --- " + title)
		}
		class := elm.SelectAttr("class").Value
		switch class {
		case "hg x_xh0": // head group
			eHG = elm
		case "sg": // meaning
			eSG = elm
		case "subEntryBlock x_xo0 t_phrases":
			ePhrases = elm
		case "subEntryBlock x_xo0 t_phrasalVerbs":
			ePhVerbs = elm
		case "subEntryBlock x_xo0 t_derivatives":
			eDerivatives = elm
		case "etym x_xo0":
			eEtym = elm
		case "note x_xo0":
			eNote = elm
		default:
			panic(fmt.Sprintf(`unexpected class: "%s" in entry "%s"`, class, title))
		}
	}
	return eHG, eSG, ePhrases, ePhVerbs, eDerivatives, eEtym, eNote
}

var EtymDebugWriter io.Writer = io.Discard

func ParseEntry(title string, body []byte) *Entry {
	eHG, eSG, ePhrases, ePhVerbs, eDerivatives, eEtym, eNote := parseTopLevelElements(title, body)
	hg := parseHG(title, eHG)
	etym, ffwords := parseEtym(title, eEtym)

	et := &Entry{
		Title:   title,
		Syll:    hg.SYL_TXT,
		IPA:     hg.PRX,
		SG:      S(eSG),
		Phr:     S(ePhrases),
		Phv:     S(ePhVerbs),
		Drv:     S(eDerivatives),
		Etym:    etym,
		Note:    S(eNote),
		FFWords: ffwords,
	}

	if strings.Contains(title, " ") || strings.Contains(title, "-") {
		// Ignore syllable in compound words
		et.NumSyll = 0
	} else {
		et.NumSyll = strings.Count(et.Syll, "·") + 1
	}
	return et
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
				//fmt.Fprintf(EtymDebugWriter, "prx children=%02d, title=%s%c", len(elm.Child), title, 10)
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

func parseFF(elm *etree.Element) string {
	assert(len(elm.Child) == 1, "ff should have 1 child")
	char := elm.Child[0].(*etree.CharData)
	return strings.TrimSpace(char.Data)
}

func collectText(indentLevel int, debugWriter io.Writer, elm *etree.Element) string {
	var r string
	for _, c := range elm.Child {
		var indent []byte
		for i := 0; i < indentLevel; i++ {
			indent = append(indent, ' ')
		}
		fmt.Fprint(debugWriter, string(indent))
		switch e := c.(type) {
		case *etree.Element:
			class := e.SelectAttrValue("class", "")
			fmt.Fprintf(EtymDebugWriter, `- <%s> Class="%s", ChildLen=%d, Text()="%s"`+"\n", e.Tag, class, len(e.Child), e.Text())
			var s string
			s = collectText(indentLevel+4, debugWriter, e)
			r += s
		case *etree.CharData:
			if e.IsWhitespace() {
				continue
			}
			s := e.Data
			fmt.Fprintf(debugWriter, "- C\"%s\"\n", s)
			r += s
		}
	}
	return r
}

func parseEtym(title string, e *etree.Element) (Etymology, []string) {
	var ees []EtymChunk
	var ffwords []string
	if e == nil {
		return nil, nil
	}
	assert(len(e.Child) == 2, "etym children should be 2")
	etym := e.Child[1].(*etree.Element)
	fmt.Fprintf(EtymDebugWriter, "[%s] ---- childlen=%d\n", title, len(etym.Child))
	for i, elem := range etym.Child {
		//fmt.Fprintf(EtymDebugWriter, "elem[%d] typ=%T, Index=%d\n", i, elem, elem.Index())
		fmt.Fprintf(EtymDebugWriter, "  - [%02d] ", i)
		switch e := elem.(type) {
		case *etree.Element:
			class := e.SelectAttr("class").Value
			if class == "gp tg_etym" && e.Text() == "." {
				// End of Etymology block ?
				fmt.Fprintf(EtymDebugWriter, ".\n")
				continue
			}
			fmt.Fprintf(EtymDebugWriter, `<%s> Class="%s", ChildLen=%d, Text()="%s"`+"\n", e.Tag, class, len(e.Child), e.Text())
			if len(e.Child) == 0 {
				panic("Unexpected structure")
			}
			var s string
			switch class {
			case "ff":
				s = parseFF(e)
				if strings.Contains(s, ":") {
					//fmt.Fprintf(os.Stderr, "dected : in '%s' . SKIP\n", s)
				} else {
					s = strings.TrimPrefix(s, "'")
					ffwords = append(ffwords, s)
					fmt.Fprintf(EtymDebugWriter, "    <ff>%s</ff>\n", s)
				}
			default:
				s = collectText(4, EtymDebugWriter, e)
			}

			ees = append(ees, EtymChunk(s))
		case *etree.CharData:
			if e.IsWhitespace() {
				continue
			}
			fmt.Fprintf(EtymDebugWriter, "C\"%s\"\n", e.Data)
			ee := EtymChunk(e.Data)
			ees = append(ees, ee)
		default:
			panic("unexpected type")
		}
	}

	return ees, ffwords
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

func assert(cnd bool, expect string) {
	if !cnd {
		panic("Assertion failed. Expect " + expect)
	}
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
