#! /bin/sh

ip_addr=`ifconfig eth0 | awk '/inet addr:/ {print $2}'| sed 's/addr://'`
ip addr add $1/255.255.255.0 dev eth0
ip addr del $ip_addr/24 dev eth0
systemctl restart demortsp
systemctl restart demortspalt
systemctl restart demowebserver
systemctl restart onvifsrvd
systemctl restart wsdd
