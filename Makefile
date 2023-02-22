PROG := ./apple-dictionary-parser
CACHE := /tmp/noad.cache
OUT_DIR := /tmp/adp

all: $(OUT_DIR)/groups/a.html $(OUT_DIR)/noad.sample1.html $(OUT_DIR)/noad.sample2.html $(OUT_DIR)/noad.txt

$(PROG): *.go cache/* extracter/*/* finder/* parser/* go.mod customize.css
	go build

$(CACHE): extracter/*/*  finder/* cache/*
	 $(PROG) dump
	mkdir -p $(OUT_DIR)

$(OUT_DIR)/noad.sample1.html: $(CACHE) $(PROG)
	$(PROG) html --words=happiness,joy,felicity,pleasure $@

$(OUT_DIR)/noad.sample2.html: $(CACHE) $(PROG) words-sample.txt
	$(PROG) html --words-file=words-sample.txt $@

$(OUT_DIR)/noad.txt: $(CACHE) $(PROG)
	$(PROG) text $@

$(OUT_DIR)/groups/a.html: $(CACHE) $(PROG) groups_index.html
	mkdir -p $(OUT_DIR)/groups
	cp groups_index.html $(OUT_DIR)/groups/index.html
	$(PROG) htmlsplit $(OUT_DIR)/groups


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
	$(PROG) debug --words=happiness,joy,felicity,pleasure,apple,mango,banana,fish
