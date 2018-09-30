FROM golang:1.11-alpine
RUN apk update && apk upgrade && apk add git

WORKDIR /go/src/github.com/cbelsole/dsw
COPY . .

RUN go build -o app cmd/server/main.go

EXPOSE 8080

CMD ["./app"]
