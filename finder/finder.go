package finder

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

func FindDictFile(baseDir string) (string, string, string, error) {
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
		return "", "", "", errors.New("Dictionary directory is not found")
	}
	bodyFilePath := foundDicDir + "/Contents/Resources/Body.data"
	_, err = os.Stat(bodyFilePath)
	if err != nil {
		return "", "", "", err
	}
	cssFilePath := foundDicDir + "/Contents/Resources/DefaultStyle.css"
	_, err = os.Stat(cssFilePath)
	if err != nil {
		return "", "", "", err
	}
	return foundDicDir, bodyFilePath, cssFilePath, nil
}

func FindDefaultCSSFile(dictDir string) (string, error) {
	cssFile := dictDir + "/Contents/Resources/DefaultStyle.css"
	_, err := os.Stat(cssFile)
	if err != nil {
		return "", err
	}
	return cssFile, nil
}