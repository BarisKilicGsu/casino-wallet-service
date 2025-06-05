package service

import (
	"errors"
	"fmt"

	"github.com/BarisKilicGsu/casino-wallet-service/internal/entities"
	"github.com/BarisKilicGsu/casino-wallet-service/internal/repository"
	"go.uber.org/zap"
)

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrDuplicateRequest    = errors.New("duplicate request id")
	ErrDuplicateRound      = errors.New("duplicate round id")
	ErrBetNotFound         = errors.New("bet not found")
	ErrInvalidRequest      = errors.New("invalid request")
	ErrGameCodeMismatch    = errors.New("game code mismatch")
	ErrWalletIDMismatch    = errors.New("wallet ID mismatch")
	ErrPlayerIDMismatch    = errors.New("player ID mismatch")
)

type IWalletService interface {
	GetPlayerBalance(playerID string) (*entities.Player, error)
	GetAllPlayers() ([]*entities.Player, error)
	ProcessTransaction(transaction *entities.Transaction) error
}

type WalletService struct {
	playerRepo      repository.IPlayerRepository
	transactionRepo repository.ITransactionRepository
	gormRepository  repository.IGormRepository
}

func NewWalletService(playerRepo repository.IPlayerRepository, transactionRepo repository.ITransactionRepository, gormRepository repository.IGormRepository) IWalletService {
	return &WalletService{
		playerRepo:      playerRepo,
		transactionRepo: transactionRepo,
		gormRepository:  gormRepository,
	}
}

func (s *WalletService) GetPlayerBalance(playerID string) (*entities.Player, error) {
	zap.L().Debug("Querying player balance", zap.String("player_id", playerID))

	player, err := s.playerRepo.GetByID(playerID, nil)
	if err != nil {
		zap.L().Error("Error while querying player balance",
			zap.String("player_id", playerID),
			zap.Error(err))
		return nil, fmt.Errorf("player not found: %w", err)
	}

	zap.L().Info("Player balance queried successfully",
		zap.String("player_id", playerID),
		zap.Float64("balance", player.Balance))
	return player, nil
}

func (s *WalletService) GetAllPlayers() ([]*entities.Player, error) {
	zap.L().Debug("Listing all players")

	players, err := s.playerRepo.GetAll(nil)
	if err != nil {
		zap.L().Error("Error while listing players", zap.Error(err))
		return nil, err
	}

	zap.L().Info("All players listed successfully",
		zap.Int("player_count", len(players)))
	return players, nil
}

