// This program parses a dictionary file in MacOS
// Basic technique is inspired by this article : https://fmentzer.github.io/posts/2020/dictionary/

// Usage:
//
//	NOAD_FILE="/System/Library/AssetsV2/com_apple_MobileAsset_DictionaryServices_dictionaryOSX/xxxxxx.asset/AssetData/New Oxford American Dictionary.dictionary/Contents/Resources/Body.data"
//	./extract $NOAD_FILE > noad.txt
package main

import (
	"fmt"
	"github.com/DQNEO/apple-dictionary/cache"
	"github.com/DQNEO/apple-dictionary/extracter"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Please specify a dictionary body file\n")
		os.Exit(1)
	}
	var dicFilePath = os.Args[1]
	var cacheFile = os.Stdout

	entries := extracter.ParseBinaryFile(dicFilePath)
	for _, e := range entries {
		s := cache.EntryToText(e.Title, e.Body)
		fmt.Fprint(cacheFile, s)
	}
}
