package service

import (
	"chaos/api/model"
	"chaos/api/system"
	"time"
)

func GetUserRef(mainId uint64) (string, error) {
	db := system.GetDb()
	var userRef model.UserRef
	db.Model(&model.UserRef{}).Where("main_id = ?", mainId).First(&userRef)
	var err error
	if userRef.ID == 0 {

		err = db.Save(&model.UserRef{
			MainID:  mainId,
			RefCode: system.GenerateNonce(8) + system.GenerateNonce(4),
			AddTime: time.Now(),
		}).Error
	}
	return userRef.RefCode, err
}
