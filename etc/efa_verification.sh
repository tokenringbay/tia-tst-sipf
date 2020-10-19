#!/bin/bash

set -u

TPVM_IP=$1
USERNAME=$2
PASSWORD=$3
HOST_IP='127.2.0.1'
LOG_FILE='/var/log/efa_deploy.log'


echo Starting EFA verification script! >> $LOG_FILE
output=$(rsh $HOST_IP DEBIAN_FRONTEND=noninteractive dpkg -i /sshpass_1.05-1_amd64.deb)
echo $output >> $LOG_FILE
ls_output=$(rsh $HOST_IP sshpass -p $PASSWORD ssh $USERNAME@$TPVM_IP 'ls /opt/efa')
echo $ls_output >> $LOG_FILE
ls_output2=$(rsh $HOST_IP sshpass -p $PASSWORD ssh $USERNAME@$TPVM_IP 'ls /var/efa')
echo $ls_output2 >> $LOG_FILE
rsh $HOST_IP sshpass -p $PASSWORD ssh -o "StrictHostKeyChecking=no" $USERNAME@$TPVM_IP 'efa fabric setting show'
