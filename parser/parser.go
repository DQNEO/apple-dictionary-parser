package parser

import (
	"fmt"
	"github.com/beevik/etree"
	"os"
	"strings"
)

type Entry struct {
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

func ParseEntry(title string, body []byte) *Entry {
	eHG, eSG, ePhrases, ePhVerbs, eDerivatives, eEtym, eNote := parseTopLevelElements(title, body)
	hg := parseHG(title, eHG)
	etym := parseEtym(title, eEtym)
	//hgDump, err := yaml.Marshal(hg)
	et := &Entry{
		Title: title,
		Syll:  hg.SYL_TXT,
		IPA:   hg.PRX,
		SG:    S(eSG),
		Phr:   S(ePhrases),
		Phv:   S(ePhVerbs),
		Drv:   S(eDerivatives),
		Etym:  etym,
		Note:  S(eNote),
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

func parseEtym(title string, e *etree.Element) string {
	if e == nil {
		return ""
	}
	assert(len(e.Child) == 2, "etym children should be 2")
	etym := e.Child[1].(*etree.Element)
	var s string
	fmt.Fprintf(os.Stderr, "num child=%d\n", len(etym.Child))
	for i, child := range etym.Child {
		//fmt.Fprintf(os.Stderr, "child[%d] typ=%T, Index=%d\n", i, child, child.Index())
		switch e := child.(type) {
		case *etree.Element:
			fmt.Fprintf(os.Stderr, "child[%d] typ=%T, Class=%s, Text()=\"%s\"\n", i, child, e.SelectAttr("class").Value, e.Text())
			if len(e.Child) == 0 {
				s += e.Text() + " "
			} else {

				for _, c := range e.Child {
					switch e := c.(type) {
					case *etree.Element:
						s += e.Text()
					case *etree.CharData:
						s += e.Data
					}
				}
			}
		case *etree.CharData:
			fmt.Fprintf(os.Stderr, "child[%d] typ=%T, Data=%s\n", i, child, e.Data)
			s += e.Data + " "
		default:
			panic("unexpected type")
		}
	}

	return s
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