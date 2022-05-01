#!/bin/bash

# This script is used to turn on IP forwarding for the tap0 interface.
network="10.88.45.0/24"

# Function to enable IP forwarding.
enable_ip_forwarding() {
    echo 1 >/proc/sys/net/ipv4/ip_forward
}

# Function to disable IP forwarding.
disable_ip_forwarding() {
    echo 0 >/proc/sys/net/ipv4/ip_forward
}

# Function to get physical interface name.
get_physical_interface_name() {
    physical_interface_name=$(ip route | grep default | awk '{print $5}')
}

# Function to add ip tables rule to forward to tap0.
add_iptables_rule() {
    # iptables -I INPUT --source $network -j ACCEPT
    # iptables -t nat -A POSTROUTING -o ${physical_interface_name} -j MASQUERADE
    iptables -I FORWARD -i ${physical_interface_name} -o tap0 -j ACCEPT
    iptables -I FORWARD -i tap0 -o ${physical_interface_name} -j ACCEPT
}

# Function to remove ip tables rule to forward to tap0.
remove_iptables_rule() {
    # iptables -D INPUT --source $network -j ACCEPT
    # iptables -t nat -D POSTROUTING -o ${physical_interface_name} -j MASQUERADE
    iptables -D FORWARD -i ${physical_interface_name} -o tap0 -j ACCEPT
    iptables -D FORWARD -i tap0 -o ${physical_interface_name} -j ACCEPT
}

enable() {
    enable_ip_forwarding
    get_physical_interface_name
    add_iptables_rule
}

disable() {
    disable_ip_forwarding
    get_physical_interface_name
    remove_iptables_rule
}

# Get command
command=$1

if [ "${command}" == "enable" ]; then
    enable
elif [ "${command}" == "disable" ]; then
    disable
else
    echo "Usage: $0 enable|disable"
fi
