package companion

import (
	"encoding/json"
	"github.com/rs/zerolog/log"
	"net/http"
	"strings"
	"time"
)

func (app *App) infoEndpoint(res http.ResponseWriter, req *http.Request) {
	log.Info().Msg("Handling info request")

	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("Access-Control-Allow-Origin", req.Header.Get("Origin"))
	res.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

	res.WriteHeader(http.StatusOK)

	infoResponse := InfoResponse{CompanionAppId: app.Reference}
	data, err := json.Marshal(infoResponse)

	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = res.Write(data)

	if err != nil {
		log.Error().Caller().Err(err).Msg("Failed to send the response")
	}
}

func (app *App) loginEndpoint(res http.ResponseWriter, req *http.Request) {
	requestedHeaders := make([]string, 0)
	for name, _ := range req.Header {
		requestedHeaders = append(requestedHeaders, name)
	}

	allowedHeaders := strings.ToLower(strings.Join(requestedHeaders, ", "))

	res.Header().Set("Access-Control-Allow-Origin", req.Header.Get("Origin"))
	res.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	res.Header().Set("Access-Control-Allow-Headers", allowedHeaders+", content-type")

	if req.Method == "OPTIONS" {
		res.WriteHeader(http.StatusOK)
		res.Write([]byte("OK"))
		return
	}

	log.Info().Msg("Handling logged_in request")

	var request LoggedInRequest
	err := json.NewDecoder(req.Body).Decode(&request)

	if err != nil {
		log.Error().Err(err).Msg("Can not record logged_in event")
	}

	app.User.Id = request.Id
	app.User.CompanyId = request.CompanyId
	app.User.CompanyName = request.CompanyName
	app.User.Name = request.Name
	app.User.LastLogin = time.Now().Unix() * 1000

	err = app.SyncBackToFirestore()

	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)

	infoResponse := LoggedInResponse{Result: "Saved!"}
	data, err := json.Marshal(infoResponse)

	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = res.Write(data)

	if err != nil {
		log.Error().Caller().Err(err).Msg("Failed to send the response")
	}
}

type InfoResponse struct {
	CompanionAppId string `json:"CompanionAppId"`
}

type LoggedInResponse struct {
	Result string `json:"result"`
}

type LoggedInRequest struct {
	CompanyId   string `json:"company_id" firestore:"company_id"`
	CompanyName string `json:"company_name" firestore:"company_name"`
	Id          string `json:"id" firestore:"id"`
	Name        string `json:"name" firestore:"name"`
}
