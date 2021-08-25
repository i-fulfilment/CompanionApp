package companion

import (
	"cloud.google.com/go/firestore"
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"
)

const AppVersion = "2.0.0"

func InitialiseApp(client *firestore.Client, config LocalConfiguration) (*App, error) {

	log.Info().Msg("Creating Companion App Instance")

	app := &App{
		Version:   AppVersion,
		Reference: config.AppId,
		firestore: client,
	}

	isNew, err := app.loadInitialConfigFromFirestore()

	if err != nil {
		log.Error().Caller().Err(err).Msg("Failed to load the app data from firestore")
		return nil, err
	}

	if isNew {
		log.Info().Msg("New app install")
		err = app.SyncBackToFirestore()

		if err != nil {
			return nil, err
		}
	}

	// Get the list of available printers
	err = app.updateAvailablePrinters()

	if err != nil {
		return nil, err
	}

	// Remove logs from any previous sessions
	err = app.deleteOldLogs(time.Now().Add(time.Minute * -1))

	if err != nil {
		return nil, err
	}

	// Remove any uncompleted print jobs from any previous sessions
	err = app.deleteOldPrintJobs()

	if err != nil {
		return nil, err
	}

	go func() {
		for range time.Tick(time.Second * 15) {
			if app.IsStarted {
				_ = app.updateAvailablePrinters()
			}
		}
	}()

	go func() {
		for range time.Tick(time.Hour * 1) {
			if app.IsStarted {
				_ = app.deleteOldLogs(time.Now().Add(time.Hour * -1))
			}
		}
	}()

	return app, nil
}

func (app *App) Start() {

	log.Info().Msg("Starting up the Companion App")

	app.IsStarted = true

	// Sync changes to our config from firestore
	go app.startReceivingConfigUpdates()

	// Deal with the incoming print jobs
	go app.startReceivingPrintJobs()

	// Deal with the incoming scale jobs
	go app.startReceivingScaleJobs()

	log.Info().Msg("Companion App is listening for new print & scales jobs to process.")

	_ = app.startWebServer()
}

func (app *App) Stop() error {

	app.IsStarted = false

	log.Info().Msg("Shutting down the Companion App")

	if app.firestorePrintJobIterator != nil {
		app.firestorePrintJobIterator.Stop()
		app.firestorePrintJobIterator = nil
		log.Info().Msg("Stopped listening to print jobs")
	}

	if app.firestoreScaleJobIterator != nil {
		app.firestoreScaleJobIterator.Stop()
		app.firestoreScaleJobIterator = nil
		log.Info().Msg("Stopped listening to scale jobs")
	}

	if app.firestoreConfigIterator != nil {
		app.firestoreConfigIterator.Stop()
		app.firestoreConfigIterator = nil
		log.Info().Msg("Stopped listening to config updates")
	}

	if app.server != nil {
		err := app.server.Shutdown(context.Background())
		if err != nil {
			log.Error().Err(err).Msg("Failed to shutdown the web server")
			return err
		}

		app.server = nil
		log.Info().Msg("Stopped local webserver")
	}

	return nil
}

func (app *App) getPrinterReference(printerType PrinterType) (*PrinterReference, error) {

	var reference *PrinterReference

	switch printerType {
	case Document:
		reference = &app.Printers.Document
	case LabelSmall:
		reference = &app.Printers.LabelSmall
	case LabelLarge:
		reference = &app.Printers.LabelLarge
	case GiftNote:
		reference = &app.Printers.GiftNote
	}

	if reference == nil {
		return nil, errors.New("no printer has been configured for this document type")
	}

	// Check the reference is able to print somewhere
	if reference.Reference == "" && reference.Forwarding == "" {
		return nil, errors.New("no device or forwarding address provided for printer type")
	}

	return reference, nil
}

func (app *App) startReceivingPrintJobs() {
	app.firestorePrintJobIterator = app.firestore.Collection("CompanionApps").Doc(app.Reference).Collection("PrintJobs").OrderBy("created", firestore.Asc).StartAfter(time.Now().Unix()).Snapshots(context.Background())

	log.Info().Msg("Started listening for inbound print jobs.")

	for {

		if app.firestorePrintJobIterator == nil {
			return
		}

		snap, err := app.firestorePrintJobIterator.Next()

		// Ignore errors related to the end of the iterator. These are expected on shutdown.
		if err != nil && errors.Is(iterator.Done, err) == false {
			log.Warn().Err(err).Msg("Error receiving print job")
			continue
		}

		if errors.Is(iterator.Done, err) {
			log.Info().Msg("Print job iterator is complete")
			return
		}

		app.handlePrintJobCollectionChanges(snap.Changes)
	}
}

