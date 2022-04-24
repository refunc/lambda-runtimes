FROM node:12-buster

COPY sources.list /etc/apt/sources.list

RUN apt-get update && \
  apt-get install -y \
  g++ \
  make \
  cmake \
  unzip \
  libcurl4-openssl-dev

RUN mkdir -p /var/task /var/runtime

ADD boot /var/runtime/boot
# RUN npm install aws-lambda-ric@2.0.0 --save
RUN cd /var/runtime/boot && npm install @javascriptbach/aws-lambda-ric@2.0.1 --registry=https://registry.npmmirror.com --save
RUN cd /var/runtime && ln -s ./boot/boot.js bootstrap

WORKDIR /var/task

ENTRYPOINT ["/var/runtime/bootstrap"]