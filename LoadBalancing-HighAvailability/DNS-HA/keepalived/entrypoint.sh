#!/bin/bash

echo params=$@

if [ "$vip" != "" ]; then
  sed -i -e "/#vip/c ${vip} #vip" /etc/keepalived/keepalived.conf
fi
if [ "$VIP" != "" ]; then
  sed -i -e "/#vip/c ${VIP} #vip" /etc/keepalived/keepalived.conf
fi


if [ "$peer1" != "" ]; then
  sed -i -e "/#peer1/c ${peer1} #peer1" /etc/keepalived/keepalived.conf
fi

if [ "$peer2" != "" ]; then
  sed -i -e "/#peer2/c ${peer2} #peer2" /etc/keepalived/keepalived.conf
fi

if [ "$peer3" != "" ]; then
  sed -i -e "/#peer3/c ${peer3} #peer3" /etc/keepalived/keepalived.conf
fi


if [ "$mcast_src_ip" != "" ]; then
  sed -i -e "/mcast_src_ip/c mcast_src_ip ${mcast_src_ip}" /etc/keepalived/keepalived.conf
fi

if [ "$interface" != "" ]; then
  sed -i -e "/interface/c interface ${interface}" /etc/keepalived/keepalived.conf
fi

if [ "$priority" != "" ]; then
  sed -i -e "/priority/c priority ${priority}" /etc/keepalived/keepalived.conf
fi

if [ "$state" != "" ]; then
  sed -i -e "/state/c state ${state}" /etc/keepalived/keepalived.conf
fi


exec /usr/sbin/keepalived --dont-fork --log-console $@


