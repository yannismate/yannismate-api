FROM golang:1.18.1-alpine as build

WORKDIR /go/src/app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o webscraper .

FROM alpine:latest
WORKDIR /root/
COPY --from=build /go/src/app/webscraper ./
COPY public ./public
CMD ["./webscraper"]