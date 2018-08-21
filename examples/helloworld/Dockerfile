FROM alpine:latest

COPY cnab/app/run /cnab/app/run
COPY cnab/bundle.json /cnab/bundle.json
COPY Dockerfile cnab/Dockerfile

CMD [ "/cnab/app/run" ]