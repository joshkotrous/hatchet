#!/bin/sh

set -eux

# run:
# ./manage-etc-hosts.sh add 10.20.1.2 test.com
# ./manage-etc-hosts.sh remove 10.20.1.2 test.com

# PATH TO YOUR HOSTS FILE
ETC_HOSTS=/etc/hosts

# Validate IP address format
validate_ip() {
    # Check if IP follows the IPv4 format
    if ! echo "$1" | grep -Eq '^([0-9]{1,3}\.){3}[0-9]{1,3}$'; then
        echo "Error: Invalid IP address format: $1" >&2
        return 1
    fi
    
    # Verify each octet is in range 0-255
    local IFS='.'
    set -- $1
    for octet; do
        if [ "$octet" -lt 0 ] || [ "$octet" -gt 255 ]; then
            echo "Error: IP octet out of range (0-255): $octet" >&2
            return 1
        fi
    done
    
    return 0
}

# Validate hostname format
validate_hostname() {
    # Check hostname follows valid format
    if ! echo "$1" | grep -Eq '^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$'; then
        echo "Error: Invalid hostname format: $1" >&2
        return 1
    fi
    
    # Check length constraint
    if [ ${#1} -gt 255 ]; then
        echo "Error: Hostname too long (max 255 chars): $1" >&2
        return 1
    fi
    
    return 0
}

function remove() {
    # IP to add/remove.
    IP=$1
    # Hostname to add/remove.
    HOSTNAME=$2
    
    # Validate inputs
    if ! validate_ip "$IP"; then
        return 1
    fi
    if ! validate_hostname "$HOSTNAME"; then
        return 1
    fi
    
    HOSTS_LINE="$IP[[:space:]]$HOSTNAME"
    if [ -n "$(grep "$HOSTS_LINE" "$ETC_HOSTS")" ]
    then
        echo "$HOSTS_LINE Found in your $ETC_HOSTS, Removing now...";
        sudo sed -i".bak" "/^$IP[[:space:]]$HOSTNAME$/d" "$ETC_HOSTS"
    else
        echo "$HOSTS_LINE was not found in your $ETC_HOSTS";
    fi
}

function add() {
    IP=$1
    HOSTNAME=$2
    
    # Validate inputs
    if ! validate_ip "$IP"; then
        return 1
    fi
    if ! validate_hostname "$HOSTNAME"; then
        return 1
    fi
    
    HOSTS_LINE="$IP[[:space:]]$HOSTNAME"
    line_content=$( printf "%s %s\n" "$IP" "$HOSTNAME" )
    
    if [ -n "$(grep "$HOSTS_LINE" "$ETC_HOSTS")" ]
        then
            echo "$line_content already exists : $(grep "$HOSTNAME" "$ETC_HOSTS")"
        else
            echo "Adding $line_content to your $ETC_HOSTS";
            # Replace vulnerable shell command with safer alternative
            printf "%s %s\n" "$IP" "$HOSTNAME" | sudo tee -a "$ETC_HOSTS" > /dev/null

            if [ -n "$(grep "$HOSTNAME" "$ETC_HOSTS")" ]
                then
                    echo "$line_content was added successfully";
                else
                    echo "Failed to Add $line_content, Try again!";
            fi
    fi
}

$@