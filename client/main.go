package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type CotacaoUSDBRL string

const serverTimeout = 300

func main() {
	cotacaoUSDBRL, err := GetCotacao()
	if err != nil {
		return
	}
	newLine := fmt.Sprintf("Dólar: %s\n", *cotacaoUSDBRL)
	WriteToFile("cotacao.txt", newLine)
}

func GetCotacao() (*CotacaoUSDBRL, error) {
	ctx, cancel := context.WithTimeout(context.Background(), serverTimeout*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/cotacao", nil)
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

func WriteToFile(filename string, stringToWrite string) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	_, err = file.WriteString(stringToWrite)
	if err != nil {
		log.Fatal(err)
	}
}
