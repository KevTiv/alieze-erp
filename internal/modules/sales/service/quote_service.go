package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/KevTiv/alieze-erp/internal/modules/sales/repository"
	"github.com/KevTiv/alieze-erp/internal/modules/sales/types"
	"github.com/KevTiv/alieze-erp/pkg/email"
	"github.com/KevTiv/alieze-erp/pkg/events"
	"github.com/KevTiv/alieze-erp/pkg/queue"
	"github.com/KevTiv/alieze-erp/pkg/storage"
	"github.com/KevTiv/alieze-erp/pkg/templates"

	"github.com/google/uuid"
)

// QuoteService handles quote/proposal generation and management
type QuoteService struct {
	salesOrderService *SalesOrderService
	salesOrderRepo    repository.SalesOrderRepository
	storage           storage.Storage
	templateEngine    *templates.Engine
	pdfGenerator      *templates.PDFGenerator
	emailService      email.EmailService
	jobQueue          queue.Queue
	eventBus          *events.Bus
	config            QuoteConfig
}

// QuoteConfig contains configuration for quote operations
type QuoteConfig struct {
	DefaultTemplate     string
	DefaultExpiryDays   int
	PublicURLBaseURL    string
	SMTPFrom            string
	EnableEmailTracking bool
}

var defaultConfig = QuoteConfig{
	DefaultTemplate:     "quotes/standard",
	DefaultExpiryDays:   30,
	PublicURLBaseURL:    "https://your-domain.com", // Override with env var
	SMTPFrom:            "noreply@your-domain.com",
	EnableEmailTracking: false,
}

// QuoteEmailRequest contains data for sending quotes via email
type QuoteEmailRequest struct {
	QuoteID        uuid.UUID
	RecipientName  string
	RecipientEmail string
	Subject        string
	Message        string
	AttachPDF      bool
}

// QuotePDFData contains all data needed for quote PDF generation
type QuotePDFData struct {
	Quote           *types.SalesOrder
	Customer        *CustomerInfo
	Company         *CompanyInfo
	Lines           []types.SalesOrderLine
	ValidUntil      string
	QuoteNumber     string
	IssuedDate      string
	TotalUntaxed    string
	TotalTax        string
	TotalAmount     string
	Notes           string
	TermsConditions string
}

type CustomerInfo struct {
	Name    string
	Email   string
	Phone   string
	Address string
	City    string
	Country string
}

type CompanyInfo struct {
	Name    string
	Logo    string
	Email   string
	Phone   string
	Address string
	Website string
	TaxID   string
}

func NewQuoteService(
	salesOrderService *SalesOrderService,
	salesOrderRepo repository.SalesOrderRepository,
	storage storage.Storage,
	templateEngine *templates.Engine,
	pdfGenerator *templates.PDFGenerator,
	emailService email.EmailService,
	jobQueue queue.Queue,
	config QuoteConfig,
) *QuoteService {
	return &QuoteService{
		salesOrderService: salesOrderService,
		salesOrderRepo:    salesOrderRepo,
		storage:           storage,
		templateEngine:    templateEngine,
		pdfGenerator:      pdfGenerator,
		emailService:      emailService,
		jobQueue:          jobQueue,
		config:            config,
	}
}

// NewQuoteServiceWithEventBus creates a quote service with event bus support
func NewQuoteServiceWithEventBus(
	salesOrderService *SalesOrderService,
	salesOrderRepo repository.SalesOrderRepository,
	storage storage.Storage,
	templateEngine *templates.Engine,
	pdfGenerator *templates.PDFGenerator,
	emailService email.EmailService,
	jobQueue queue.Queue,
	config QuoteConfig,
	eventBus *events.Bus,
) *QuoteService {
	service := NewQuoteService(salesOrderService, salesOrderRepo, storage, templateEngine, pdfGenerator, emailService, jobQueue, config)
	service.eventBus = eventBus
	return service
}

