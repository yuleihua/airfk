package common

import (
	"net"
)

// GetListener
func GetListener(address string) (net.Listener, error) {
	ld, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, err
	}

	ln, e := net.ListenTCP("tcp", ld)
	if e != nil {
		return nil, e
	}
	return ln, nil
}

// GetLocalAddress
func GetLocalAddress() ([]string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return []string{"localhost"}, err
	}

	var ipList []string
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ipList = append(ipList, ipnet.IP.String())
			}
		}
	}
	return ipList, nil
}
