FROM golang:1.7beta2-wheezy

ADD . /go/src/sevki.org/cloud/gcs

RUN go get -v sevki.org/cloud/gcs

ENTRYPOINT gcs

EXPOSE 8080