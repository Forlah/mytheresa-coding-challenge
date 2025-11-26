package database

import (
	"github.com/mytheresa/go-hiring-challenge/models"
	"gorm.io/gorm"
)

type categoriesRepository struct {
	db *gorm.DB
}

func NewCategoriesRepository(db *gorm.DB) *categoriesRepository {
	return &categoriesRepository{
		db: db,
	}
}

//go:generate mockgen -source=categories_repository.go -destination=../mocks/categories_repository_mock.go -package=mocks
type CategoriesStore interface {
	GetAllCategories() ([]models.Category, error)
	CreateCategory(category models.Category) (*models.Category, error)
}

func (r *categoriesRepository) GetAllCategories() ([]models.Category, error) {
	var categories []models.Category
	if err := r.db.Find(&categories).Error; err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *categoriesRepository) CreateCategory(category models.Category) (*models.Category, error) {
	if err := r.db.Create(&category).Error; err != nil {
		return nil, err
	}

	// Retrieve the latest created category from the database
	result := &models.Category{}
	if err := r.db.Order("id desc").First(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}
