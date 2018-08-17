FROM alpine

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories

RUN apk --update -t add drill bind-tools bash bash-completion nginx nginx-mod-stream
RUN rm -rf /var/cache/apk/*


RUN rm -rf /etc/nginx/nginx.conf.default /etc/nginx/mime.types /etc/nginx/nginx.conf
RUN mkdir /etc/nginx/ -p
ADD nginx.conf.default /etc/nginx/nginx.conf.default
ADD mime.types /etc/nginx/mime.types

ADD ./entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# expose both the HTTP (80) and HTTPS (443) ports
EXPOSE 80 443

ENTRYPOINT ["/entrypoint.sh"]
