package main

import (
	"net"
	// "net/netip"
	"os"
	// "strconv"
	"time"
	"fmt"
	"io"
	"path/filepath"

	"github.com/jedisct1/dlog"
	"github.com/miekg/dns"
	// "github.com/qdm12/dns/v2/pkg/nameserver"
	"github.com/nmrshll/go-cp"
	"github.com/rs/zerolog"

	watchdog "github.com/Control-D-Inc/ctrld/cmd/cli"

)

func testDNS() bool {
	testDomains := [2]string{"connectivity-check.ubuntu.com.", "dns.msftncsi.com."}
	ipv4Address := "127.0.0.3:53"
	ipv6Address := "[::1]:53"
	var hostDnsServers []string
	hostDnsServers = append(hostDnsServers, ipv4Address, ipv6Address)
	var testOK []string
	
	for i := 0; i < len(testDomains); i++ {

		m := new(dns.Msg)
		m.SetQuestion(testDomains[i], dns.TypeA)

		c := new(dns.Client)
		c.Dialer = &net.Dialer{
			Timeout: 5 * time.Second,
		}

		for i := 0; i < len(hostDnsServers); i++ {
			dlog.Notice("Testing address " + hostDnsServers[i])
			in, _, err := c.Exchange(m, hostDnsServers[i]) //This should not be hard-coded
			switch {
				case err != nil || len(in.Answer) <= 0:
					dlog.Error("Error resolving testdomain:" + testDomains[i] + err.Error())
					continue
				default:
					dlog.Info("Successfully resolved testdomain testdomain: " + testDomains[i])
					testOK = append(testOK, testDomains[i])
			}
		
		}
	}


	if len(testOK) > 0 {
		dlog.Info("One or more testdomains resolved successfully, TestDNS true (successfull):")
		return true
	} else {
		dlog.Error("Unable to resolve any test domain successfully, TestDNS false (unsuccessfull)")
		return false
	}

}

// func getHostDnsServersIPv4() ([]netip.Addr, []netip.Addr, error) {

// 	var ipv4HostDnsConfig []netip.Addr
// 	var ipv6HostDnsConfig []netip.Addr

// 	hostDnsConfig, err := nameserver.GetDNSServers()
// 	if err != nil {
// 		dlog.Error("Error retriving host DNS config!")
// 		return nil, nil, err
// 		}

// 	for i := 0; i < len(hostDnsConfig); i++ {

// 		interfaceIndexAsString := strconv.Itoa(i)
// 		dlog.Notice("on interface " + interfaceIndexAsString + " in interface array")

// 		if  hostDnsConfig[i].Is6() == true {
// 			dlog.Notice("Found ipv6 interface:" + interfaceIndexAsString)
// 			ipv6HostDnsConfig = append(ipv6HostDnsConfig, hostDnsConfig[i])
			
// 		}

// 		if  hostDnsConfig[i].Is4() == true {
// 			dlog.Notice("Found ipv4 interface:" + interfaceIndexAsString)
// 			ipv4HostDnsConfig = append(ipv4HostDnsConfig, hostDnsConfig[i])
			
// 		}
// 	}
// 	return ipv4HostDnsConfig, ipv6HostDnsConfig, nil
// }


func cleanupLogs(CurrentLog string, OldLog string) (error) {

	fmt.Fprintln(os.Stdout, "Running log cleanup")

	if _, err := os.Stat(OldLog); err == nil {
		fmt.Fprintln(os.Stdout, "Log cleanup: Deleting current .old file")
		err := os.Remove(OldLog)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error deleting .old file!" + err.Error())
			return err
		}
	}

	if _, err := os.Stat(CurrentLog); err == nil {
		fmt.Fprintln(os.Stdout, "Log cleanup: .log file becomes .old file")
		errCopy := cp.CopyFile(CurrentLog, OldLog)
		if errCopy != nil {
			fmt.Fprintln(os.Stderr, "Error copying .log file to .old file!" + errCopy.Error())
			return err
		}
		fmt.Fprintln(os.Stdout, "Log cleanup: deleting .log file")
		errRemove := os.Remove(CurrentLog)
		if errRemove != nil {
			fmt.Fprintln(os.Stderr, "Error deleting .log file! Logs will not rotate correctly" + errRemove.Error())
			return err
		}
	}	
	return nil
}


func getLogFilePaths() (string, string) {
	executable, err := os.Executable()
	if err != nil { 
		dlog.Error("getSafeBrowsingLogPath: Could not get root path of executable file! Logging will be console only." + err.Error())
		return "", ""
	} 

	currentLogPath := filepath.Dir(executable) + "\\scamjam-dns-watcher.log"
	oldLogPath := filepath.Dir(executable) + "\\scamjam-dns-watcher.old"
	

	return currentLogPath, oldLogPath
}



	

func ctrldInit() { //move this into custom ctrld
	l := zerolog.New(io.Discard)
	watchdog.MainLog.Store(&l)
}

