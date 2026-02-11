package main

import (
	"os"
	"io"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	watchdog "github.com/Control-D-Inc/ctrld/cmd/cli"

)

func ctrldInit() { //move this into custom ctrld
	l := zerolog.New(io.Discard)
	watchdog.MainLog.Store(&l)
}


func CopyFile(src string, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		log.Error().Msg("Error opening file at " + src + "Error: " + err.Error())
		return err
	}

	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		log.Error().Msg("Error creating file at " + dst + "! Error: " + err.Error())
		return err
	}

	defer destFile.Close()

	_, err = io.Copy(sourceFile, destFile)
	if err != nil {
		log.Error().Msg("Error copying file " + dst + "! Error: " + err.Error())
		return err
	}

	return nil

}


