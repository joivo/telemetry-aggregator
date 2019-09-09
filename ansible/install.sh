#!/bin/bash

MY_PATH="`dirname \"$0\"`"              
MY_PATH="`( cd \"$MY_PATH\" && pwd )`" 

if [ -z "$MY_PATH" ] ; then
  # For some reason, the path is not accessible
  # to the script (e.g. permissions re-evaled after suid)
  exit 1
fi

echo "Current directory path: "$MY_PATH

HOSTS_CONF_FILE=$MY_PATH"/hosts.conf"

DEPLOYED_HOST_IP_PATTERN="deployed_host_ip"
DEPLOYED_HOST_IP=$(grep $DEPLOYED_HOST_IP_PATTERN $HOSTS_CONF_FILE | awk -F "=" '{print $2}')

# Remote user of the host
REMOTE_USER_PATTERN="remote_user"
REMOTE_USER=$(grep $REMOTE_USER_PATTERN $HOSTS_CONF_FILE | awk -F "=" '{print $2}')

# Ansible ssh private key file path
PRIVATE_KEY_FILE_PATH_PATTERN="ansible_ssh_private_key_file"
PRIVATE_KEY_FILE_PATH=$(grep $PRIVATE_KEY_FILE_PATH_PATTERN $HOSTS_CONF_FILE | awk -F "=" '{print $2}')

# Playbook path
ANSIBLE_FILES_PATH_PATTERN="ansible_files_path"
ANSIBLE_FILES_PATH=$(grep $ANSIBLE_FILES_PATH_PATTERN $HOSTS_CONF_FILE | awk -F "=" '{print $2}')

ANSIBLE_PLAYBOOK_FILE=$ANSIBLE_FILES_PATH/"ansible-playbook"
ANSIBLE_HOSTS_FILE=$ANSIBLE_FILES_PATH/"hosts"
ANSIBLE_CFG_FILE=$ANSIBLE_FILES_PATH/"ansible.cfg"

echo "Ansible SSH private key file path: $PRIVATE_KEY_FILE_PATH"
echo "Remote user of host: $REMOTE_USER"
echo "Host IP: $DEPLOYED_HOST_IP"
echo "Remote user: $REMOTE_USER"

#Testing ssh-port
SSH_PORT_STATUS=$(nmap $DEPLOYED_HOST_IP -PN -p SSH | egrep 'open|closed|filtered')
echo "SSH port status: " $SSH_PORT_STATUS

#Testing Connection
echo "ssh -i $PRIVATE_KEY_FILE_PATH -q $REMOTE_USER@$DEPLOYED_HOST_IP exit"
ssh -i $PRIVATE_KEY_FILE_PATH -q $REMOTE_USER@$DEPLOYED_HOST_IP exit
RETURN_TEST=$?
echo "Status of the connectivity: " $RETURN_TEST

if [ $RETURN_TEST -eq 0 ]; then
   echo "Successful connection!"
else
   echo "The SSH connection failed!"
   exit 1
fi

# Delete everthing between[aggregator-machine] and [aggregator-machine:vars]
sed -i 's/\[aggregator-machine\].*\[aggregator-machine:vars\]/[aggregator-machine]\n\n[aggregator-machine:vars]/' $ANSIBLE_HOSTS_FILE
# Writes the IP address in Ansible hosts file
sed -i "2s/.*/$DEPLOYED_HOST_IP/" $ANSIBLE_HOSTS_FILE

# Writes the path of the private key file in Ansible hosts file
sed -i "s#.*$PRIVATE_KEY_FILE_PATH_PATTERN=.*#$PRIVATE_KEY_FILE_PATH_PATTERN=$PRIVATE_KEY_FILE_PATH#g" $ANSIBLE_HOSTS_FILE

# Writes the remote user name in Ansible hosts file
sed -i "s/.*$REMOTE_USER_PATTERN = .*/$REMOTE_USER_PATTERN = $REMOTE_USER/" $ANSIBLE_CFG_FILE

PATH_VM="/home/$REMOTE_USER"
echo "Path to add configuration files: $PATH_VM"

# Run ansible
readonly DEPLOY_YML_FILE="deploy-aggregator.yml"

(cd $ANSIBLE_FILES_PATH && sudo ansible-playbook -vvv $DEPLOY_YML_FILE)