// CreateQuote creates a new quote (sales order with quotation status)
func (s *QuoteService) CreateQuote(ctx context.Context, order types.SalesOrder) (*types.SalesOrder, error) {
	// Force status to quotation
	order.Status = types.SalesOrderStatusQuotation

	// Set validity date if not provided (use config or default)
	if order.ValidityDate == nil {
		expiryDays := s.config.DefaultExpiryDays
		if expiryDays == 0 {
			expiryDays = defaultConfig.DefaultExpiryDays
		}
		validUntil := time.Now().AddDate(0, 0, expiryDays)
		order.ValidityDate = &validUntil
	}

	// Create the sales order
	createdOrder, err := s.salesOrderService.CreateSalesOrder(ctx, order)
	if err != nil {
		return nil, fmt.Errorf("failed to create quote: %w", err)
	}

	// Generate access token for public viewing
	token, err := s.generateAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Update with access token using SQL function
	err = s.salesOrderRepo.ExecuteSQL(ctx, "SELECT generate_quote_access_token($1, $2)", createdOrder.ID, token)
	if err != nil {
		return nil, fmt.Errorf("failed to set access token: %w", err)
	}

	// Publish event
	s.publishEvent(ctx, "quote.created", createdOrder)

	return createdOrder, nil
}

// GenerateQuotePDF generates a PDF for a quote
func (s *QuoteService) GenerateQuotePDF(ctx context.Context, quoteID uuid.UUID, template string) ([]byte, error) {
	// Get the quote
	order, err := s.salesOrderService.GetSalesOrder(ctx, quoteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}
	if order == nil {
		return nil, fmt.Errorf("quote not found")
	}

	// Validate it's a quote
	if order.Status != types.SalesOrderStatusQuotation {
		return nil, fmt.Errorf("order is not a quote")
	}

	// Get customer and company info (placeholder - would fetch from repositories)
	customer := &CustomerInfo{
		Name:    "Customer Name", // TODO: Fetch from customer repository
		Email:   "customer@example.com",
		Address: "Customer Address",
	}

	company := &CompanyInfo{
		Name:    "Your Company",
		Email:   "contact@yourcompany.com",
		Phone:   "+1-555-0100",
		Address: "123 Business St",
		Website: "www.yourcompany.com",
	}

	// Prepare PDF data
	pdfData := QuotePDFData{
		Quote:           order,
		Customer:        customer,
		Company:         company,
		Lines:           order.Lines,
		ValidUntil:      formatDate(order.ValidityDate),
		QuoteNumber:     order.Reference,
		IssuedDate:      formatDate(&order.OrderDate),
		TotalUntaxed:    formatCurrency(order.AmountUntaxed),
		TotalTax:        formatCurrency(order.AmountTax),
		TotalAmount:     formatCurrency(order.AmountTotal),
		Notes:           order.Note,
		TermsConditions: "Standard terms and conditions apply",
	}

	// Use template if not specified
	if template == "" {
		template = "quotes/standard"
	}

	// Generate PDF
	pdfBytes, err := s.pdfGenerator.RenderPDF(template, pdfData, &templates.PDFOptions{
		PageSize:     "A4",
		Orientation:  "portrait",
		MarginTop:    "10mm",
		MarginRight:  "10mm",
		MarginBottom: "10mm",
		MarginLeft:   "10mm",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return pdfBytes, nil
}

// SaveQuotePDF generates and saves a quote PDF to storage
func (s *QuoteService) SaveQuotePDF(ctx context.Context, quoteID uuid.UUID, template string) (string, error) {
	// Generate PDF
	pdfBytes, err := s.GenerateQuotePDF(ctx, quoteID, template)
	if err != nil {
		return "", err
	}

	// Prepare storage key
	storageKey := fmt.Sprintf("quotes/%s/%s.pdf", quoteID.String(), time.Now().Format("20060102-150405"))

	// Upload to storage
	metadata, err := s.storage.Upload(ctx, storage.UploadOptions{
		Key:         storageKey,
		Data:        pdfBytes,
		ContentType: "application/pdf",
		Metadata: map[string]string{
			"quote_id":  quoteID.String(),
			"generated": time.Now().Format(time.RFC3339),
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload PDF: %w", err)
	}

	return metadata.Key, nil
}

// SendQuoteByEmail sends a quote via email
func (s *QuoteService) SendQuoteByEmail(ctx context.Context, request QuoteEmailRequest) error {
	// Get the quote
	order, err := s.salesOrderService.GetSalesOrder(ctx, request.QuoteID)
	if err != nil {
		return fmt.Errorf("failed to get quote: %w", err)
	}
	if order == nil {
		return fmt.Errorf("quote not found")
	}

	// Validate it's a quote
	if order.Status != types.SalesOrderStatusQuotation {
		return fmt.Errorf("order is not a quote")
	}

	// Generate PDF if attachment requested
	var attachmentData []byte
	if request.AttachPDF {
		pdfBytes, err := s.GenerateQuotePDF(ctx, request.QuoteID, "")
		if err != nil {
			return fmt.Errorf("failed to generate PDF attachment: %w", err)
		}
		attachmentData = pdfBytes
	}

	// Prepare email data
	emailData := map[string]interface{}{
		"RecipientName": request.RecipientName,
		"QuoteNumber":   order.Reference,
		"Amount":        formatCurrency(order.AmountTotal),
		"ValidUntil":    formatDate(order.ValidityDate),
		"ViewURL":       s.getPublicQuoteURL(order.ID),
		"Message":       request.Message,
	}

	// Render email body from template
	emailBody, err := s.templateEngine.RenderString(ctx, "emails/quote_sent", emailData)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	// Send email
	emailReq := email.SendEmailRequest{
		To:      []string{request.RecipientEmail},
		Subject: request.Subject,
		Body:    emailBody,
		IsHTML:  true,
	}

	if request.AttachPDF {
		emailReq.Attachments = []email.Attachment{
			{
				Filename:    fmt.Sprintf("quote-%s.pdf", order.Reference),
				ContentType: "application/pdf",
				Data:        attachmentData,
			},
		}
	}

	err = s.emailService.Send(ctx, emailReq)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	// Mark quote as sent using SQL function
	err = s.salesOrderRepo.ExecuteSQL(ctx, "SELECT mark_quote_sent($1)", order.ID)
	if err != nil {
		return fmt.Errorf("failed to update quote sent status: %w", err)
	}

	// Publish event
	s.publishEvent(ctx, "quote.sent", map[string]interface{}{
		"quote_id":  order.ID,
		"recipient": request.RecipientEmail,
		"sent_at":   time.Now(),
	})

	return nil
}

// SendQuoteByEmailAsync queues a quote email to be sent in background
func (s *QuoteService) SendQuoteByEmailAsync(ctx context.Context, request QuoteEmailRequest) error {
	// Enqueue job
	job := queue.Job{
		Type: "quote:send_email",
		Payload: map[string]interface{}{
			"quote_id":        request.QuoteID.String(),
			"recipient_name":  request.RecipientName,
			"recipient_email": request.RecipientEmail,
			"subject":         request.Subject,
			"message":         request.Message,
			"attach_pdf":      request.AttachPDF,
		},
		Priority:   1,
		MaxRetries: 3,
	}

	err := s.jobQueue.Enqueue(ctx, job)
	if err != nil {
		return fmt.Errorf("failed to enqueue quote email job: %w", err)
	}

	return nil
}

// GetPublicQuote retrieves a quote by its public access token
func (s *QuoteService) GetPublicQuote(ctx context.Context, token string) (*types.SalesOrder, error) {
	// Query by access token
	var quote *types.SalesOrder
	err := s.salesOrderRepo.QueryRow(ctx,
		"SELECT * FROM sales_orders WHERE quote_access_token = $1 AND status = 'quotation'",
		&quote,
		token,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote by token: %w", err)
	}
	if quote == nil {
		return nil, fmt.Errorf("quote not found or invalid token")
	}

	// Check if expired
	if quote.ValidityDate != nil && quote.ValidityDate.Before(time.Now()) {
		return nil, fmt.Errorf("quote has expired")
	}

	return quote, nil
}

// TrackQuoteView tracks when a quote is viewed
func (s *QuoteService) TrackQuoteView(ctx context.Context, quoteID uuid.UUID) error {
	// Update viewed timestamp using SQL function
	err := s.salesOrderRepo.ExecuteSQL(ctx, "SELECT track_quote_view($1)", quoteID)
	if err != nil {
		return fmt.Errorf("failed to track quote view: %w", err)
	}

	// Publish event
	s.publishEvent(ctx, "quote.viewed", map[string]interface{}{
		"quote_id":  quoteID,
		"viewed_at": time.Now(),
	})

	return nil
}

// AcceptQuote converts a quote to a confirmed order
func (s *QuoteService) AcceptQuote(ctx context.Context, quoteID uuid.UUID) (*types.SalesOrder, error) {
	// Use SQL function to accept quote (handles expiry validation)
	var result bool
	err := s.salesOrderRepo.QueryRow(ctx, "SELECT accept_quote($1)", &result, quoteID)
	if err != nil {
		return nil, fmt.Errorf("failed to accept quote: %w", err)
	}

	if !result {
		return nil, fmt.Errorf("quote cannot be accepted (may be expired or invalid)")
	}

	// Get updated order
	order, err := s.salesOrderService.GetSalesOrder(ctx, quoteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get accepted quote: %w", err)
	}

	// Publish event
	s.publishEvent(ctx, "quote.accepted", order)

	return order, nil
}

// CheckQuoteExpiry checks if a quote has expired
func (s *QuoteService) CheckQuoteExpiry(ctx context.Context, quoteID uuid.UUID) (bool, error) {
	var isExpired bool
	err := s.salesOrderRepo.QueryRow(ctx, "SELECT check_quote_expiry($1)", &isExpired, quoteID)
	if err != nil {
		return false, fmt.Errorf("failed to check quote expiry: %w", err)
	}

	return isExpired, nil
}

// GetQuoteAnalytics retrieves quote analytics from the view
func (s *QuoteService) GetQuoteAnalytics(ctx context.Context, organizationID uuid.UUID, dateFrom, dateTo *time.Time) (interface{}, error) {
	query := `
		SELECT
			total_quotes,
			quotes_sent,
			quotes_viewed,
			quotes_accepted,
			view_rate,
			conversion_rate,
			avg_accept_hours
		FROM quote_analytics
		WHERE organization_id = $1
	`

	var analytics struct {
		TotalQuotes    int64   `db:"total_quotes"`
		QuotesSent     int64   `db:"quotes_sent"`
		QuotesViewed   int64   `db:"quotes_viewed"`
		QuotesAccepted int64   `db:"quotes_accepted"`
		ViewRate       float64 `db:"view_rate"`
		ConversionRate float64 `db:"conversion_rate"`
		AvgAcceptHours float64 `db:"avg_accept_hours"`
	}

	err := s.salesOrderRepo.QueryRow(ctx, query, &analytics, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote analytics: %w", err)
	}

	return analytics, nil
}

// generateAccessToken generates a secure random token for public access
func (s *QuoteService) generateAccessToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// getPublicQuoteURL generates the public URL for a quote
func (s *QuoteService) getPublicQuoteURL(quoteID uuid.UUID) string {
	baseURL := s.config.PublicURLBaseURL
	if baseURL == "" {
		baseURL = defaultConfig.PublicURLBaseURL
	}
	return fmt.Sprintf("%s/quotes/view/%s", baseURL, quoteID.String())
}

// publishEvent publishes an event to the event bus if available
func (s *QuoteService) publishEvent(ctx context.Context, eventType string, payload interface{}) {
	if s.eventBus != nil {
		if err := s.eventBus.Publish(ctx, eventType, payload); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Failed to publish event %s: %v\n", eventType, err)
		}
	}
}

// Helper functions
func formatDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("January 2, 2006")
}

func formatCurrency(amount float64) string {
	return fmt.Sprintf("$%.2f", amount)
}
