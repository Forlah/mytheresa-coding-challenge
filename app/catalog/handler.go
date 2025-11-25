package catalog

import (
	"net/http"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/app/database"
)

type Response struct {
	Products []Product `json:"products"`
}

type Product struct {
	Code  string  `json:"code"`
	Price float64 `json:"price"`
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
	res, err := h.productsRepo.GetAllProducts()
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve products")
		return
	}

	// Map response
	products := make([]Product, len(res))
	for i, p := range res {
		products[i] = Product{
			Code:  p.Code,
			Price: p.Price.InexactFloat64(),
		}
	}

	// Return the products as a JSON response
	response := Response{
		Products: products,
	}

	api.OKResponse(w, response)
}
