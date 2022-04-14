FROM golang:1.18.1-alpine as build

WORKDIR /go/src/yannismate-api

COPY go.mod ./
COPY go.sum ./
COPY libs ./libs

RUN go mod download

WORKDIR /go/src/yannismate-api/services/webscraper

COPY services/webscraper/* ./

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o webscraper .

FROM alpine:latest
WORKDIR /root/
COPY --from=build /go/src/yannismate-api/services/webscraper/webscraper ./
COPY --from=build /go/src/yannismate-api/services/webscraper/config.json ./
CMD ["./webscraper"]