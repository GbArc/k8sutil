FROM debian:buster-slim

RUN apt update && apt install -y wget
RUN wget https://go.dev/dl/go1.18.2.linux-amd64.tar.gz && tar zxf go1.18.2.linux-amd64.tar.gz
RUN mkdir -p /app
RUN mkdir -p /root/.kube
WORKDIR /app
COPY . /app
RUN echo "export PATH=/go/bin:$PATH" >> /root/.bashrc
RUN export PATH=/go/bin:$PATH && cd /app && go get
