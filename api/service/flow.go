package service

import (
	"chaos/api/model"
	"chaos/api/system"
)

func QueryAccountBalance(mainId, assetId uint64) (model.AccountBalance, error) {
	db := system.GetDb()
	var accountBalance model.AccountBalance
	err := db.Model(&model.AccountBalance{}).Where("main_id = ? and asset_id = ?", mainId, assetId).First(&accountBalance).Error
	return accountBalance, err
}
