all: build

.PHONY: build
build:
	go build -o matnet

run: build
	sudo ./matnet


.PHONY: clean
clean:
	rm -f matnet