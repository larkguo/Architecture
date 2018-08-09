#!/bin/bash

WarningFile=$1
HostsFrom=$2
MonitorFile=$3
DnsmasqName=$4
PingTimeout=$5
ModuleName="HostsCheck"
LOCALIPv4="127.0.0.1"
LOCALIPv6="::1"

echo WarningFile=$WarningFile,HostsFrom=$HostsFrom,MonitorFile=$MonitorFile,DnsmasqName=$DnsmasqName,PingTimeout=$PingTimeout 

WarningPath=${WarningFile%/*}
MonitorPath=${MonitorFile%/*}
echo WarningPath=$WarningPath,MonitorPath=$MonitorPath
mkdir $WarningPath -p
mkdir $MonitorPath -p
touch $WarningFile
touch $MonitorFile

alias cp='cp'
cp $HostsFrom $MonitorFile -f

while true
    do
    {
        LINE_NUMBER=0
        cat $MonitorFile | while read LINE
        do
        {
             LINE_NUMBER=$(($LINE_NUMBER+1))
             if [ "$LINE" == "" ]; then
                continue
             fi
             IP=`echo $LINE|sed "s/^[ \t#]*//g"|awk '{print $1}'`

             if [ "$IP" == "$LOCALIPv4" -o "$IP" == "$LOCALIPv6" ]; then
               continue
             fi
	
             ping $IP -c 1 -W $PingTimeout > /dev/null 2>&1
             res=$?
             if [ $res  -eq 0 ]; then
                    if [ "${LINE:0:1}" == "#" -o "${LINE:0:1}" == " " -o "${LINE:0:1}" == "\t" ]; then
			time=$(date "+%Y-%m-%d %H:%M:%S")
                        echo "($time) (line=$LINE_NUMBER) ($IP UP)"
                        sed -i "${LINE_NUMBER}s/^[ \t#]*//g" $MonitorFile
                        killall -s SIGHUP $DnsmasqName
			echo "[$time] [$ModuleName] [$IP UP]" >> $WarningFile
                    fi
             else
                    if [ "${LINE:0:1}" != "#" ]; then
			time=$(date "+%Y-%m-%d %H:%M:%S")
                        echo "($time) (line=$LINE_NUMBER) ($IP DOWN)"
                        sed -i "${LINE_NUMBER}s/^/#&/g" $MonitorFile
                        killall -s SIGHUP $DnsmasqName
                        echo "[$time] [$ModuleName] [$IP DOWN]" >> $WarningFile
                    fi
             fi

        }
        done
        sleep 1
    }
done


