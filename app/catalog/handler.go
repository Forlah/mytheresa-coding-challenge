package catalog

import (
	"net/http"
	"strconv"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/app/database"
)

type Response struct {
	Products []Product `json:"products"`
	Total    *int64    `json:"total,omitempty"`
}

type Product struct {
	Code     string  `json:"code"`
	Price    float64 `json:"price"`
	Category string  `json:"category"`
}

type CatalogHandler struct {
	productsRepo database.ProductsStore
}

func NewCatalogHandler(r database.ProductsStore) *CatalogHandler {
	return &CatalogHandler{
		productsRepo: r,
	}
}

func (h *CatalogHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	filter := database.ProductFilter{}

	query := r.URL.Query()

	strOffset := query.Get("offset")
	strLimit := query.Get("limit")

	if strOffset != "" {
		if parsedOffset, err := strconv.ParseInt(strOffset, 10, 0); err == nil {
			offset := int(parsedOffset)
			filter.Offset = &offset
		}
	}

	if strLimit != "" {
		if parsedLimit, err := strconv.ParseInt(strLimit, 10, 0); err == nil {
			limit := int(parsedLimit)
			filter.Limit = &limit
		}
	}

	res, count, err := h.productsRepo.GetAllProducts(filter)
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve products")
		return
	}

	// Map response
	products := make([]Product, len(res))
	for i, p := range res {
		products[i] = Product{
			Code:     p.Code,
			Price:    p.Price.InexactFloat64(),
			Category: p.ProductCategory.Name,
		}
	}

	// Return the products as a JSON response
	response := Response{
		Products: products,
		Total:    count,
	}

	api.OKResponse(w, response)
}
