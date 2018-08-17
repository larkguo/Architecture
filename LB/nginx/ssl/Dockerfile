FROM alpine

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories

RUN apk --update -t add drill bind-tools bash bash-completion nginx nginx-mod-stream
RUN rm -rf /var/cache/apk/*

RUN rm -fr /etc/nginx/ssl/*
RUN mkdir /etc/nginx/ssl -p
ADD nginx.conf.default /etc/nginx/ssl/nginx.conf.default
ADD mime.types /etc/nginx/ssl/mime.types
ADD server.key /etc/nginx/ssl/server.key
ADD server.crt /etc/nginx/ssl/server.crt
ADD ./entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# expose both the HTTP (80) and HTTPS (443) ports
EXPOSE 80 443 

ENTRYPOINT ["/entrypoint.sh"]
