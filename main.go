package main

import (
	"github.com/rokusei/gopass-server/db"
	"github.com/rokusei/gopass-server/server"
	"gorm.io/gorm"

	"gorm.io/driver/sqlite"
)

func main() {
	gdb, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	gdb.AutoMigrate(&db.User{}, &db.Vault{}, &db.VaultEntry{})
	if err != nil {
		panic("failed to connect database")
	}

	server.Run(gdb)
}
