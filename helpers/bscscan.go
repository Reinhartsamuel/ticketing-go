package helpers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type EtherscanBalanceResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  string `json:"result"` // Balance in Wei
}

func CheckBscTestnetBalance(address string) (string, error) {
	apiKey := os.Getenv("ETHERSCAN_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("missing ETHERSCAN_API_KEY in env")
	}

	url := fmt.Sprintf(
		"https://api.etherscan.io/v2/api?chainid=97&module=account&action=balance&address=%s&tag=latest&apikey=%s",
		address, apiKey,
	)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("error calling Etherscan: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}

	var parsed EtherscanBalanceResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("error unmarshaling JSON: %w", err)
	}

	if parsed.Status != "1" {
		return "", fmt.Errorf("etherscan error: %s", parsed.Message)
	}

	return parsed.Result, nil // Still in wei
}
