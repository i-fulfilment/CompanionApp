package companion

import (
	"cloud.google.com/go/firestore"
	"context"
	"github.com/rs/zerolog/log"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type PrintJob struct {
	PrinterType        PrinterType            `json:"printer_type" firestore:"printer_type"`
	Quantity           int                    `json:"quantity" firestore:"quantity"`
	Created            time.Time              `json:"created" firestore:"created"`
	Url                string                 `json:"url" firestore:"url"`
	File               *os.File               `json:"-" firestore:"-"`
	Printer            *PrinterReference      `json:"-" firestore:"-"`
	FirestoreReference *firestore.DocumentRef `json:"-" firestore:"-"`
}

func (job *PrintJob) Handle() {

	startPrintRoutineTime := time.Now()

	// Always clean up at the end
	defer job.clean()

	// Get the File
	downloadDuration, err := job.downloadFile()
	if err != nil {
		log.Error().Err(err).Msg("Failed to download file")
		return
	}

	// Print the File
	printDuration, err := job.print()
	if err != nil {
		log.Error().Err(err).Msg("Failed to print file")
		return
	}

	log.Debug().Dur("Download (ms)", downloadDuration).Dur("Print (ms)", printDuration).Dur("Total Time Taken (ms)", time.Now().Sub(startPrintRoutineTime)).Msg("Completed print request")

	_, err = job.FirestoreReference.Delete(context.Background())

	if err != nil {
		log.Error().Err(err).Msg("Failed to remove the print job from firestore")
		return
	}

	log.Info().Msg("Print job removed from firestore")

	job.FirestoreReference = nil
	job.File = nil
	job.Printer = nil
}

func (job *PrintJob) downloadFile() (time.Duration, error) {

	log.Info().Msg("Downloading file to print")

	startDownload := time.Now()

	resp, err := http.Get(job.Url)

	if err != nil {
		return 0, err
	}

	//goland:noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	file, err := ioutil.TempFile("", "print_job_*.pdf")

	if err != nil {
		return 0, err
	}

	_, err = io.Copy(file, resp.Body)

	if err != nil {
		return 0, err
	}

	//goland:noinspection GoUnhandledErrorResult
	defer file.Close()

	job.File = file

	return time.Now().Sub(startDownload), nil
}

func (job *PrintJob) print() (time.Duration, error) {

	log.Info().Msg("Send the print command")

	startPrintTime := time.Now()

	err := PrintFile(job.Printer.Reference, job.Printer.Tray, job.File, job.Quantity)
	if err != nil {
		return 0, err
	}

	return time.Now().Sub(startPrintTime), nil
}

func (job *PrintJob) clean() {

	if job.File != nil {

		log.Info().Msg("Removing temp file")

		err := os.Remove(job.File.Name())

		if err != nil {
			log.Warn().Str("Error", err.Error()).Msg("Failed to remove the job's file")
		}
	}
}
