CSS_FILES := /tmp/DefaultStyle.css /tmp/customize.css
CACHE := /tmp/noad.cache
PROG := ./adp

all: out/groups/a.html out/noad.sample1.html out/noad.sample2.html out/noad.txt

/tmp/DefaultStyle.css:
	DIR=`dirname "${DICT_FILE}"`; cp "$$DIR/DefaultStyle.css" $@

/tmp/customize.css: customize.css
	cp $< $@

$(PROG): main.go html_template.go  cache/* extracter/*/* finder/* parser/* go.mod
	go build -o $@ main.go html_template.go

$(CACHE): $(PROG)
	 $(PROG) --mode=dump

out/noad.sample1.html: $(CACHE) $(CSS_FILES) $(PROG)
	$(PROG) --mode=html --words=happiness,joy,felicity,pleasure > $@

out/noad.sample2.html: $(CACHE) $(PROG) words-sample.txt $(CSS_FILES)
	$(PROG) --mode=html --words-file=words-sample.txt  > $@

out/noad.txt: $(CACHE) $(PROG)
	$(PROG) --mode=text > $@

out/groups/a.html: $(CACHE) $(PROG) $(CSS_FILES) groups_index.html
	mkdir -p out/groups
	#cp out/*.css out/groups/
	cp groups_index.html out/groups/index.html
	$(PROG) --mode=htmlsplit out/groups


.PHONY: etym
etym: $(CACHE) $(PROG)
	$(PROG) --mode=etym out

clean:
	rm -fr out/* ; rm -f $(CACHE) $(PROG)


.PHONY: debug
debug: $(CACHE) adp
	$(PROG) --mode=debug --words-file=../lexicon/passtan1/data/2100plus.txt >out/etym.txt
