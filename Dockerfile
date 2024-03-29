FROM golang as builder

ADD . /luet-k8s
RUN cd /luet-k8s && CGO_ENABLED=0 go build

FROM ubuntu:20.04
ADD . /luet-k8s
RUN apt-get update
ENV LUET_YES=true
ENV LUET_DEBUG=true
RUN apt-get install -y uidmap curl wget git libcap2-bin
RUN curl https://luet.io/install.sh | sh
RUN chmod u-s /usr/bin/new[gu]idmap && \
    setcap cap_setuid+eip /usr/bin/newuidmap && \
    setcap cap_setgid+eip /usr/bin/newgidmap 
RUN useradd -u 1000 -d /luet -ms /bin/bash luet
RUN luet upgrade && luet install repository/mocaccino-extra && luet install container/img net-fs/minio-client container/docker

RUN mkdir -p /run/runc  && chmod 777 /run/runc

USER luet
WORKDIR /luet
COPY --from=builder /luet-k8s/luet-k8s /usr/bin/luet-k8s
RUN chmod -R 777 /luet
ENTRYPOINT "/usr/bin/luet-k8s"
