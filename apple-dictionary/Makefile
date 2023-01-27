DICT_FILE = "/System/Library/AssetsV2/com_apple_MobileAsset_DictionaryServices_dictionaryOSX/6b98409a6f704b07449c95dead92a7911dba87d6.asset/AssetData/New Oxford American Dictionary.dictionary/Contents/Resources/Body.data"

all: out/noad.html out/noad.parsed.txt

out/noad.html: out/noad.raw.txt parse.go
	go run parse.go $< html   > $@

out/noad.parsed.txt: out/noad.raw.txt parse.go
	go run parse.go $< text  > $@

out/noad.raw.txt: extract.go
	 go run extract.go $(DICT_FILE) > $@

clean:
	rm out/*.html out/*.txt
