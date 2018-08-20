FROM node:10-alpine

COPY ./cnab/app/run /cnab/app/run
COPY ./Dockerfile /cnab/Dockerfile
COPY ./cnab/manifest.json /cnab/manifest.json
COPY ./index.js /cnab/app/index.js
COPY ./package.json /cnab/app/package.json

RUN chmod 755 /cnab/app/run

CMD ["/cnab/app/run"]