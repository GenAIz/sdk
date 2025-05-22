FROM golang:1.24.2-alpine3.21

WORKDIR /var/tmp/build
ADD go.mod go.sum /var/tmp/build/

RUN go mod download

COPY . .

RUN go build && \
    go test ./... && \
    go install && \
    cd / &&  \
    rm -rf /var/tmp/build

RUN apk add --no-cache bash
RUN addgroup -S genaiz && \
    adduser -S genaiz -G genaiz -s /bin/bash && \
	mkdir -p /home/genaiz/.config/genaiz

ENTRYPOINT ["genaiz"]