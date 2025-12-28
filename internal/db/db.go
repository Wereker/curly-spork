package db

import (
	catalog "app/internal/models/catalog"
	order "app/internal/models/order"
	sitecontent "app/internal/models/sitecontent"
	user "app/internal/models/user"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect() *gorm.DB {
	if os.Getenv("ENV") != "production" {
		err := godotenv.Load("deployments/.env.dev")
		if err != nil {
			log.Println("Не удалось загрузить .env.dev:", err)
		}
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Не удалось подключиться к БД: " + err.Error())
	}

	db.AutoMigrate(
		&catalog.Category{},
		&catalog.Manufacturer{},
		&catalog.Attribute{},
		&catalog.Product{},
		&catalog.AttributeValue{},
		&catalog.ProductImage{},
		&catalog.ProductWarehouse{},

		&order.Order{},
		&order.OrderItem{},
		&order.Payment{},
		&order.PaymentMethod{},
		&order.PaymentStatus{},
		&order.Request{},
		&order.RequestStatus{},
		&order.Shipment{},
		&order.ShipmentMethod{},
		&order.ShipmentStatus{},

		&user.Right{},
		&user.User{},

		&sitecontent.CommonBlock{},
		&sitecontent.New{},
		&sitecontent.Story{},
	)

	return db
}
