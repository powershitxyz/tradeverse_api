package service

import (
	"chaos/api/chain"
	"chaos/api/log"
	"chaos/api/model"
	"chaos/api/system"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func UpdateAccountBalance(top *chain.TopupTxInfo) error {
	var mainId = top.MainID
	var txHash = top.TxHash
	var amount = top.Amount
	var status = top.Status
	var blockNumber = top.BlockNumber
	var blockTime = top.BlockTime
	var userFromLog = top.UserFromLog

	var refFlowID = top.RefFlowID
	var op = top.Op

	db := system.GetDb()

	tx := db.Begin()
	committed := false
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			log.Error("panic", r)
			return
		}
		if !committed {
			_ = tx.Rollback()
		}
	}()

	var withdrawBalanceFlow model.AccountBalanceFlow
	var refLockBalanceFlow model.AccountBalanceFlow
	if op == model.BalanceFlowOpWithdraw || op == model.BalanceFlowOpUnfreeze {
		tx.Where("id = ?", refFlowID).First(&withdrawBalanceFlow)
		if withdrawBalanceFlow.ID == 0 {
			return errors.New("withdraw flow id not found")
		}
		tx.Where("id = ?", withdrawBalanceFlow.RefFlowID).First(&refLockBalanceFlow)
		if refLockBalanceFlow.ID == 0 {
			return errors.New("ref flow id not found")
		}
	}

	var targetBalanceFlow model.AccountBalanceFlow
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("main_id = ? and tx_hash = ?", mainId, txHash).
		First(&targetBalanceFlow).Error; err != nil {
		// defer tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("account balance flow not found")
		}
		return err
	}
	if targetBalanceFlow.Status != model.BalanceFlowStatusPending {
		// tx.Rollback()
		return errors.New("account balance flow not pending, already processed")
	}

	if status == "success" {
		if op == model.BalanceFlowOpFreeze {
			targetBalanceFlow.Status = model.BalanceFlowStatusPendingWithDraw
		} else {
			targetBalanceFlow.Status = model.BalanceFlowStatusSuccess
		}
		if op == model.BalanceFlowOpWithdraw {
			if refLockBalanceFlow.Status == model.BalanceFlowStatusPendingWithDraw {
				refLockBalanceFlow.Status = model.BalanceFlowStatusSuccess
				refLockBalanceFlow.UpdateTime = time.Now()
				if err := tx.Save(&refLockBalanceFlow).Error; err != nil {
					log.Error("update account balance flow failed", err)
					return err
				}
			}
		} else if op == model.BalanceFlowOpUnfreeze {
			if refLockBalanceFlow.Status == model.BalanceFlowStatusPendingWithDraw {
				refLockBalanceFlow.Status = model.BalanceFlowStatusCanceled
				refLockBalanceFlow.UpdateTime = time.Now()
				if err := tx.Save(&refLockBalanceFlow).Error; err != nil {
					log.Error("update account balance flow failed", err)
					return err
				}
			}
		}
	} else {
		targetBalanceFlow.Status = model.BalanceFlowStatusFailed
	}
	targetBalanceFlow.UpdateTime = time.Now()
	targetBalanceFlow.BlockHeight = int(blockNumber)
	targetBalanceFlow.BlockTimestamp = blockTime
	targetBalanceFlow.LogIndex = 0
	targetBalanceFlow.FromAddr = userFromLog
	targetBalanceFlow.ToAddr = top.To
	targetBalanceFlow.RealAmount = amount.Uint64()
	targetBalanceFlow.ChainID = fmt.Sprintf("%d", top.ChainID)

	if err := tx.Save(&targetBalanceFlow).Error; err != nil {
		log.Error("update account balance flow failed", err)
		// tx.Rollback()
		return err
	}
	if targetBalanceFlow.Status != model.BalanceFlowStatusSuccess && targetBalanceFlow.Status != model.BalanceFlowStatusPendingWithDraw {
		if err := tx.Commit().Error; err != nil {
			return err
		}
		committed = true
		return nil
	}

	var accountBalance model.AccountBalance
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("main_id = ?", mainId).
		First(&accountBalance).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// create account balance
			accountBalance = model.AccountBalance{
				MainID:     mainId,
				AssetID:    0,
				Available:  0,
				Frozen:     0,
				UpdateTime: time.Now(),
			}
			if err := tx.Create(&accountBalance).Error; err != nil {
				log.Error("create account balance failed", err)
				// tx.Rollback()
				return err
			}
		} else {
			// defer tx.Rollback()
			return err
		}
	}

	switch op {
	case model.BalanceFlowOpRecharge:
		accountBalance.Available = accountBalance.Available + amount.Uint64()
	case model.BalanceFlowOpWithdraw:
		var amountUint64 = amount.Uint64()
		if accountBalance.Withdrawal < amountUint64 {
			return errors.New("withdrawal amount is greater than available balance")
		}
		accountBalance.Withdrawal = accountBalance.Withdrawal - amountUint64
	case model.BalanceFlowOpUnfreeze:
		var amountUint64 = amount.Uint64()
		if accountBalance.Withdrawal < amountUint64 {
			return errors.New("withdrawal amount is greater than available balance")
		}
		accountBalance.Withdrawal = accountBalance.Withdrawal - amountUint64
		accountBalance.Available = accountBalance.Available + amountUint64
	}

	accountBalance.UpdateTime = time.Now()

	if err := tx.Save(&accountBalance).Error; err != nil {
		// defer tx.Rollback()
		log.Error("update account balance failed", err)
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}
	committed = true
	return nil
}
