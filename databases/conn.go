package databases

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/RehanAthallahAzhar/shopeezy-inventory-cart/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct{}

func NewDB() *DB {
	return &DB{}
}

func (d *DB) Connect(ctx context.Context, credential *models.Credential) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Jakarta",
		credential.Host, credential.Username, credential.Password, credential.DatabaseName, credential.Port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		/*
			NamingStrategy: schema.NamingStrategy{  Contoh penggunaan NamingStrategy
				SingularTable: true,  Menggunakan nama tabel tunggal (misal 'user' bukan 'users')
			},

			Di sini TIDAK perlu ada Dialector lain seperti `Dialector: postgres.Open(dsn),`
			karena `postgres.Open(dsn)` adalah Dialector itu sendiri yang diteruskan sebagai argumen pertama.
		*/
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool properties
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Database connection established successfully.")
	return db, nil
}
