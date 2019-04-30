FROM golang:alpine AS dev-env

WORKDIR /usr/local/go/src/{{ .RelDir }}
COPY . /usr/local/go/src/{{ .RelDir }}

RUN apk update && apk upgrade && \
    apk add --no-cache bash git

RUN go get ./...

RUN go build -o dist/{{.Name}} &&\
    cp -f dist/{{.Name}} /usr/local/bin/ &&\
    cp -f dist/{{.Name}}.json /usr/local/etc/ &&\

RUN ls -l && ls -l dist

CMD ["/usr/local/bin/{{.Name}}", "-c", "/usr/local/etc/{{.Name}}.json" ]