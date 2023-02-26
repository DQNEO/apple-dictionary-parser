package main

import (
	_ "embed"
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

	<script>%s</script>
</head>
<body>
<div id="widget" style="display:none;"></div>
`

//go:embed customize.css
var customCss string

//go:embed myapp.js
var javaScriptExample string

func GetInternalCssBlock(defaultCssPath string) string {
	defaultCss, err := os.ReadFile(defaultCssPath)
	if err != nil {
		panic(err)
	}

	return "<style>" + string(defaultCss) + "\n" + customCss + "</style>"
}

func GetExternalCssBlock() string {
	return `<link rel="stylesheet" href="DefaultStyle.css">
    <link rel="stylesheet" href="customize.css">`
}

func GenHtmlHeader(title string, inlineCss bool, defaultCssPath string, javaScript string) string {
	var cssBlock string
	if inlineCss {
		cssBlock = GetInternalCssBlock(defaultCssPath)
	} else {
		cssBlock = GetExternalCssBlock()
	}
	return fmt.Sprintf(htmlHeader, title, cssBlock, javaScript)
}

const htmlFooter = `</body>
</html>
`