func (app *App) startReceivingScaleJobs() {
	app.firestoreScaleJobIterator = app.firestore.Collection("CompanionApps").Doc(app.Reference).Collection("ScaleJobs").OrderBy("created", firestore.Asc).StartAfter(time.Now().Unix()).Snapshots(context.Background())

	log.Info().Msg("Started listening for inbound scale jobs.")

	for {
		if app.firestoreScaleJobIterator == nil {
			return
		}

		snap, err := app.firestoreScaleJobIterator.Next()

		// Ignore errors related to the end of the iterator. These are expected on shutdown.
		if err != nil && errors.Is(iterator.Done, err) == false {
			log.Warn().Err(err).Msg("Error receiving scale job")
			continue
		}

		if errors.Is(iterator.Done, err) {
			log.Info().Msg("Scale job iterator is complete")
			return
		}

		app.handleScaleJobCollectionChanges(snap.Changes)
	}
}

func (app *App) handlePrintJobCollectionChanges(changes []firestore.DocumentChange) {

	for _, change := range changes {
		if change.Kind != firestore.DocumentAdded {
			continue
		}

		log.Info().Interface("job", change.Doc.Data()).Msg("New document added to the print jobs collection")

		record := change.Doc.Data()

		var printerType PrinterType

		printerTypeRaw := (record["printer_type"]).(string)
		switch printerTypeRaw {
		case "document":
			printerType = Document
			break
		case "gift_note":
			printerType = GiftNote
			break
		case "label_small":
			printerType = LabelSmall
			break
		case "label_large":
			printerType = LabelLarge
			break
		}

		reference, err := app.getPrinterReference(printerType)

		if err != nil {
			log.Error().Err(err).Caller().Msg("Failed to get the printer reference")
			return
		}

		quantity, _ := strconv.Atoi((record["quantity"]).(string))

		printJob := PrintJob{
			PrinterType:        printerType,
			Quantity:           quantity,
			Created:            time.Unix((record["created"]).(int64), 0),
			Url:                (record["url"]).(string),
			Printer:            reference,
			FirestoreReference: change.Doc.Ref,
		}

		// Keep record of our recent print job
		app.LastPrintJob = &printJob

		app.SyncBackToFirestore()

		// Do the print
		go printJob.Handle()
	}
}

func (app *App) handleScaleJobCollectionChanges(changes []firestore.DocumentChange) {

	for _, change := range changes {
		if change.Kind != firestore.DocumentAdded {
			continue
		}

		log.Info().Interface("job", change.Doc.Data()).Msg("New document added to the scale jobs collection")

		record := change.Doc.Data()

		scaleJob := ScaleJob{
			Created:            time.Unix((record["created"]).(int64), 0),
			FirestoreReference: change.Doc.Ref,
		}

		// Do the print
		go scaleJob.Handle()
	}
}

func (app *App) stopReceivingPrintJobs() {
	if app.firestorePrintJobIterator != nil {
		app.firestorePrintJobIterator.Stop()
	}

	app.firestorePrintJobIterator = nil
}

func (app *App) SyncBackToFirestore() error {

	app.JavaVersion, _ = GetJavaVersion()
	app.OperatingSystem = runtime.GOOS
	app.Hostname, _ = os.Hostname()

	_, err := app.firestore.Collection("CompanionApps").Doc(app.Reference).Set(context.Background(), app)

	if err != nil {
		log.Error().Err(err).Caller().Msg("Failed to sync to firestore")
		return err
	}

	return nil
}

func (app *App) loadInitialConfigFromFirestore() (bool, error) {

	log.Info().Str("AppId", app.Reference).Msg("Attempting to load app data from Firestore")

	document, err := app.firestore.Collection("CompanionApps").Doc(app.Reference).Get(context.Background())

	log.Info().Msg("Got a response from firestore")

	if status.Code(err) == codes.NotFound {
		log.Info().Msg("App data does not exist on firestore yet")
		return true, nil
	}

	if err != nil {
		log.Error().Msg("Failed to load app data from firestore")
		return false, err
	}

	record := &App{}
	err = document.DataTo(record)

	if err != nil {
		log.Error().Msg("Failed to Marshal the app data from firestore")
		return false, err
	}

	app.updateAppFromFirestoreData(record)

	return false, nil
}

