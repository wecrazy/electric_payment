package model

import (
	"electric_payment/config"

	"gorm.io/gorm"
)

type WAMessageReply struct {
	gorm.Model
	ID          uint       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	LanguageID  uint       `gorm:"column:language_id" json:"language_id"`     // Foreign key to Language table
	Language    string     `gorm:"-" json:"language"`                         // Language name, not stored in DB
	Keywords    string     `gorm:"column:keywords;type:text" json:"keywords"` // Store keywords as comma-separated string
	ReplyText   string     `gorm:"column:reply_text;type:text" json:"reply_text"`
	ForUserType WAUserType `gorm:"column:for_user_type;type:varchar(20);not null;default:'common'" json:"for_user_type"`
	UserOf      WAUserOf   `gorm:"column:user_of;type:enum('pltmh_palesan','common');default:'common'" json:"user_of"`
}

func (WAMessageReply) TableName() string {
	return config.GetConfig().Database.TbWAMsgReply
}
