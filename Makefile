# run make command as follows
# make DICT_FILE="/System/Library/AssetsV2/com_apple_MobileAsset_DictionaryServices_dictionaryOSX/xxxx.asset/AssetData/New Oxford American Dictionary.dictionary/Contents/Resources/Body.data"
CSS_FILES := out/DefaultStyle.css out/customize.css
CACHE := /tmp/.noad.cache

all: out/groups/a.html out/noad.sample1.html out/noad.sample2.html out/noad.txt

out/DefaultStyle.css:
	DIR=`dirname "${DICT_FILE}"`; cp "$$DIR/DefaultStyle.css" $@

out/customize.css: customize.css
	cp $< $@

extract: extract.go
	go build -o $@ $<

format: format.go html_template.go parser/*
	go build -o $@ format.go html_template.go

$(CACHE): extract
	 ./extract "${DICT_FILE}" > $@

out/noad.sample1.html: $(CACHE) $(CSS_FILES) format
	./format --mode=html --words=happiness,joy,felicity,pleasure $<   > $@

out/noad.sample2.html: $(CACHE) words-sample.txt $(CSS_FILES) format
	./format --mode=html --words-file=words-sample.txt  $<   > $@

out/noad.txt: $(CACHE) format
	./format --mode=text  $< > $@

out/groups/a.html: $(CACHE) format $(CSS_FILES) groups_index.html
	mkdir -p out/groups
	cp out/*.css out/groups/
	cp groups_index.html out/groups/index.html
	./format --mode=htmlsplit $< out/groups


.PHONY: etym
etym: $(CACHE) format
	./format --mode=etym $< out

clean:
	rm -fr out/* ; rm -f $(CACHE) extract format


.PHONY: debug
debug: $(CACHE) format
	./format --mode=debug --words-file=../lexicon/passtan1/data/2100plus.txt $< >out/etym.txt
