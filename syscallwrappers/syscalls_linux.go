package syscallwrappers

/*
#cgo CFLAGS: -I/usr/include
#cgo LDFLAGS: -L/usr/lib
#include <sys/types.h>
#include <sys/socket.h>
#include <linux/if_packet.h>
#include <linux/filter.h>
#include <net/ethernet.h>
#include <string.h>
#include <arpa/inet.h>
int setsockopt(int sockfd, int level, int optname, const void *optval, socklen_t optlen);
uint8_t ALL_L1_ISS[6] = {0x01, 0x80, 0xC2, 0x00, 0x00, 0x14};
uint8_t ALL_L2_ISS[6] = {0x01, 0x80, 0xC2, 0x00, 0x00, 0x15};
uint8_t ALL_P2P_ISS[6] = {0x09, 0x00, 0x2b, 0x00, 0x00, 0x5b};
uint8_t ALL_ISS[6] = {0x09, 0x00, 0x2B, 0x00, 0x00, 0x05};
uint8_t ALL_ESS[6] = {0x09, 0x00, 0x2B, 0x00, 0x00, 0x04};
static struct sock_filter isisfilter[] = {
	//{ 0x28, 0, 0, 0x0000000c }, { 0x25, 5, 0, 0x000005dc },
	{ 0x28, 0, 0, 0x0000000e - 14 }, { 0x15, 0, 3, 0x0000fefe },
	{ 0x30, 0, 0, 0x00000011 - 14 }, { 0x15, 0, 1, 0x00000083 },
	{ 0x6, 0, 0, 0x00040000 }, { 0x6, 0, 0, 0x00000000 },
};
static struct sock_fprog bpf = {
	.len = 6,
	.filter = isisfilter,
};
int reg_bpf(int fd) {
	return setsockopt(fd, SOL_SOCKET, SO_ATTACH_FILTER, &bpf, sizeof(bpf));
}
int bind_to_interface(int fd, int ifindex) {
	struct sockaddr_ll s_addr;
	memset(&s_addr, 0, sizeof(struct sockaddr_ll));
	s_addr.sll_family = AF_PACKET;
	s_addr.sll_protocol = htons(ETH_P_ALL);
	s_addr.sll_ifindex = ifindex;
	return bind(fd, (struct sockaddr *)(&s_addr), sizeof(struct sockaddr_ll));
}
int isis_multicast_join(int fd, int registerto, int ifindex)
{
	struct packet_mreq mreq;
	memset(&mreq, 0, sizeof(mreq));
	mreq.mr_ifindex = ifindex;
	if (registerto) {
		mreq.mr_type = PACKET_MR_MULTICAST;
		mreq.mr_alen = ETH_ALEN;
		if (registerto == 1)
			memcpy(&mreq.mr_address, ALL_L1_ISS, ETH_ALEN);
		else if (registerto == 2)
			memcpy(&mreq.mr_address, ALL_L2_ISS, ETH_ALEN);
		else if (registerto == 3)
			memcpy(&mreq.mr_address, ALL_ISS, ETH_ALEN);
		else if (registerto == 4)
			memcpy(&mreq.mr_address, ALL_P2P_ISS, ETH_ALEN);
		else
			memcpy(&mreq.mr_address, ALL_ESS, ETH_ALEN);
	} else {
		mreq.mr_type = PACKET_MR_ALLMULTI;
	}
	return setsockopt(fd, SOL_PACKET, PACKET_ADD_MEMBERSHIP, &mreq, sizeof(struct packet_mreq));
}
*/
import "C"

import (
	"unsafe"
)

func SetBPFFilter(sockfd int) int {
	return int(C.reg_bpf(C.int(sockfd)))
}

func SetSockOpt(sockfd int, level int, optName int, optVal uintptr, optLen int) int {
	ptr := unsafe.Pointer(optVal)
	return int(C.setsockopt(C.int(sockfd), C.int(level), C.int(optName), ptr, C.uint(optLen)))
}

func JoinISISMcast(sockfd int, ifIndex int) int {
	return int(C.isis_multicast_join(C.int(sockfd), 4, C.int(ifIndex)))
}

func BindToInterface(sockfd int, ifIndex int) int {
	return int(C.bind_to_interface(C.int(sockfd), C.int(ifIndex)))
}
