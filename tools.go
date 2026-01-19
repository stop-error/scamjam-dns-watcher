package main

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/pkg/nameserver"
)

func TestDNS() error {
	testdomain := "internetbeacon.msedge.net."
	m := new(dns.Msg)
	m.SetQuestion(testdomain, dns.TypeA)

	c := new(dns.Client)
	c.Dialer = &net.Dialer{
		Timeout: 5 * time.Second,
	}

	in, _, err := c.Exchange(m, "127.0.0.3:53") //This should not be hard-coded
	if err != nil {
		return err
	}
	if len(in.Answer) <= 0 {
		return err
	}
	return nil
}

func GetHostDnsServersIPv4() ([]netip.Addr, error) {

	var ipv4HostDnsConfig []netip.Addr

	hostDnsConfig, err := nameserver.GetDNSServers()
	if err != nil {
		return nil, errors.New("Error retriving host DNS config!")
		}

	for i := 0; i < len(hostDnsConfig); i++ {

		interfaceIndexAsString := strconv.Itoa(i)
		fmt.Fprintln(os.Stdout, "on interface " + interfaceIndexAsString)

		if  hostDnsConfig[i].Is6() == true {
			fmt.Fprintln(os.Stdout, "Skipping ipv6 address with interface index" + interfaceIndexAsString)
			
		}

		if  hostDnsConfig[i].Is4() == true {
			ipv4HostDnsConfig = append(ipv4HostDnsConfig, hostDnsConfig[i])
			
		}
	}
	return ipv4HostDnsConfig, nil
}

func TestHostDnsServersScamJam (dnsConfig []netip.Addr) (bool) { //probably want to return an error

	proxyAddr, _ := netip.ParseAddr("127.0.0.3")
	var interfacesSetToScamJamDNS []netip.Addr

	for i := 0; i < len(dnsConfig); i++ {
		if dnsConfig[i] == proxyAddr {
			fmt.Fprintln(os.Stdout, "interface index " + strconv.Itoa(i) + " is set to scamjam-dns-server")
			interfacesSetToScamJamDNS = append(interfacesSetToScamJamDNS, dnsConfig[i])
		}
	}

	return reflect.DeepEqual(dnsConfig, interfacesSetToScamJamDNS)

}

