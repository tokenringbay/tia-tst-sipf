#!/usr/bin/env python
"""Module for SIPF App deployment for Extreme SLX switches.

Copyright 2018 Extreme Networks, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
"""
import logging
import subprocess
import sys
import time

from CLI import CLI

APP_NAME = 'EFA'


def deploy_app(username, password):
    # Set up logging
    logging.basicConfig(filename='/var/log/efa_deploy.log', level=logging.DEBUG)
    # Initialize and Start TPVM and check status. Once the TPVM is running
    print('Step 1 Checking if TPVM is already setup')
    if not is_tpvm_installed():
        # Initialize TPVM
        logging.info('Initializing TPVM, since it was not setup')
        output = CLI('tpvm install', do_print=False).get_output()[0]
        logging.info(output)
        if 'ERROR: ADVANCED_FEATURES license is missing.' in output:
            print('Adding ADVANCED_FEATURES license')
            output = CLI('license eula accept ADVANCED_FEATURES', do_print=False).get_output()[0]
            logging.info(output)
            output = CLI('tpvm install', do_print=False).get_output()[0]
            logging.info(output)
        # Sleep for 20 seconds to ensure TPVM is installed
        time.sleep(20)
    print('Done')
    print('Step 2 Checking if TPVM is already running')
    if not is_tpvm_running():
        logging.info('Starting TPVM since its not running')
        output = CLI('tpvm start', do_print=False)
        logging.info(output.get_output())
        # Seep for 20 seconds to ensure that the TPVM is started
        time.sleep(20)
    print('Done')
    # Set auto-boot for TPVM
    print('Step 3 Setting auto-boot for TPVM')
    set_auto_boot_tpvm()
    print('Done')
    if is_tpvm_running():
        print('Step 4 Stopping TPVM for EFA deployment')
        output = CLI('tpvm stop', do_print=False)
        print('Done')
        logging.info(output.get_output())
        subprocess.check_call(["/fabos/sbin/efa_deployment.sh"])
        print('Step 9 Starting TPVM after EFA deployment')
        output = CLI('tpvm start', do_print=False)
        logging.info(output.get_output())
        # Seep for 30 seconds to ensure that the TPVM is started
        time.sleep(40)
        print('Done')
        print('Step 10 Get IP Address assigned to TPVM to deploy the app')
        ipv4_addresses, ipv6_addresses = get_tpvm_ip()
        ip_address = ''
        if ipv4_addresses:
            ip_address = ipv4_addresses[0]
        if not ipv4_addresses and ipv6_addresses:
            ip_address = ipv6_addresses[0]
        logging.info('IP Address of the TPVM: %s', ip_address)
        print('Done')

        if ip_address != '':
            verify_deployment(ip_address, username, password)
            print('Application is Installed and ready for use on the TPVM with IP: ', ip_address)
        else:
            print("Completed application deployment but couldn't verify due to IP not assigned to TPVM")
    else:
        print('Failed to start TPVM to deploy the application')
        print('Please clear any warnings/errors with the TPVM startup.')
        print('The output of \'show tpvm status\' should be as follows:')
        print('TPVM is running, and AutoStart is enabled on this host.')
        print('Any warnings can be cleared using tpvm show status clear-tag <tag>')
        print('After clearing all error/warning messages run the \'efa deploy\' command again')


def set_auto_boot_tpvm():
    # Set auto-boot option for TPVM. This will ensure that the VM boots up on switch reboot
    output = CLI('tpvm auto-boot enable', do_print=False)
    logging.info(output.get_output())


def is_tpvm_running():
    """
    This method returns true is TPVM is already running else returns false.
    :return: boolean - Returns true if TPVM is running else returns False
    """
    is_running = False
    output = CLI('show tpvm status', do_print=False).get_output()[0]
    logging.info(output)
    if 'TPVM is running' in output:
        is_running = True
    return is_running


def is_tpvm_installed():
    """
    This method returns truw if TPVM is installed else returns false
    :return: boolean
    """
    is_installed = True
    output = CLI('show tpvm status', do_print=False).get_output()[0]
    logging.info(output)
    if 'TPVM is not installed' in output or output.startswith('this host cannot be reached with'):
        is_installed = False
    return is_installed


def get_tpvm_ip():
    """
    Returns the IP address of the device. The 1st choice would be to return the IPv4 address and
    if that is not setup
    IPv6 address of the TPVM
    :return:
    """
    output = CLI('show tpvm ip-address', do_print=False).get_output()
    logging.info(output)
    ipv4_index = output.index('IPv4:')
    ipv6_index = output.index('IPv6:')
    ipv4_addresses = []
    start_index = ipv4_index + 1
    if start_index < ipv6_index:
        for i in range(start_index, ipv6_index):
            ipv4_addresses.append(output[i].split()[1])
    start_index = ipv6_index + 1
    ipv6_addresses = []
    if start_index <= len(output):
        for i in range(start_index, len(output)):
            ipv6_addresses.append(output[i].split()[1])

    return ipv4_addresses, ipv6_addresses


def verify_deployment(tpvm_ip, username, password):
    """
    Run some checks to ensure that the app server and client are deployed and running
    :param tpvm_ip:
    :return:
    """
    print("Step 11 Verifying Client and Server deployment on the TPVM")
    if is_tpvm_running():
        subprocess.check_call(["/fabos/sbin/efa_verification.sh", tpvm_ip, username, password])
    else:
        print('Failed to verify EFA deployment, since TPVM is not running')


def main():
    username = 'admin'
    password = 'password'
    if len(sys.argv) > 1:
        username = sys.argv[1].split("--user", 1)[1]
        username = username.strip()
    if len(sys.argv) > 2:
        password = sys.argv[2].split("--password", 1)[1]
        password = password.strip()
    deploy_app(username=username, password=password)


if __name__ == '__main__':
    main()
