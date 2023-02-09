FROM golang:1.20-alpine

RUN apk add --no-cache git

RUN mkdir /output

WORKDIR /usr/src/filament
COPY . .
RUN go generate -v ./...
RUN go build -v -o /output/filament .

###

FROM alpine:3.17

RUN apk add mailcap ca-certificates

COPY --from=0 /output/filament /usr/local/bin/filament

RUN adduser -h /config -S -D -k /var/empty -g "" -s /sbin/nologin app
USER app

WORKDIR /config

ENTRYPOINT ["filament"]
