package repository

import (
	"time"

	"github.com/BarisKilicGsu/casino-wallet-service/internal/entities"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IPlayerRepository interface {
	GetByID(id string, outTx *gorm.DB) (*entities.Player, error)
	GetByIDWithLock(id string, outTx *gorm.DB) (*entities.Player, error)
	GetAll(outTx *gorm.DB) ([]*entities.Player, error)
	UpdateBalance(id string, amount float64, outTx *gorm.DB) error
	Create(player *entities.Player, outTx *gorm.DB) error
}

type playerRepository struct {
	IGormRepository
}

func NewPlayerRepository(repository IGormRepository) IPlayerRepository {
	return &playerRepository{
		IGormRepository: repository,
	}
}

func (r *playerRepository) GetByID(id string, outTx *gorm.DB) (*entities.Player, error) {
	if outTx == nil {
		outTx = r.GetDB()
	}
	var player entities.Player
	if err := outTx.First(&player, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &player, nil
}

func (r *playerRepository) GetByIDWithLock(id string, outTx *gorm.DB) (*entities.Player, error) {
	if outTx == nil {
		outTx = r.GetDB()
	}
	var player entities.Player
	if err := outTx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&player, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &player, nil
}

func (r *playerRepository) GetAll(outTx *gorm.DB) ([]*entities.Player, error) {
	if outTx == nil {
		outTx = r.GetDB()
	}
	var players []*entities.Player
	if err := outTx.Find(&players).Error; err != nil {
		return nil, err
	}
	return players, nil
}

func (r *playerRepository) UpdateBalance(id string, amount float64, outTx *gorm.DB) error {
	if outTx == nil {
		outTx = r.GetDB()
	}
	return outTx.Model(&entities.Player{}).
		Where("id = ?", id).
		UpdateColumn("balance", gorm.Expr("balance + ?", amount)).
		Error
}

func (r *playerRepository) Create(player *entities.Player, outTx *gorm.DB) error {
	if outTx == nil {
		outTx = r.GetDB()
	}
	player.CreatedAt = time.Now()
	player.UpdatedAt = time.Now()
	return outTx.Create(player).Error
}
