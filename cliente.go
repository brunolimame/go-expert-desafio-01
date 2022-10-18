package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type Cotacao struct {
	Bid       string `json:"bid"`
	Timestamp int    `json:"timestamp,string"`
}

func main() {
	salvarCotacao()
}

func salvarCotacao() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		panic(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Erro realizar a requisição %v\n", err)
	}
	defer res.Body.Close()

	var c Cotacao
	err = json.NewDecoder(res.Body).Decode(&c)
	if err != nil {
		fmt.Printf("Erro ao ler os dados da API %v\n", err)
	}

	f, err := os.Create("cotacao.txt")
	if err != nil {
		fmt.Printf("Erro ao criar arquivo cotacao.txt %v\n", err)
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("Dolar: %s", c.Bid))
	if err != nil {
		fmt.Printf("Erro ao gravar dados no arquivo %v\n", err)
	}
	fmt.Println("Cotação salva")
}
