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

const htmlHeader = `
<!doctype html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">

    <title>A Basic HTML5 Template</title>
    <meta name="description" content="A simple HTML5 Template for new projects.">
    <meta name="author" content="SitePoint">

    <meta property="og:title" content="A Basic HTML5 Template">
    <meta property="og:type" content="website">
    <meta property="og:url" content="https://www.sitepoint.com/a-basic-html5-template/">
    <meta property="og:description" content="A simple HTML5 Template for new projects.">
    <meta property="og:image" content="image.png">

    <link rel="icon" href="/favicon.ico">
    <link rel="icon" href="/favicon.svg" type="image/svg+xml">
    <link rel="apple-touch-icon" href="/apple-touch-icon.png">

    <link rel="stylesheet" href="DefaultStyle.css">

</head>
<body>
`

func main() {
	rawDumpFile := os.Args[1]
	all, err := os.ReadFile(rawDumpFile)
	if err != nil {
		panic(err)
	}
	entries := bytes.Split(all, []byte{'\n'})
	for _, ent := range entries[:] {
		if len(ent) == 0 {
			// Possibly end of file
			continue
		}
		ttlBytes, rawBody, found := bytes.Cut(ent, []byte(":::"))
		if !found {
			panic("failed to Cut:" + (string(ent)))
		}
		title := string(ttlBytes)

		doc := etree.NewDocument()
		err = doc.ReadFromBytes(rawBody)
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
				panic("unexpected class:" + class + " --- " + title)
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
