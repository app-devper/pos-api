package usecase

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"pos/app/data/entities"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type stockCountByIdStub struct {
	repositories.IStockCount
	gotBranch string
}

func (s *stockCountByIdStub) GetStockCountById(id string, branchId string) (*entities.StockCount, error) {
	s.gotBranch = branchId
	return nil, mongo.ErrNoDocuments
}

// A Count of another branch, with its system quantities, used to be returned
// to anyone who knew its id.
func TestGetStockCountByIdReadsOnlyTheCallersBranch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	branch := primitive.NewObjectID().Hex()
	repo := &stockCountByIdStub{}
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/stock-counts/x", nil)
	ctx.Params = gin.Params{{Key: "id", Value: primitive.NewObjectID().Hex()}}
	ctx.Set("BranchId", branch)

	GetStockCountById(repo)(ctx)

	if repo.gotBranch != branch {
		t.Fatalf("read with branch %q, want the caller's %q", repo.gotBranch, branch)
	}
	if w.Code != http.StatusNotFound {
		t.Fatalf("status %d, want 404", w.Code)
	}
}
