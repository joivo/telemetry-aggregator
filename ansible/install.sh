#!/bin/sh

MY_PATH="`dirname \"$0\"`"              
MY_PATH="`( cd \"${MY_PATH}\" && pwd )`"

if [[ -z "$MY_PATH" ]] ; then
  # For some reason, the path is not accessible
  # to the script (e.g. permissions re-evaled after suid)
  exit 1
fi

echo "Current directory path: "${MY_PATH}

readonly HOSTS_CONF_FILE=${MY_PATH}"/hosts.conf"

readonly DEPLOYED_HOST_IP_PATTERN="deployed_host_ip"
readonly DEPLOYED_HOST_IP=$(grep ${DEPLOYED_HOST_IP_PATTERN} ${HOSTS_CONF_FILE} | awk -F "=" '{print $2}')

# Remote user of the host
readonly REMOTE_USER_PATTERN="remote_user"
readonly REMOTE_USER=$(grep ${REMOTE_USER_PATTERN} ${HOSTS_CONF_FILE} | awk -F "=" '{print $2}')

# Ansible ssh private key file path
readonly PRIVATE_KEY_FILE_PATH_PATTERN="ansible_ssh_private_key_file"
readonly PRIVATE_KEY_FILE_PATH=$(grep ${PRIVATE_KEY_FILE_PATH_PATTERN} ${HOSTS_CONF_FILE} | awk -F "=" '{print $2}')

# Playbook path
readonly ANSIBLE_FILES_PATH_PATTERN="ansible_files_path"
readonly ANSIBLE_FILES_PATH=$(grep ${ANSIBLE_FILES_PATH_PATTERN} ${HOSTS_CONF_FILE} | awk -F "=" '{print $2}')

readonly ANSIBLE_PLAYBOOK_FILE=${ANSIBLE_FILES_PATH}/"ansible-playbook"
readonly ANSIBLE_HOSTS_FILE=${ANSIBLE_FILES_PATH}/"hosts"
readonly ANSIBLE_CFG_FILE=${ANSIBLE_FILES_PATH}/"ansible.cfg"

echo "Ansible SSH private key file path: ${PRIVATE_KEY_FILE_PATH}"
echo "Remote user of host: ${REMOTE_USER}"
echo "Host IP: ${DEPLOYED_HOST_IP}"

#Testing ssh-port
readonly SSH_PORT=22
SSH_STATUS=$(nmap ${DEPLOYED_HOST_IP} -PN -p ${SSH_PORT} | egrep 'open|closed|filtered')
echo "SSH port status: " ${SSH_STATUS}

#Testing Connection
echo "ssh -i ${PRIVATE_KEY_FILE_PATH} -q ${REMOTE_USER}@${DEPLOYED_HOST_IP} exit"
ssh -i ${PRIVATE_KEY_FILE_PATH} -q ${REMOTE_USER}@${DEPLOYED_HOST_IP} exit
EXIT_CODE=$?
echo "Status of the connectivity: " ${EXIT_CODE}

if [[ ${EXIT_CODE} -eq 0 ]]; then
   echo "Successful connection!"
else
   echo "The SSH connection failed!"
   exit 1
fi

# Delete everything between[aggregator-machine] and [aggregator-machine:vars]
sed -i 's/\[aggregator-machine\].*\[aggregator-machine:vars\]/[aggregator-machine]\n\n[aggregator-machine:vars]/' ${ANSIBLE_HOSTS_FILE}
# Writes the IP address in Ansible hosts file
sed -i "2s/.*/${DEPLOYED_HOST_IP}/" ${ANSIBLE_HOSTS_FILE}

# Writes the path of the private key file in Ansible hosts file
sed -i "s#.*${PRIVATE_KEY_FILE_PATH_PATTERN}=.*#${PRIVATE_KEY_FILE_PATH_PATTERN}=${PRIVATE_KEY_FILE_PATH}#g" ${ANSIBLE_HOSTS_FILE}

# Writes the remote user name in Ansible hosts file
sed -i "s/.*${REMOTE_USER_PATTERN} = .*/${REMOTE_USER_PATTERN} = ${REMOTE_USER}/" ${ANSIBLE_CFG_FILE}

PATH_VM="/home/$REMOTE_USER"
echo "Path to add configuration files: $PATH_VM"

# Run ansible
readonly DEPLOY_YML_FILE="deploy-aggregator.yml"

(cd ${ANSIBLE_FILES_PATH} && sudo ansible-playbook -vvv ${DEPLOY_YML_FILE})

exit 0
