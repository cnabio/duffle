FROM alpine:latest

RUN apk update
RUN apk add -u bash

COPY Dockerfile /cnab/Dockerfile
COPY app /cnab/app

CMD ["/cnab/app/run"]
