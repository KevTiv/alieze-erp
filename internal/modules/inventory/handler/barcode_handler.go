package handler

import (
	"encoding/json"
	"net/http"

	"github.com/KevTiv/alieze-erp/internal/modules/inventory/service"
	"github.com/KevTiv/alieze-erp/internal/modules/inventory/types"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

type BarcodeHandler struct {
	service *service.BarcodeService
}

func NewBarcodeHandler(service *service.BarcodeService) *BarcodeHandler {
	return &BarcodeHandler{
		service: service,
	}
}

func (h *BarcodeHandler) RegisterRoutes(router *httprouter.Router) {
	// Barcode scanning endpoints
	router.POST("/api/inventory/mobile/scan", h.ScanBarcode)
	router.GET("/api/inventory/mobile/scans/:scan_id", h.GetScanByID)
	router.GET("/api/inventory/mobile/scans", h.ListScans)

	// Mobile scanning session endpoints
	router.POST("/api/inventory/mobile/sessions", h.CreateScanningSession)
	router.GET("/api/inventory/mobile/sessions/:session_id", h.GetScanningSession)
	router.GET("/api/inventory/mobile/sessions", h.ListScanningSessions)
	router.POST("/api/inventory/mobile/sessions/:session_id/scans", h.AddScanToSession)
	router.POST("/api/inventory/mobile/sessions/:session_id/complete", h.CompleteScanningSession)
	router.GET("/api/inventory/mobile/sessions/:session_id/lines", h.GetSessionLines)

	// Barcode generation endpoints
	router.POST("/api/inventory/mobile/barcodes/generate", h.GenerateBarcode)
	router.POST("/api/inventory/mobile/barcodes/products/generate", h.GenerateBarcodesForProducts)

	// Barcode lookup endpoints
	router.GET("/api/inventory/mobile/barcodes/:barcode", h.FindEntityByBarcode)
	router.POST("/api/inventory/mobile/barcodes/validate", h.ValidateBarcodeFormat)

	// Barcode label endpoints
	router.POST("/api/inventory/mobile/labels", h.GenerateBarcodeLabel)
}

// ScanBarcode handles barcode scanning requests
func (h *BarcodeHandler) ScanBarcode(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Parse request
	var request types.BarcodeScanRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set organization ID
	request.OrganizationID = orgID

	// Perform scan
	response, err := h.service.ScanBarcode(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetScanByID retrieves a specific barcode scan
func (h *BarcodeHandler) GetScanByID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Get scan ID from URL
	scanID, err := uuid.Parse(ps.ByName("scan_id"))
	if err != nil {
		http.Error(w, "Invalid scan ID", http.StatusBadRequest)
		return
	}

	// Get scan
	scan, err := h.service.GetScanByID(r.Context(), orgID, scanID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if scan == nil {
		http.Error(w, "Scan not found", http.StatusNotFound)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(scan)
}

// ListScans retrieves barcode scans for an organization
func (h *BarcodeHandler) ListScans(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	limit := 100
	offset := 0

	if r.URL.Query().Has("limit") {
		fmt.Sscanf(r.URL.Query().Get("limit"), "%d", &limit)
	}
	if r.URL.Query().Has("offset") {
		fmt.Sscanf(r.URL.Query().Get("offset"), "%d", &offset)
	}

	// Get scans
	scans, err := h.service.ListScans(r.Context(), orgID, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(scans)
}

// CreateScanningSession creates a new mobile scanning session
func (h *BarcodeHandler) CreateScanningSession(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Parse request
	var request types.CreateScanningSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set organization ID
	request.OrganizationID = orgID

	// Create session
	session, err := h.service.CreateScanningSession(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(session)
}

// GetScanningSession retrieves a scanning session
func (h *BarcodeHandler) GetScanningSession(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Get session ID from URL
	sessionID, err := uuid.Parse(ps.ByName("session_id"))
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	// Get session
	session, err := h.service.GetScanningSession(r.Context(), orgID, sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if session == nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(session)
}

// ListScanningSessions retrieves scanning sessions for an organization
func (h *BarcodeHandler) ListScanningSessions(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	limit := 100
	offset := 0
	status := r.URL.Query().Get("status")

	if r.URL.Query().Has("limit") {
		fmt.Sscanf(r.URL.Query().Get("limit"), "%d", &limit)
	}
	if r.URL.Query().Has("offset") {
		fmt.Sscanf(r.URL.Query().Get("offset"), "%d", &offset)
	}

	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	// Get sessions
	sessions, err := h.service.ListScanningSessions(r.Context(), orgID, statusPtr, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sessions)
}

// AddScanToSession adds a scan to an existing session
func (h *BarcodeHandler) AddScanToSession(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Get session ID from URL
	sessionID, err := uuid.Parse(ps.ByName("session_id"))
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	// Parse request
	var request types.AddScanToSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set session and organization IDs
	request.SessionID = sessionID
	request.OrganizationID = orgID

	// Add scan to session
	response, err := h.service.AddScanToSession(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CompleteScanningSession marks a session as completed
func (h *BarcodeHandler) CompleteScanningSession(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Get session ID from URL
	sessionID, err := uuid.Parse(ps.ByName("session_id"))
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	// Create request
	request := types.CompleteSessionRequest{
		SessionID:    sessionID,
		OrganizationID: orgID,
	}

	// Complete session
	success, err := h.service.CompleteScanningSession(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if success {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Session completed successfully",
		})
	} else {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Session not found or already completed",
		})
	}
}

// GetSessionLines retrieves all lines for a scanning session
func (h *BarcodeHandler) GetSessionLines(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Get session ID from URL
	sessionID, err := uuid.Parse(ps.ByName("session_id"))
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	// Get session lines
	lines, err := h.service.GetSessionLines(r.Context(), orgID, sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(lines)
}

// GenerateBarcode generates a barcode for an entity
func (h *BarcodeHandler) GenerateBarcode(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Parse request
	var request types.BarcodeGenerationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set organization ID
	request.OrganizationID = orgID

	// Generate barcode
	response, err := h.service.GenerateBarcode(r.Context(), request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GenerateBarcodesForProducts generates barcodes for multiple products
func (h *BarcodeHandler) GenerateBarcodesForProducts(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Parse request
	var request struct {
		ProductIDs []uuid.UUID `json:"product_ids"`
		Prefix     *string     `json:"prefix,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Generate barcodes
	barcodes, err := h.service.GenerateBarcodesForProducts(r.Context(), orgID, request.ProductIDs, request.Prefix)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(barcodes)
}

// FindEntityByBarcode finds an entity by its barcode
func (h *BarcodeHandler) FindEntityByBarcode(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Get barcode from URL
	barcode := ps.ByName("barcode")
	if barcode == "" {
		http.Error(w, "Barcode is required", http.StatusBadRequest)
		return
	}

	// Find entity
	entity, err := h.service.FindEntityByBarcode(r.Context(), orgID, barcode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if entity == nil {
		http.Error(w, "Barcode not found", http.StatusNotFound)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(entity)
}

// ValidateBarcodeFormat validates a barcode format
func (h *BarcodeHandler) ValidateBarcodeFormat(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Parse request
	var request struct {
		Barcode string `json:"barcode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate format
	valid, err := h.service.ValidateBarcodeFormat(r.Context(), request.Barcode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":   valid,
		"barcode": request.Barcode,
	})
}

// GenerateBarcodeLabel generates a printable barcode label
func (h *BarcodeHandler) GenerateBarcodeLabel(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Get organization ID from context
	orgID, ok := r.Context().Value("organizationID").(uuid.UUID)
	if !ok {
		http.Error(w, "Organization ID not found in context", http.StatusUnauthorized)
		return
	}

	// Parse request
	var request struct {
		EntityType string `json:"entity_type"`
		EntityID   string `json:"entity_id"`
		Format     string `json:"format"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Generate label
	label, err := h.service.GenerateBarcodeLabel(r.Context(), orgID, request.EntityType, request.EntityID, request.Format)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"label":   label,
	})
}
