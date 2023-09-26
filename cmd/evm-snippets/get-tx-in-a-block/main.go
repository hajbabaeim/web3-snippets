package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type rpcRequest struct {
	ID      int           `json:"id"`
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type rpcResponse struct {
	ID      int             `json:"id"`
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   *rpcError       `json:"error"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Block struct {
	Transactions []Transaction `json:"transactions"`
}

type Transaction struct {
	Hash  string `json:"hash"`
	To    string `json:"to"`
	From  string `json:"from"`
	Value string `json:"value"`
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	infuraKey := os.Getenv("INFURA_KEY")
	nodeURL := fmt.Sprintf("https://mainnet.infura.io/v3/%s", infuraKey)
	receiverAddress := os.Getenv("RECEIVER_ADDRESS")
	fromBlockNumber := 18204230
	toBlockNumber := 18204232

	for blockNumber := fromBlockNumber; blockNumber <= toBlockNumber; blockNumber++ {
		blockHex := intToHex(blockNumber)
		body, err := json.Marshal(rpcRequest{
			ID:      1,
			JSONRPC: "2.0",
			Method:  "eth_getBlockByNumber",
			Params:  []interface{}{blockHex, true},
		})

		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		resp, err := http.Post(nodeURL, "application/json", bytes.NewBuffer(body))
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer resp.Body.Close()

		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		var response rpcResponse
		err = json.Unmarshal(responseBody, &response)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if response.Error != nil {
			fmt.Printf("Error: %d, %s\n", response.Error.Code, response.Error.Message)
			return
		}

		var block Block
		err = json.Unmarshal(response.Result, &block)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		for _, tx := range block.Transactions {
			if tx.To == strings.ToLower(receiverAddress) {
				fmt.Printf("Transaction Hash: %s\n", tx.Hash)
				fmt.Printf("From: %s\n", tx.From)
				amount, _ := hexToBigInt(tx.Value)
				floatValue := new(big.Float).SetInt(amount)
				divisor := new(big.Float).SetFloat64(1e18)
				result := new(big.Float).Quo(floatValue, divisor)
				fmt.Printf("Value: %f\n", result)
				fmt.Printf("Block Number: %d\n", blockNumber)
			}
		}
	}

	fmt.Println("Finished processing blocks.")
}

func hexToInt(hex string) (int, error) {
	var n int
	_, err := fmt.Sscanf(hex, "%x", &n)
	return n, err
}

func hexToBigInt(hex string) (*big.Int, error) {
	value, success := new(big.Int).SetString(hex, 0)
	if !success {
		return nil, fmt.Errorf("failed to convert hex to big.Int")
	}
	return value, nil
}

func intToHex(n int) string {
	return fmt.Sprintf("0x%X", n)
}
