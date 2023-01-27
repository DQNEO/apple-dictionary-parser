package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/beevik/etree"
)

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

func main() {
	rawDumpFile := os.Args[1]
	all, err := os.ReadFile(rawDumpFile)
	if err != nil {
		panic(err)
	}
	entries := bytes.Split(all, []byte{'\n'})
	fmt.Printf("children,hg,sg,phrases,phverbse,derivatives,text\n")
	for _, ent := range entries[:] {
		if len(ent) == 0 {
			// Possibly end of file
			continue
		}
		title, rawBody, found := bytes.Cut(ent, []byte(":::"))
		if !found {
			panic("failed to Cut:" + (string(ent)))
		}

		_ = title
		//var d interface{}
		doc := etree.NewDocument()
		err = doc.ReadFromBytes(rawBody)
		if err != nil {
			panic(err)
		}

		//os.Stdout.Write(title)
		//fmt.Print(":::")
		//		children := doc.Child[1].(*etree.Element).Child
		child := doc.Child[0].(*etree.Element)
		numChildren := len(child.Child)
		if numChildren < 2 || 7 < numChildren {
			panic("Unexpected number of children")
		}
		fmt.Printf("children=%d,", len(child.Child))
		var ss [7]string
		for _, ch := range child.Child {
			elm := ch.(*etree.Element)
			if elm.Tag != "span" {
				panic("unexpected tag:" + elm.Tag + " --- " + string(title))
			}
			class := elm.SelectAttr("class").Value
			pe := &ParsedEntry{Title: string(title)}
			switch class {
			case "hg x_xh0":
				ss[0] = class
				pe.HG = elm
			case "sg":
				ss[1] = class
				pe.SG = elm
			case "subEntryBlock x_xo0 t_phrases":
				ss[2] = class
				pe.Phrases = elm
			case "subEntryBlock x_xo0 t_phrasalVerbs":
				ss[3] = class
				pe.PhVerbs = elm
			case "subEntryBlock x_xo0 t_derivatives":
				ss[4] = class
				pe.Derivatives = elm
			case "etym x_xo0":
				ss[5] = class
				pe.Etym = elm
			case "note x_xo0":
				ss[6] = class
				pe.Note = elm
			default:
				panic("unexpected class:" + class + " --- " + string(title))
			}

		}
		fmt.Printf("%s,%s,%s,%s,%s,%s,%s,%s", ss[0], ss[1], ss[2], ss[3], ss[4], ss[5], ss[6], string(title))
		//fmt.Printf("[%s]", title)
		//		dump.P(child.Child)
		fmt.Print("\n")
	}
}
