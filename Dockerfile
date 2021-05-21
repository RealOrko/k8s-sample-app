FROM ubuntu:latest

ENV DEBIAN_FRONTEND=noninteractive
ENV PORT=8000
ENV FAIL_HEALTH_CHECK=false

RUN apt update \
  && apt upgrade -y \
  && apt install dnsutils -y \
  && apt install netcat -y \
  && apt install traceroute -y \
  && apt install curl -y \
  && apt install wget -y \
  && apt install unzip -y \
  && apt install nano -y \
  && apt install busybox-static -y \
  && apt install psmisc -y

COPY ./bin/sample-app /

ENTRYPOINT [ "/sample-app" ]