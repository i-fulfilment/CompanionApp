package companion

import (
	"cloud.google.com/go/firestore"
	"context"
	"github.com/rs/zerolog/log"
	"time"
)

type ScaleJob struct {
	Created            time.Time              `json:"created" firestore:"created"`
	Message            string                 `json:"message" firestore:"message"`
	Status             string                 `json:"status" firestore:"status"`
	Weight             float64                `json:"weight" firestore:"weight"`
	FirestoreReference *firestore.DocumentRef `json:"-" firestore:"-"`
}

type ScalesOutput struct {
	Error  string `json:"error"`
	Weight int    `json:"weight"`
}

func (job *ScaleJob) Handle() {

	startRoutineTime := time.Now()

	// Get the File
	grams, err := job.readScales()

	if err != nil {
		log.Error().Err(err).Msg("Failed to read scales")
		_, _ = job.FirestoreReference.Update(context.Background(), []firestore.Update{
			{
				Path:  "message",
				Value: err.Error(),
			},
			{
				Path:  "status",
				Value: "error",
			},
		})

		return
	}

	log.Debug().Dur("Total Time Taken (ms)", time.Now().Sub(startRoutineTime)).Msg("Completed scale request")

	_, err = job.FirestoreReference.Update(context.Background(), []firestore.Update{
		{
			Path:  "message",
			Value: "Value read okay.",
		},
		{
			Path:  "error",
			Value: "",
		},
		{
			Path:  "status",
			Value: "complete",
		},
		{
			Path:  "weight",
			Value: grams,
		},
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to save the weight back to firestore")
	}
}

func (job *ScaleJob) readScales() (int, error) {
	return ReadScales()
}
