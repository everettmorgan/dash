FROM ubuntu:latest

RUN mkdir -p /dashd

WORKDIR /dashd

COPY . .

ENTRYPOINT ./dash