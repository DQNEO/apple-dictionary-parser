PROG := ./apple-dictionary-parser
CACHE := /tmp/noad.cache
OUT_DIR := /tmp/adp

all: $(PROG) $(OUT_DIR)/groups/a.html $(OUT_DIR)/noad.sample1.html $(OUT_DIR)/noad.sample2.html $(OUT_DIR)/noad.txt

$(PROG): *.go cache/* extracter/*/* finder/* parser/* go.mod customize.css myapp.js
	go build

$(CACHE): $(PROG) cache/* extracter/*/* finder/*
	 $(PROG) dump
	mkdir -p $(OUT_DIR)

$(OUT_DIR)/noad.sample1.html: $(CACHE) $(PROG)
	$(PROG) html --words=happiness,joy,felicity,pleasure --out-file $@

$(OUT_DIR)/noad.sample2.html: $(CACHE) $(PROG) words-sample.txt
	$(PROG) html --words-file=words-sample.txt --out-file $@

$(OUT_DIR)/noad.txt: $(CACHE) $(PROG)
	$(PROG) text --out-file $@

$(OUT_DIR)/groups/a.html: $(CACHE) $(PROG) groups_index.html
	mkdir -p $(OUT_DIR)/groups
	cp groups_index.html $(OUT_DIR)/groups/index.html
	$(PROG) html-split --out-dir $(OUT_DIR)/groups


.PHONY: phonetics
phonetics: $(CACHE) $(PROG)
	$(PROG)  phonetics --words=happiness,joy,felicity,pleasure,apple,mango,banana,fish

.PHONY: etym
etym: $(CACHE) $(PROG)
	$(PROG) etym $(OUT_DIR)

clean:
	rm -fr $(CACHE) $(PROG) $(OUT_DIR)

.PHONY: debug
debug: $(CACHE) $(PROG)
	$(PROG) phonetics --word-regex='cally$$'

/tmp/adp/my/a.html: $(CACHE) $(PROG) mywords.txt
	rm -rf /tmp/adp/my/*
	mkdir -p /tmp/adp/my
	$(PROG) html-split --words-file=mywords.txt --out-dir /tmp/adp/my
	open /tmp/adp/my
