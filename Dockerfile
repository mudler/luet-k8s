FROM quay.io/mocaccino/extra as builder

ADD . /luet-k8s
RUN luet install -y lang/go
RUN cd /luet-k8s && CGO_ENABLED=0 go build

FROM ubuntu:20.04
RUN apt-get update
ENV LUET_YES=true

RUN apt-get install -y uidmap curl wget git libcap2-bin
RUN curl https://get.mocaccino.org/luet/get_luet_root.sh | sh
RUN luet install repository/mocaccino-extra
RUN luet install container/img net-fs/minio-client
RUN luet upgrade
RUN chmod u-s /usr/bin/new[gu]idmap && \
    setcap cap_setuid+eip /usr/bin/newuidmap && \
    setcap cap_setgid+eip /usr/bin/newgidmap 

RUN mkdir -p /run/runc  && chmod 777 /run/runc

RUN useradd -u 1000 -d /luet -ms /bin/bash luet
USER luet
WORKDIR /luet
COPY --from=builder /luet-k8s/luet-k8s /usr/bin/luet-k8s
RUN chmod -R 777 /luet
ENTRYPOINT "/usr/bin/luet-k8s"
