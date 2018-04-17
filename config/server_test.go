package config

import (
	"net"
	"reflect"
	"testing"
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
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := addrIsGreater(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("addrIsGreater() = %v, want %v", got, tt.want)
			}
		})
	}
}
