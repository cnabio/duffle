FROM alpine:latest

RUN apk update
RUN apk add -u bash jq

COPY Dockerfile /cnab/Dockerfile
COPY app /cnab/app

CMD ["/cnab/app/run"]
