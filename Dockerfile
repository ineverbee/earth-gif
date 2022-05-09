#build stage
FROM golang:1.18.1-alpine3.15 AS builder
RUN apk add --no-cache git
WORKDIR /go/src/app
COPY . .
RUN go get -d -v ./...
RUN go build -o /go/bin/app -v ./main.go

#final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
ENV API_KEY=q5efKXwGHYJpub1h9bUqHmOS1ERPYQCRAh5wic3H
COPY --from=builder /go/bin/app /app
COPY ./Helvetica.ttf .
ENTRYPOINT ["/app"]
LABEL Name=earthgif Version=0.0.1
EXPOSE 8080