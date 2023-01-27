# run make command as follows
# make DICT_FILE="/System/Library/AssetsV2/com_apple_MobileAsset_DictionaryServices_dictionaryOSX/xxxx.asset/AssetData/New Oxford American Dictionary.dictionary/Contents/Resources/Body.data"

all: out/noad.sample.html out/noad.parsed.txt

out/DefaultStyle.css:
	DIR=`dirname "${DICT_FILE}"`; cp "$$DIR/DefaultStyle.css" out/

out/noad.dump.txt: extract.go
	 go run extract.go "${DICT_FILE}" > $@

out/noad.sample.html: out/noad.dump.txt out/DefaultStyle.css parse.go
	go run parse.go --words=happiness,joy,felicity,pleasure --mode=html $<   > $@

out/noad.parsed.txt: out/noad.dump.txt parse.go
	go run parse.go --mode=text  $< text  > $@

clean:
	rm -f out/*.html out/*.txt out/*.css
