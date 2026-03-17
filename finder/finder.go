package finder

import (
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

var BaseDir = "/System/Library/AssetsV2/com_apple_MobileAsset_DictionaryServices_dictionaryOSX"
var assetsV3BaseDir = "/System/Library/AssetsV2/com_apple_MobileAsset_DictionaryServices_dictionary3macOS"
var legacyBaseDir = "/Library/Dictionaries"

func resolveBaseDir() string {
	major, err := macMajorVersion()
	if err != nil {
		return BaseDir
	}
	if major < 11 {
		return legacyBaseDir
	}
	if _, err := os.Stat(assetsV3BaseDir); err == nil {
		return assetsV3BaseDir
	}
	return BaseDir
}

func macMajorVersion() (int, error) {
	cmd := exec.Command("sw_vers", "-productVersion")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	versionStr := strings.TrimSpace(string(output))
	dot := strings.Index(versionStr, ".")
	if dot == -1 {
		return strconv.Atoi(versionStr)
	}
	majorStr := strings.TrimLeft(versionStr[:dot], "0")
	if majorStr == "" {
		majorStr = "0"
	}
	return strconv.Atoi(majorStr)
}

func FindDictFile() (string, string, string, error) {
	baseDir := resolveBaseDir()
	candidates := []string{}
	appendCandidate := func(dir string) {
		if dir == "" {
			return
		}
		for _, existing := range candidates {
			if existing == dir {
				return
			}
		}
		candidates = append(candidates, dir)
	}
	appendCandidate(baseDir)
	appendCandidate(assetsV3BaseDir)
	appendCandidate(BaseDir)
	appendCandidate(legacyBaseDir)

	var foundDicDir string
	var walkErr error
	walk := func(root string) error {
		return filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() && info.Name() == "New Oxford American Dictionary.dictionary" {
				foundDicDir = path
			}
			return nil
		})
	}

	for _, dir := range candidates {
		if _, err := os.Stat(dir); err != nil {
			continue
		}
		walkErr = walk(dir)
		if walkErr != nil || foundDicDir != "" {
			break
		}
	}

	if walkErr != nil {
		return "", "", "", walkErr
	}
	if foundDicDir == "" {
		return "", "", "", errors.New("Dictionary directory is not found")
	}
	bodyFilePath := foundDicDir + "/Contents/Resources/Body.data"
	if _, err := os.Stat(bodyFilePath); err != nil {
		return "", "", "", err
	}
	cssFilePath := foundDicDir + "/Contents/Resources/DefaultStyle.css"
	if _, err := os.Stat(cssFilePath); err != nil {
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
