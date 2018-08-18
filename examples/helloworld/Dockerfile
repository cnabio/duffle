FROM alpine:latest

COPY cnab/app/run /cnab/app/run
COPY cnab/manifest.json /cnab/manifest.json
COPY cnab/parameters.json /cnab/parameters.json
COPY Dockerfile cnab/Dockerfile

CMD [ "/cnab/app/run" ]