package main

import "fmt"

const htmlHeader = `<!doctype html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">

    <title>%s</title>

    <link rel="stylesheet" href="DefaultStyle.css">
    <link rel="stylesheet" href="customize.css">
</head>
<body>
`

func GenHtmlHeader(title string) string {
	return fmt.Sprintf(htmlHeader, title)
}

const htmlFooter = `</body>
</html>
`
