package companion

import (
	"encoding/json"
	"github.com/kardianos/service"
	"time"
)

func NewSystemLogger(systemLogger service.Logger) *SystemLogger {
	return &SystemLogger{
		serviceLogger: systemLogger,
	}
}

type SystemLogger struct {
	serviceLogger service.Logger
}

func (logger *SystemLogger) Write(p []byte) (int, error) {

	record := make(map[string]interface{})
	err := json.Unmarshal(p, &record)

	record["timestamp"] = time.Now().Unix()

	err = logger.serviceLogger.Info(record)

	if err != nil {
		return 0, err
	}

	if err != nil {
		return 0, err
	}

	return len(string(p)), nil
}
