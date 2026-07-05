package usecase

import "testing"

func TestIsValidCustomerStatus(t *testing.T) {
	validStatuses := []string{"DRAFT", "ACTIVE", "INACTIVE", "ARCHIVED"}
	for _, status := range validStatuses {
		if !isValidCustomerStatus(status) {
			t.Fatalf("expected status %s to be valid", status)
		}
	}

	invalidStatuses := []string{"", "DELETED", "PENDING"}
	for _, status := range invalidStatuses {
		if isValidCustomerStatus(status) {
			t.Fatalf("expected status %s to be invalid", status)
		}
	}
}
