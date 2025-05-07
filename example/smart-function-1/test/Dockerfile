FROM alpine:latest

RUN apk add --no-cache bash

RUN mkdir -p /opt/genaiz/smart-function-1/bin
ADD start.sh /opt/genaiz/smart-function-1/bin/
RUN chmod +x /opt/genaiz/smart-function-1/bin/start.sh

ENTRYPOINT ["/opt/genaiz/smart-function-1/bin/start.sh", "test file"]