func (app *App) comparePrinter(current *PrinterReference, compare PrinterReference) {
	if current.Reference != compare.Reference {
		current.Reference = compare.Reference
	}
	if current.Tray != compare.Tray {
		current.Tray = compare.Tray
	}
	if current.Forwarding != compare.Forwarding {
		current.Forwarding = compare.Forwarding
	}
	if current.Name != compare.Name {
		current.Name = compare.Name
	}
}

func (app *App) updateAppFromFirestoreData(record *App) {

	if app.User.Id == "" {
		app.User.Id = record.User.Id
		app.User.Name = record.User.Name
		app.User.CompanyName = record.User.CompanyName
		app.User.CompanyId = record.User.CompanyId
		app.User.LastLogin = record.User.LastLogin
	}

	if app.Bay.Name != record.Bay.Name {
		app.Bay.Name = record.Bay.Name
	}

	if app.Paused != record.Paused {
		app.Paused = record.Paused
	}

	if app.Bay.Name != record.Bay.Name {
		app.Bay.Name = record.Bay.Name
	}

	if app.Scale.Forwarding != record.Scale.Forwarding {
		app.Scale.Forwarding = record.Scale.Forwarding
	}

	if app.Scale.Name != record.Scale.Name {
		app.Scale.Name = record.Scale.Name
	}

	if app.Scale.VendorId != record.Scale.VendorId {
		app.Scale.VendorId = record.Scale.VendorId
	}

	if app.Scale.ProductId != record.Scale.ProductId {
		app.Scale.ProductId = record.Scale.ProductId
	}

	app.comparePrinter(&app.Printers.Document, record.Printers.Document)
	app.comparePrinter(&app.Printers.GiftNote, record.Printers.GiftNote)
	app.comparePrinter(&app.Printers.LabelLarge, record.Printers.LabelLarge)
	app.comparePrinter(&app.Printers.LabelSmall, record.Printers.LabelSmall)

	log.Info().Msg("Updated the app data with data from firestore")
}

func (app *App) startReceivingConfigUpdates() {

	log.Info().Str("AppId", app.Reference).Msg("Attempting to subscribe to future config changes from Firestore")

	app.firestoreConfigIterator = app.firestore.Collection("CompanionApps").Doc(app.Reference).Snapshots(context.Background())

	for {
		document, err := app.firestoreConfigIterator.Next()

		if err != nil {
			log.Error().Err(err).Msg("Failed to get config update from firestore")
		}

		record := &App{}
		err = document.DataTo(record)

		if err != nil {
			log.Error().Msg("Failed to Marshal the app data from firestore")
			continue
		}

		app.updateAppFromFirestoreData(record)
	}
}

func (app *App) deleteOldLogs(lastValidLog time.Time) error {

	log.Info().Time("Clear Logs Before", lastValidLog).Msg("Cleaning up old logs")

	iterator := app.firestore.Collection("CompanionApps").Doc(app.Reference).Collection("Logs").OrderBy("timestamp", firestore.Asc).EndBefore(lastValidLog.Unix()).Documents(context.Background())

	docs, err := iterator.GetAll()

	if err != nil {
		log.Error().Err(err).Msg("Failed to capture old logs")
		return err
	}

	for _, doc := range docs {
		_, err = app.firestore.Collection("CompanionApps").Doc(app.Reference).Collection("Logs").Doc(doc.Ref.ID).Delete(context.Background())

		if err != nil {
			log.Error().Err(err).Msg("Failed to delete old logs")
			return err
		}
	}

	return nil
}

func (app *App) deleteOldPrintJobs() error {

	log.Info().Msg("Cleaning up old print jobs")

	iterator := app.firestore.Collection("CompanionApps").Doc(app.Reference).Collection("PrintJobs").OrderBy("timestamp", firestore.Asc).EndBefore(time.Now().Unix()).Documents(context.Background())

	docs, err := iterator.GetAll()

	if err != nil {
		log.Error().Err(err).Msg("Failed to capture old print jobs")
		return err
	}

	log.Info().Msg(fmt.Sprintf("Removing %d old print jobs", len(docs)))

	for _, doc := range docs {
		_, err = app.firestore.Collection("CompanionApps").Doc(app.Reference).Collection("PrintJobs").Doc(doc.Ref.ID).Delete(context.Background())

		if err != nil {
			log.Error().Err(err).Msg("Failed to delete old print job")
			return err
		}
	}

	return nil
}

