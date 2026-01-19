// Copyright (c) 2015 Daniel Theophanes

// This software is provided 'as-is', without any express or implied
// warranty. In no event will the authors be held liable for any damages
// arising from the use of this software.

// Permission is granted to anyone to use this software for any purpose,
// including commercial applications, and to alter it and redistribute it
// freely, subject to the following restrictions:

//    1. The origin of this software must not be misrepresented; you must not
//    claim that you wrote the original software. If you use this software
//    in a product, an acknowledgment in the product documentation would be
//    appreciated but is not required.

//    2. Altered source versions must be plainly marked as such, and must not be
//    misrepresented as being the original software.

//    3. This notice may not be removed or altered from any source
//    distribution.

// simple does nothing except block while running the service.

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/kardianos/service"
	"github.com/nextdns/nextdns/host"

)

type program struct{
	exit chan struct{}
}

func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	go p.run()
	return nil
}

func (p *program) run() {

	ticker := time.NewTicker(10 * time.Second)
	for {
		fmt.Fprintln(os.Stdout, "Going to sleep for 10 seconds")
		select {
		case tm := <-ticker.C:

				fmt.Fprintln(os.Stdout, "Tick! " + tm.String())
			
				hostDns, err := GetHostDnsServersIPv4()
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error from GetHostDnsServersIPv4, will not make changes to host dns config")
					continue
				}
				
				isHostDnsScamJam := TestHostDnsServersScamJam(hostDns)
				
				err = TestDNS()
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error in response from scamjam-dns-server! Resetting host DNS config to DHCP")
					if isHostDnsScamJam == false {
						fmt.Fprintln(os.Stderr, "Host is not configured to use scamjam-dns-server, no need to reset dns to dhcp.")
						break
					}
					host.ResetDNS()
					break
				}
				
				
				switch isHostDnsScamJam {
				case true:
					fmt.Fprintln(os.Stdout, "All interfaces configured for scamjam-dns-server")
				case false:
					fmt.Fprintln(os.Stdout, "Host not configured for scamjam-dns-server on all interfaces, setting DNS.")
					host.SetDNS("127.0.0.3")
				
				}
			
	
		case <-p.exit:
			ticker.Stop()
		}
	}

	// Do work here
}

func (p *program) Stop(s service.Service) error {
	// Stop should not block. Return with a few seconds.
	return nil
}

func main() {
	svcConfig := &service.Config{
		Name:        "scamjam-dns-watcher",
		DisplayName: "ScamJam DNS Watcher",
		Description: "Service monitors availability of scamjam-dns-server and sets host dns servers accordingly",
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
	}
	err = s.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
	}
}