# run make command as follows
# make DICT_FILE="/System/Library/AssetsV2/com_apple_MobileAsset_DictionaryServices_dictionaryOSX/xxxx.asset/AssetData/New Oxford American Dictionary.dictionary/Contents/Resources/Body.data"

all: out/a.html out/noad.sample1.html out/noad.sample2.html out/noad.parsed.txt

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
	./format --mode=html --words=happiness,joy,felicity,pleasure $<   > $@

out/noad.sample2.html: out/noad.dump out/DefaultStyle.css out/customize.css format
	./format --mode=html --words-file=words-sample.txt  $<   > $@

out/noad.parsed.txt: out/noad.dump format
	./format --mode=text  $< > $@

out/a.html: out/noad.dump format
	./format --mode=htmlsplit $< out

clean:
	rm -f out/*.html out/*.txt out/*.css out/*.dump extract format