func (app *App) updateAvailablePrinters() error {

	printers, err := ListAvailablePrinters()

	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch the list of available printers")
		return err
	}

	isDirty := len(app.AvailablePrinters) != len(printers)

	if !isDirty {
		for _, availablePrinter := range app.AvailablePrinters {
			match := false
			for _, printer := range printers {
				if printer.Name == availablePrinter.Name {
					match = true
					break
				}
			}

			if match == false {
				isDirty = true
				break
			}
		}
	}

	if isDirty {
		log.Info().Msg("Available printers have changed")
		app.AvailablePrinters = printers
		err = app.SyncBackToFirestore()

		if err != nil {
			return err
		}
	}

	return nil
}

func (app *App) startWebServer() error {

	mux := http.NewServeMux()
	app.server = &http.Server{Addr: ":62222", Handler: mux}

	mux.HandleFunc("/info", app.infoEndpoint)
	mux.HandleFunc("/logged_in", app.loginEndpoint)

	log.Info().Int("Port", 62222).Msg("Starting Local Server")

	err := app.server.ListenAndServe()

	if err != nil && errors.Is(http.ErrServerClosed, err) == false {
		log.Error().Err(err).Msg("Failed to start the local server")
		return err
	}

	return nil
}

type App struct {
	Version                   string    `json:"version" firestore:"version"`
	Reference                 string    `json:"-" firestore:"-"`
	Bay                       Bay       `json:"bay" firestore:"bay"`
	Paused                    bool      `json:"paused" firestore:"paused"`
	Printers                  Printers  `json:"printers" firestore:"printers"`
	Scale                     Scale     `json:"scale" firestore:"scale"`
	User                      User      `json:"user" firestore:"user"`
	JavaVersion               string    `json:"java_version" firestore:"java_version"`
	OperatingSystem           string    `json:"operating_system" firestore:"operating_system"`
	Hostname                  string    `json:"hostname" firestore:"hostname"`
	AvailablePrinters         []Printer `json:"available_printers" firestore:"available_printers"`
	LastPrintJob              *PrintJob `json:"last_print_job" firestore:"last_print_job"`
	IsStarted                 bool      `json:"is_started" firestore:"is_started"`
	firestore                 *firestore.Client
	firestorePrintJobIterator *firestore.QuerySnapshotIterator
	firestoreScaleJobIterator *firestore.QuerySnapshotIterator
	firestoreConfigIterator   *firestore.DocumentSnapshotIterator
	server                    *http.Server
}

type Printers struct {
	Document   PrinterReference `json:"document" firestore:"document"`
	GiftNote   PrinterReference `json:"gift_note" firestore:"gift_note"`
	LabelLarge PrinterReference `json:"label_large" firestore:"label_large"`
	LabelSmall PrinterReference `json:"label_small" firestore:"label_small"`
}

type PrinterReference struct {
	Forwarding string `json:"forwarding" firestore:"forwarding"`
	Name       string `json:"name" firestore:"name"`
	Reference  string `json:"reference" firestore:"reference"`
	Tray       string `json:"tray" firestore:"tray"`
}

type User struct {
	// Has to be string for old app compatibility
	CompanyId   string `json:"company_id" firestore:"company_id"`
	CompanyName string `json:"company_name" firestore:"company_name"`
	// Has to be string for old app compatibility
	Id        string `json:"id" firestore:"id"`
	LastLogin int64 `json:"last_login" firestore:"last_login"`
	Name      string `json:"name" firestore:"name"`
}

type Bay struct {
	Name string `json:"name" firestore:"name"`
}

type Scale struct {
	Forwarding string `json:"forwarding" firestore:"forwarding"`
	Name       string `json:"name" firestore:"name"`
	ProductId  int `json:"product_id" firestore:"product_id"`
	VendorId   int `json:"vendor_id" firestore:"vendor_id"`
}

type PrinterType int

const (
	Document PrinterType = iota
	LabelSmall
	LabelLarge
	GiftNote
)
