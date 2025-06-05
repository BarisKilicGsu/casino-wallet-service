package entities

import (
	"time"

	"github.com/BarisKilicGsu/casino-wallet-service/models"
	"gorm.io/gorm"
)

type Player struct {
	ID        string         `json:"id" gorm:"primaryKey"`
	WalletID  string         `json:"wallet_id" gorm:"uniqueIndex"`
	Balance   float64        `json:"balance"`
	Currency  string         `json:"currency"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (p *Player) ToApiResponse() *models.PlayerResponse {
	return &models.PlayerResponse{
		ID:       p.ID,
		WalletID: p.WalletID,
		Balance:  p.Balance,
		Currency: p.Currency,
	}
}
