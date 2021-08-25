package companion

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"os"
	"runtime"
)

type LocalConfiguration struct {
	AppId string `json:"appId"`
}

func GetConfig() (LocalConfiguration, error) {

	configFilePath, err := GetConfigFilePath()

	if err != nil {
		return LocalConfiguration{}, err
	}

	// Check to see if we have a config file already
	if _, err := os.Stat(configFilePath); err == nil {
		log.Info().Msg("We have an existing config")
		return readConfig()
	}

	log.Info().Msg("We do not have an existing config")

	// Create a new config
	return generateConfig()
}

func generateConfig() (LocalConfiguration, error) {

	uuidRef, err := uuid.NewUUID()

	log.Info().Str("UUID", uuidRef.String()).Msg("Generated the new UUID")

	if err != nil {
		return LocalConfiguration{}, err
	}

	config := LocalConfiguration{
		AppId: uuidRef.String(),
	}

	data, err := json.Marshal(config)

	if err != nil {
		return LocalConfiguration{}, err
	}

	log.Info().Msg("Checking to see if we have a valid directory path or not")

	dir, err := GetConfigDirectory()

	if err != nil {
		log.Error().Msg("Failed to calculate the config directory")
		return LocalConfiguration{}, err
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {

		log.Info().Str("path", dir).Msg("Config folder does not exist, creating directory")

		err := os.MkdirAll(dir, os.ModePerm)

		if err != nil {

			log.Error().Msg("Failed to create new required config directory")

			return LocalConfiguration{}, err
		}
	}

	log.Info().Msg("Config directory exists, about to write the config file")

	configFilePath, err := GetConfigFilePath()

	if err != nil {
		return LocalConfiguration{}, err
	}

	err = ioutil.WriteFile(configFilePath, data, os.ModePerm)

	if err != nil {
		log.Error().Msg("Failed to write the config file")
		return LocalConfiguration{}, err
	}

	log.Info().Msg("Config file was created successfully")

	return config, nil
}

func readConfig() (LocalConfiguration, error) {

	configFilePath, err := GetConfigFilePath()

	if err != nil {
		return LocalConfiguration{}, err
	}

	data, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return LocalConfiguration{}, err
	}

	log.Info().RawJSON("config", data).Msg("Config file loaded")

	var config LocalConfiguration
	err = json.Unmarshal(data, &config)

	if err != nil {
		return LocalConfiguration{}, err
	}

	return config, nil
}

func GetConfigDirectory() (string, error) {

	if runtime.GOOS == "windows" {
		progData := os.Getenv("ProgramData")

		if progData == "" {
			return "", errors.New("failed to find ProgramData env")
		}

		return fmt.Sprintf("%s\\blade", progData), nil
	}

	return "/etc/blade", nil
}

func GetConfigFilePath() (string, error) {

	dir, err := GetConfigDirectory()

	if err != nil {
		return "", err
	}

	if runtime.GOOS == "windows" {
		return fmt.Sprintf("%s\\config.json", dir), nil
	}

	return fmt.Sprintf("%s/config.json", dir), nil
}
