package companion

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
)

func ListAvailablePrinters() ([]Printer, error) {

	dir, err := GetConfigDirectory()

	if err != nil {
		return nil, err
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("java", "-jar", fmt.Sprintf("%s\\PrinterTools.jar", dir))
	} else {
		cmd = exec.Command("java", "-jar", fmt.Sprintf("%s/PrinterTools.jar", dir))
	}

	var errorBuffer bytes.Buffer
	cmd.Stderr = &errorBuffer

	output, err := cmd.Output()
	if err != nil {
		log.Error().Err(err).Str("Error Output", string(errorBuffer.Bytes())).Msg("Failed to list the printers")
		return nil, err
	}

	var printers []Printer
	err = json.Unmarshal(output, &printers)

	return printers, nil
}

func PrintFile(printerName string, printerTray string, file *os.File, quantity int) error {

	if printerName == "" {
		return errors.New("no printer name specified")
	}

	if quantity <= 0 {
		return errors.New("invalid print quantity specified")
	}

	if file == nil {
		return errors.New("no file to print specified")
	}

	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {

		log.Info().Msg("Windows Runtime detected. Printing via SumatraPDF")

		dir, err := GetConfigDirectory()

		if err != nil {
			return err
		}

		// Specify the print quantity
		settings := fmt.Sprintf("%dx", quantity)

		// Do we have a specific print tray?
		if printerTray != "" {
			settings += fmt.Sprintf(" bin=%s", printerTray)
		}

		// Generate the print command
		cmd = exec.Command(fmt.Sprintf("%s\\SumatraPDF.exe", dir), "-print-to", fmt.Sprintf("%s", printerName), "-print-settings", settings, fmt.Sprintf("%s", file.Name()))
	} else {

		log.Info().Msg("Unix runtime detected. Printing via lp")

		cmd = exec.Command("lp", "-d", printerName, "-n", strconv.Itoa(quantity), file.Name())

	}

	log.Info().Str("Command", cmd.String()).Msg("About to run print command")

	var errBuff bytes.Buffer
	cmd.Stderr = &errBuff

	// Run the command and get the output
	output, err := cmd.Output()

	log.Info().Str("output", string(output)).Msg("Read the printer output")

	if err != nil {
		log.Error().Str("Error Output", errBuff.String()).Msg("Could not printer")
		return err
	}

	return nil
}

type Printer struct {
	Name  string `json:"name" firestore:"name"`
	Trays []Tray `json:"trays" firestore:"trays"`
}

type Tray struct {
	Name string `json:"name" firestore:"name"`
}
