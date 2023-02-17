package main

import (
	"fmt"
	"os"
)

const htmlHeader = `<!doctype html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">

    <title>%s</title>

    %s

</head>
<body>
`

const defaultCssFile = "/tmp/DefaultStyle.css"
const cstmCssFile = "/tmp/customize.css"

func GetInternalCssBlock() string {
	defaultCss, err := os.ReadFile(defaultCssFile)
	if err != nil {
		panic(err)
	}
	cstmCss, err := os.ReadFile(cstmCssFile)
	if err != nil {
		panic(err)
	}

	return "<style>" + string(defaultCss) + "\n" + string(cstmCss) + "</style>"
}

func GetExternalCssBlock() string {
	return `<link rel="stylesheet" href="DefaultStyle.css">
    <link rel="stylesheet" href="customize.css">`
}

func GenHtmlHeader(title string, inlineCss bool) string {
	var cssBlock string
	if inlineCss {
		cssBlock = GetInternalCssBlock()
	} else {
		cssBlock = GetExternalCssBlock()
	}
	return fmt.Sprintf(htmlHeader, title, cssBlock)
}

const htmlFooter = `</body>
</html>
`
