package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
)

// --- minimal fakes ---------------------------------------------------------

type fakeOrganizerRepo struct {
	byID    map[entities.OrganizerID]*entities.Organizer
	byEmail map[string]*entities.Organizer
}

func newFakeOrganizerRepo() *fakeOrganizerRepo {
	return &fakeOrganizerRepo{
		byID:    make(map[entities.OrganizerID]*entities.Organizer),
		byEmail: make(map[string]*entities.Organizer),
	}
}

func (r *fakeOrganizerRepo) Save(_ context.Context, o *entities.Organizer) error {
	if _, exists := r.byEmail[o.Email()]; exists {
		return domainerrors.ErrEmailAlreadyExists
	}
	r.byID[o.ID()] = o
	r.byEmail[o.Email()] = o
	return nil
}

func (r *fakeOrganizerRepo) FindByID(_ context.Context, id entities.OrganizerID) (*entities.Organizer, error) {
	o, ok := r.byID[id]
	if !ok {
		return nil, domainerrors.ErrOrganizerNotFound
	}
	return o, nil
}

func (r *fakeOrganizerRepo) FindByEmail(_ context.Context, email string) (*entities.Organizer, error) {
	o, ok := r.byEmail[strings.ToLower(email)]
	if !ok {
		return nil, domainerrors.ErrOrganizerNotFound
	}
	return o, nil
}

// fakeHasher: hash is just "hashed:" + plain. Compare checks the prefix.
// Deterministic, enables exact assertions in tests.
type fakeHasher struct{}

func (fakeHasher) Hash(plain string) (string, error) { return "hashed:" + plain, nil }
func (fakeHasher) Compare(hash, plain string) error {
	if hash == "hashed:"+plain {
		return nil
	}
	return domainerrors.ErrInvalidCredentials
}

type fakeSigner struct {
	tokenForID map[entities.OrganizerID]string
}

func newFakeSigner() *fakeSigner {
	return &fakeSigner{tokenForID: make(map[entities.OrganizerID]string)}
}

func (s *fakeSigner) Sign(id entities.OrganizerID) (string, error) {
	token := "fake-token-for-" + string(id)
	s.tokenForID[id] = token
	return token, nil
}

func (s *fakeSigner) Verify(token string) (entities.OrganizerID, error) {
	for id, t := range s.tokenForID {
		if t == token {
			return id, nil
		}
	}
	return "", domainerrors.ErrInvalidToken
}

type fakeIDGen struct{ id string }

func (g *fakeIDGen) Generate() string { return g.id }

type fakeClock struct{ now time.Time }

func (c *fakeClock) Now() time.Time { return c.now }

// --- register tests --------------------------------------------------------

func TestRegisterOrganizerUseCase_HappyPath(t *testing.T) {
	t.Parallel()
	repo := newFakeOrganizerRepo()
	uc := NewRegisterOrganizerUseCase(repo, fakeHasher{}, &fakeIDGen{id: "org-1"},
		&fakeClock{now: time.Date(2026, 4, 27, 10, 0, 0, 0, time.UTC)})

	o, err := uc.Execute(context.Background(), RegisterOrganizerInput{
		Email: "alice@example.com", Password: "valid-password-12+", DisplayName: "Alice",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if o.ID() != "org-1" {
		t.Errorf("expected id 'org-1', got %q", o.ID())
	}
	if o.PasswordHash() != "hashed:valid-password-12+" {
		t.Errorf("expected hashed password, got %q", o.PasswordHash())
	}
}

func TestRegisterOrganizerUseCase_RejectsShortPassword(t *testing.T) {
	t.Parallel()
	repo := newFakeOrganizerRepo()
	uc := NewRegisterOrganizerUseCase(repo, fakeHasher{}, &fakeIDGen{id: "org-1"},
		&fakeClock{now: time.Now()})

	_, err := uc.Execute(context.Background(), RegisterOrganizerInput{
		Email: "alice@example.com", Password: "short", DisplayName: "Alice",
	})

	if !errors.Is(err, domainerrors.ErrInvalidPassword) {
		t.Errorf("expected ErrInvalidPassword, got %v", err)
	}
}

func TestRegisterOrganizerUseCase_RejectsDuplicateEmail(t *testing.T) {
	t.Parallel()
	repo := newFakeOrganizerRepo()
	clock := &fakeClock{now: time.Now()}
	first := NewRegisterOrganizerUseCase(repo, fakeHasher{}, &fakeIDGen{id: "org-1"}, clock)
	_, _ = first.Execute(context.Background(), RegisterOrganizerInput{
		Email: "alice@example.com", Password: "valid-password-12+", DisplayName: "Alice",
	})

	second := NewRegisterOrganizerUseCase(repo, fakeHasher{}, &fakeIDGen{id: "org-2"}, clock)
	_, err := second.Execute(context.Background(), RegisterOrganizerInput{
		Email: "alice@example.com", Password: "another-valid-password", DisplayName: "Bob",
	})

	if !errors.Is(err, domainerrors.ErrEmailAlreadyExists) {
		t.Errorf("expected ErrEmailAlreadyExists, got %v", err)
	}
}

// --- login tests -----------------------------------------------------------

func seedOrganizer(t *testing.T, repo *fakeOrganizerRepo) {
	t.Helper()
	hasher := fakeHasher{}
	hash, _ := hasher.Hash("correct-password-12+")
	o, err := entities.NewOrganizer("org-1", "alice@example.com", "Alice", hash, time.Now())
	if err != nil {
		t.Fatalf("seed: NewOrganizer: %v", err)
	}
	if saveErr := repo.Save(context.Background(), o); saveErr != nil {
		t.Fatalf("seed: Save: %v", saveErr)
	}
}

func TestLoginOrganizerUseCase_HappyPath_ReturnsTokenAndOrganizer(t *testing.T) {
	t.Parallel()
	repo := newFakeOrganizerRepo()
	signer := newFakeSigner()
	seedOrganizer(t, repo)
	uc := NewLoginOrganizerUseCase(repo, fakeHasher{}, signer)

	out, err := uc.Execute(context.Background(), LoginOrganizerInput{
		Email: "alice@example.com", Password: "correct-password-12+",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if out.Token != "fake-token-for-org-1" {
		t.Errorf("expected token 'fake-token-for-org-1', got %q", out.Token)
	}
	if out.Organizer.ID() != "org-1" {
		t.Errorf("expected organizer org-1, got %q", out.Organizer.ID())
	}
}

func TestLoginOrganizerUseCase_ReturnsErrInvalidCredentials_WhenPasswordIsWrong(t *testing.T) {
	t.Parallel()
	repo := newFakeOrganizerRepo()
	seedOrganizer(t, repo)
	uc := NewLoginOrganizerUseCase(repo, fakeHasher{}, newFakeSigner())

	_, err := uc.Execute(context.Background(), LoginOrganizerInput{
		Email: "alice@example.com", Password: "wrong-password",
	})

	if !errors.Is(err, domainerrors.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLoginOrganizerUseCase_ReturnsErrInvalidCredentials_WhenEmailDoesNotExist(t *testing.T) {
	t.Parallel()
	repo := newFakeOrganizerRepo()
	uc := NewLoginOrganizerUseCase(repo, fakeHasher{}, newFakeSigner())

	_, err := uc.Execute(context.Background(), LoginOrganizerInput{
		Email: "nonexistent@example.com", Password: "any-password",
	})

	// Critical security property: a missing email must produce the same
	// error as a wrong password, to prevent email enumeration.
	if !errors.Is(err, domainerrors.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials (no enumeration), got %v", err)
	}
}
