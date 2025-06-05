package seed

import (
	"fmt"
	"log"
	"time"

	"github.com/BarisKilicGsu/casino-wallet-service/internal/entities"
	"gorm.io/gorm"
)

const (
	SamplePlayerCount = 30
)

// SeedPlayers creates sample players in the database
func SeedPlayers(db *gorm.DB) error {
	// Koleksiyonda oyuncu var mı kontrol et
	var count int64
	if err := db.Model(&entities.Player{}).Count(&count).Error; err != nil {
		return err
	}

	// If players exist, skip seeding
	if count > 0 {
		log.Println("Players already exist in the database, skipping seeding...")
		return nil
	}

	samplePlayers := []entities.Player{}
	for i := 0; i < SamplePlayerCount; i++ {
		samplePlayers = append(samplePlayers, entities.Player{
			ID:       fmt.Sprintf("player%d", i+1),
			WalletID: fmt.Sprintf("wallet%d", i+1),
			Balance:  100000.00,
			Currency: "INR",
		})
	}

	// Add players
	for i := range samplePlayers {
		samplePlayers[i].CreatedAt = time.Now()
		samplePlayers[i].UpdatedAt = time.Now()
	}

	if err := db.Create(&samplePlayers).Error; err != nil {
		return err
	}

	log.Printf("%d sample players successfully added", len(samplePlayers))
	return nil
}
