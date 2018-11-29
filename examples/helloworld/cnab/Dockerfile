FROM alpine:latest

RUN apk add -u bash

COPY Dockerfile /cnab/Dockerfile
COPY app /cnab/app

CMD ["/cnab/app/run"]
