package workers

import (
	"Repos/ticketing-go/models"
	"context"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"gorm.io/gorm"
)

// Standard ERC20 Transfer event ABI
const transferABI = `[{"anonymous":false,"inputs":[
	{"indexed":true,"name":"from","type":"address"},
	{"indexed":true,"name":"to","type":"address"},
	{"indexed":false,"name":"value","type":"uint256"}],
	"name":"Transfer","type":"event"}]`

var (
	usdtAddress  = common.HexToAddress("0xCD60747D9Bbb1da2AfB2F834391f0FF6ccb15f1a") // USDT on BSC testnet
	rpcURL       = "https://bnb-testnet.g.alchemy.com/v2/51MRDeFHeLtd5FrWrTMv0bsusLfs5n8r"
	pollInterval = 15 * time.Second
)

func StartVerificationPoller(db *gorm.DB) {
	log.Println("🚀 Poller started")
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
				Addresses: []common.Address{usdtAddress},
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

				var unpacked []any
				unpacked, err := parsedABI.Unpack("Transfer", vLog.Data)
				if err != nil {
					log.Printf("❌ Error unpacking value: %v", err)
					continue
				}
				amount, ok := unpacked[0].(*big.Int)
				if !ok {
					log.Printf("❌ Error unpacking value: %v", unpacked[0])
					continue
				}

				log.Println("📦 Detected USDT Transfer")
				log.Printf("   ├─ From   : %s", from.Hex())
				log.Printf("   ├─ To     : %s", to.Hex())
				log.Printf("   └─ Amount : %s (raw wei)", amount.String())

				// OPTIONAL: Convert to float if needed
				// amtFloat := new(big.Float).Quo(new(big.Float).SetInt(&amount), big.NewFloat(1e18))
				// log.Printf("   └─ Amount : %f USDT", amtFloat)

				// If you're just testing, comment out DB logic
				// match, err := CheckExpectedPayment(db, to.Hex(), &amount)
				// if err != nil {
				// 	 log.Printf("❌ DB error: %v", err)
				// } else if match {
				// 	 log.Printf("✅ Matched expected payment for: %s", to.Hex())
				// }
			}
		}

		latestBlock = newBlock
	}
}

func CheckExpectedPayment(db *gorm.DB, to string, amount *big.Int) (bool, error) {
	// Do the DB lookup here
	var count int64
	err := db.Table("expected_payments").
		Where("lower(expected_to_address) = lower(?)", to).
		Where("expected_amount = ?", amount.String()).
		Where("paid = FALSE AND expires_at > now()").
		Count(&count).Error

	if err != nil {
		return false, err
	}
	if count > 0 {
		// mark as paid
		err := db.Exec(`
			UPDATE expected_payments
			SET paid = TRUE, paid_at = now()
			WHERE lower(expected_to_address) = lower(?)
				AND expected_amount = ?
				AND paid = FALSE AND expires_at > now()
			LIMIT 1
		`, to, amount.String()).Error
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func StartReservationsPoller(db *gorm.DB) {
	log.Println("🚀 Reservations poller started")

	// get all from Reservations table where payment due is passed
	// payment due <= now()
	var reservations []models.Reservation
	err := db.Where("payment_due <= now()").Find(&reservations).Error
	if err != nil {
		log.Printf("❌ Error fetching reservations: %v", err)
		return
	}

	// update record to payment status == EXPIRED and return the hold ticket quota back
	// to events table
	// for _, reservation := range reservations {
	// 	err := db.Model(&reservation).Updates(models.Reservation{
	// 		PaymentStatus: "EXPIRED",
	// 	}).Error
	// 	if err != nil {
	// 		log.Printf("❌ Error updating reservation: %v", err)
	// 		continue
	// 	}
	// 	// return the hold ticket quota back to events table
	// 	var event models.Event
	// 	err = db.First(&event, reservation.EventID).Error
	// 	if err != nil {
	// 		log.Printf("❌ Error fetching event: %v", err)
	// 		continue
	// 	}
	// 	event.HoldTicketQuota += 1
	// 	err = db.Save(&event).Error
	// 	if err != nil {
	// 		log.Printf("❌ Error updating event: %v", err)
	// 		continue
	// 	}
	// }
}
