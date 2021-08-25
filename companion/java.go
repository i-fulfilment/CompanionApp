package companion

import (
	"bytes"
	"github.com/rs/zerolog/log"
	"os/exec"
	"strings"
)

func GetJavaVersion() (string, error) {
	var cmd *exec.Cmd
	cmd = exec.Command("java", "--version")

	log.Info().Str("Command", cmd.String()).Msg("Checking what version of java is installed")

	var errBuffer bytes.Buffer
	cmd.Stderr = &errBuffer

	result, err := cmd.Output()

	if err != nil {
		log.Error().Err(err).Str("Output", errBuffer.String()).Msg("Failed to look up Java version")
		return "", err
	}

	fullText := string(result)

	log.Info().Str("output", fullText).Msg("java -version")

	if fullText == "" {
		log.Warn().Msg("Unknown Java Version")
	}

	return strings.Split(fullText, "\n")[0], nil
}
