package main

import (
	"context"
	"fmt"
	"time"

	"alieze-erp/pkg/events"
)

// This example demonstrates how the event system works across modules

func main() {
	fmt.Println("=== Event Flow Test ===\n")

	// Create an event bus
	bus := events.NewBus(false) // Synchronous for testing

	// Simulate module event handlers
	bus.Subscribe("contact.created", func(ctx context.Context, event events.Event) error {
		fmt.Printf("[CRM Handler] Received %s event\n", event.Type)
		fmt.Printf("  Payload: %v\n", event.Payload)
		return nil
	})

	bus.Subscribe("contact.created", func(ctx context.Context, event events.Event) error {
		fmt.Printf("[Sales Handler] Received %s event - syncing customer data\n", event.Type)
		return nil
	})

	bus.Subscribe("order.confirmed", func(ctx context.Context, event events.Event) error {
		fmt.Printf("[Accounting Handler] Received %s event - creating invoice\n", event.Type)
		return nil
	})

	bus.Subscribe("order.confirmed", func(ctx context.Context, event events.Event) error {
		fmt.Printf("[CRM Handler] Received %s event - updating customer activity\n", event.Type)
		return nil
	})

	bus.Subscribe("invoice.paid", func(ctx context.Context, event events.Event) error {
		fmt.Printf("[Sales Handler] Received %s event - marking order as paid\n", event.Type)
		return nil
	})

	ctx := context.Background()

	// Simulate a business flow
	fmt.Println("1. Creating a contact...")
	contactData := map[string]interface{}{
		"id":   "12345",
		"name": "Acme Corporation",
		"email": "contact@acme.com",
	}
	bus.Publish(ctx, "contact.created", contactData)
	fmt.Println()

	time.Sleep(100 * time.Millisecond)

	fmt.Println("2. Confirming a sales order...")
	orderData := map[string]interface{}{
		"id":          "ORDER-001",
		"customer_id": "12345",
		"total":       1500.00,
	}
	bus.Publish(ctx, "order.confirmed", orderData)
	fmt.Println()

	time.Sleep(100 * time.Millisecond)

	fmt.Println("3. Recording invoice payment...")
	invoiceData := map[string]interface{}{
		"id":       "INV-001",
		"order_id": "ORDER-001",
		"amount":   1500.00,
	}
	bus.Publish(ctx, "invoice.paid", invoiceData)
	fmt.Println()

	fmt.Println("=== Event Flow Complete ===")
	fmt.Println("\nThis demonstrates how modules communicate through events:")
	fmt.Println("- CRM publishes contact.created → Sales syncs customer data")
	fmt.Println("- Sales publishes order.confirmed → Accounting creates invoice & CRM updates activity")
	fmt.Println("- Accounting publishes invoice.paid → Sales marks order as paid")
}
