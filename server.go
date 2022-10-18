package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Moeda struct {
	Moeda CotacaoMoeda `json:"USDBRL"`
}

type CotacaoMoeda struct {
	Bid       string `json:"bid"`
	Timestamp int    `json:"timestamp,string"`
}

func main() {
	prepararBancoDeDados()
	http.HandleFunc("/cotacao", cotacaoHandle)
	http.ListenAndServe(":8080", nil)
}

const (
	dbName     = "cotacao.db"
	dbDriver   = "sqlite3"
	urlCotacao = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
)

func cotacaoHandle(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", urlCotacao, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro durante a requisição da cotação: %v\n", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	var c Moeda
	err = json.NewDecoder(res.Body).Decode(&c)
	if err != nil {
		panic(err)
	}

	salvarCotacaoBanco("USDBRL", c.Moeda)

	w.Header().Set("Contet-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(c.Moeda)

}

func conectarDB() (*sql.DB, error) {
	return sql.Open(dbDriver, "./"+dbName)
}

func prepararBancoDeDados() {
	_, err := os.Stat(dbName)
	if err != nil {
		f, err := os.Create(dbName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao criar arquivo do banco de dados: %v\n", err)
		}
		defer f.Close()

		db, err := conectarDB()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao conectar ao banco de dados: %v\n", err)
			return
		}

		//CRIANDO TABELA NO BANCO DE DADOS
		stmt, _ := db.Prepare("CREATE TABLE IF NOT EXISTS cotacoes (id INTEGER PRIMARY KEY,moeda VARCHAR(64) NULL,valor VARCHAR(50) NULL,timestamp INTEGER NULL)")
		_, err = stmt.Exec()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao criar a tabela no banco de dados: %v\n", err)
			return
		}
	}
}

func salvarCotacaoBanco(moeda string, c CotacaoMoeda) {

	db, err := conectarDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro de conexão ao banco de dados: %v\n", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	stmt, err := db.PrepareContext(ctx, "INSERT INTO cotacoes (moeda, valor,timestamp) VALUES (?, ?, ?)")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao preparar novo registro: %v\n", err)
	}
	_, err = stmt.Exec(moeda, c.Bid, c.Timestamp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro criar um novo registro: %v\n", err)
	}
}
