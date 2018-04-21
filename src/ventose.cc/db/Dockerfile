FROM golang

ADD . /go/src/ventose.cc/db

RUN go get -t -d ./...
RUN go install ventose.cc/db
RUN go test ventose.cc/db

ENTRYPOINT /go/bin/db