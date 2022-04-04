package models

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DBConfig struct {
	Addr     string `env:"DB_ADDR" envDefault:"127.0.0.1"`
	Port     string `env:"DB_PORT" envDefault:"5432"`
	Database string `env:"DB_USER" envDefault:"user"` //TODO: to do for production
	User     string `env:"DB_USER" envDefault:"user"`
	Password string `env:"DB_PASSWORD" envDefault:"password"`

	MaxIdleConns int `env:"MAX_IDLE_CONNS" envDefault:"20"`
	MaxOpenConns int `env:"MAX_OPEN_CONNS" envDefault:"20"`
}

func connectPostgres(cfg *DBConfig) (*gorm.DB, error) {

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Taipei",
		cfg.Addr, cfg.Port, cfg.User, cfg.Password, cfg.Database)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	// db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{ //logger for debug
	// 	Logger: logger.Default.LogMode(logger.Info),
	// }) //TODO
	if err != nil {
		return nil, err
	}

	return db, nil
}

func GormOpen(cfg *DBConfig) (*gorm.DB, error) {
	db, err := connectPostgres(cfg)
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)

	return db, nil
}
