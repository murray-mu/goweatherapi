FROM golang:1.13-stretch AS build
WORKDIR /src/github.com/murray-mu/goweatherapi
ADD . .
RUN make


FROM debian:stretch

ENV DEBIAN_FRONTEND=noninteractive \
    TERM=xterm

MAINTAINER murray-mu

LABEL name=goweatherapi
LABEL version=1.0.0
LABEL architecrture=amd64
LABEL source="https://github.com/murray-mu/goweatherapi.git"

COPY --from=build /src/github.com/murray-mu/goweatherapi/bin /usr/local/bin
COPY ./docs /usr/local/share/doc/goweatherapi
