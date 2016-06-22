FROM golang:1.7beta1-wheezy

ADD . /go/src/sevki.org/cloud/

RUN go get sevki.org/cloud/wiki



ENTRYPOINT wiki

EXPOSE 8080