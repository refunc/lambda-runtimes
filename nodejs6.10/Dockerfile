FROM lambci/lambda-base

ENV PATH=/var/lang/bin:$PATH \
  LD_LIBRARY_PATH=/var/lang/lib:$LD_LIBRARY_PATH \
  AWS_EXECUTION_ENV=AWS_Lambda_nodejs6.10 \
  NODE_PATH=/opt/nodejs/node6/node_modules:/opt/nodejs/node_modules:/var/runtime/node_modules:/var/runtime:/var/task:/var/runtime/node_modules

RUN rm -rf /var/runtime /var/lang && \
  curl https://lambci.s3.amazonaws.com/fs/nodejs6.10.tgz | tar -zx -C /

COPY bootstrap /var/runtime/bootstrap
COPY awslambda-mock.js /var/runtime/node_modules/awslambda/build/Release/awslambda.js

USER sbx_user1051

ENTRYPOINT ["/var/runtime/bootstrap"]
