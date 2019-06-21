FROM golang:1.12

WORKDIR /build

COPY . .

RUN go get ./...

RUN go build -o saas ./cmd

CMD ["/build/saas"]
