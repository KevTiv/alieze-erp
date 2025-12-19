package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/KevTiv/alieze-erp/internal/modules/sales/service"
	"github.com/KevTiv/alieze-erp/internal/modules/sales/types"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

type QuoteHandler struct {
	quoteService *service.QuoteService
}

func NewQuoteHandler(quoteService *service.QuoteService) *QuoteHandler {
	return &QuoteHandler{
		quoteService: quoteService,
	}
}

func (h *QuoteHandler) RegisterRoutes(router *httprouter.Router) {
	// Quote management endpoints
	router.POST("/api/v1/sales/quotes", h.CreateQuote)
	router.GET("/api/v1/sales/quotes/:id", h.GetQuote)
	router.POST("/api/v1/sales/quotes/:id/generate", h.GenerateQuotePDF)
	router.POST("/api/v1/sales/quotes/:id/send", h.SendQuote)
	router.POST("/api/v1/sales/quotes/:id/accept", h.AcceptQuote)
	router.GET("/api/v1/sales/quotes/:id/analytics", h.GetQuoteAnalytics)

	// Public endpoints (no auth required)
	router.GET("/api/v1/sales/quotes/public/:token", h.GetPublicQuote)
	router.POST("/api/v1/sales/quotes/public/:token/view", h.TrackQuoteView)
}

// CreateQuoteRequest represents the request body for creating a quote
type CreateQuoteRequest struct {
	types.SalesOrder
	Template      string `json:"template,omitempty"`
	SendImmediately bool `json:"send_immediately,omitempty"`
	RecipientEmail string `json:"recipient_email,omitempty"`
	RecipientName  string `json:"recipient_name,omitempty"`
}

// CreateQuote creates a new quote
func (h *QuoteHandler) CreateQuote(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req CreateQuoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create the quote
	quote, err := h.quoteService.CreateQuote(r.Context(), req.SalesOrder)
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send immediately if requested
	if req.SendImmediately && req.RecipientEmail != "" {
		emailReq := service.QuoteEmailRequest{
			QuoteID:        quote.ID,
			RecipientName:  req.RecipientName,
			RecipientEmail: req.RecipientEmail,
			Subject:        "Your Quote is Ready - " + quote.Reference,
			Message:        "Please review the attached quote at your convenience.",
			AttachPDF:      true,
		}

		// Send asynchronously
		if err := h.quoteService.SendQuoteByEmailAsync(r.Context(), emailReq); err != nil {
			// Log error but don't fail the quote creation
			respondJSON(w, map[string]interface{}{
				"quote": quote,
				"email_error": err.Error(),
			}, http.StatusCreated)
			return
		}
	}

	respondJSON(w, quote, http.StatusCreated)
}

// GetQuote retrieves a quote by ID
func (h *QuoteHandler) GetQuote(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		respondError(w, "Invalid quote ID", http.StatusBadRequest)
		return
	}

	quote, err := h.quoteService.salesOrderService.GetSalesOrder(r.Context(), id)
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if quote == nil {
		respondError(w, "Quote not found", http.StatusNotFound)
		return
	}

	// Verify it's a quote
	if quote.Status != types.SalesOrderStatusQuotation {
		respondError(w, "Order is not a quote", http.StatusBadRequest)
		return
	}

	respondJSON(w, quote, http.StatusOK)
}

// GeneratePDFRequest represents the request body for PDF generation
type GeneratePDFRequest struct {
	Template string `json:"template,omitempty"`
	Download bool   `json:"download,omitempty"`
}

// GenerateQuotePDF generates a PDF for the quote
func (h *QuoteHandler) GenerateQuotePDF(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		respondError(w, "Invalid quote ID", http.StatusBadRequest)
		return
	}

	var req GeneratePDFRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Default to empty request if body is empty
		req = GeneratePDFRequest{Template: "", Download: false}
	}

	// Generate PDF
	pdfBytes, err := h.quoteService.GenerateQuotePDF(r.Context(), id, req.Template)
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// If download requested, return PDF directly
	if req.Download {
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", "attachment; filename=quote-"+id.String()+".pdf")
		w.Write(pdfBytes)
		return
	}

	// Otherwise, save to storage and return URL
	storageKey, err := h.quoteService.SaveQuotePDF(r.Context(), id, req.Template)
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]string{
		"storage_key": storageKey,
		"message":     "PDF generated successfully",
	}, http.StatusOK)
}

