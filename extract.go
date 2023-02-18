// This program parses a dictionary file in MacOS
// Basic technique is inspired by this article : https://fmentzer.github.io/posts/2020/dictionary/

// Usage:
//
//	DICT_FILE="/System/Library/AssetsV2/com_apple_MobileAsset_DictionaryServices_dictionaryOSX/xxxxxx.asset/AssetData/New Oxford American Dictionary.dictionary/Contents/Resources/Body.data"
//	./extract -o noad.txt $DICT_FILE
package main

import (
	"flag"
	"fmt"
	"github.com/DQNEO/apple-dictionary/cache"
	"github.com/DQNEO/apple-dictionary/extracter"
	"io"
	"os"
)

var outFilePath = flag.String("o", "", "output file")

func main() {
	flag.Parse()
	dicFilePath := flag.Arg(0)
	if dicFilePath == "" {
		fmt.Fprintf(os.Stderr, "Please specify a dictionary body file.\n For example: '/System/Library/AssetsV2/com_apple_MobileAsset_DictionaryServices_dictionaryOSX/xxxxxx.asset/AssetData/New Oxford American Dictionary.dictionary/Contents/Resources/Body.data'\n")
		os.Exit(1)
	}
	var cacheFile io.Writer
	if *outFilePath == "" {
		cacheFile = os.Stdout
	} else {
		oFile, err := os.Create(*outFilePath)
		if err != nil {
			panic(err)
		}
		cacheFile = oFile
	}
	entries := extracter.ParseBinaryFile(dicFilePath)
	cache.SaveEntries(cacheFile, entries)
}
