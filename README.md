# Apple Dictionary Parser

Apple Dictionary Parser is a command line tool and a library to parse and analyze MacOS's built-in dictionary files.

Currently only "New Oxford American Dictionary" is supported.

# Installation

```
$ go install github.com/DQNEO/apple-dictionary-parser@latest
```

The command line tool `apple-dictionary-parser ` will be installed in your `$GOPATH/bin` directory.


# Usage

## Export dictionary contents in raw format

```
$ apple-dictionary-parser dump
```

This `dump` subcommand automatically finds the location of the dictionary file in your MacOS, and extract the binary content into a raw dump file (`/tmp/noad.cache`).

The format of the dump file is TSV (tab separated values),  each line representing a word with definition.

```
<world title>\t<definition of the word in XML>\n
```

If you are just interested in the raw contents of the dictionary and want to process the data on your own, this will be all you want.

## Export dictionary contents into a text file
```
$ apple-dictionary-parser text > /tmp/all-words.txt
```

## Export dictionary contents into a json file
```
$ apple-dictionary-parser json > /tmp/all-words.json
```

## Export dictionary contents into a HTML file
```
$ apple-dictionary-parser html > /tmp/all-words.html
```

## Export dictionary contents into alphabetically separated HTML files

```
$ apple-dictionary-parser html-split --out-dir /tmp/
```

This generates a.html, b.html, ..., z.html files in a given directory.

### Useful options to filter words

If you want to filter words to extract, you can use filtering options such as `--words` or `--words-file`

```
$ apple-dictionary-parser text --words=--words=happiness,joy,pleasure

$ apple-dictionary-parser html --words-file=your-words.txt
```

These filtering options are applicable to most of subcommands.

## Analyze etymology data

```
$ apple-dictionary-parser etym /tmp/
```

This analyzes etymology graph and make outputs in various formats (yaml, html)

## Show phonetics information

```
$ apple-dictionary-parser phonetics
```

# License
MIT

# Author
@DQNEO
