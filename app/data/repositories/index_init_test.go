package repositories

import (
	"errors"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
)

type indexCreatorStub struct {
	createIndexFn func() (string, error)
}

func (s *indexCreatorStub) CreateIndex() (string, error) {
	return s.createIndexFn()
}

func TestEnsureCollectionIndexCallsCreateIndex(t *testing.T) {
	called := false
	entity := &indexCreatorStub{
		createIndexFn: func() (string, error) {
			called = true
			return "idx", nil
		},
	}

	ensureCollectionIndex(entity, "test")

	if !called {
		t.Fatal("expected CreateIndex to be called")
	}
}

func TestEnsureCollectionIndexIgnoresCreateIndexError(t *testing.T) {
	entity := &indexCreatorStub{
		createIndexFn: func() (string, error) {
			return "", errors.New("index failed")
		},
	}

	ensureCollectionIndex(entity, "test")
}

func TestCreateCollectionIndexHandlesNilRepo(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("expected no panic, got %v", r)
		}
	}()

	createCollectionIndex(nil, "test", mongo.IndexModel{})
}
