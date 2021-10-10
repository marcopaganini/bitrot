.PHONY: clean

bin := bitrot
bindir := /usr/local/bin
src := $(wildcard *.go)
git_tag := $(shell git describe --tags)

# Default target
${bin}: Makefile ${src}
	go build -v -ldflags "-X main.Build=${git_tag}" -o "${bin}"

clean:
	rm -f "${bin}"

install: ${bin}
	install -m 755 "${bin}" "${bindir}"
