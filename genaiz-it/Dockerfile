FROM golang:1.24.2-alpine3.21 AS build

ADD genaiz/go.mod genaiz/go.sum /var/tmp/build/genaiz/
ADD genaiz-it/go.mod genaiz-it/go.sum /var/tmp/build/genaiz-it/
ADD genaiz-lib/go.mod genaiz-lib/go.sum /var/tmp/build/genaiz-lib/
ADD genaiz-oauth/go.mod genaiz-oauth/go.sum /var/tmp/build/genaiz-oauth/
ADD go.work go.work.sum /var/tmp/build/

WORKDIR /var/tmp/build
RUN go mod download

WORKDIR /var/tmp/build/genaiz-lib
COPY genaiz-lib/ .

WORKDIR /var/tmp/build/genaiz-oauth
COPY genaiz-oauth/ .

WORKDIR /var/tmp/build/genaiz-it
COPY genaiz-it/ .
RUN go build

FROM alpine:3.21 AS runtime

COPY --from=build /var/tmp/build/genaiz-it/genaiz-it /usr/local/bin/genaiz-it

RUN apk add --no-cache bash
RUN addgroup -g 1000 -S genaiz && \
    adduser -u 1000 -S genaiz -G genaiz -s /bin/bash -h /home/genaiz && \
	mkdir -p /home/genaiz/.config/genaiz && \
    mkdir -p /home/genaiz/.cache/genaiz && \
    chown -R genaiz:genaiz /home/genaiz

USER genaiz
WORKDIR /home/genaiz

ENTRYPOINT ["genaiz-it"]