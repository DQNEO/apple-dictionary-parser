package main

import "fmt"

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

func GetInternalCssBlock() string {
	return ""
}

func GetExternalCssBlock() string {
	return `<link rel="stylesheet" href="DefaultStyle.css">
    <link rel="stylesheet" href="customize.css">`
}

func GenHtmlHeader(title string, cssBlock string) string {
	return fmt.Sprintf(htmlHeader, title, cssBlock)
}

const htmlFooter = `</body>
</html>
`
