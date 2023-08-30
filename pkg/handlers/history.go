package handlers

import (
	"database/sql"
	"encoding/json"
	"featuretester/pkg/errors"
	"featuretester/pkg/history"
	"log"
	"net/http"
	"os"
)

type HistoryHandler struct {
	HistoryRepo history.Repository
	InfoLog     *log.Logger
	ErrLog      *log.Logger
}

func NewHistoryHandler(db *sql.DB) *HistoryHandler {
	return &HistoryHandler{
		HistoryRepo: history.NewHistoryRepo(db),
		InfoLog:     log.New(os.Stdout, "INFO\tHistory HANDLER\t", log.Ldate|log.Ltime),
		ErrLog:      log.New(os.Stdout, "ERROR\tHistory HANDLER\t", log.Ldate|log.Ltime),
	}
}

func (rh *HistoryHandler) GetFeatureHistory(w http.ResponseWriter, r *http.Request) {
	receivedRequest := &history.Request{}

	err := errors.ValidateAndParseJSON(r, receivedRequest)
	if err != nil {
		rh.ErrLog.Printf("%s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	dates, err := rh.HistoryRepo.ParseAndValidateDates(receivedRequest.StartDate, receivedRequest.EndDate)
	if err != nil {
		rh.ErrLog.Printf("%s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userHistory, err := rh.HistoryRepo.GetUserHistory(r.Context(), receivedRequest.UserID, dates)
	if err != nil {
		rh.ErrLog.Printf("%s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	url, err := rh.HistoryRepo.CreateCSV(userHistory)
	if err != nil {
		rh.ErrLog.Printf("%s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(struct {
		CsvURL string `json:"csv_url"`
	}{
		CsvURL: url,
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(resp)
	if err != nil {
		rh.ErrLog.Printf("%s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
