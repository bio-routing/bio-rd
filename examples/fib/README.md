# Example: FIB
This example aims to demonstrate how to move paths from the locRIB to the Linux Kernel routing table

## How to build


    go build .



## Run
1. Prepare your environment with `sudo ./prepare.sh`
3. Run the demo with `sudo ./fib`
4. Run `ip r ls` to verify if the new routeis added to the kernel routing table
5. Tear-Down environment with `sudo ./teardown.sh`