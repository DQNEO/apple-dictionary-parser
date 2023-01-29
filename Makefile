# run make command as follows
# make DICT_FILE="/System/Library/AssetsV2/com_apple_MobileAsset_DictionaryServices_dictionaryOSX/xxxx.asset/AssetData/New Oxford American Dictionary.dictionary/Contents/Resources/Body.data"

all: out/noad.sample1.html out/noad.sample2.html out/noad.parsed.txt

out/DefaultStyle.css:
	DIR=`dirname "${DICT_FILE}"`; cp "$$DIR/DefaultStyle.css" out/

out/customize.css: customize.css
	cp customize.css $@

out/noad.dump.txt: extract.go
	 go run extract.go "${DICT_FILE}" > $@

out/noad.sample1.html: out/noad.dump.txt out/DefaultStyle.css out/customize.css parse.go html_template.go
	go run parse.go html_template.go --words=happiness,joy,felicity,pleasure --mode=html $<   > $@

out/noad.sample2.html: out/noad.dump.txt out/DefaultStyle.css out/customize.css parse.go html_template.go
	go run parse.go  html_template.go --words-file=words-sample.txt --mode=html $<   > $@

out/noad.parsed.txt: out/noad.dump.txt parse.go html_template.go
	go run parse.go  html_template.go --mode=text  $< text  > $@

clean:
	rm -f out/*.html out/*.txt out/*.css
