package workers

import (
	"Repos/ticketing-go/models"
	"context"
	"log"
	"math"
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

func roundTo6(f float64) float64 {
	return math.Round(f*1e6) / 1e6
}

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
				txHash := vLog.TxHash.Hex()

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
				amtFloat := new(big.Float).Quo(new(big.Float).SetInt(amount), big.NewFloat(1e6))
				log.Printf("   └─ Amount : %f USDT", amtFloat)

				// If you're just testing, comment out DB logic
				match, err := CheckExpectedPayment(db, strings.ToLower(from.Hex()), strings.ToLower(to.Hex()), amount, txHash)
				if err != nil {
					log.Printf("❌ DB error: %v", err)
				} else if match {
					log.Printf("✅ Matched expected payment for: %s", to.Hex())
				}
			}
		}

		latestBlock = newBlock
	}
}

func CheckExpectedPayment(db *gorm.DB, from string, to string, amount *big.Int, txHash string) (bool, error) {
	log.Printf("🔍 Checking expected payment: FROM %s → TO %s | AMOUNT (raw): %s", from, to, amount.String())

	// Convert to USDT float, rounded to 6 decimals
	amtFloat := new(big.Float).Quo(new(big.Float).SetInt(amount), big.NewFloat(1e6))
	usdRaw, _ := amtFloat.Float64()
	usdRounded := roundTo6(usdRaw)

	var reservation models.Reservation
	err := db.
		Where("lower(customer_wallet) = lower(?)", from).
		Where("lower(merchant_wallet) = lower(?)", to).
		Where("ROUND(payment_amount_usd, 6) = ?", usdRounded).
		Where("payment_status = ?", "PENDING").
		First(&reservation).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Println("❌ No matching reservation found")
			return false, nil
		}
		log.Printf("❌ DB query error: %v", err)
		return false, err
	}

	// Pretty-print the reservation data
	log.Println("✅ Matching reservation found:")
	log.Printf("   ├─ ID               : %d", reservation.ID)
	log.Printf("   ├─ UserID           : %d", reservation.UserID)
	log.Printf("   ├─ Amount (USD)     : %.6f", reservation.PaymentAmountUsd)
	log.Printf("   ├─ Amount (IDR)     : %.2f", reservation.PaymentAmountIdr)
	log.Printf("   ├─ Customer Wallet  : %s", reservation.CustomerWallet)
	log.Printf("   ├─ Merchant Wallet  : %s", reservation.MerchantWallet)
	log.Printf("   ├─ Payment Due      : %s", reservation.PaymentDue)
	log.Printf("   ├─ Created At       : %s", reservation.CreatedAt)
	log.Printf("   └─ Payment Status   : %s", reservation.PaymentStatus)

	// Update the reservation to PAID
	err = db.Exec(`
	UPDATE reservations
	SET payment_status = 'PAID',
	    payment_confirmed_at = now(),
	    payment_tx_hash = ?,
	    payment_tx_url = ?
	WHERE id = ?
`, txHash, "https://testnet.bscscan.com/tx/"+txHash, reservation.ID).Error

	if err != nil {
		log.Printf("❌ Failed to update reservation %d to PAID: %v", reservation.ID, err)
		return false, err
	}

	log.Printf("🎉 Reservation %d marked as PAID", reservation.ID)
	return true, nil
}

func StartReservationsPoller(db *gorm.DB) {
	log.Println("🚀 Reservations poller started")

	// get all from Reservations table where payment due is passed
	// payment due <= now()
	var reservations []models.Reservation
	err := db.Preload("ReservationItems").Where("payment_due <= now() AND payment_status = ?", "PENDING").Find(&reservations).Error
	if err != nil {
		log.Printf("❌ Error fetching reservations: %v", err)
		return
	}

	// update record to payment status == EXPIRED and return the hold ticket quota back
	// to events table
	for _, reservation := range reservations {
		err := db.Model(&reservation).Update("payment_status", "EXPIRED").Error
		if err != nil {
			log.Printf("❌ Error updating reservation status: %v", err)
			continue
		}

		for _, item := range reservation.ReservationItems {
			var event models.Event
			err := db.First(&event, item.EventID).Error
			if err != nil {
				log.Printf("❌ Error fetching event %d: %v", item.EventID, err)
				continue
			}

			event.EventTicketAmount += item.Quantity
			if err := db.Save(&event).Error; err != nil {
				log.Printf("❌ Error returning quota for event %d: %v", item.EventID, err)
				continue
			}
			log.Printf("✅ Returned %d tickets back to event %d", item.Quantity, item.EventID)
		}
	}
}
