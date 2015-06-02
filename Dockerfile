FROM golang:1.4

RUN go get github.com/jstemmer/go-junit-report && \
	go get github.com/axw/gocov/gocov && \
	go get github.com/AlekSi/gocov-xml && \
	go get golang.org/x/tools/cmd/cover

VOLUME /go/src/github.com/Synthace
WORKDIR /go/src/github.com/Synthace/goflow
