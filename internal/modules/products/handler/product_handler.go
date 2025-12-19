package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/KevTiv/alieze-erp/internal/modules/products/types"
	"github.com/KevTiv/alieze-erp/internal/modules/products/service"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

type ProductHandler struct {
	service *service.ProductService
}

func NewProductHandler(service *service.ProductService) *ProductHandler {
	return &ProductHandler{
		service: service,
	}
}

func (h *ProductHandler) RegisterRoutes(router *httprouter.Router) {
	router.POST("/api/products", h.CreateProduct)
	router.GET("/api/products/:id", h.GetProduct)
	router.GET("/api/products", h.ListProducts)
	router.PUT("/api/products/:id", h.UpdateProduct)
	router.DELETE("/api/products/:id", h.DeleteProduct)
	router.GET("/api/categories/:category_id/products", h.GetProductsByCategory)
	router.GET("/api/product-types/:product_type", h.GetProductsByType)
	router.GET("/api/products-by-active/:active", h.GetProductsByActiveStatus)
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req types.Product
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdProduct, err := h.service.CreateProduct(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdProduct)
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	product, err := h.service.GetProduct(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if product == nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(product)
}

func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Parse query parameters
	name := r.URL.Query().Get("name")
	defaultCode := r.URL.Query().Get("default_code")
	barcode := r.URL.Query().Get("barcode")
	productType := r.URL.Query().Get("product_type")
	categoryIDStr := r.URL.Query().Get("category_id")
	activeStr := r.URL.Query().Get("active")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	filters := types.ProductFilter{
		Limit:  10,
		Offset: 0,
	}

	if name != "" {
		filters.Name = &name
	}

	if defaultCode != "" {
		filters.DefaultCode = &defaultCode
	}

	if barcode != "" {
		filters.Barcode = &barcode
	}

	if productType != "" {
		filters.ProductType = &productType
	}

	if categoryIDStr != "" {
		categoryID, err := uuid.Parse(categoryIDStr)
		if err != nil {
			http.Error(w, "Invalid category ID", http.StatusBadRequest)
			return
		}
		filters.CategoryID = &categoryID
	}

	if activeStr != "" {
		active, err := strconv.ParseBool(activeStr)
		if err != nil {
			http.Error(w, "Invalid active value", http.StatusBadRequest)
			return
		}
		filters.Active = &active
	}

	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			http.Error(w, "Invalid limit", http.StatusBadRequest)
			return
		}
		filters.Limit = limit
	}

	if offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			http.Error(w, "Invalid offset", http.StatusBadRequest)
			return
		}
		filters.Offset = offset
	}

	products, _, err := h.service.ListProducts(r.Context(), filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(products)
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var req types.Product
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.ID = id

	updatedProduct, err := h.service.UpdateProduct(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedProduct)
}

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteProduct(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductHandler) GetProductsByCategory(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	categoryID, err := uuid.Parse(ps.ByName("category_id"))
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	filters := types.ProductFilter{
		CategoryID: &categoryID,
	}

	products, _, err := h.service.ListProducts(r.Context(), filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(products)
}

func (h *ProductHandler) GetProductsByType(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	productType := ps.ByName("product_type")

	filters := types.ProductFilter{
		ProductType: &productType,
	}

	products, _, err := h.service.ListProducts(r.Context(), filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(products)
}

func (h *ProductHandler) GetProductsByActiveStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	activeStr := ps.ByName("active")
	active, err := strconv.ParseBool(activeStr)
	if err != nil {
		http.Error(w, "Invalid active value", http.StatusBadRequest)
		return
	}

	filters := types.ProductFilter{
		Active: &active,
	}

	products, _, err := h.service.ListProducts(r.Context(), filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(products)
}
