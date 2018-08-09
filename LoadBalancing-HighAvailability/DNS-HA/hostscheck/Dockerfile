FROM alpine:edge

RUN apk add bash && apk add bash-doc && apk add bash-completion
RUN rm -rf /var/cache/apk/*

ADD ./entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
CMD ["/cyphyroot/log/cyphy_warning.log","/etc/hosts","/etc/dnsmasq/hosts","dnsmasq","1"]  