// SendQuoteRequest represents the request body for sending a quote
type SendQuoteRequest struct {
	RecipientName  string `json:"recipient_name"`
	RecipientEmail string `json:"recipient_email"`
	Subject        string `json:"subject,omitempty"`
	Message        string `json:"message,omitempty"`
	AttachPDF      bool   `json:"attach_pdf"`
	Async          bool   `json:"async"`
}

// SendQuote sends a quote via email
func (h *QuoteHandler) SendQuote(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		respondError(w, "Invalid quote ID", http.StatusBadRequest)
		return
	}

	var req SendQuoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.RecipientEmail == "" {
		respondError(w, "Recipient email is required", http.StatusBadRequest)
		return
	}

	// Set defaults
	if req.Subject == "" {
		req.Subject = "Your Quote is Ready"
	}
	if req.RecipientName == "" {
		req.RecipientName = "Valued Customer"
	}

	emailReq := service.QuoteEmailRequest{
		QuoteID:        id,
		RecipientName:  req.RecipientName,
		RecipientEmail: req.RecipientEmail,
		Subject:        req.Subject,
		Message:        req.Message,
		AttachPDF:      req.AttachPDF,
	}

	// Send asynchronously or synchronously
	if req.Async {
		err = h.quoteService.SendQuoteByEmailAsync(r.Context(), emailReq)
	} else {
		err = h.quoteService.SendQuoteByEmail(r.Context(), emailReq)
	}

	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]string{
		"message": "Quote sent successfully",
	}, http.StatusOK)
}

// AcceptQuote accepts a quote and converts it to a confirmed order
func (h *QuoteHandler) AcceptQuote(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := uuid.Parse(ps.ByName("id"))
	if err != nil {
		respondError(w, "Invalid quote ID", http.StatusBadRequest)
		return
	}

	order, err := h.quoteService.AcceptQuote(r.Context(), id)
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, order, http.StatusOK)
}

// GetPublicQuote retrieves a quote by its public access token
func (h *QuoteHandler) GetPublicQuote(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	token := ps.ByName("token")
	if token == "" {
		respondError(w, "Invalid token", http.StatusBadRequest)
		return
	}

	quote, err := h.quoteService.GetPublicQuote(r.Context(), token)
	if err != nil {
		respondError(w, err.Error(), http.StatusNotFound)
		return
	}

	respondJSON(w, quote, http.StatusOK)
}

// TrackQuoteView tracks when a quote is viewed via public URL
func (h *QuoteHandler) TrackQuoteView(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	token := ps.ByName("token")
	if token == "" {
		respondError(w, "Invalid token", http.StatusBadRequest)
		return
	}

	// Get quote by token to get ID
	quote, err := h.quoteService.GetPublicQuote(r.Context(), token)
	if err != nil {
		respondError(w, err.Error(), http.StatusNotFound)
		return
	}

	// Track the view
	if err := h.quoteService.TrackQuoteView(r.Context(), quote.ID); err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]string{
		"message": "View tracked successfully",
	}, http.StatusOK)
}

// GetQuoteAnalytics retrieves analytics for quotes
func (h *QuoteHandler) GetQuoteAnalytics(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Parse organization ID from context or query params
	organizationIDStr := r.URL.Query().Get("organization_id")
	if organizationIDStr == "" {
		respondError(w, "Organization ID is required", http.StatusBadRequest)
		return
	}

	organizationID, err := uuid.Parse(organizationIDStr)
	if err != nil {
		respondError(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}

	// Parse optional date filters
	var dateFrom, dateTo *time.Time
	if dateFromStr := r.URL.Query().Get("date_from"); dateFromStr != "" {
		if t, err := time.Parse(time.RFC3339, dateFromStr); err == nil {
			dateFrom = &t
		}
	}
	if dateToStr := r.URL.Query().Get("date_to"); dateToStr != "" {
		if t, err := time.Parse(time.RFC3339, dateToStr); err == nil {
			dateTo = &t
		}
	}

	analytics, err := h.quoteService.GetQuoteAnalytics(r.Context(), organizationID, dateFrom, dateTo)
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, analytics, http.StatusOK)
}

// Helper functions

func respondJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, message string, statusCode int) {
	respondJSON(w, map[string]string{"error": message}, statusCode)
}
