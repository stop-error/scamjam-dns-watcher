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
	"os"
	"fmt"

	watchdog "github.com/Control-D-Inc/ctrld/cmd/cli"
	"github.com/kardianos/service"
	"github.com/rs/zerolog"
)

var watchdogRunning bool

var ctrldProg watchdog.Prog

var MainWatchdogStopCh = make(chan bool)

var logger zerolog.Logger

type program struct{
	exit chan struct{}
}

func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.

	ctrldInit() //move this into custom ctrld
	watchdog.InitConsoleLogging()
	ctrldProg.InitLogging(false)
	ctrldProg.SetMainWatchdogStopCh(MainWatchdogStopCh)
	ctrldProg.PreRun()

	logger.Info().Msg("Resetting DNS for watchdog start")
	ctrldProg.ResetDNS(false, true)

	logger.Info().Msg("Starting watchdog goroutine")
	go ctrldProg.SetDNS()
	watchdogRunning = true

	go p.run()

	return nil
	
}

func (p *program) run() {

	ticker := time.NewTicker(25 * time.Second)
	for {
		logger.Info().Msg("Going to sleep for 25 seconds")
		select {
		case tm := <-ticker.C:

				logger.Info().Msg("Tick! " + tm.String())
				
				if !testDNS() {
					logger.Error().Msg("Error in response from scamjam-dns-server!")
					
					if watchdogRunning == true {
						MainWatchdogStopCh <- true
						logger.Info().Msg("Watchdog should now be closed")
						logger.Info().Msg("Waiting for waitgroup to return...")
						ctrldProg.DnsWg.Wait()
						watchdogRunning = false
						logger.Info().Msg("Resetting DNS...")
						ctrldProg.ResetDNS(false, true)
					}
					

				} else {
					logger.Info().Msg("TestDNS successful")
					if watchdogRunning == false {
						logger.Info().Msg("Restarting watchdog")
						ctrldProg.PreRun()
						go ctrldProg.SetDNS()
						watchdogRunning = true
					}
					

				}
		case <-p.exit:
			logger.Info().Msg("scamjam-dns-watchdog has recieved exit signal!")
				if watchdogRunning == true {
					MainWatchdogStopCh <- true
					logger.Info().Msg("Sending close signal to watchdog")
					logger.Info().Msg("Waiting for waitgroup to return...")
					ctrldProg.DnsWg.Wait()
					watchdogRunning = false
					logger.Info().Msg("Resetting host DNS settings")
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

	
	useLogFile, logPath := initLogger()
	if useLogFile == true {
		logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0664)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error opening log file! logging will be console only.")
			logger = zerolog.New(os.Stderr).With().Caller().Logger()
		} else {
			multi := zerolog.MultiLevelWriter(os.Stdout, logFile)
			logger = zerolog.New(multi).With().Caller().Logger()
		}
		
	}
	
	svcConfig := &service.Config{
		Name:        "scamjam-dns-watcher",
		DisplayName: "ScamJam DNS Watcher",
		Description: "Service monitors availability of scamjam-dns-server and sets host dns servers accordingly",
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		logger.Error().Msg(err.Error())
	}
	err = s.Run()
	if err != nil {
		logger.Error().Msg(err.Error())
	}
}