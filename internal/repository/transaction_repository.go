package repository

import (
	"time"

	"github.com/BarisKilicGsu/casino-wallet-service/internal/entities"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ITransactionRepository interface {
	Create(transaction *entities.Transaction, outTx *gorm.DB) error
	GetByReqID(reqID string, outTx *gorm.DB) (*entities.Transaction, error)
	GetByRoundID(roundID string, outTx *gorm.DB) (*entities.Transaction, error)
	GetByPlayerID(playerID string, outTx *gorm.DB) ([]*entities.Transaction, error)
	GetByReqIDWithLock(reqID string, outTx *gorm.DB) (*entities.Transaction, error)
	GetByRoundIDAndPlayerIDAndWalletIDWithLock(roundID, walletID string, transactionType entities.TransactionType, outTx *gorm.DB) (*entities.Transaction, error)
}

type transactionRepository struct {
	IGormRepository
}

func NewTransactionRepository(repository IGormRepository) ITransactionRepository {
	return &transactionRepository{
		repository,
	}
}

func (r *transactionRepository) Create(transaction *entities.Transaction, outTx *gorm.DB) error {
	if outTx == nil {
		outTx = r.GetDB()
	}
	transaction.CreatedAt = time.Now()
	transaction.UpdatedAt = time.Now()
	return outTx.Create(transaction).Error
}

func (r *transactionRepository) GetByReqID(reqID string, outTx *gorm.DB) (*entities.Transaction, error) {

	if outTx == nil {
		outTx = r.GetDB()
	}

	var transaction entities.Transaction
	if err := outTx.Where("req_id = ?", reqID).First(&transaction).Error; err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *transactionRepository) GetByRoundID(roundID string, outTx *gorm.DB) (*entities.Transaction, error) {
	if outTx == nil {
		outTx = r.GetDB()
	}
	var transaction entities.Transaction
	if err := outTx.Where("round_id = ?", roundID).First(&transaction).Error; err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *transactionRepository) GetByPlayerID(playerID string, outTx *gorm.DB) ([]*entities.Transaction, error) {
	if outTx == nil {
		outTx = r.GetDB()
	}
	var transactions []*entities.Transaction
	if err := outTx.Where("player_id = ?", playerID).Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

func (r *transactionRepository) GetByReqIDWithLock(reqID string, outTx *gorm.DB) (*entities.Transaction, error) {
	if outTx == nil {
		outTx = r.GetDB()
	}
	var transaction entities.Transaction
	if err := outTx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("req_id = ?", reqID).
		First(&transaction).Error; err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *transactionRepository) GetByRoundIDAndPlayerIDAndWalletIDWithLock(roundID, walletID string, transactionType entities.TransactionType, outTx *gorm.DB) (*entities.Transaction, error) {
	if outTx == nil {
		outTx = r.GetDB()
	}
	var transaction entities.Transaction
	if err := outTx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("round_id = ? AND wallet_id = ? AND type = ?",
			roundID, walletID, transactionType).
		First(&transaction).Error; err != nil {
		return nil, err
	}
	return &transaction, nil
}
