package main

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const DATABASE_FILE string = "srpmain.sqlite"

var (
	db *gorm.DB
)

type Losses struct {
	ID       uint   `gorm:"primaryKey"`
	UserName string `gorm:"index; not null; size:40"`
	Url      string `gorm:"uniqueIndex;size:256; not null"`
	Paid     bool   `gorm:"default:false"`
	Batch    uint
	Srp      uint `gorm:"not null; default 1"`
}

func init() {
	var err error

	db, err = gorm.Open(sqlite.Open(DATABASE_FILE), &gorm.Config{})

	if err != nil {
		log.Panicf("Unable to connect to database %v", err)
	}

	db.AutoMigrate(&Losses{})
}
