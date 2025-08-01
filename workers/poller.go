package main

import (
	"context"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Standard ERC20 Transfer event ABI
const transferABI = `[{"anonymous":false,"inputs":[
	{"indexed":true,"name":"from","type":"address"},
	{"indexed":true,"name":"to","type":"address"},
	{"indexed":false,"name":"value","type":"uint256"}],
	"name":"Transfer","type":"event"}]`

var (
	usdcAddress  = common.HexToAddress("0xfA02ee4D1B9D8316f4682F71e3E26bD00f0eCF3e") // Example: USDC on Arbitrum Sepolia
	rpcURL       = "https://sepolia-rollup.arbitrum.io/rpc"
	pollInterval = 15 * time.Second // Adjust as needed
)

func main() {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatalf("Failed to connect to RPC: %v", err)
	}
	defer client.Close()

	parsedABI, err := abi.JSON(strings.NewReader(transferABI))
	if err != nil {
		log.Fatalf("Failed to parse ABI: %v", err)
	}
	transferSigHash := parsedABI.Events["Transfer"].ID

	// Block tracking
	latestBlock, err := client.BlockNumber(context.Background())
	if err != nil {
		log.Fatalf("Failed to fetch block: %v", err)
	}

	log.Println("Starting poller at block", latestBlock)

	for {
		time.Sleep(pollInterval)

		newBlock, err := client.BlockNumber(context.Background())
		if err != nil {
			log.Printf("Error fetching new block: %v", err)
			continue
		}

		if newBlock <= latestBlock {
			continue
		}

		// Poll from latestBlock+1 to newBlock
		for block := latestBlock + 1; block <= newBlock; block++ {
			log.Printf("Checking block %d...", block)

			query := ethereum.FilterQuery{
				FromBlock: big.NewInt(int64(block)),
				ToBlock:   big.NewInt(int64(block)),
				Addresses: []common.Address{usdcAddress},
				Topics:    [][]common.Hash{{transferSigHash}},
			}

			logs, err := client.FilterLogs(context.Background(), query)
			if err != nil {
				log.Printf("Error filtering logs: %v", err)
				continue
			}

			for _, vLog := range logs {
				if len(vLog.Topics) < 3 {
					continue
				}

				from := common.HexToAddress(vLog.Topics[1].Hex())
				to := common.HexToAddress(vLog.Topics[2].Hex())

				var amount big.Int
				err := parsedABI.UnpackIntoInterface(&amount, "Transfer", vLog.Data)
				if err != nil {
					log.Printf("Error unpacking value: %v", err)
					continue
				}

				// 🔍 Replace this with your actual Redis or DB check
				if isMonitoredAddress(to) {
					log.Printf("🔔 Payment Detected: from %s to %s, amount %s", from.Hex(), to.Hex(), amount.String())
					// Update DB / Redis payment status here
				}
			}
		}

		latestBlock = newBlock
	}
}

// Dummy address checker
func isMonitoredAddress(addr common.Address) bool {
	// Replace with actual Redis or database check
	monitored := map[string]bool{
		"0x1234567890abcdef1234567890abcdef12345678": true,
	}
	return monitored[strings.ToLower(addr.Hex())]
}
