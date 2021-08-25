package companion

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"os/exec"
	"runtime"
)

func ReadScales() (int, error) {
	log.Info().Msg("Sending the read scales command")

	dir, err := GetConfigDirectory()

	if err != nil {
		return 0, err
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("java", "-jar", fmt.Sprintf("%s\\ScaleTools.jar", dir))
	} else {
		cmd = exec.Command("java", "-jar", fmt.Sprintf("%s/ScaleTools.jar", dir))
	}

	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	log.Info().Str("output", string(output)).Msg("Read the scales")

	var result ScalesOutput
	err = json.Unmarshal(output, &result)

	if err != nil {
		log.Error().Msg("Could not run command")
		return 0, err
	}

	if result.Error != "" {
		return 0, errors.New("Failed to get weight: " + result.Error)
	}

	return result.Weight, nil
}
