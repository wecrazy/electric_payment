package model

import (
	"electric_payment/config"

	"gorm.io/gorm"
)

type AdminPasswordChangeLog struct {
	gorm.Model
	Email    string `json:"email" gorm:"column:email"`
	Password string `json:"password" gorm:"column:password"`
}

func (AdminPasswordChangeLog) TableName() string {
	return config.GetConfig().Database.TbAdminPwdChangelog
}
