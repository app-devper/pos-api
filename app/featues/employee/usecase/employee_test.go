package usecase

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pos/app/data/entities"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type employeeRepoStub struct {
	repositories.IEmployee
	createFn      func(form request.Employee) (*entities.Employee, error)
	getAllFn      func() ([]entities.Employee, error)
	getByBranchFn func(branchId string) ([]entities.Employee, error)
	getByIDFn     func(id string) (*entities.Employee, error)
	updateByIDFn  func(id string, form request.UpdateEmployee) (*entities.Employee, error)
	removeByIDFn  func(id string) (*entities.Employee, error)
}

func (s *employeeRepoStub) CreateEmployee(form request.Employee) (*entities.Employee, error) {
	return s.createFn(form)
}

func (s *employeeRepoStub) GetEmployees() ([]entities.Employee, error) {
	return s.getAllFn()
}

func (s *employeeRepoStub) GetEmployeesByBranchId(branchId string) ([]entities.Employee, error) {
	return s.getByBranchFn(branchId)
}

func (s *employeeRepoStub) GetEmployeeById(id string) (*entities.Employee, error) {
	return s.getByIDFn(id)
}

func (s *employeeRepoStub) UpdateEmployeeById(id string, form request.UpdateEmployee) (*entities.Employee, error) {
	return s.updateByIDFn(id, form)
}

func (s *employeeRepoStub) RemoveEmployeeById(id string) (*entities.Employee, error) {
	return s.removeByIDFn(id)
}

func TestCreateEmployeePassesCreatedBy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var gotForm request.Employee
	repo := &employeeRepoStub{
		createFn: func(form request.Employee) (*entities.Employee, error) {
			gotForm = form
			return &entities.Employee{UserId: form.UserId, Role: form.Role}, nil
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/employees", strings.NewReader(`{"branchId":"`+primitive.NewObjectID().Hex()+`","userId":"user-2","role":"STAFF"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("UserId", "admin-1")

	CreateEmployee(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotForm.CreatedBy != "admin-1" {
		t.Fatalf("expected CreatedBy admin-1, got %s", gotForm.CreatedBy)
	}
}

func TestGetEmployeesByBranchIdPassesParam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	branchID := primitive.NewObjectID().Hex()
	var gotBranchID string
	repo := &employeeRepoStub{
		getByBranchFn: func(branchId string) ([]entities.Employee, error) {
			gotBranchID = branchId
			return []entities.Employee{}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/employees/branch/"+branchID, nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "branchId", Value: branchID}}

	GetEmployeesByBranchId(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotBranchID != branchID {
		t.Fatalf("expected branch id %s, got %s", branchID, gotBranchID)
	}
}

func TestUpdateEmployeeByIdPassesUpdatedBy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	employeeID := primitive.NewObjectID().Hex()
	var gotID string
	var gotForm request.UpdateEmployee
	repo := &employeeRepoStub{
		updateByIDFn: func(id string, form request.UpdateEmployee) (*entities.Employee, error) {
			gotID = id
			gotForm = form
			return &entities.Employee{Id: primitive.NewObjectID()}, nil
		},
	}

	req := httptest.NewRequest(http.MethodPut, "/employees/"+employeeID, strings.NewReader(`{"branchId":"`+primitive.NewObjectID().Hex()+`","role":"ADMIN"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "employeeId", Value: employeeID}}
	ctx.Set("UserId", "admin-1")

	UpdateEmployeeById(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotID != employeeID {
		t.Fatalf("expected employee id %s, got %s", employeeID, gotID)
	}
	if gotForm.UpdatedBy != "admin-1" {
		t.Fatalf("expected UpdatedBy admin-1, got %s", gotForm.UpdatedBy)
	}
}

func TestGetEmployeeByIdPassesParam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	employeeID := primitive.NewObjectID().Hex()
	var gotID string
	repo := &employeeRepoStub{
		getByIDFn: func(id string) (*entities.Employee, error) {
			gotID = id
			return &entities.Employee{Id: primitive.NewObjectID()}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/employees/"+employeeID, nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "employeeId", Value: employeeID}}

	GetEmployeeById(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotID != employeeID {
		t.Fatalf("expected employee id %s, got %s", employeeID, gotID)
	}
}

func TestDeleteEmployeeByIdPassesParam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	employeeID := primitive.NewObjectID().Hex()
	var gotID string
	repo := &employeeRepoStub{
		removeByIDFn: func(id string) (*entities.Employee, error) {
			gotID = id
			return &entities.Employee{Id: primitive.NewObjectID()}, nil
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/employees/"+employeeID, nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "employeeId", Value: employeeID}}

	DeleteEmployeeById(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotID != employeeID {
		t.Fatalf("expected employee id %s, got %s", employeeID, gotID)
	}
}

func TestGetEmployeesCallsRepository(t *testing.T) {
	gin.SetMode(gin.TestMode)

	called := false
	repo := &employeeRepoStub{
		getAllFn: func() ([]entities.Employee, error) {
			called = true
			return []entities.Employee{}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/employees", nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req

	GetEmployees(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if !called {
		t.Fatal("expected repository GetEmployees to be called")
	}
}
