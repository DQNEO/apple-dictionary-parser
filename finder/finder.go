package finder

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func FindFiles(baseDir string) (string, string, string, error) {
	var foundDicDir string
	err := filepath.Walk(baseDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if info.Name() == "New Oxford American Dictionary.dictionary" {
				foundDicDir = path
			} else {

			}
		} else {
			return nil
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	if foundDicDir == "" {
		fmt.Fprintf(os.Stderr, "Dictionary directory is not found")
		os.Exit(1)
	}
	bodyFile := foundDicDir + "/Contents/Resources/Body.data"
	cssFile := foundDicDir + "/Contents/Resources/DefaultStyle.css"
	return foundDicDir, bodyFile, cssFile, nil
}
