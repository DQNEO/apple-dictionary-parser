PROG := ./adp
CACHE := /tmp/noad.cache
OUT_DIR := /tmp/adp

all: $(OUT_DIR)/groups/a.html $(OUT_DIR)/noad.sample1.html $(OUT_DIR)/noad.sample2.html $(OUT_DIR)/noad.txt

$(PROG): *.go cache/* extracter/*/* finder/* parser/* go.mod customize.css
	go build -o $@

$(CACHE): $(PROG)
	 $(PROG) --mode=dump
	mkdir -p $@

$(OUT_DIR)/noad.sample1.html: $(CACHE) $(PROG)
	$(PROG) --mode=html --words=happiness,joy,felicity,pleasure > $@

$(OUT_DIR)/noad.sample2.html: $(CACHE) $(PROG) words-sample.txt
	$(PROG) --mode=html --words-file=words-sample.txt  > $@

$(OUT_DIR)/noad.txt: $(CACHE) $(PROG)
	$(PROG) --mode=text > $@

$(OUT_DIR)/groups/a.html: $(CACHE) $(PROG) groups_index.html
	mkdir -p $(OUT_DIR)/groups
	cp groups_index.html $(OUT_DIR)/groups/index.html
	$(PROG) --mode=htmlsplit $(OUT_DIR)/groups


.PHONY: etym
etym: $(CACHE) $(PROG)
	$(PROG) --mode=etym $(OUT_DIR)

clean:
	rm -fr $(CACHE) $(PROG) $(OUT_DIR)

.PHONY: debug
debug: $(CACHE) adp
	$(PROG) --mode=debug --words-file=../lexicon/passtan1/data/2100plus.txt > $(OUT_DIR)/debug.txt
