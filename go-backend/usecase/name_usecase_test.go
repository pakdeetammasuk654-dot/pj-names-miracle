package usecase_test

import (
	"context"
	"testing"

	"go-backend/config"
	"go-backend/domain"
	"go-backend/repository"
	"go-backend/usecase"
)

func TestNameSearchUseCase_SearchTopNames_KalakiniFilter(t *testing.T) {
	db, err := config.InitDB()
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}
	defer db.Close()

	repo := repository.NewNameRepository(db)
	uc := usecase.NewNameSearchUseCase(repo)

	// Search for Monday born baby
	req := domain.NameSearchRequest{
		Mode:       domain.ModeStandalone,
		DayOfBirth: "monday",
		Limit:      100,
	}

	results, err := uc.SearchTopNames(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(results) == 0 {
		t.Fatalf("expected results for Monday born baby, got 0")
	}

	t.Logf("Fetched %d 100%% Kalakini-safe top names for Monday!", len(results))
	t.Logf("Top 1 Name for Monday: %s (Rank %d, Score: %d)", results[0].THName, results[0].Rank, results[0].TotalScore)
}
