package main

import (
	"context"
	"database/sql"
	"encoding/json"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"net/http"
	"time"
)

type CotacaoUSDBRL struct {
	USDBRL struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

const economiaApiTimeoutMs = 200
const databaseTimeoutMs = 10

func main() {
	http.HandleFunc("/cotacao", CotacaoHandler)
	http.ListenAndServe(":8080", nil)
}

func CotacaoHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/cotacao" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	cotacaoUSDBRL, err := GetCotacaoUSDBRL()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = SaveCotacaoUSDBRLWithTimeout(cotacaoUSDBRL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cotacaoUSDBRL.USDBRL.Bid)
}

func GetCotacaoUSDBRL() (*CotacaoUSDBRL, error) {
	ctx, cancel := context.WithTimeout(context.Background(), economiaApiTimeoutMs*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var cotacaoUSDBRL CotacaoUSDBRL
	err = json.Unmarshal(body, &cotacaoUSDBRL)
	if err != nil {
		return nil, err
	}
	return &cotacaoUSDBRL, nil
}

func SaveCotacaoUSDBRLWithTimeout(cotacaoUSDBRL *CotacaoUSDBRL) error {
	ctx, cancel := context.WithTimeout(context.Background(), databaseTimeoutMs*time.Millisecond)
	defer cancel()

	done := make(chan error)
	go func() {
		done <- SaveCotacaoUSDBRL(cotacaoUSDBRL)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}

func SaveCotacaoUSDBRL(cotacaoUSDBRL *CotacaoUSDBRL) error {
	db, err := sql.Open("sqlite3", "./server/db.sqlite")
	if err != nil {
		return err
	}

	stmt, err := db.Prepare("insert into cotacoes(code, codein, bid, timestamp) values(?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(cotacaoUSDBRL.USDBRL.Code, cotacaoUSDBRL.USDBRL.Codein, cotacaoUSDBRL.USDBRL.Bid, cotacaoUSDBRL.USDBRL.Timestamp)
	if err != nil {
		return err
	}
	return nil
}
