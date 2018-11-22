package config

import (
	"net"
	"reflect"
	"testing"

	bnet "github.com/bio-routing/bio-rd/net"
)

func TestGetLoopbackIP(t *testing.T) {
	type args struct {
		iface *net.Interface
	}
	tests := []struct {
		name    string
		args    args
		want    net.IP
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := _getLoopbackIP(tt.args.iface)
			if (err != nil) != tt.wantErr {
				t.Errorf("_getLoopbackIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("_getLoopbackIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetHighestIP(t *testing.T) {
	type args struct {
		ifs []net.Interface
	}
	tests := []struct {
		name    string
		args    args
		want    net.IP
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := _getHighestIP(tt.args.ifs)
			if (err != nil) != tt.wantErr {
				t.Errorf("_getHighestIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("_getHighestIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddrIsGreater(t *testing.T) {
	type args struct {
		a net.IP
		b net.IP
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "192.168.0.0 lower than 192.168.0.1",
			args: args{
				a: net.IPv4(192, 168, 0, 0),
				b: net.IPv4(192, 168, 0, 1),
			},
			want: false,
		},
		{
			name: "192.168.0.1 higher than 192.168.0.0",
			args: args{
				a: net.IPv4(192, 168, 0, 1),
				b: net.IPv4(192, 168, 0, 0),
			},
			want: true,
		},
		{
			name: "10.0.0.0 lower than 172.12.0.0",
			args: args{
				a: net.IPv4(10, 0, 0, 0),
				b: net.IPv4(172, 12, 0, 0),
			},
			want: false,
		},
		{
			name: "172.12.0.0 higher than 10.0.0.0",
			args: args{
				a: net.IPv4(172, 12, 0, 0),
				b: net.IPv4(10, 0, 0, 0),
			},
			want: true,
		},
		{
			name: "::7d2:0:0:0:1c8 higher than ::7d1:0:0:0:1c8",
			args: args{
				a: bnet.IPv6(2002, 456).Bytes(),
				b: bnet.IPv6(2001, 456).Bytes(),
			},
			want: true,
		},
		{
			name: "::7d1:0:0:0:1c8 higher than ::7d2:0:0:0:1c8",
			args: args{
				a: bnet.IPv6(2001, 456).Bytes(),
				b: bnet.IPv6(2002, 456).Bytes(),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := addrIsGreater(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("addrIsGreater() = %v, want %v", got, tt.want)
			}
		})
	}
}
