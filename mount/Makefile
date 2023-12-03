
%: %.go
	go build -o bin/$@ $<

all: 
	mkdir -p bin
	for f in $(wildcard *.go); do \
		make $${f%.go}; \
	done
	