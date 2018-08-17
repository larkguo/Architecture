#!/bin/bash
#set -e

echo params=$@

DefaultFile=/etc/nginx/ssl/nginx.conf.default
ConfigFile=/etc/nginx/ssl/nginx.conf

echo ConfigFile=$ConfigFile

alias cp='cp'
cp $DefaultFile $ConfigFile  -f

if [ "$server1" != "" ]; then
  sed -i -e "/#server1/c server ${server1}; #server1" $ConfigFile
fi
if [ "$server2" != "" ]; then
  sed -i -e "/#server2/c server ${server2}; #server2" $ConfigFile
fi
if [ "$server3" != "" ]; then
  sed -i -e "/#server3/c server ${server3}; #server3" $ConfigFile
fi
if [ "$server4" != "" ]; then
  sed -i -e "/#server4/c server ${server4}; #server4" $ConfigFile
fi
if [ "$server5" != "" ]; then
  sed -i -e "/#server5/c server ${server5}; #server5" $ConfigFile
fi

if [ "$listen" != "" ]; then
  sed -i -e "/#listen1/c listen ${listen} ssl; #listen1" $ConfigFile
fi
if [ "$listen1" != "" ]; then
  sed -i -e "/#listen1/c listen ${listen1} ssl; #listen1" $ConfigFile
fi
if [ "$listen2" != "" ]; then
  sed -i -e "/#listen2/c listen ${listen2} ssl; #listen2" $ConfigFile
fi
if [ "$listen3" != "" ]; then
  sed -i -e "/#listen3/c listen ${listen3} ssl; #listen3" $ConfigFile
fi

exec nginx -c $ConfigFile -g 'daemon off;' $@
