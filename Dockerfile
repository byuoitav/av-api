FROM alpine

RUN mkdir -p /go
ADD . /go

WORKDIR /go

CMD ["/go/av-api"]

EXPOSE 8000
