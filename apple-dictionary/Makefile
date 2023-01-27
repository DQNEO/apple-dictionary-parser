DICT_FILE = "/System/Library/AssetsV2/com_apple_MobileAsset_DictionaryServices_dictionaryOSX/6b98409a6f704b07449c95dead92a7911dba87d6.asset/AssetData/New Oxford American Dictionary.dictionary/Contents/Resources/Body.data"

all: out/noad.sample.html out/noad.parsed.txt

out/noad.sample.html: out/noad.dump.txt parse.go
	go run parse.go --words=happiness,joy,felicity,pleasure --mode=html $<   > $@

out/noad.parsed.txt: out/noad.dump.txt parse.go
	go run parse.go --mode=text  $< text  > $@

out/noad.dump.txt: extract.go
	 go run extract.go $(DICT_FILE) > $@

clean:
	rm -f out/*.html out/*.txt
