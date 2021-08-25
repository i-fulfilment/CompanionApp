package main

import (
	"github.com/kardianos/service"
	"github.com/rs/zerolog/log"
	"main/companion"
)

type Program struct{}

var app *companion.App

// Start should not block. Do the actual work async.
func (p *Program) Start(s service.Service) error {

	log.Info().Msg("Service Start() Requested")

	serviceLogger, err := s.Logger(nil)

	isInteractive := service.Interactive()

	client, cfg, err := Setup(serviceLogger, isInteractive)

	if err != nil {
		return err
	}

	app, err = companion.InitialiseApp(client, cfg)

	if err != nil {
		return err
	}

	go app.Start()

	return nil
}

// Stop should not block. Return with a few seconds.
func (p *Program) Stop(s service.Service) error {
	log.Info().Msg("Service Stop() Requested")
	err := app.Stop()
	if err != nil {
		return err
	}
	return nil
}
