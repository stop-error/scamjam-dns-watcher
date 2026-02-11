package main

import (
	"net"
	"time"
	"strings"

	"github.com/miekg/dns"
)


func testDNS() bool {
	testDomains := [2]string{"connectivity-check.ubuntu.com.", "dns.msftncsi.com."}
	var testOK []string
	
	for i := 0; i < len(testDomains); i++ {
		localAddr, err := getLocalAddr()
		if err == nil {
			var testResult bool

			logger.Info().Msg("Updated localAddr to " + localAddr)

			switch {
			case IsIPv4(localAddr):
				logger.Info().Msg("localAddr is IPv4, will test with IPv4 DNS server address")
				 testResult = runTestQuery(testDomains[i], false)
			case IsIPv6(localAddr):
				logger.Info().Msg("localAddr is IPv6, will test with IPv6 DNS server address")
				testResult = runTestQuery(testDomains[i], true)
			}
			
			if testResult == true {
				testOK = append(testOK, testDomains[i])
			}

		} else {
			logger.Warn().Msg("Skipped testDNS due to error getting active IP address")
		}
	}
	if len(testOK) > 0 {
		logger.Info().Msg("One or more testdomains resolved successfully, TestDNS true (successfull):")
		return true
	} else {
		logger.Error().Msg("Unable to resolve any test domain successfully, TestDNS false (unsuccessfull)")
		return false
	}
}


func IsIPv4(address string) bool {
    return strings.Count(address, ":") < 2
}

func IsIPv6(address string) bool {
    return strings.Count(address, ":") >= 2
}

func getLocalAddr() (string, error) {

    conn, err := net.Dial("udp", "9.9.9.9:80")
    if err != nil {
        logger.Error().Msg("Unable to get active host IP address!")
		return "", err
    }
    defer conn.Close()

	localAddrWithPort := conn.LocalAddr()
    localAddr, _, err  := net.SplitHostPort(localAddrWithPort.String())
	if err != nil {
		logger.Error().Msg("Unable to split active host IP from dialed connection port")
		return "", err
	}

    return localAddr, nil
}
	

func runTestQuery(hostname string, isIPv6 bool) bool {
	ipV4Server := "127.0.0.3:53"
	ipV6Server := "[::1]:53"
	var ipAddressToUse string

	m := new(dns.Msg)
	m.SetQuestion(hostname, dns.TypeA)

	c := new(dns.Client)
	c.Dialer = &net.Dialer{
		Timeout: 2 * time.Second,
	}


	if (isIPv6 == true) {
		ipAddressToUse = ipV6Server
	} else {
		ipAddressToUse = ipV4Server
	}


	logger.Info().Msg("Testing address " + ipAddressToUse)
		in, _, err := c.Exchange(m, ipAddressToUse) //This should not be hard-coded
		switch {
		case err != nil:
			logger.Error().Msg("Error resolving testdomain:" + m.String() + err.Error())
			return false
		case in == nil: 
			logger.Error().Msg("Error resolving testdomain" + m.String() + " (value of dns.Msg is nil)")
			return false
		case len(in.Answer) <= 0: 
			logger.Error().Msg("Error resolving testdomain:" + m.String() + " (length of Answer is zero or less)")
			return false
		default:
			logger.Info().Msg("Successfully resolved testdomain: " + m.String())
			return true
		}
}
	
	
	

