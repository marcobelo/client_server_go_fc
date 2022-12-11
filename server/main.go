package main

import (
	"database/sql"
	"encoding/json"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"net/http"
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
	err = SaveCotacaoUSDBRL(cotacaoUSDBRL)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cotacaoUSDBRL.USDBRL.Bid)
}

func GetCotacaoUSDBRL() (*CotacaoUSDBRL, error) {
	res, err := http.Get("https://economia.awesomeapi.com.br/json/last/USD-BRL")
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

func SaveCotacaoUSDBRL(cotacaoUSDBRL *CotacaoUSDBRL) error {
	db, err := sql.Open("sqlite3", "./server/db.sqlite")
	if err != nil {
		return err
	}
	//TODO: (timeout max 10ms) save the cotation in a sqlite database
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
