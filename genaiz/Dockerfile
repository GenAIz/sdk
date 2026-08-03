FROM golang:1.26.5-alpine3.24 AS build

ADD genaiz/go.mod genaiz/go.sum /var/tmp/build/genaiz/
ADD genaiz-lib/go.mod genaiz-lib/go.sum /var/tmp/build/genaiz-lib/
ADD go.work go.work.sum /var/tmp/build/

WORKDIR /var/tmp/build
RUN go mod download

WORKDIR /var/tmp/build/genaiz-lib
COPY genaiz-lib/ .

WORKDIR /var/tmp/build/genaiz
COPY genaiz/ .
RUN go build -tags prod

FROM alpine:3.24 AS runtime

COPY --from=build /var/tmp/build/genaiz/genaiz /usr/local/bin/genaiz

RUN apk add --no-cache bash
RUN addgroup -g 1000 -S genaiz && \
    adduser -u 1000 -S genaiz -G genaiz -s /bin/bash -h /home/genaiz && \
	mkdir -p /home/genaiz/.config/genaiz && \
    mkdir -p /home/genaiz/.cache/genaiz && \
    chown -R genaiz:genaiz /home/genaiz

USER genaiz
WORKDIR /home/genaiz

ENTRYPOINT ["genaiz"]