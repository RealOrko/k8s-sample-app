FROM ubuntu:latest

ENV DEBIAN_FRONTEND=noninteractive
ENV PORT=8000
ENV FAIL_HEALTH_CHECK=false

COPY ./bin/sample-app /

ENTRYPOINT [ "/sample-app" ]