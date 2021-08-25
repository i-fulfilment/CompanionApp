package main

import (
	"cloud.google.com/go/firestore"
	"context"
	_ "embed"
	"fmt"
	"github.com/kardianos/service"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/option"
	"io"
	"io/ioutil"
	"main/companion"
	"os"
	"runtime"
)

//go:embed service-account.json
var credentials []byte

//go:embed libs/SumatraPDF.exe
var sumatraPDF []byte

//go:embed SUMATRAPDF_AUTHORS.txt
var sumatraAuthors []byte

//go:embed SUMATRAPDF_COPYING_BSD_LICENSE.txt
var sumatraCopyingBsd []byte

//go:embed SUMATRAPDF_COPYING_LICENSE.txt
var sumatraCopying []byte

//go:embed LICENSE.txt
var license []byte

//go:embed OPEN_SOURCE.txt
var openSource []byte

//go:embed libs/printer-tools.jar
var printerTools []byte

//go:embed libs/scale-tools.jar
var scaleTools []byte

func Setup(serviceLogger service.Logger, isInteractive bool) (*firestore.Client, companion.LocalConfiguration, error) {

	log.Print("Setting up CompanionApp prerequisites.")

	client, err := getFirestoreClient()

	if err != nil {
		return nil, companion.LocalConfiguration{}, err
	}

	cfg, err := getConfig()

	if err != nil {
		return nil, companion.LocalConfiguration{}, err
	}

	if runtime.GOOS == "windows" {
		err = setupPDFPrinting()

		if err != nil {
			return nil, companion.LocalConfiguration{}, err
		}
	}

	err = setupPrinterTools()

	if err != nil {
		return nil, companion.LocalConfiguration{}, err
	}

	err = setupScaleTools()

	if err != nil {
		return nil, companion.LocalConfiguration{}, err
	}

	SetupLogger(cfg.AppId, client, serviceLogger, isInteractive)

	return client, cfg, nil
}

func SetupLogger(reference string, client *firestore.Client, serviceLogger service.Logger, isInteractive bool) {

	loggers := make([]io.Writer, 0)

	// Add our firestore logger
	loggers = append(loggers, companion.NewFirestoreLogger(reference, client))

	// Optionally add our system logger if we are not running in interactive mode
	if isInteractive == false {
		loggers = append(loggers, companion.NewSystemLogger(serviceLogger))
	} else {
		// Add our console logger
		loggers = append(loggers, zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: zerolog.TimeFormatUnix})
	}

	// Pass our log output writers to our logging library
	log.Logger = zerolog.New(io.MultiWriter(loggers...))
}

func setupPDFPrinting() error {

	dir, err := companion.GetConfigDirectory()

	if err != nil {
		log.Error().Err(err).Msg("Failed to get config directory for copying PDF program")
		return err
	}

	err = ioutil.WriteFile(fmt.Sprintf("%s\\SumatraPDF.exe", dir), sumatraPDF, os.ModePerm)

	if err != nil {
		log.Error().Err(err).Msg("Failed to copy the PDF program")
		return err
	}

	err = ioutil.WriteFile(fmt.Sprintf("%s\\SUMATRAPDF_AUTHORS.txt", dir), sumatraAuthors, os.ModePerm)
	if err != nil {
		log.Error().Err(err).Msg("Failed to copy the SUMATRAPDF_AUTHORS.txt license file")
		return err
	}

	err = ioutil.WriteFile(fmt.Sprintf("%s\\SUMATRAPDF_COPYING_BSD_LICENSE.txt", dir), sumatraCopyingBsd, os.ModePerm)
	if err != nil {
		log.Error().Err(err).Msg("Failed to copy the SUMATRAPDF_COPYING_BSD_LICENSE.txt license file")
		return err
	}

	err = ioutil.WriteFile(fmt.Sprintf("%s\\SUMATRAPDF_COPYING_LICENSE.txt", dir), sumatraCopying, os.ModePerm)
	if err != nil {
		log.Error().Err(err).Msg("Failed to copy the SUMATRAPDF_COPYING_LICENSE.txt license file")
		return err
	}

	return nil
}

func setupPrinterTools() error {

	dir, err := companion.GetConfigDirectory()

	if err != nil {
		log.Error().Err(err).Msg("Failed to get config directory for copying Printer Tools program")
		return err
	}

	if runtime.GOOS == "windows" {
		err = ioutil.WriteFile(fmt.Sprintf("%s\\PrinterTools.jar", dir), printerTools, os.ModePerm)
	} else {
		err = ioutil.WriteFile(fmt.Sprintf("%s/PrinterTools.jar", dir), printerTools, os.ModePerm)
	}

	if err != nil {
		log.Error().Err(err).Msg("Failed to copy the Printer Tools program")
		return err
	}

	return nil
}

func setupLicense() error {

	dir, err := companion.GetConfigDirectory()

	if err != nil {
		log.Error().Err(err).Msg("Failed to get config directory for copying license files program")
		return err
	}

	if runtime.GOOS == "windows" {
		err = ioutil.WriteFile(fmt.Sprintf("%s\\LICENSE.txt", dir), license, os.ModePerm)
	} else {
		err = ioutil.WriteFile(fmt.Sprintf("%s/LICENSE.txt", dir), license, os.ModePerm)
	}

	if err != nil {
		log.Error().Err(err).Msg("Failed to copy the license.txt file")
		return err
	}

	if runtime.GOOS == "windows" {
		err = ioutil.WriteFile(fmt.Sprintf("%s\\OPEN_SOURCE.txt", dir), openSource, os.ModePerm)
	} else {
		err = ioutil.WriteFile(fmt.Sprintf("%s/OPEN_SOURCE.txt", dir), openSource, os.ModePerm)
	}

	if err != nil {
		log.Error().Err(err).Msg("Failed to copy the open_source.txt file")
		return err
	}

	return nil
}

func setupScaleTools() error {

	dir, err := companion.GetConfigDirectory()

	if err != nil {
		log.Error().Err(err).Msg("Failed to get config directory for copying Scale Tools program")
		return err
	}

	if runtime.GOOS == "windows" {
		err = ioutil.WriteFile(fmt.Sprintf("%s\\ScaleTools.jar", dir), scaleTools, os.ModePerm)
	} else {
		err = ioutil.WriteFile(fmt.Sprintf("%s/ScaleTools.jar", dir), scaleTools, os.ModePerm)
	}

	if err != nil {
		log.Error().Err(err).Msg("Failed to copy the Scale Tools program")
		return err
	}

	return nil
}

func getConfig() (companion.LocalConfiguration, error) {

	log.Info().Msg("Getting the config")

	cfg, err := companion.GetConfig()

	if err != nil {
		log.Error().Err(err).Msg("Failed to load the config")
		return companion.LocalConfiguration{}, err
	}

	return cfg, nil
}

func getFirestoreClient() (*firestore.Client, error) {

	log.Info().Msg("Getting the firebase client connection")

	projectId := "**REPLACE_ME**"
	client, err := firestore.NewClient(context.Background(), projectId, option.WithCredentialsJSON(credentials))

	if err != nil {
		log.Error().Str("Error", err.Error()).Caller().Msg("Failed to load the config")
		return nil, err
	}

	return client, nil
}
