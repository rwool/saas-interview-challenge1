FROM golang:1.12-alpine

RUN apk --no-cache add ca-certificates git
WORKDIR /build
COPY . .
RUN go get -mod=readonly ./...
RUN go build -mod=readonly -ldflags='-s -w' -o saas ./cmd

FROM alpine

RUN apk --no-cache add ca-certificates
RUN adduser -D -H saas
USER saas
WORKDIR /saas
COPY --from=0 /build/saas .
CMD ["./saas"]
