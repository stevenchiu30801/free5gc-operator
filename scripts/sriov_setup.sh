#!/bin/bash

usage() {
    echo "***Experimental script for SR-IOV setup***"
    echo "Configure virtual functions on SR-IOV device"
    echo ""
    echo "Usage: ./sriov_setup.sh SRIOV-INTF NUM_VF"
    echo "Arguments:"
    echo "    SRIOV_INTF        Interface to be enabled SR-IOV"
    echo "    NUM_VF            Number of VFs to be created"
}

if [[ $# -eq 2 ]]; then
    if [[ $1 == "-h" ]] || [[ $1 == "--help" ]]; then
        usage
        exit 0
    fi

    # Check if input NUM_VF is an integer
    if ! [[ $2 =~ ^[0-9]+$ ]]; then
        usage
        exit 1
    fi
else
    usage
    exit 1
fi

# Check if number of existing VFs meets requested number
EXISTING_VF=$( ip link show $1 | grep -c vf )
if [[ ${EXISTING_VF} -eq $2 ]]; then
    echo "Number of existing VFs is exactly $2"
    exit 0
elif [[ ${EXISTING_VF} -ne 0 ]]; then
    # Reset VFs
    echo 0 | sudo tee /sys/class/net/$1/device/sriov_numvfs >/dev/null
fi

# Check if IOMMU support for Linux kernel is enable
if ! dmesg | grep 'DMAR: IOMMU enabled' >/dev/null; then
    echo -e "Please enable IOMMU support for Linux kernel\n\
        1. Append 'intel_iommu=on' to the 'GRUM_CMDLINE_LINUX' entry in /etc/default/grub\n\
        2. Update grub with command 'update-grub'\n\
        3. Reboot for IOMMU change to take effect"

    while true; do
        read -p "Continue SR-IOV setup [Y/n] " continue_setup
        if [[ $continue_setup == "n" ]] || [[ $continue_setup == "N" ]]; then
            exit 1
        elif [[ $continue_setup == "" ]] || [[ $continue_setup == "y" ]] || [[ $continue_setup == "Y" ]]; then
            break
        fi
    done
fi

# Get the driver used by the device
DRIVER=$( ethtool -i $1 | awk '$1=="driver:" {print $2}' )

if [[ -z "${DRIVER}" ]]; then
    echo "No driver information for device '$1'"
    exit 1
fi

# Check if device's kernel module is loaded
if ! lsmod | awk '{print $1}' | grep ${DRIVER} >/dev/null; then
    # Check if driver is included in Linux kernel
    if ! find /lib/modules/$(uname -r) -type f -name '*.ko' | grep /${DRIVER}.ko >/dev/null; then
        echo "PF driver '${DRIVER}' is not provided in kernel"
        echo "Please install the driver '${DRIVER}'"
        exit 1
    elif ! find /lib/modules/$(uname -r) -type f -name '*.ko' | grep /${DRIVER}vf.ko >/dev/null; then
        echo "VF driver '${DRIVER}vf' is not provided in kernel"
        echo "Please install the driver '${DRIVER}vf'"
        exit 1
    fi

    # Load device's kernel module
    modprobe ${DRIVER}
fi

# Check if the requested number of VFs exceeds the maximum number of supported VFs
TOTAL_VF=$( cat /sys/class/net/$1/device/sriov_totalvfs )
if [[ $2 -gt ${TOTAL_VF} ]]; then
    echo "Number of VFs should not exceed the maximum number of VFs supported by device '$1'"
    echo "Maximum number: ${TOTAL_VF}"
    exit 1
fi

# Create VFs
echo $2 | sudo tee /sys/class/net/$1/device/sriov_numvfs >/dev/null

# # Assign MAC address to VFs if VF driver is i40e
# if [[ ${DRIVER} -eq "i40e" ]]; then
#     for i in $(seq 0 $(($2 - 1)));
#     do
#         sudo ip link set $1 vf $i mac aa:bb:cc:dd:ee:$(printf "%02d" $i)
#     done
# fi
