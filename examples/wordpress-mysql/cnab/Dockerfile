FROM alpine:latest

ENV HELM_LATEST_VERSION="v2.10.0"

# install helm and azure-cli
RUN apk add --update ca-certificates \
 && apk add --update -t deps wget \
 && wget https://storage.googleapis.com/kubernetes-helm/helm-${HELM_LATEST_VERSION}-linux-amd64.tar.gz \
 && tar -xvf helm-${HELM_LATEST_VERSION}-linux-amd64.tar.gz \
 && mv linux-amd64/helm /usr/local/bin \
 && rm -f /helm-${HELM_LATEST_VERSION}-linux-amd64.tar.gz \
 && apk add bash py-pip \
 && apk add --virtual=build gcc libffi-dev musl-dev openssl-dev python-dev make \
 && pip install --upgrade pip \
 && pip install azure-cli \
 && apk del --purge deps \
 && rm /var/cache/apk/*

COPY app/run /cnab/app/run
COPY Dockerfile cnab/Dockerfile

CMD [ "/cnab/app/run" ]
