package main

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func DBConnection(env *Env) *gorm.DB {
	uri := fmt.Sprintf(
		"host=%s user=%s dbname=%s password=%s sslmode=%s port=5432",
		env.DB_HOST, env.DB_USER, env.DB_NAME, env.DB_PASSWORD, env.DB_SSLMODE,
	)

	db, err := gorm.Open(postgres.Open(uri), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		panic("failed to connect database: " + err.Error())
	}
	fmt.Printf("Connected to the database")
	if err := db.AutoMigrate(&Candle{}); err != nil {
		panic("failed to make table: " + err.Error())
	}

	return db
}
