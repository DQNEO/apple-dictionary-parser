# run make command as follows
# make DICT_FILE="/System/Library/AssetsV2/com_apple_MobileAsset_DictionaryServices_dictionaryOSX/xxxx.asset/AssetData/New Oxford American Dictionary.dictionary/Contents/Resources/Body.data"
CSS_FILES := /tmp/DefaultStyle.css /tmp/customize.css
CACHE := /tmp/.noad.cache

all: out/groups/a.html out/noad.sample1.html out/noad.sample2.html out/noad.txt

/tmp/DefaultStyle.css:
	DIR=`dirname "${DICT_FILE}"`; cp "$$DIR/DefaultStyle.css" $@

/tmp/customize.css: customize.css
	cp $< $@

extract: extract.go
	go build -o $@ $<

adp: main.go html_template.go parser/*
	go build -o $@ main.go html_template.go

$(CACHE): extract
	 ./extract  -o $@ "${DICT_FILE}"

out/noad.sample1.html: $(CACHE) $(CSS_FILES) adp
	./adp --mode=html --words=happiness,joy,felicity,pleasure > $@

out/noad.sample2.html: $(CACHE) words-sample.txt $(CSS_FILES) adp
	./adp --mode=html --words-file=words-sample.txt  > $@

out/noad.txt: $(CACHE) adp
	./adp --mode=text > $@

out/groups/a.html: $(CACHE) adp $(CSS_FILES) groups_index.html
	mkdir -p out/groups
	cp out/*.css out/groups/
	cp groups_index.html out/groups/index.html
	./adp --mode=htmlsplit out/groups


.PHONY: etym
etym: $(CACHE) adp
	./adp --mode=etym out

clean:
	rm -fr out/* ; rm -f $(CACHE) extract adp


.PHONY: debug
debug: $(CACHE) adp
	./adp --mode=debug --words-file=../lexicon/passtan1/data/2100plus.txt >out/etym.txt
