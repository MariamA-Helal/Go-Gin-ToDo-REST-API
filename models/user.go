package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	//UserID        uint   `gorm:"primaryKey" json:"id"`
	Username      string `json:"username" gorm:"unique"`
	Password      string `json:"-"`
	Role          string `json:"role" gorm:"default:'user'"`
	SecretKey     string `json:"-"`
	UpgradeStatus string `json:"upgrade_status" gorm:"default:'none'"`
}
