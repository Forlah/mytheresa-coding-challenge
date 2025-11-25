package database

import (
	"github.com/mytheresa/go-hiring-challenge/models"
	"gorm.io/gorm"
)

type productsRepository struct {
	db *gorm.DB
}

type ProductsStore interface {
	GetAllProducts() ([]models.Product, error)
}

func NewProductsRepository(db *gorm.DB) *productsRepository {
	return &productsRepository{
		db: db,
	}
}

func (r *productsRepository) GetAllProducts() ([]models.Product, error) {
	var products []models.Product
	if err := r.db.Preload("Variants").Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}