func (s *WalletService) ProcessTransaction(transaction *entities.Transaction) error {
	zap.L().Debug("Processing transaction",
		zap.String("req_id", transaction.ReqID),
		zap.String("type", string(transaction.Type)),
		zap.String("player_id", transaction.PlayerID))

	// Transaction'ı SERIALIZABLE izolasyon seviyesinde başlat
	tx, err := s.gormRepository.StartTransaction()
	if err != nil {
		zap.L().Error("Error while starting transaction", zap.Error(err))
		return err
	}

	// Check for duplicate request
	existingTx, err := s.transactionRepo.GetByReqIDWithLock(transaction.ReqID, tx)
	if err == nil && existingTx != nil {
		zap.L().Warn("Duplicate request detected",
			zap.String("req_id", transaction.ReqID))
		s.gormRepository.RollbackTransaction(tx)
		return ErrDuplicateRequest
	}

	// SELECT FOR UPDATE ile player'ı kilitle
	player, err := s.playerRepo.GetByIDWithLock(transaction.PlayerID, tx)
	if err != nil {
		zap.L().Error("Player not found",
			zap.String("player_id", transaction.PlayerID),
			zap.Error(err))
		s.gormRepository.RollbackTransaction(tx)
		return fmt.Errorf("player not found: %w", err)
	}

	switch transaction.Type {
	case entities.TransactionTypeBet:
		// Check for existing bet with same round_id, player_id and wallet_id
		existingBet, err := s.transactionRepo.GetByRoundIDAndPlayerIDAndWalletIDWithLock(
			transaction.RoundID,
			transaction.WalletID,
			entities.TransactionTypeBet,
			tx,
		)
		if err == nil && existingBet != nil {
			zap.L().Warn("Duplicate round detected for bet",
				zap.String("round_id", transaction.RoundID),
				zap.String("wallet_id", transaction.WalletID))
			s.gormRepository.RollbackTransaction(tx)
			return ErrDuplicateRound
		}

		if player.Balance < transaction.Amount {
			zap.L().Warn("Insufficient balance",
				zap.String("player_id", transaction.PlayerID),
				zap.Float64("current_balance", player.Balance),
				zap.Float64("requested_amount", transaction.Amount))
			s.gormRepository.RollbackTransaction(tx)
			return ErrInsufficientBalance
		}
		if err := s.playerRepo.UpdateBalance(player.ID, -transaction.Amount, tx); err != nil {
			zap.L().Error("Error while updating balance",
				zap.String("player_id", transaction.PlayerID),
				zap.Float64("amount", transaction.Amount),
				zap.Error(err))
			s.gormRepository.RollbackTransaction(tx)
			return fmt.Errorf("balance update failed: %w", err)
		}
		zap.L().Info("Bet transaction completed successfully",
			zap.String("player_id", transaction.PlayerID),
			zap.Float64("amount", transaction.Amount))

	case entities.TransactionTypeResult:
		// Check for bet transaction with same round_id, player_id and wallet_id
		betTx, err := s.transactionRepo.GetByRoundIDAndPlayerIDAndWalletIDWithLock(
			transaction.RoundID,
			transaction.WalletID,
			entities.TransactionTypeBet,
			tx,
		)
		if err != nil || betTx == nil {
			zap.L().Warn("Bet not found",
				zap.String("round_id", transaction.RoundID),
				zap.String("player_id", transaction.PlayerID),
				zap.String("wallet_id", transaction.WalletID))
			s.gormRepository.RollbackTransaction(tx)
			return ErrBetNotFound
		}

		// Check if bet and result transactions match
		if betTx.GameCode != transaction.GameCode {
			zap.L().Warn("Game code mismatch between bet and result",
				zap.String("bet_game_code", betTx.GameCode),
				zap.String("result_game_code", transaction.GameCode))
			s.gormRepository.RollbackTransaction(tx)
			return ErrGameCodeMismatch
		}

		if betTx.WalletID != transaction.WalletID {
			zap.L().Warn("Wallet ID mismatch between bet and result",
				zap.String("bet_wallet_id", betTx.WalletID),
				zap.String("result_wallet_id", transaction.WalletID))
			s.gormRepository.RollbackTransaction(tx)
			return ErrWalletIDMismatch
		}

		if betTx.PlayerID != transaction.PlayerID {
			zap.L().Warn("Player ID mismatch between bet and result",
				zap.String("bet_player_id", betTx.PlayerID),
				zap.String("result_player_id", transaction.PlayerID))
			s.gormRepository.RollbackTransaction(tx)
			return ErrPlayerIDMismatch
		}

		// Check for existing result with same round_id
		existingResult, err := s.transactionRepo.GetByRoundIDAndPlayerIDAndWalletIDWithLock(
			transaction.RoundID,
			transaction.WalletID,
			entities.TransactionTypeResult,
			tx,
		)
		if err == nil && existingResult != nil {
			zap.L().Warn("Duplicate round detected for result",
				zap.String("round_id", transaction.RoundID),
				zap.String("player_id", transaction.PlayerID),
				zap.String("wallet_id", transaction.WalletID))
			s.gormRepository.RollbackTransaction(tx)
			return ErrDuplicateRound
		}

		if transaction.Amount > 0 {
			if err := s.playerRepo.UpdateBalance(player.ID, transaction.Amount, tx); err != nil {
				zap.L().Error("Error while updating balance during win transaction",
					zap.String("player_id", transaction.PlayerID),
					zap.Float64("amount", transaction.Amount),
					zap.Error(err))
				s.gormRepository.RollbackTransaction(tx)
				return fmt.Errorf("balance update failed: %w", err)
			}
		} else {
			zap.L().Info("Balance update is not allowed, amount is 0",
				zap.String("player_id", transaction.PlayerID),
				zap.Float64("amount", transaction.Amount))
			s.gormRepository.RollbackTransaction(tx)
			return ErrInvalidRequest
		}

		zap.L().Info("Win transaction completed successfully",
			zap.String("player_id", transaction.PlayerID),
			zap.Float64("amount", transaction.Amount))
	}

	// Save transaction
	if err := s.transactionRepo.Create(transaction, tx); err != nil {
		zap.L().Error("Error while saving transaction",
			zap.String("req_id", transaction.ReqID),
			zap.Error(err))
		s.gormRepository.RollbackTransaction(tx)
		return err
	}

	if err := s.gormRepository.FinishTransaction(tx, err); err != nil {
		zap.L().Error("Error while finishing transaction", zap.Error(err))
		return err
	}

	return nil
}
