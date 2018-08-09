FROM alpine

RUN apk --update -t add keepalived drill bind-tools bash bash-completion
#RUN apk update && apk add bash && apk add bash-doc && apk add bash-completion && apk add keepalived && apk add drill && apk add bind-tools
RUN rm -rf /var/cache/apk/*

ADD ./check_dns.sh  /check_dns.sh
RUN chmod +x /check_dns.sh

ADD ./entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]

