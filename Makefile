BINARY_NAME=gophinator

all:
	mkdir -p bin
	go build -o bin/$(BINARY_NAME) -v

clean:
	rm -rf bin

.PHONY: all clean
