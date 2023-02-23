package finder

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

var BaseDir = "/System/Library/AssetsV2/com_apple_MobileAsset_DictionaryServices_dictionaryOSX"

func FindDictFile() (string, string, string, error) {
	baseDir := BaseDir
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
