package main

import (
	_ "embed"
	"errors"
	"github.com/jessevdk/go-flags"
	"github.com/kardianos/service"
	"github.com/rs/zerolog/log"
	"os"
)

func main() {

	println("* Source code for this software can be found at: https://github.com/i-fulfilment/CompanionApp")
	println("* Licenses are stored alongside the program data (Windows: %ProgramData%, Linux/Mac: /etc/blade)")
	println("* Windows silent PDF printing support via www.sumatrapdfreader.org")
	println()


	var opts FlagOptions

	// Parse any flags that were provided
	var parser = flags.NewParser(&opts, flags.Default)
	if _, err := parser.Parse(); err != nil {
		switch flagsErr := err.(type) {
		case flags.ErrorType:
			if flagsErr == flags.ErrHelp {
				os.Exit(0)
			}
			log.Error().Err(err).Msg("Failed to parse the flags")
			os.Exit(1)
		default:
			log.Error().Err(err).Msg("Failed to parse the flags")
			os.Exit(1)
		}
	}

	if opts.ListPrinters {
		listPrinters()
		return
	}

	if opts.ReadScales {
		readScales()
		return
	}

	if opts.Info {
		info()
		return
	}

	if opts.PrintTestPage != "" {
		printTestPage(opts.PrintTestPage)
		return
	}

	serviceOptions := make(service.KeyValue, 0)

	if service.Platform() == "linux-systemd" {
		// Ask Systemd to restart the service if it errors and quits
		serviceOptions["Restart"] = "on-failure"

		// Do not keep trying if it fails 3 times within 30 secs
		serviceOptions["StartLimitIntervalSec"] = "30s"
		serviceOptions["StartLimitBurst"] = 3
	}

	svcConfig := &service.Config{
		Name:        "CompanionApp",
		DisplayName: "Blade IMS Companion App",
		Description: "A support service to allow Blade IMS to print files and read USB devices.",
		Option:      serviceOptions,
	}

	// Create an instance of our Program
	prg := &Program{}

	s, err := service.New(prg, svcConfig)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	// When the service flag is provided then we don't want to run the app, but instead we
	// want to control the related service. The options are start, stop, restart, install & uninstall
	if opts.Service != "" && opts.Service != "status" {
		err := service.Control(s, opts.Service)
		if err != nil {
			println(err.Error())
			os.Exit(1)
		}
		return
	}

	// The status is handled in a special way
	if opts.Service == "status" {
		status, err := s.Status()
		if errors.Is(service.ErrNotInstalled, err) {
			println("Companion App is not installed as a service. Run using the flag `--service install` to install the Companion App as a service.")
			os.Exit(1)
		}

		if err != nil {
			println(err.Error())
			os.Exit(1)
		}

		switch status {
		case service.StatusUnknown:
			println("Failed to check the status for the Companion App")
			os.Exit(1)
		case service.StatusStopped:
			println("STOPPED")
			os.Exit(0)
		case service.StatusRunning:
			println("RUNNING")
			os.Exit(0)
		default:
			println("Unexpected service status")
			os.Exit(1)
		}
		return
	}

	err = s.Run()

	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}

type FlagOptions struct {
	Service       string `short:"s" long:"service" description:"Control the CompanionApp service." choice:"start" choice:"stop" choice:"status" choice:"restart" choice:"install" choice:"uninstall"`
	ListPrinters  bool   `short:"l" long:"list-printers" description:"List the available printers."`
	ReadScales    bool   `short:"r" long:"read-scales" description:"Read the weight from attached USB scales."`
	PrintTestPage string `short:"p" long:"print-test-page" description:"Print test page. Provide a printer name."`
	Info          bool   `short:"i" long:"info" description:"Get some info regarding the companion app's setup'."`
}
