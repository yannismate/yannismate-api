FROM golang:1.18.1-alpine as build

WORKDIR /go/src/yannismate-api

COPY go.mod ./
COPY go.sum ./
COPY libs ./libs

RUN go mod download

WORKDIR /go/src/yannismate-api/services/api

COPY services/api/* ./

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api .

FROM alpine:latest
WORKDIR /root/
COPY --from=build /go/src/yannismate-api/services/api/api ./
COPY --from=build /go/src/yannismate-api/services/api/config.json ./
CMD ["./api"]