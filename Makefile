# run make command as follows
# make DICT_FILE="/System/Library/AssetsV2/com_apple_MobileAsset_DictionaryServices_dictionaryOSX/xxxx.asset/AssetData/New Oxford American Dictionary.dictionary/Contents/Resources/Body.data"

all: out/noad.sample1.html out/noad.sample2.html out/noad.parsed.txt

out/DefaultStyle.css:
	DIR=`dirname "${DICT_FILE}"`; cp "$$DIR/DefaultStyle.css" out/

out/customize.css: customize.css
	cp customize.css $@


extract: extract.go
	go build -o $@ $<

format: format.go html_template.go
	go build -o $@ format.go html_template.go

out/noad.dump: extract
	 ./extract "${DICT_FILE}" > $@

out/noad.sample1.html: out/noad.dump out/DefaultStyle.css out/customize.css format
	./format --words=happiness,joy,felicity,pleasure --mode=html $<   > $@

out/noad.sample2.html: out/noad.dump out/DefaultStyle.css out/customize.css format
	./format --words-file=words-sample.txt --mode=html $<   > $@

out/noad.parsed.txt: out/noad.dump format.go html_template.go
	./format --mode=text  $< text  > $@

clean:
	rm -f out/*.html out/*.txt out/*.css out/*.dump extract format
