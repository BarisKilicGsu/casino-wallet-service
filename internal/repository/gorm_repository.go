package repository

import "gorm.io/gorm"

type IGormRepository interface {
	StartTransaction() (*gorm.DB, error)
	StartTransactionWithIsolation(isolationLevel string) (*gorm.DB, error)
	FinishTransaction(tx *gorm.DB, err error) error
	RollbackTransaction(tx *gorm.DB)
	GetDB() *gorm.DB
	CommitTransaction(tx *gorm.DB) error
}

type gormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) IGormRepository {
	return &gormRepository{db: db}
}

func (r *gormRepository) StartTransaction() (*gorm.DB, error) {
	tx := r.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return tx, nil
}

func (r *gormRepository) StartTransactionWithIsolation(isolationLevel string) (*gorm.DB, error) {
	tx := r.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	// Set isolation level
	if err := tx.Exec("SET TRANSACTION ISOLATION LEVEL " + isolationLevel).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	return tx, nil
}

func (r *gormRepository) FinishTransaction(tx *gorm.DB, err error) error {
	if err != nil {
		tx.Rollback()
	} else {
		tx.Commit()
	}
	return nil
}

func (r *gormRepository) RollbackTransaction(tx *gorm.DB) {
	tx.Rollback()
}

func (r *gormRepository) CommitTransaction(tx *gorm.DB) error {
	return tx.Commit().Error
}

func (r *gormRepository) GetDB() *gorm.DB {
	return r.db
}
