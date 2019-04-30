## makefile for airfk.

all: build install

appname=fktool

build:
	go build -o dist/${appname} cmd/tool/tool.go

install:
	cp -f dist/${appname} ${GOBIN}/

clean:
	rm -f dist/${appname}
