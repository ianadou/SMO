package match

import (
	"context"
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
)

func TestListMatchesByGroupUseCase_Execute_ReturnsAllMatchesInGroup(t *testing.T) {
	t.Parallel()

	repo := newFakeMatchRepository()
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		m, _ := entities.NewMatch(
			entities.MatchID([]string{"m-1", "m-2", "m-3"}[i]),
			"group-1",
			"Match",
			"Venue",
			time.Now().Add(time.Duration(i)*time.Hour),
			entities.MatchStatusDraft,
			time.Now(),
		)
		_ = repo.Save(ctx, m)
	}

	useCase := NewListMatchesByGroupUseCase(repo)

	matches, err := useCase.Execute(ctx, "group-1")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(matches) != 3 {
		t.Errorf("expected 3 matches, got %d", len(matches))
	}
}

func TestListMatchesByGroupUseCase_Execute_ReturnsEmpty_WhenNoMatches(t *testing.T) {
	t.Parallel()

	repo := newFakeMatchRepository()
	useCase := NewListMatchesByGroupUseCase(repo)

	matches, err := useCase.Execute(context.Background(), "empty-group")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(matches) != 0 {
		t.Errorf("expected 0 matches, got %d", len(matches))
	}
}

func TestListMatchesByGroupUseCase_Execute_OnlyReturnsMatchesFromRequestedGroup(t *testing.T) {
	t.Parallel()

	repo := newFakeMatchRepository()
	ctx := context.Background()

	m1, _ := entities.NewMatch("m-1", "group-A", "Match A", "V", time.Now().Add(time.Hour), entities.MatchStatusDraft, time.Now())
	m2, _ := entities.NewMatch("m-2", "group-B", "Match B", "V", time.Now().Add(time.Hour), entities.MatchStatusDraft, time.Now())
	_ = repo.Save(ctx, m1)
	_ = repo.Save(ctx, m2)

	useCase := NewListMatchesByGroupUseCase(repo)
	matches, _ := useCase.Execute(ctx, "group-A")

	if len(matches) != 1 {
		t.Errorf("expected 1 match for group-A, got %d", len(matches))
	}
	if len(matches) > 0 && matches[0].Title() != "Match A" {
		t.Errorf("expected match 'Match A', got %q", matches[0].Title())
	}
}
