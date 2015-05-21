FROM golang:1.4

#RUN apt-get update && apt-get install -y libglpk-dev libsqlite3-dev && rm -rf /var/lib/apt/lists/*
RUN go get github.com/jstemmer/go-junit-report
RUN go get github.com/axw/gocov/gocov
RUN go get github.com/AlekSi/gocov-xml
RUN go get golang.org/x/tools/cmd/cover

VOLUME /go/src/github.com/Synthace
WORKDIR /go/src/github.com/Synthace/goflow
