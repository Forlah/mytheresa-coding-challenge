package database

import (
	"github.com/mytheresa/go-hiring-challenge/models"
	"gorm.io/gorm"
)

type productsRepository struct {
	db *gorm.DB
}

type ProductFilter struct {
	Offset        *int
	Limit         *int
	Category      *string
	PriceLessThan *float64
}

type ProductsStore interface {
	GetAllProducts(filters ProductFilter) ([]models.Product, *int64, error)
	GetProductByCode(code string) (*models.Product, error)
}

func NewProductsRepository(db *gorm.DB) *productsRepository {
	return &productsRepository{
		db: db,
	}
}

func (r *productsRepository) GetAllProducts(filters ProductFilter) ([]models.Product, *int64, error) {
	const (
		maxLimit     = 100
		minLimit     = 1
		defaultLimit = 10
	)

	offset := 0
	limit := defaultLimit

	if filters.Offset != nil {
		offset = *filters.Offset
	}

	if filters.Limit != nil {
		limit = *filters.Limit
	}

	if limit > maxLimit {
		limit = maxLimit
	} else if limit < 0 {
		limit = minLimit
	}

	var products []models.Product
	var productCount int64

	query := r.db.
		Preload("Variants").
		Preload("ProductCategory")

	if filters.Category != nil {
		query = query.Joins("JOIN product_categories ON product_categories.product_id = products.id").
			Where("product_categories.name = ?", *filters.Category)
	}

	if filters.PriceLessThan != nil {
		query = query.Where("price < ?", *filters.PriceLessThan)
	}

	if err := query.Model(&models.Product{}).Count(&productCount).Error; err != nil {
		return nil, nil, err
	}

	if int64(offset) > productCount {
		return []models.Product{}, &productCount, nil
	}

	if err := query.
		Order("id ASC").
		Offset(offset).
		Limit(limit).
		Find(&products).Error; err != nil {
		return nil, nil, err
	}

	return products, &productCount, nil
}

func (r *productsRepository) GetProductByCode(code string) (*models.Product, error) {
	var product models.Product
	if err := r.db.
		Preload("Variants").
		Preload("ProductCategory").
		Where("code = ?", code).
		First(&product).Error; err != nil {
		return nil, err
	}

	return &product, nil
}
