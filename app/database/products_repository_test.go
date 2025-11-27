package database

import (
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	dialector := postgres.New(postgres.Config{
		Conn:       db,
		DriverName: "postgres",
	})

	gdb, err := gorm.Open(dialector, &gorm.Config{})
	assert.NoError(t, err)

	cleanup := func() {
		db.Close()
	}

	return gdb, mock, cleanup
}

func Test_productsRepository_GetAllProducts(t *testing.T) {
	gdb, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewProductsRepository(gdb)

	offset := 0
	limit := 10

	t.Run("success with default filters", func(t *testing.T) {
		filters := ProductFilter{
			Offset: &offset,
			Limit:  &limit,
		}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "products"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		selectProductQuery := regexp.QuoteMeta(`SELECT * FROM "products" ORDER BY id ASC LIMIT $1`)
		mock.ExpectQuery(selectProductQuery).
			WithArgs(limit).
			WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name", "price"}).
				AddRow(1, "PROD001", "Product 1", 100.0).
				AddRow(2, "PROD002", "Product 2", 200.0),
			)

		selectProductCategoryQuery := regexp.QuoteMeta(`SELECT * FROM "product_categories" WHERE "product_categories"."product_id" IN ($1,$2)`)
		mock.ExpectQuery(selectProductCategoryQuery).
			WithArgs(1, 2).
			WillReturnRows(sqlmock.NewRows([]string{"id", "product_id", "code", "name"}).
				AddRow(1, 1, uuid.New(), "Shoes").
				AddRow(2, 2, uuid.New(), "Clothings"),
			)

		selectProductVariantsQuery := regexp.QuoteMeta(`SELECT * FROM "product_variants" WHERE "product_variants"."product_id" IN ($1,$2)`)
		mock.ExpectQuery(selectProductVariantsQuery).
			WithArgs(1, 2).
			WillReturnRows(sqlmock.NewRows([]string{"id", "product_id", "name", "sku", "price"}).
				AddRow(1, 1, "VARIANT A", "SKU001A", 10.0).
				AddRow(2, 2, "VARIANT B", "SKU001B", 20.0),
			)

		products, count, err := repo.GetAllProducts(filters)
		assert.NoError(t, err)
		assert.NotNil(t, count)
		assert.Equal(t, int64(2), *count)
		assert.Len(t, products, 2)
		assert.Equal(t, uint(1), products[0].ID)
		assert.Equal(t, uint(2), products[1].ID)
	})

	t.Run("success with category filter", func(t *testing.T) {
		category := "Shoes"

		filters := ProductFilter{
			Offset:   &offset,
			Limit:    &limit,
			Category: &category,
		}

		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT count(*) FROM "products" JOIN product_categories ON product_categories.product_id = products.id WHERE product_categories.name = $1`)).
			WithArgs(category).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT "products"."id","products"."code","products"."price" FROM "products" JOIN product_categories ON product_categories.product_id = products.id WHERE product_categories.name = $1 ORDER BY id ASC LIMIT $2`)).
			WithArgs(category, limit).
			WillReturnRows(sqlmock.NewRows([]string{"id", "code", "price"}).
				AddRow(3, "code3", 150.0),
			)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "product_categories" WHERE "product_categories"."product_id" = $1`)).
			WithArgs(3).
			WillReturnRows(sqlmock.NewRows([]string{"id", "product_id", "code", "name"}).
				AddRow(3, 3, uuid.New(), "Shoes"),
			)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "product_variants" WHERE "product_variants"."product_id" = $1`)).
			WithArgs(3).
			WillReturnRows(sqlmock.NewRows([]string{"id", "product_id", "name", "sku", "price"}))

		products, count, err := repo.GetAllProducts(filters)
		assert.NoError(t, err)
		assert.NotNil(t, count)
		assert.Equal(t, int64(1), *count)
		assert.Len(t, products, 1)
	})

	t.Run("success with price filter", func(t *testing.T) {
		price := 120.0
		filters := ProductFilter{
			Offset:        &offset,
			Limit:         &limit,
			PriceLessThan: &price,
		}

		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT count(*) FROM "products" WHERE price < $1`)).
			WithArgs(price).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT * FROM "products" WHERE price < $1 ORDER BY id ASC LIMIT $2`)).
			WithArgs(price, limit).
			WillReturnRows(sqlmock.NewRows([]string{"id", "code", "price"}).
				AddRow(4, "PROD004", 100.0),
			)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "product_categories" WHERE "product_categories"."product_id" = $1`)).
			WithArgs(4).
			WillReturnRows(sqlmock.NewRows([]string{"id", "product_id", "code", "name"}).
				AddRow(4, 4, uuid.New(), "Accessories"),
			)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "product_variants" WHERE "product_variants"."product_id" = $1`)).
			WithArgs(4).
			WillReturnRows(sqlmock.NewRows([]string{"id", "product_id", "name", "sku", "price"}))

		products, count, err := repo.GetAllProducts(filters)
		assert.NoError(t, err)
		assert.NotNil(t, count)
		assert.Equal(t, int64(1), *count)
		assert.Len(t, products, 1)
	})

	t.Run("offset greater than count", func(t *testing.T) {
		off := 5
		filters := ProductFilter{
			Offset: &off,
			Limit:  &limit,
		}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "products"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		products, count, err := repo.GetAllProducts(filters)
		assert.NoError(t, err)
		assert.NotNil(t, count)
		assert.Equal(t, int64(2), *count)
		assert.Len(t, products, 0)
	})

	t.Run("error on count query", func(t *testing.T) {
		filters := ProductFilter{
			Offset: &offset,
			Limit:  &limit,
		}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "products"`)).
			WillReturnError(assert.AnError)

		products, count, err := repo.GetAllProducts(filters)
		assert.Error(t, err)
		assert.Nil(t, products)
		assert.Nil(t, count)
	})
}

func Test_productsRepository_GetProductByCode(t *testing.T) {
	gdb, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewProductsRepository(gdb)

	t.Run("success", func(t *testing.T) {
		code := "PROD001"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "products" WHERE code = $1 ORDER BY "products"."id" LIMIT $2`)).
			WithArgs(code, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name", "price"}).
				AddRow(1, code, "Product 1", 100.0),
			)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "product_categories" WHERE "product_categories"."product_id" = $1`)).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "product_id", "code", "name"}).
				AddRow(1, 1, uuid.New(), "Shoes"),
			)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "product_variants" WHERE "product_variants"."product_id" = $1`)).
			WithArgs(1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "product_id", "name", "sku", "price"}).
				AddRow(1, 1, "VARIANT A", "SKU001A", 10.0),
			)

		product, err := repo.GetProductByCode(code)
		assert.NoError(t, err)
		assert.NotNil(t, product)
		assert.Equal(t, uint(1), product.ID)
		assert.Equal(t, code, product.Code)
	})

	t.Run("not found", func(t *testing.T) {
		code := "PROD404"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "products" WHERE code = $1 ORDER BY "products"."id" LIMIT 1`)).
			WithArgs(code).
			WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name", "price"}))

		product, err := repo.GetProductByCode(code)
		assert.Error(t, err)
		assert.Nil(t, product)
	})

	t.Run("db error", func(t *testing.T) {
		code := "PRODERR"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "products" WHERE code = $1 ORDER BY "products"."id" LIMIT 1`)).
			WithArgs(code).
			WillReturnError(assert.AnError)

		product, err := repo.GetProductByCode(code)
		assert.Error(t, err)
		assert.Nil(t, product)
	})
}

