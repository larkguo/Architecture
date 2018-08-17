FROM alpine

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories


RUN apk --update -t add drill bind-tools bash bash-completion nginx 
RUN apk --update -t add nginx-mod-stream  nginx-mod-mail
RUN rm -rf /var/cache/apk/*


RUN rm -rf /etc/nginx/tcp/*
RUN mkdir /etc/nginx/tcp -p
ADD nginx.conf.default /etc/nginx/tcp/nginx.conf.default

ADD ./entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# expose both the HTTP (80) and HTTPS (443) ports
EXPOSE 80 443

ENTRYPOINT ["/entrypoint.sh"]
