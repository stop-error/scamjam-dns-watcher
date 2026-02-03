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
	"time"

	"github.com/jedisct1/dlog"
	"github.com/kardianos/service"

	watchdog "github.com/Control-D-Inc/ctrld/cmd/cli"
)

var currentLogPath, oldLogPath = getLogFilePaths()

var watchdogRunning bool

var ctrldProg watchdog.Prog

var MainWatchdogStopCh = make(chan bool)
// var ctrldConfig ctrld.Config

type program struct{
	exit chan struct{}
}

func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	dlog.Init("scamjam-dns-watcher", dlog.SeverityNotice, "")

	switch {
		case len(currentLogPath) <= 0 || len(oldLogPath) <= 0:
			dlog.Error("Could not get log file path! Will not run log cleanup. Logging will be console only.")
		default:
			err := CleanupLogs(currentLogPath, oldLogPath)
			if err != nil {
				dlog.Error("Error cleaning up log files! Will try to continue, but log files may grow to unmanagable size.")
			}
			dlog.UseLogFile(currentLogPath)
	}


	if len(currentLogPath) <= 0 || len(oldLogPath) <= 0 {
		dlog.Error("Could not get log file path! Will not run log cleanup. Logging will be console only.")
	} else {

	}


	ctrldInit() //move this into custom ctrld
	watchdog.InitConsoleLogging()
	ctrldProg.InitLogging(false)
	ctrldProg.SetMainWatchdogStopCh(MainWatchdogStopCh)
	ctrldProg.PreRun()

	dlog.Notice("Resetting DNS for watchdog start")
	ctrldProg.ResetDNS(false, true)

	dlog.Notice("Starting watchdog goroutine")
	go ctrldProg.SetDNS()
	watchdogRunning = true

	go p.run()

	return nil
	
}

func (p *program) run() {

	ticker := time.NewTicker(40 * time.Second)
	for {
		dlog.Notice("Going to sleep for 40 seconds")
		select {
		case tm := <-ticker.C:

				dlog.Notice("Tick! " + tm.String())
				
				if !TestDNS() {
					dlog.Error("Error in response from scamjam-dns-server!")
					
					if watchdogRunning == true {
						MainWatchdogStopCh <- true
						dlog.Notice("Watchdog should now be closed")
						dlog.Notice("Waiting for waitgroup to return...")
						ctrldProg.DnsWg.Wait()
						watchdogRunning = false
						dlog.Notice("Resetting DNS...")
						ctrldProg.ResetDNS(false, true)
					}
					

				} else {
					dlog.Notice("TestDNS successful")
					if watchdogRunning == false {
						dlog.Notice("Restarting watchdog")
						ctrldProg.PreRun()
						go ctrldProg.SetDNS()
						watchdogRunning = true
					}
					

				}
		case <-p.exit:
			dlog.Notice("scamjam-dns-watchdog has recieved exit signal!")
				if watchdogRunning == true {
					MainWatchdogStopCh <- true
					dlog.Notice("Sending close signal to watchdog")
					dlog.Notice("Waiting for waitgroup to return...")
					ctrldProg.DnsWg.Wait()
					watchdogRunning = false
					dlog.Notice("Resetting host DNS settings")
					ctrldProg.ResetDNS(false, true)
				}
			ticker.Stop()
		}
	}

	// Do work here
}

func (p *program) Stop(s service.Service) error {
	// Stop should not block. Return with a few seconds.
	close(p.exit)
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
		dlog.Error(err.Error())
	}
	err = s.Run()
	if err != nil {
		dlog.Error(err.Error())
	}
}