FROM golang:1.24.2-alpine3.21 AS build

WORKDIR /var/tmp/build
ADD go.mod go.sum /var/tmp/build/

RUN go mod download

COPY . .

RUN go build

FROM alpine:3.21 AS runtime

COPY --from=build /var/tmp/build/genaiz /usr/local/bin/genaiz

RUN apk add --no-cache bash
RUN addgroup -g 1000 -S genaiz && \
    adduser -u 1000 -S genaiz -G genaiz -s /bin/bash -h /home/genaiz && \
	mkdir -p /home/genaiz/.config/genaiz && \
    mkdir -p /home/genaiz/.cache/genaiz && \
    chown -R genaiz:genaiz /home/genaiz

USER genaiz
WORKDIR /home/genaiz

ENTRYPOINT ["genaiz"]