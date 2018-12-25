FROM lambci/lambda-base

ENV AWS_EXECUTION_ENV=AWS_Lambda_python2.7

RUN rm -rf /var/runtime /var/lang && \
  curl https://lambci.s3.amazonaws.com/fs/python2.7.tgz | tar -zx -C /

RUN rm /var/runtime/awslambda/runtime.so

COPY bootstrap /var/runtime/bootstrap
COPY runtime-mock.py /var/runtime/awslambda/runtime.py
COPY lambda_runtime_client.py /var/runtime/awslambda/lambda_runtime_client.py

USER sbx_user1051

ENTRYPOINT ["/var/runtime/bootstrap"]
