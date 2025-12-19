package service_test

import (
	"time"

	"github.com/google/uuid"
)

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func floatPtr(f float64) *float64 {
	return &f
}

func boolPtr(b bool) *bool {
	return &b
}

func uuidPtr(u uuid.UUID) *uuid.UUID {
	return &u
}

func timePtr(t time.Time) *time.Time {
	return &t
}
