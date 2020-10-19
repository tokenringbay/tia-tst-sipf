#!/bin/bash

set -u

HOST_IP='127.2.0.1'
LOG_FILE='/var/log/efa_deploy.log'

echo Starting EFA Client and Server deployment script! >> $LOG_FILE
# Create directory to mount TPVM.img in HOST
rsh $HOST_IP mkdir -p /temp
# export TPVM.img by qemu-nbd
rsh $HOST_IP qemu-nbd -c /dev/nbd0 /TPVM/TPVM.img
# dev-map /dev/nbd0
rsh $HOST_IP kpartx -a /dev/nbd0
# mount TPVM.img via device-mapper
rsh $HOST_IP mount /dev/mapper/nbd0p1 /temp
# Check mount status
# rsh $HOST_IP ls /temp/

echo Step 5 Creating required directories
rsh $HOST_IP mkdir -p /temp/var/efa
rsh $HOST_IP mkdir -p /temp/opt/efa
rsh $HOST_IP mkdir -p /temp/var/log/efa
echo Done

echo Step 6 Deploying EFA client
rcp /fabos/sbin/efa root@$HOST_IP:/temp/opt/efa/
echo Done

echo Step 7 Deploying EFA server
rcp /fabos/sbin/efa-server root@$HOST_IP:/temp/opt/efa/
rcp /fabos/sbin/efa-server.sh root@$HOST_IP:/temp/etc/init.d/efa-server
rcp /fabos/sbin/efa-server.conf root@$HOST_IP:/temp/etc/init/efa-server.conf
echo Done

echo Step 8 Adding EFA binaries to path
environ=$(rsh $HOST_IP cat /temp/etc/environment)
if echo "$environ" | grep -q "/opt/efa/"; then
    echo "EFA binaries are in path"  >> $LOG_FILE;
else
    echo "Setting up path for EFA binaries" >> $LOG_FILE;
    path="${environ::-1}:/opt/efa/\""
    if [[ ! -f "/fabos/sbin/environment" ]]; then
        touch "/fabos/sbin/environment"
    fi
    echo "$path" > "/fabos/sbin/environment"
    rcp /fabos/sbin/environment root@$HOST_IP:/temp/etc/environment
fi
echo Done
#echo Testing EFA deployment
rcp /fabos/sbin/sshpass_1.05-1_amd64.deb root@$HOST_IP:/

# dev-map /dev/nbd0
output=$(rsh $HOST_IP umount /dev/mapper/nbd0p1)
# delete dev-map /dev/nbd0
output=$(rsh $HOST_IP kpartx -d /dev/nbd0)
#disconnect TPVM.img by qemu-nbd
output=$(rsh $HOST_IP qemu-nbd -d /dev/nbd0)

