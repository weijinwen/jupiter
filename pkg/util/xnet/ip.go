// Copyright 2020 Douyu
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package xnet

import (
	"errors"
	"fmt"
	"net"
	"time"
)

// GetLocalIP ...
func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", errors.New("unable to determine local ip")
}

// GetLocalIP ...
func GetLocalMainIP() (string, int, error) {
	// UDP Connect, no handshake
	conn, err := net.Dial("udp", "8.8.8.8:8")
	if err != nil {
		return "", 0, err
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String(), localAddr.Port, nil
}

// GetMacAddrs ...
func GetMacAddrs() (macAddrs []string) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		fmt.Printf("fail to get net interfaces: %v", err)
		return macAddrs
	}

	for _, netInterface := range netInterfaces {
		macAddr := netInterface.HardwareAddr.String()
		if len(macAddr) == 0 {
			continue
		}

		macAddrs = append(macAddrs, macAddr)
	}
	return macAddrs
}

// GetIPs ...
func GetIPs() (ips []string) {
	interfaceAddr, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Printf("fail to get net interface addrs: %v", err)
		return ips
	}

	for _, address := range interfaceAddr {
		ipNet, isValidIPNet := address.(*net.IPNet)
		if isValidIPNet && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ips = append(ips, ipNet.IP.String())
			}
		}
	}
	return ips
}

// IsPortAvailable checks if a port is available on the given host.
func IsPortAvailable(host string, port int) bool {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return true
	}
	conn.Close()
	return false
}

// IsTCPPortOpen checks if a TCP port is open on the given host.
func IsTCPPortOpen(host string, port int) bool {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// IsUDPPortOpen checks if a UDP port is open on the given host.
func IsUDPPortOpen(host string, port int) bool {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("udp", addr, 3*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// GetAvailablePort returns an available port.
func GetAvailablePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// GetAvailablePorts returns n available ports.
func GetAvailablePorts(n int) ([]int, error) {
	ports := make([]int, 0, n)
	for i := 0; i < n; i++ {
		port, err := GetAvailablePort()
		if err != nil {
			return nil, err
		}
		ports = append(ports, port)
	}
	return ports, nil
}

// IsIPv4 checks if the given string is a valid IPv4 address.
func IsIPv4(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil && parsedIP.To4() != nil
}

// IsIPv6 checks if the given string is a valid IPv6 address.
func IsIPv6(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil && parsedIP.To4() == nil
}

// IsPrivateIP checks if the given IP is a private IP address.
func IsPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	return parsedIP.IsPrivate()
}

// IsLoopbackIP checks if the given IP is a loopback address.
func IsLoopbackIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	return parsedIP.IsLoopback()
}

// IsLinkLocalIP checks if the given IP is a link-local address.
func IsLinkLocalIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	return parsedIP.IsLinkLocalUnicast() || parsedIP.IsLinkLocalMulticast()
}

// NormalizeIPv6 normalizes an IPv6 address.
func NormalizeIPv6(ip string) string {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return ip
	}
	return parsedIP.String()
}

// IsCIDR checks if the given string is a valid CIDR notation.
func IsCIDR(cidr string) bool {
	_, _, err := net.ParseCIDR(cidr)
	return err == nil
}

// IsMACAddress checks if the given string is a valid MAC address.
func IsMACAddress(mac string) bool {
	_, err := net.ParseMAC(mac)
	return err == nil
}

// GetInterfaceByName returns the interface by name.
func GetInterfaceByName(name string) (*net.Interface, error) {
	return net.InterfaceByName(name)
}

// GetInterfaces returns all network interfaces.
func GetInterfaces() ([]net.Interface, error) {
	return net.Interfaces()
}

// GetHostByName returns the IP addresses for the given hostname.
func GetHostByName(hostname string) ([]string, error) {
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return nil, err
	}
	result := make([]string, len(ips))
	for i, ip := range ips {
		result[i] = ip.String()
	}
	return result, nil
}

// GetAddrInfo returns the address information for the given hostname.
func GetAddrInfo(hostname string) (*net.TCPAddr, error) {
	return net.ResolveTCPAddr("tcp", hostname+":0")
}