func Test_productsRepository_GetProductByID(t *testing.T) {
	gdb, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewProductsRepository(gdb)

	t.Run("success", func(t *testing.T) {
		id := uint(1)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "products" WHERE id = $1 ORDER BY "products"."id" LIMIT $2`)).
			WithArgs(id, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name", "price"}).
				AddRow(1, "PROD001", "Product 1", 100.0),
			)

		product, err := repo.GetProductByID(id)
		assert.NoError(t, err)
		assert.NotNil(t, product)
		assert.Equal(t, uint(1), product.ID)
		assert.Equal(t, "PROD001", product.Code)
	})

	t.Run("not found", func(t *testing.T) {
		id := uint(404)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "products" WHERE id = $1 ORDER BY "products"."id" LIMIT 1`)).
			WithArgs(id).
			WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name", "price"}))

		product, err := repo.GetProductByID(id)
		assert.Error(t, err)
		assert.Nil(t, product)
	})

	t.Run("db error", func(t *testing.T) {
		id := uint(2)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "products" WHERE id = $1 ORDER BY "products"."id" LIMIT 1`)).
			WithArgs(id).
			WillReturnError(assert.AnError)

		product, err := repo.GetProductByID(id)
		assert.Error(t, err)
		assert.Nil(t, product)
	})
}
