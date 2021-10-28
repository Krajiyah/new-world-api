package internal

import (
	"fmt"
	"os"

	"github.com/gofrs/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewProdDB() (*gorm.DB, error) {
	return gorm.Open(postgres.Open(os.Getenv("DATABASE_URL")), &gorm.Config{})
}

func NewUnitDB() (*gorm.DB, string, error) {
	u, _ := uuid.NewV4()
	fileName := fmt.Sprintf("unit-test-%s.db", u.String())
	db, err := gorm.Open(sqlite.Open(fileName), &gorm.Config{})
	return db, fileName, err
}
