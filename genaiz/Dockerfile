FROM golang:1.24.2-alpine3.21 AS build

WORKDIR /var/tmp/build
ADD go.mod go.sum /var/tmp/build/

RUN go mod download

COPY . .

RUN go build && go test ./...

FROM alpine:3.21 AS runtime

COPY --from=build /var/tmp/build/genaiz /usr/local/bin/genaiz

RUN apk add --no-cache bash
RUN addgroup -S genaiz && \
    adduser -S genaiz -G genaiz -s /bin/bash && \
	mkdir -p /home/genaiz/.config/genaiz

ENTRYPOINT ["genaiz"]