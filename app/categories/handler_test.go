package categories

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/mytheresa/go-hiring-challenge/app/mocks"
	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func Test_CategoriesHandler_HandleGet(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	categoriesRepoMock := mocks.NewMockCategoriesStore(controller)
	productsRepoMock := mocks.NewMockProductsStore(controller)

	h := NewCategoriesHandler(categoriesRepoMock, productsRepoMock)
	t.Run("success", func(t *testing.T) {
		dbAllCategoriesResultMock := []models.Category{
			{
				ID:        1,
				ProductID: 2,
				Code:      uuid.New(),
				Name:      "Shoes",
			},
			{
				ID:        2,
				ProductID: 3,
				Code:      uuid.New(),
				Name:      "Clothing",
			},
		}

		categoriesRepoMock.
			EXPECT().
			GetAllCategories().
			Return(dbAllCategoriesResultMock, nil)

		r := httptest.NewRequest(http.MethodGet, "/categories", nil)
		w := httptest.NewRecorder()

		h.HandleGet(w, r)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		expectedResponse := []Response{
			{
				ID:        1,
				ProductID: 2,
				Code:      dbAllCategoriesResultMock[0].Code.String(),
				Name:      "Shoes",
			},
			{
				ID:        2,
				ProductID: 3,
				Code:      dbAllCategoriesResultMock[1].Code.String(),
				Name:      "Clothing",
			},
		}

		assert.Equal(t, expectedResponse, response)
	})

	t.Run("error from repository", func(t *testing.T) {
		categoriesRepoMock.
			EXPECT().
			GetAllCategories().
			Return(nil, assert.AnError)

		r := httptest.NewRequest(http.MethodGet, "/categories", nil)
		w := httptest.NewRecorder()

		h.HandleGet(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func Test_CategoriesHandler_HandleCreate(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	categoriesRepoMock := mocks.NewMockCategoriesStore(controller)
	productsRepoMock := mocks.NewMockProductsStore(controller)

	h := NewCategoriesHandler(categoriesRepoMock, productsRepoMock)

	t.Run("success", func(t *testing.T) {
		request := CreateCategoryRequest{
			ProductID: 1,
			Name:      "New Category",
		}

		validCategories := []models.Category{
			{
				ID:        1,
				ProductID: 1,
				Code:      uuid.New(),
				Name:      "Category 1",
			},

			{
				ID:        2,
				ProductID: 2,
				Code:      uuid.New(),
				Name:      "Category 2",
			},
		}

		categoriesRepoMock.
			EXPECT().
			GetAllCategories().
			Return(validCategories, nil)

		productsRepoMock.
			EXPECT().
			GetProductByID(request.ProductID).
			Return(&models.Product{ID: request.ProductID}, nil)

		categoriesRepoMock.
			EXPECT().
			CreateCategory(gomock.Any()).
			Return(&models.Category{
				ID:        1,
				ProductID: request.ProductID,
				Code:      uuid.New(),
				Name:      request.Name,
			}, nil)

		reqBodyMock, err := json.Marshal(request)
		assert.NoError(t, err)

		r := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(reqBodyMock))
		w := httptest.NewRecorder()

		h.HandlePost(w, r)

		assert.Equal(t, http.StatusOK, w.Code)

		responseBody := new(Response)
		decodeErr := json.NewDecoder(w.Result().Body).Decode(responseBody)
		assert.NoError(t, decodeErr)

		assert.NoError(t, err)

		assert.Equal(t, uint(1), responseBody.ID)
		assert.Equal(t, request.Name, responseBody.Name)
		assert.Equal(t, request.ProductID, responseBody.ProductID)
	})

	t.Run("duplicate category name", func(t *testing.T) {
		request := CreateCategoryRequest{
			ProductID: 1,
			Name:      "Category 1",
		}

		validCategories := []models.Category{
			{
				ID:        1,
				ProductID: 1,
				Code:      uuid.New(),
				Name:      "Category 1",
			},
		}

		categoriesRepoMock.
			EXPECT().
			GetAllCategories().
			Return(validCategories, nil)

		reqBodyMock, err := json.Marshal(request)
		assert.NoError(t, err)

		r := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(reqBodyMock))
		w := httptest.NewRecorder()

		h.HandlePost(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid product ID", func(t *testing.T) {
		request := CreateCategoryRequest{
			ProductID: 999,
			Name:      "New Category",
		}

		validCategories := []models.Category{}

		categoriesRepoMock.
			EXPECT().
			GetAllCategories().
			Return(validCategories, nil)

		productsRepoMock.
			EXPECT().
			GetProductByID(request.ProductID).
			Return(nil, assert.AnError)

		reqBodyMock, err := json.Marshal(request)
		assert.NoError(t, err)

		r := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(reqBodyMock))
		w := httptest.NewRecorder()

		h.HandlePost(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("error when creating category", func(t *testing.T) {

		request := CreateCategoryRequest{
			ProductID: 1,
			Name:      "New Category",
		}

		validCategories := []models.Category{
			{
				ID:        1,
				ProductID: 1,
				Code:      uuid.New(),
				Name:      "Category 1",
			},

			{
				ID:        2,
				ProductID: 2,
				Code:      uuid.New(),
				Name:      "Category 2",
			},
		}

		categoriesRepoMock.
			EXPECT().
			GetAllCategories().
			Return(validCategories, nil)

		productsRepoMock.
			EXPECT().
			GetProductByID(request.ProductID).
			Return(&models.Product{ID: request.ProductID}, nil)

		categoriesRepoMock.
			EXPECT().
			CreateCategory(gomock.Any()).
			Return(nil, assert.AnError)

		reqBodyMock, err := json.Marshal(request)
		assert.NoError(t, err)

		r := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(reqBodyMock))
		w := httptest.NewRecorder()

		h.HandlePost(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
