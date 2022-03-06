FROM golang:stretch
  
ADD . /k8s-deployment-watcher
WORKDIR /k8s-deployment-watcher

RUN make dwserver && \
    make dwcl
