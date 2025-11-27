package database

import (
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/stretchr/testify/assert"
)

func Test_categoriesRepository_GetAllCategories(t *testing.T) {
	gdb, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewCategoriesRepository(gdb)

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "product_categories"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(1, "Shoes").
				AddRow(2, "Bags"),
			)

		categories, err := repo.GetAllCategories()
		assert.NoError(t, err)
		assert.Len(t, categories, 2)
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "product_categories"`)).
			WillReturnError(assert.AnError)

		categories, err := repo.GetAllCategories()
		assert.Error(t, err)
		assert.Nil(t, categories)
	})
}

func Test_categoriesRepository_CreateCategory(t *testing.T) {
	gdb, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewCategoriesRepository(gdb)

	t.Run("success", func(t *testing.T) {
		category := models.Category{
			Name: "Accessories",
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "product_categories"`)).
			WithArgs().
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectCommit()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "product_categories" ORDER BY id desc,"product_categories"."id" LIMIT $1`)).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(1, "Accessories"),
			)

		createdCategory, err := repo.CreateCategory(category)
		assert.NoError(t, err)
		assert.NotNil(t, createdCategory)
		assert.Equal(t, uint(1), createdCategory.ID)
		assert.Equal(t, "Accessories", createdCategory.Name)
	})
	t.Run("db error on insert", func(t *testing.T) {
		category := models.Category{
			Name: "Accessories",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "categories" ("name") VALUES ($1)`)).
			WithArgs(category.Name).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		createdCategory, err := repo.CreateCategory(category)
		assert.Error(t, err)
		assert.Nil(t, createdCategory)
	})
}
