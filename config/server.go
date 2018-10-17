package config

import (
	"fmt"
	"net"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/taktv6/tflow2/convert"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

//Global struct for global configuration
type Global struct {
	LocalASN         uint32
	RouterID         uint32
	Port             uint16
	LocalAddressList []net.IP
	Listen           bool
	LoopbackIface    string
}

//BGPPORT default bgp port for default which is tcp 179
const BGPPORT = uint16(179)

//SetDefaultGlobalConfigValues maps default configuration values to Global struct
func (g *Global) SetDefaultGlobalConfigValues() error {
	if g.LocalAddressList == nil {
		g.LocalAddressList = make([]net.IP, 0)
		g.LocalAddressList = append(g.LocalAddressList, net.ParseIP("0.0.0.0"))
		g.LocalAddressList = append(g.LocalAddressList, net.ParseIP("::"))
	}

	//Set router Id if no override is set via config file
	if g.RouterID == 0 {
		rtrid, err := generateRouterID(g.LoopbackIface)
		if err != nil {
			return fmt.Errorf("Unable to determine router ID: %v", err)
		}
		g.RouterID = rtrid
	}

	//Set default port if no override is set via config file
	if g.Port == 0 {
		g.Port = BGPPORT
	}

	return nil
}

//ReadGlobalConfig Search and read global config files in path ./, /etc/bio-rd/ or $HOME/.bio-rd/
func (g *Global) ReadGlobalConfig() error {

	viper.SetConfigName("global")
	viper.AddConfigPath("/etc/bio-rd/")
	home, err := homedir.Dir()
	if err != nil {
		logrus.Infof("Can't find home directory")
	} else {
		viper.AddConfigPath(home + "/.bio-rd")
	}
	viper.AddConfigPath(".")

	viper.SetDefault("LoopbackIface", "lo")

	err = viper.ReadInConfig()

	if err != nil {

		return fmt.Errorf("unable to read the global config file: %#v", err)
	}

	err = viper.Unmarshal(g)
	if err != nil {
		return fmt.Errorf("unable to decode into config struct, %#v", err)
	}
	return nil
}

func generateRouterID(loopbackIfaceName string) (uint32, error) {
	addr, err := getLoopbackIP(loopbackIfaceName)
	if err == nil {
		return convert.Uint32b([]byte(addr)[12:16]), nil
	}

	return 0, fmt.Errorf("Unable to determine router id")
}

func getHighestIP() (net.IP, error) {
	ifs, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("Unable to ")
	}

	return _getHighestIP(ifs)
}

func _getHighestIP(ifs []net.Interface) (net.IP, error) {
	candidates := make([]net.IP, 0)
	for _, iface := range ifs {
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, fmt.Errorf("Unable to get interface addrs for %s: %v", iface.Name, err)
		}

		for _, addr := range addrs {
			a := net.ParseIP(addr.String())
			if addr.String() != "127.0.0.1" && a.To4() != nil {
				candidates = append(candidates, a)
			}
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("No IPv4 address found on any interface")
	}

	max := candidates[0]
	for _, c := range candidates[1:] {
		if addrIsGreater(c, max) {
			max = c
		}
	}

	return nil, fmt.Errorf("No non localhost IPv4 address found on interface lo")
}

func getLoopbackIP(loopbackIfaceName string) (net.IP, error) {
	iface, err := net.InterfaceByName(loopbackIfaceName)
	if err != nil {
		return nil, fmt.Errorf("Unable to get interface lo: %v", err)
	}

	return _getLoopbackIP(iface)
}

func _getLoopbackIP(iface *net.Interface) (net.IP, error) {
	addrs, err := iface.Addrs()
	if err != nil {
		return nil, fmt.Errorf("Unable to get interface addresses: %v", err)
	}

	candidates := make([]net.IP, 0)
	for _, addr := range addrs {
		a := net.ParseIP(strings.Split(addr.String(), "/")[0])
		if a.String() != "127.0.0.1" && a.To4() != nil {
			candidates = append(candidates, a)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("No non localhost IPv4 address found on interface lo")
	}

	max := candidates[0]
	for _, c := range candidates {
		if addrIsGreater(c, max) {
			max = c
		}
	}

	return max, nil
}

func addrIsGreater(a net.IP, b net.IP) bool {
	/*
	 * FIXME: Implement proper comparison
	 */
	if a.String() > b.String() {
		return true
	}
	return false
}
