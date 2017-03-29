FROM alpine

RUN mkdir -p /go
ADD . /go

WORKDIR /go

CMD ["/go/av-api-x86"]

EXPOSE 8000
