package catalog

import (
	"net/http"
	"strconv"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/app/database"
	"github.com/mytheresa/go-hiring-challenge/models"
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

type ProductDetail struct {
	Code     string           `json:"code"`
	Price    float64          `json:"price"`
	Category string           `json:"category"`
	Variants []models.Variant `json:"variants"`
}

type CatalogHandler struct {
	productsRepo database.ProductsStore
}

func NewCatalogHandler(r database.ProductsStore) *CatalogHandler {
	return &CatalogHandler{
		productsRepo: r,
	}
}

func (h *CatalogHandler) buildProductFilter(r *http.Request) database.ProductFilter {
	filter := database.ProductFilter{}
	query := r.URL.Query()

	category := query.Get("category")
	if category != "" {
		filter.Category = &category
	}

	priceLessThanStr := query.Get("priceLessThan")

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

	if priceLessThanStr != "" {
		if parsedPrice, err := strconv.ParseFloat(priceLessThanStr, 64); err == nil {
			filter.PriceLessThan = &parsedPrice
		}
	}

	const (
		maxLimit     = 100
		minLimit     = 1
		defaultLimit = 10
	)

	offset := 0
	limit := defaultLimit

	if filter.Offset != nil {
		offset = *filter.Offset
	}

	if filter.Limit != nil {
		limit = *filter.Limit
	}

	if limit > maxLimit {
		limit = maxLimit
	} else if limit < 0 {
		limit = minLimit
	}

	filter.Offset = &offset
	filter.Limit = &limit

	return filter
}

func (h *CatalogHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	filter := h.buildProductFilter(r)

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

func (h *CatalogHandler) HandleGetByCode(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		api.ErrorResponse(w, http.StatusBadRequest, "Product code is required")
		return
	}

	product, err := h.productsRepo.GetProductByCode(code)
	if err != nil {
		api.ErrorResponse(w, http.StatusNotFound, "Product not found")
		return
	}

	for index, variant := range product.Variants {
		// if variant price is zero, set it to the product price
		if variant.Price.IsZero() {
			product.Variants[index].Price = product.Price
		}
	}

	response := ProductDetail{
		Code:     product.Code,
		Price:    product.Price.InexactFloat64(),
		Category: product.ProductCategory.Name,
		Variants: product.Variants,
	}

	api.OKResponse(w, response)
}
