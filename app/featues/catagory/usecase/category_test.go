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

type categoryRepoStub struct {
	repositories.ICategory
	createFn        func(form request.Category) (*entities.Category, error)
	getAllFn        func() ([]entities.Category, error)
	getByIDFn       func(id string) (*entities.Category, error)
	updateByIDFn    func(id string, form request.Category) (*entities.Category, error)
	updateDefaultFn func(id string) (*entities.Category, error)
	removeByIDFn    func(id string, updatedBy string) (*entities.Category, error)
}

func (s *categoryRepoStub) CreateCategory(form request.Category) (*entities.Category, error) {
	return s.createFn(form)
}

func (s *categoryRepoStub) GetCategoryAll() ([]entities.Category, error) {
	return s.getAllFn()
}

func (s *categoryRepoStub) GetCategoryById(id string) (*entities.Category, error) {
	return s.getByIDFn(id)
}

func (s *categoryRepoStub) UpdateCategoryById(id string, form request.Category) (*entities.Category, error) {
	return s.updateByIDFn(id, form)
}

func (s *categoryRepoStub) UpdateDefaultCategoryById(id string) (*entities.Category, error) {
	return s.updateDefaultFn(id)
}

func (s *categoryRepoStub) RemoveCategoryById(id string, updatedBy string) (*entities.Category, error) {
	return s.removeByIDFn(id, updatedBy)
}

func TestCreateCategoryPassesPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var gotForm request.Category
	repo := &categoryRepoStub{
		createFn: func(form request.Category) (*entities.Category, error) {
			gotForm = form
			return &entities.Category{Name: form.Name, Value: form.Value}, nil
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(`{"name":"ยา","value":"drug"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req

	CreateCategory(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotForm.Name != "ยา" || gotForm.Value != "drug" {
		t.Fatalf("unexpected form passed to repository: %+v", gotForm)
	}
}

func TestGetCategoriesCallsRepository(t *testing.T) {
	gin.SetMode(gin.TestMode)

	called := false
	repo := &categoryRepoStub{
		getAllFn: func() ([]entities.Category, error) {
			called = true
			return []entities.Category{}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/categories", nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req

	GetCategories(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if !called {
		t.Fatal("expected repository GetCategoryAll to be called")
	}
}

func TestGetCategoryByIdPassesParam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	categoryID := primitive.NewObjectID().Hex()
	var gotID string
	repo := &categoryRepoStub{
		getByIDFn: func(id string) (*entities.Category, error) {
			gotID = id
			return &entities.Category{Id: primitive.NewObjectID()}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/categories/"+categoryID, nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "categoryId", Value: categoryID}}

	GetCategoryById(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotID != categoryID {
		t.Fatalf("expected category id %s, got %s", categoryID, gotID)
	}
}

func TestUpdateCategoryByIdPassesParamAndPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	categoryID := primitive.NewObjectID().Hex()
	var gotID string
	var gotForm request.Category
	repo := &categoryRepoStub{
		updateByIDFn: func(id string, form request.Category) (*entities.Category, error) {
			gotID = id
			gotForm = form
			return &entities.Category{Id: primitive.NewObjectID()}, nil
		},
	}

	req := httptest.NewRequest(http.MethodPut, "/categories/"+categoryID, strings.NewReader(`{"name":"ยาใหม่","value":"new-drug"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "categoryId", Value: categoryID}}

	UpdateCategoryById(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotID != categoryID {
		t.Fatalf("expected category id %s, got %s", categoryID, gotID)
	}
	if gotForm.Name != "ยาใหม่" || gotForm.Value != "new-drug" {
		t.Fatalf("unexpected form passed to repository: %+v", gotForm)
	}
}

func TestUpdateDefaultCategoryByIdPassesParam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	categoryID := primitive.NewObjectID().Hex()
	var gotID string
	repo := &categoryRepoStub{
		updateDefaultFn: func(id string) (*entities.Category, error) {
			gotID = id
			return &entities.Category{Id: primitive.NewObjectID()}, nil
		},
	}

	req := httptest.NewRequest(http.MethodPatch, "/categories/"+categoryID+"/default", nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "categoryId", Value: categoryID}}

	UpdateDefaultCategoryById(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotID != categoryID {
		t.Fatalf("expected category id %s, got %s", categoryID, gotID)
	}
}

func TestDeleteCategoryByIdPassesParam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	categoryID := primitive.NewObjectID().Hex()
	var gotID string
	var gotUpdatedBy string
	repo := &categoryRepoStub{
		removeByIDFn: func(id string, updatedBy string) (*entities.Category, error) {
			gotID = id
			gotUpdatedBy = updatedBy
			return &entities.Category{Id: primitive.NewObjectID()}, nil
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/categories/"+categoryID, nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "categoryId", Value: categoryID}}
	ctx.Set("UserId", "admin-1")

	DeleteCategoryById(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotID != categoryID {
		t.Fatalf("expected category id %s, got %s", categoryID, gotID)
	}
	if gotUpdatedBy != "admin-1" {
		t.Fatalf("expected UpdatedBy admin-1, got %s", gotUpdatedBy)
	}
}
