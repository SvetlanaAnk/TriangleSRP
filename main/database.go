package main

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
)

type Losses struct {
	Url      string `gorm:"primaryKey"`
	NickName string `gorm:"index; size:60"`
	UserId   string `gorm:"index; not null; size:100"`
	Paid     bool   `gorm:"default:false"`
	Batch    uint
	Srp      uint64 `gorm:"not null; default 1"`
	ShipId   uint   `gorm:"not null; default 1"`
	ShipName string `gorm:"default '"`
	Warnings string `gorm:"default ''"`
}

type DoctrineShips struct {
	Ship_ID uint   `gorm:"primaryKey"`
	Name    string `gorm:"default:''"`
	Srp     uint64 `gorm:"not null"`
}

type ServerConfiguration struct {
	GuildId    string `gorm:"primaryKey"`
	SrpChannel string `gorm:"uniqueIndex"`
}
type Administrators struct {
	UserId       string `gorm:"primaryKey"`
	UserName     string
	IsSuperAdmin bool `gorm:"default: false"`
}

type Ships struct {
	Ship_ID uint   `gorm:"primaryKey"`
	Name    string `gorm:"default:''"`
}

func init() {
	var err error

	db, err = gorm.Open(sqlite.Open(DATABASE_FILE), &gorm.Config{})

	if err != nil {
		log.Panicf("Unable to connect to database %v", err)
	}

	db.AutoMigrate(&Losses{})
	db.AutoMigrate(&DoctrineShips{})
	db.AutoMigrate(&ServerConfiguration{})
	db.AutoMigrate(&Administrators{})
	db.AutoMigrate(&Ships{})

	serverConfigurations := []ServerConfiguration{}
	db.Find(&serverConfigurations)
	for _, config := range serverConfigurations {
		SRP_CHANNEL_MAP[config.GuildId] = config.SrpChannel
	}

	res := db.Where("user_id = ?", "1064094675310477353").First(&Administrators{})
	if res.Error != nil {
		db.Create(&Administrators{UserId: "1064094675310477353", UserName: "theblob8584", IsSuperAdmin: true})
	}
	res = db.Where("user_id = ?", "416767410788630558").First(&Administrators{})
	if res.Error != nil {
		db.Create(&Administrators{UserId: "416767410788630558", UserName: "jinxdecaire", IsSuperAdmin: true})
	}
}
