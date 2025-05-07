FROM alpine:latest

RUN apk add --no-cache bash

ADD start.sh /opt/genaiz/smart-function-1/bin/
RUN chmod +x /opt/genaiz/smart-function-1/bin/start.sh

CMD ["/opt/genaiz/smart-function-1/bin/start.sh", "default file"]