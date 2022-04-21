FROM golang:1.17-buster as build-image

COPY bootstrap.go /go/bootstrap.go

RUN go build -o bootstrap /go/bootstrap.go

FROM golang:1.17-buster

RUN mkdir -p /var/task /var/runtime

COPY --from=build-image /go/bootstrap /var/runtime/bootstrap

WORKDIR /var/task

ENTRYPOINT ["/var/runtime/bootstrap"]