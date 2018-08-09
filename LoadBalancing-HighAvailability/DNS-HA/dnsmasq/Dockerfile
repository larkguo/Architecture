FROM alpine:latest
RUN apk update

WORKDIR /usr/local/src
COPY ./dnsmasq-2.79-answercount.tar.gz /usr/local/src/dnsmasq-2.79-answercount.tar.gz
RUN tar xfz dnsmasq-2.79-answercount.tar.gz

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories
RUN apk add bash bash-completion alpine-sdk linux-headers 
RUN rm -rf /var/cache/apk/*

WORKDIR /usr/local/src/dnsmasq-2.79
RUN make
RUN make install
RUN make clean

EXPOSE 53 53/udp

ENTRYPOINT ["dnsmasq", "-k", "-q", "--no-hosts", "--addn-hosts=/etc/dnsmasq/hosts", "--log-facility=-","--conf-file=/etc/dnsmasq/dnsmasq.conf"]

