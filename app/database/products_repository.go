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

//go:generate mockgen -source=products_repository.go -destination=../mocks/products_repository_mock.go -package=mocks
type ProductsStore interface {
	GetAllProducts(filters ProductFilter) ([]models.Product, *int64, error)
	GetProductByCode(code string) (*models.Product, error)
	GetProductByID(id uint) (*models.Product, error)
}

func NewProductsRepository(db *gorm.DB) *productsRepository {
	return &productsRepository{
		db: db,
	}
}

func (r *productsRepository) GetAllProducts(filters ProductFilter) ([]models.Product, *int64, error) {
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

	if int64(*filters.Offset) > productCount {
		return []models.Product{}, &productCount, nil
	}

	if err := query.
		Order("id ASC").
		Offset(*filters.Offset).
		Limit(*filters.Limit).
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

func (r *productsRepository) GetProductByID(id uint) (*models.Product, error) {
	var product models.Product
	if err := r.db.
		Where("id = ?", id).
		First(&product).Error; err != nil {
		return nil, err
	}

	return &product, nil
}
