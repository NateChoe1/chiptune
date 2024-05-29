TUNES=$(wildcard tunes/*.tune)
OUT=$(subst tunes/,examples/,$(subst .tune,.wav,$(TUNES)))

all: $(OUT)

chiptune: $(wildcard *.go)
	go build .

examples/%.wav: tunes/%.tune chiptune
	./chiptune $< $@

.PHONY: all
