#!/bin/bash

# add new dummy interface
echo "[1/3] Adding new dummy interface: dummy1"
ip link add dummy1 type dummy

sleep 2

# add address
echo "[2/3] Assigning new IP to dummy1 interface: 169.254.1.1/24"
ip addr add dev dummy1 169.254.1.1/24

sleep 2

# Up interface
echo "[3/3] Set dummy1 interface link to up"
ip link set up dev dummy1

sleep 2

echo "Output current link informations"
ip --brief addr ls
ip r ls