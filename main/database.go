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
	UserName string `gorm:"index; not null; size:40"`
	Paid     bool   `gorm:"default:false"`
	Batch    uint
	Srp      uint64 `gorm:"not null; default 1"`
	ShipId   uint   `gorm:"not null; default 1"`
}

type DoctrineShips struct {
	ShipID uint32 `gorm:"primaryKey"`
	Name   string `gorm:"default:''"`
	Srp    uint64 `gorm:"not null"`
}

func init() {
	var err error

	db, err = gorm.Open(sqlite.Open(DATABASE_FILE), &gorm.Config{})

	if err != nil {
		log.Panicf("Unable to connect to database %v", err)
	}

	db.AutoMigrate(&Losses{})
	db.AutoMigrate(&DoctrineShips{})
}
