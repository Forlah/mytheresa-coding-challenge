package catalog

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/app/mocks"
	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_CatalogHandler_HandleGet(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	productsRepoMock := mocks.NewMockProductsStore(controller)

	h := NewCatalogHandler(productsRepoMock)
	t.Run("success", func(t *testing.T) {
		dbAllProductsResultMock := []models.Product{
			{
				ID:    1,
				Code:  "PROD001",
				Price: decimal.NewFromFloat(100.0),
				ProductCategory: models.Category{
					ID:        1,
					ProductID: 2,
					Code:      uuid.New(),
					Name:      "Shoes",
				},
			},
			{
				ID:    2,
				Code:  "PROD002",
				Price: decimal.NewFromFloat(200.0),
				ProductCategory: models.Category{
					ID:        1,
					ProductID: 3,
					Code:      uuid.New(),
					Name:      "Clothing",
				},
			},
			{
				ID:    3,
				Code:  "PROD003",
				Price: decimal.NewFromFloat(300.0),
				ProductCategory: models.Category{
					ID:        1,
					ProductID: 4,
					Code:      uuid.New(),
					Name:      "Accessories",
				},
			},
		}

		available := int64(3)

		productsRepoMock.EXPECT().
			GetAllProducts(gomock.Any()).
			Return(dbAllProductsResultMock, &available, nil).
			Times(1)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/catalog?offset=0&limit=10", nil)

		h.HandleGet(w, r)

		assert.Equal(t, http.StatusOK, w.Code)

		responseBody := new(Response)
		decodeErr := json.NewDecoder(w.Result().Body).Decode(responseBody)
		assert.NoError(t, decodeErr)

		response := w.Result()

		// Check the response status code
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.Equal(t, &available, responseBody.Total)
		assert.Len(t, responseBody.Products, int(3))
	})
}

func Test_BuildProductFilter(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	productsRepoMock := mocks.NewMockProductsStore(controller)

	h := NewCatalogHandler(productsRepoMock)

	t.Run("with no query params", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/catalog", nil)

		filter := h.buildProductFilter(r)

		assert.NotNil(t, filter)
		assert.Equal(t, 0, *filter.Offset)
		assert.Equal(t, 10, *filter.Limit)
		assert.Nil(t, filter.Category)
		assert.Nil(t, filter.PriceLessThan)
	})

	t.Run("with all query params", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/catalog?offset=5&limit=20&category=Shoes&priceLessThan=150.5", nil)

		filter := h.buildProductFilter(r)

		assert.NotNil(t, filter)
		assert.Equal(t, 5, *filter.Offset)
		assert.Equal(t, 20, *filter.Limit)
		assert.NotNil(t, filter.Category)
		assert.Equal(t, "Shoes", *filter.Category)
		assert.NotNil(t, filter.PriceLessThan)
		assert.Equal(t, 150.5, *filter.PriceLessThan)
	})

	t.Run("with invalid price_less_than query param", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/catalog?priceLessThan=invalid", nil)

		filter := h.buildProductFilter(r)

		assert.NotNil(t, filter)
		assert.Equal(t, 0, *filter.Offset)
		assert.Equal(t, 10, *filter.Limit)
		assert.Nil(t, filter.Category)
		assert.Nil(t, filter.PriceLessThan)
	})

	t.Run("when maximum limit in filter is exceeded", func(t *testing.T) {

		r := httptest.NewRequest(http.MethodGet, "/catalog?limit=150", nil)

		filter := h.buildProductFilter(r)

		assert.NotNil(t, filter)
		assert.Equal(t, 0, *filter.Offset)
		assert.Equal(t, 100, *filter.Limit)
		assert.Nil(t, filter.Category)
		assert.Nil(t, filter.PriceLessThan)
	})

	t.Run("when negative limit in filter is provided", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/catalog?limit=-10", nil)
		filter := h.buildProductFilter(r)

		assert.NotNil(t, filter)
		assert.Equal(t, 0, *filter.Offset)
		assert.Equal(t, 1, *filter.Limit)
		assert.Nil(t, filter.Category)
		assert.Nil(t, filter.PriceLessThan)
	})
}

func Test_CatalogHandler_HandleGetByCode(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	productsRepoMock := mocks.NewMockProductsStore(controller)

	h := NewCatalogHandler(productsRepoMock)

	t.Run("success", func(t *testing.T) {
		productCode := "PROD001"
		dbProductResultMock := models.Product{
			ID:    1,
			Code:  productCode,
			Price: decimal.NewFromFloat(100.0),
			ProductCategory: models.Category{
				ID:        1,
				ProductID: 2,
				Code:      uuid.New(),
				Name:      "Shoes",
			},
		}

		productsRepoMock.EXPECT().
			GetProductByCode(productCode).
			Return(&dbProductResultMock, nil).
			Times(1)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/catalog/%s", productCode), nil)

		h.HandleGetByCode(w, r)

		assert.Equal(t, http.StatusOK, w.Code)

		responseBody := new(ProductDetail)
		decodeErr := json.NewDecoder(w.Result().Body).Decode(responseBody)
		assert.NoError(t, decodeErr)

		response := w.Result()

		// Check the response status code
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.Equal(t, productCode, responseBody.Code)
	})

	t.Run("missing product code", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/catalog/", nil)

		h.HandleGetByCode(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		responseBody := new(api.ErrorAPIResponse)
		decodeErr := json.NewDecoder(w.Result().Body).Decode(responseBody)
		assert.NoError(t, decodeErr)

		response := w.Result()

		// Check the response status code
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, http.StatusBadRequest, response.StatusCode)
		assert.Equal(t, "Product code is required", responseBody.Error)
	})

	t.Run("product not found", func(t *testing.T) {
		productCode := "NON_EXISTENT_PRODUCT"

		productsRepoMock.EXPECT().
			GetProductByCode(productCode).
			Return(nil, fmt.Errorf("product not found")).
			Times(1)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/catalog/%s", productCode), nil)

		h.HandleGetByCode(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)

		responseBody := new(api.ErrorAPIResponse)
		decodeErr := json.NewDecoder(w.Result().Body).Decode(responseBody)
		assert.NoError(t, decodeErr)

		response := w.Result()

		// Check the response status code
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, http.StatusNotFound, response.StatusCode)
		assert.Equal(t, "Product not found", responseBody.Error)
	})
}
