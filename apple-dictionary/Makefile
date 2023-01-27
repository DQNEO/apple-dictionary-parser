DICT_FILE = "/System/Library/AssetsV2/com_apple_MobileAsset_DictionaryServices_dictionaryOSX/7e66d0bf940535a6ed4e0b6b29b6879cecc18477.asset/AssetData/New Oxford American Dictionary.dictionary/Contents/Resources/Body.data"

out/noad.txt: extract.go
	 go run extract.go $(DICT_FILE) > $@
