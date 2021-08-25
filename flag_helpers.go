package main

import (
	_ "embed"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"main/companion"
	"os"
	"runtime"
	"strings"
	"time"
)

/**
This file contains some additional help functions that should ONLY be called from
the main.go flag options.
*/

//go:embed resources/sample.pdf
var samplePdf []byte

func listPrinters() {

	log.Info().Msg("Fetching the list of printers")

	printers, err := companion.ListAvailablePrinters()

	if err != nil {
		log.Error().Caller().Err(err).Msg("Failed to list printers")
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Printer", "Trays"})

	for _, printer := range printers {
		trays := make([]string, 0)
		for _, tray := range printer.Trays {
			trays = append(trays, tray.Name)
		}
		table.Append([]string{printer.Name, strings.Join(trays, ", ")})
	}
	table.Render()
}

func readScales() {

	weight, err := companion.ReadScales()

	if err != nil {
		log.Error().Caller().Err(err).Msg("Failed to read scale")
	}

	println(fmt.Sprintf("%dkg", weight))
}

func printTestPage(printerName string) {

	dir, err := companion.GetConfigDirectory()

	if err != nil {
		log.Error().Err(err).Msg("Failed to copy the sample.pdf file")
		os.Exit(1)
	}

	if runtime.GOOS == "windows" {
		err = ioutil.WriteFile(fmt.Sprintf("%s\\sample.pdf", dir), samplePdf, os.ModePerm)
	} else {
		err = ioutil.WriteFile(fmt.Sprintf("%s/sample.pdf", dir), samplePdf, os.ModePerm)
	}

	if err != nil {
		log.Error().Err(err).Msg("Failed to copy the sample.pdf file")
		os.Exit(1)
	}

	var file *os.File
	if runtime.GOOS == "windows" {
		file, err = os.Open(fmt.Sprintf("%s\\sample.pdf", dir))
	} else {
		file, err = os.Open(fmt.Sprintf("%s/sample.pdf", dir))
	}

	if err != nil {
		log.Error().Err(err).Msg("Failed to open the sample.pdf file")
		os.Exit(1)
	}

	printers, err := companion.ListAvailablePrinters()

	if err != nil {
		log.Error().Caller().Err(err).Msg("Failed to list printers")
		os.Exit(1)
	}

	for _, printer := range printers {
		if printer.Name == printerName {

			before := time.Now()
			err := companion.PrintFile(printerName, "", file, 1)

			if err != nil {
				log.Error().Caller().Err(err).Msg("Failed to printer test page")
				os.Exit(1)
			}

			log.Info().Msg(fmt.Sprintf("Test page printed successfully in %s", time.Now().Sub(before)))
			os.Exit(0)
		}
	}

	log.Error().Msg("Printer name specified could not be found. Use --list-printers to find available printer names.")
}

func info() {

	javaVersion, err := companion.GetJavaVersion()

	var javaVersionString string
	if err != nil {
		javaVersionString = "Failed to lookup java version."
	} else {
		javaVersionString = javaVersion
	}

	configDirectory, err := companion.GetConfigDirectory()

	var configDirectoryString string
	if err != nil {
		configDirectoryString = "Failed to lookup config directory path"
	} else {
		configDirectoryString = configDirectory
	}

	var config string
	configFilePath, err := companion.GetConfigFilePath()

	if err == nil {
		configContents, err := ioutil.ReadFile(configFilePath)
		if err != nil {
			config = "Failed to read config file."
		} else {
			config = string(configContents)
		}
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Value"})

	table.Append([]string{"Config", config})
	table.Append([]string{"Companion App Version", companion.AppVersion})
	table.Append([]string{"Config Directory", configDirectoryString})
	table.Append([]string{"Java Version", javaVersionString})
	table.Append([]string{"Golang Version", runtime.Version()})
	table.Append([]string{"GOARCH", runtime.GOARCH})
	table.Append([]string{"GOOS", runtime.GOOS})

	table.Render()
}
