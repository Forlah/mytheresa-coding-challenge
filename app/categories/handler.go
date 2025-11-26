package categories

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/app/database"
	"github.com/mytheresa/go-hiring-challenge/models"
)

type Response struct {
	ID        uint   `json:"id"`
	ProductID uint   `json:"product_id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
}

type CreateCategoryRequest struct {
	ProductID uint   `json:"product_id"`
	Name      string `json:"name"`
}

type CategoriesHandler struct {
	categoriesRepo database.CategoriesStore
	productRepo    database.ProductsStore
}

func NewCategoriesHandler(categoryStore database.CategoriesStore, productStore database.ProductsStore) *CategoriesHandler {
	return &CategoriesHandler{
		categoriesRepo: categoryStore,
		productRepo:    productStore,
	}
}

func (h *CategoriesHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	categories, err := h.categoriesRepo.GetAllCategories()
	if err != nil {
		fmt.Println("Error fetching categories:", err)
		api.ErrorResponse(w, http.StatusInternalServerError, "Error fetching categories")

		return
	}

	result := make([]Response, len(categories))
	for i, category := range categories {
		result[i] = Response{
			ID:        category.ID,
			ProductID: category.ProductID,
			Code:      category.Code.String(),
			Name:      category.Name,
		}
	}

	api.OKResponse(w, result)
}

func (h *CategoriesHandler) HandlePost(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading request body")
		api.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")

		return
	}

	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Println("Error closing request body")
		}
	}()

	var payload CreateCategoryRequest
	err = json.Unmarshal(body, &payload)
	if err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, "Invalid JSON payload")

		return
	}

	validCategories, err := h.categoriesRepo.GetAllCategories()
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, "Error fetching categories")

		return
	}

	isExistingCategory := false
	for _, cat := range validCategories {
		if strings.EqualFold(cat.Name, payload.Name) && cat.ProductID == payload.ProductID {
			isExistingCategory = true

			break
		}
	}

	if isExistingCategory {
		api.ErrorResponse(w, http.StatusBadRequest, "Duplicate category name")

		return
	}

	_, err = h.productRepo.GetProductByID(payload.ProductID)
	if err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, "Invalid product ID")

		return
	}

	code := uuid.New()

	category := models.Category{
		ProductID: payload.ProductID,
		Code:      code,
		Name:      payload.Name,
	}

	created, err := h.categoriesRepo.CreateCategory(category)
	if err != nil {
		log.Println("Error creating category:", err)
		api.ErrorResponse(w, http.StatusInternalServerError, "Error creating category")

		return
	}

	response := Response{
		ID:        created.ID,
		ProductID: created.ProductID,
		Code:      created.Code.String(),
		Name:      created.Name,
	}

	api.OKResponse(w, response)
}
