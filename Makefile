.PHONY: build install clean test scan view

BINARY := map-axi
CMD := ./cmd/map-axi

build:
	go build -o $(BINARY) $(CMD)

install: build
	mkdir -p $(HOME)/.local/bin
	cp $(BINARY) $(HOME)/.local/bin/

clean:
	rm -f $(BINARY)
	rm -f architecture.json

test:
	go test ./...

scan:
	$(BINARY) scan .

view:
	$(BINARY) view

view-d2:
	$(BINARY) view --format d2

view-mermaid:
	$(BINARY) view --format mermaid

diff:
	$(BINARY) diff $(BASE)

# watchexec -w . -e go -- "make scan && make view"