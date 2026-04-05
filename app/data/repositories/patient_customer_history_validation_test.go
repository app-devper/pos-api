package repositories

import (
	"testing"

	"pos/app/domain/request"
)

func TestCreatePatientRejectsInvalidBranchId(t *testing.T) {
	entity := &patientEntity{}

	if _, err := entity.CreatePatient(request.Patient{BranchId: "invalid-branch-id"}); err == nil {
		t.Fatal("expected invalid branch id error")
	}
}

func TestGetPatientsRejectsInvalidBranchId(t *testing.T) {
	entity := &patientEntity{}

	if _, err := entity.GetPatients("invalid-branch-id"); err == nil {
		t.Fatal("expected invalid branch id error")
	}
}

func TestGetPatientByCustomerCodeRejectsInvalidBranchId(t *testing.T) {
	entity := &patientEntity{}

	if _, err := entity.GetPatientByCustomerCode("C001", "invalid-branch-id"); err == nil {
		t.Fatal("expected invalid branch id error")
	}
}

func TestCreateCustomerHistoryRejectsInvalidBranchId(t *testing.T) {
	entity := &customerHistoryEntity{}

	if _, err := entity.CreateCustomerHistory(request.CustomerHistory{BranchId: "invalid-branch-id"}); err == nil {
		t.Fatal("expected invalid branch id error")
	}
}

func TestGetCustomerHistoriesRejectsInvalidBranchId(t *testing.T) {
	entity := &customerHistoryEntity{}

	if _, err := entity.GetCustomerHistories("C001", "invalid-branch-id"); err == nil {
		t.Fatal("expected invalid branch id error")
	}
}
