all: build

.PHONY: build
build:
	go build -o matnet

.PHONY: clean
clean:
	rm -f matnet