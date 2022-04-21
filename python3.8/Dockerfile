FROM python:3.8-buster

COPY sources.list /etc/apt/sources.list

RUN apt-get update && \
  apt-get install -y \
  g++ \
  make \
  cmake \
  unzip \
  libcurl4-openssl-dev

RUN pip install awslambdaric==2.0.2

RUN mkdir -p /var/task /var/runtime

COPY bootstrap /var/runtime/bootstrap

WORKDIR /var/task

ENTRYPOINT ["/var/runtime/bootstrap"]