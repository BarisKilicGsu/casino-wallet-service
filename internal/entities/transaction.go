package entities

import (
	"time"

	"github.com/BarisKilicGsu/casino-wallet-service/models"
	"gorm.io/gorm"
)

type TransactionType string

const (
	TransactionTypeBet    TransactionType = "bet"
	TransactionTypeResult TransactionType = "result"
)

type Transaction struct {
	ID        uint64          `json:"id" gorm:"primaryKey;AUTO_INCREMENT"`
	ReqID     string          `json:"req_id" gorm:"uniqueIndex"`
	PlayerID  string          `json:"player_id" gorm:"index"`
	WalletID  string          `json:"wallet_id" gorm:"index"`
	RoundID   string          `json:"round_id" gorm:"index"`
	SessionID string          `json:"session_id" gorm:"index"`
	GameCode  string          `json:"game_code" gorm:"index"`
	Type      TransactionType `json:"type"`
	Amount    float64         `json:"amount"`
	Currency  string          `json:"currency"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt gorm.DeletedAt  `json:"-" gorm:"index"`
}

func (t *Transaction) CreateFromEventRequest(eventRequest models.EventRequest) {
	t.ReqID = *eventRequest.ReqID
	t.PlayerID = *eventRequest.PlayerID
	t.WalletID = *eventRequest.WalletID
	t.RoundID = *eventRequest.RoundID
	t.SessionID = *eventRequest.SessionID
	t.GameCode = *eventRequest.GameCode
	t.Type = TransactionType(*eventRequest.Type)
	t.Amount = *eventRequest.Amount
	t.Currency = *eventRequest.Currency
}
