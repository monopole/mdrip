# THIS IS JUST PLACEHOLDER, NOT YET WORKING

FROM ubuntu
RUN apt-get update && apt-get install -y git
# FROM gcr.io/cloud-builders/go:alpine
# RUN apk update && apk add --no-cache bash git
COPY go/bin/mdrip /mdrip
EXPOSE 8080
CMD ["/mdrip",\
    "demo",\
    "--port=8080",\
    "."]
