package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/infrastructure/persistence/postgres/generated"
	"github.com/ianadou/smo/infrastructure/persistence/postgres/mappers"
)

const invitationRepoOp = "postgres invitation repository"

// PostgresInvitationRepository is the PostgreSQL implementation of
// InvitationRepository.
//
// It holds the pool (not just a generated.DBTX) because
// RespondWithCapacityGuard runs inside an explicit transaction:
// serializing on the match row requires BEGIN/COMMIT, which the
// sqlc DBTX interface does not expose.
type PostgresInvitationRepository struct {
	pool    *pgxpool.Pool
	queries *generated.Queries
}

// NewPostgresInvitationRepository builds the repository.
func NewPostgresInvitationRepository(pool *pgxpool.Pool) *PostgresInvitationRepository {
	return &PostgresInvitationRepository{pool: pool, queries: generated.New(pool)}
}

// Save persists a new invitation. FK violations on match_id become
// ErrReferencedEntityNotFound.
func (r *PostgresInvitationRepository) Save(ctx context.Context, inv *entities.Invitation) error {
	params := mappers.InvitationToCreateParams(inv)
	if _, err := r.queries.CreateInvitation(ctx, params); err != nil {
		if isInvitationForeignKeyViolation(err) {
			return fmt.Errorf(invitationRepoOp+": save %q: %w",
				inv.ID(), domainerrors.ErrReferencedEntityNotFound)
		}
		return fmt.Errorf(invitationRepoOp+": save %q: %w", inv.ID(), err)
	}
	return nil
}

// FindByID looks up an invitation by identifier.
func (r *PostgresInvitationRepository) FindByID(ctx context.Context, id entities.InvitationID) (*entities.Invitation, error) {
	row, err := r.queries.GetInvitationByID(ctx, string(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf(invitationRepoOp+": find %q: %w",
				id, domainerrors.ErrInvitationNotFound)
		}
		return nil, fmt.Errorf(invitationRepoOp+": find %q: %w", id, err)
	}
	inv, mapErr := mappers.InvitationToDomain(row)
	if mapErr != nil {
		return nil, fmt.Errorf(invitationRepoOp+": map %q: %w", id, mapErr)
	}
	return inv, nil
}

// FindByTokenHash looks up an invitation by its token hash. Used by
// RespondToInvitationUseCase to find an invitation from a submitted
// token.
func (r *PostgresInvitationRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*entities.Invitation, error) {
	row, err := r.queries.GetInvitationByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf(invitationRepoOp+": find by hash: %w",
				domainerrors.ErrInvitationNotFound)
		}
		return nil, fmt.Errorf(invitationRepoOp+": find by hash: %w", err)
	}
	inv, mapErr := mappers.InvitationToDomain(row)
	if mapErr != nil {
		return nil, fmt.Errorf(invitationRepoOp+": map by hash: %w", mapErr)
	}
	return inv, nil
}

// ListByMatch returns all invitations for the given match, newest first.
func (r *PostgresInvitationRepository) ListByMatch(ctx context.Context, matchID entities.MatchID) ([]*entities.Invitation, error) {
	rows, err := r.queries.ListInvitationsByMatchID(ctx, string(matchID))
	if err != nil {
		return nil, fmt.Errorf(invitationRepoOp+": list by match %q: %w", matchID, err)
	}
	invitations := make([]*entities.Invitation, 0, len(rows))
	for _, row := range rows {
		inv, mapErr := mappers.InvitationToDomain(row)
		if mapErr != nil {
			return nil, fmt.Errorf(invitationRepoOp+": map %q: %w", row.ID, mapErr)
		}
		invitations = append(invitations, inv)
	}
	return invitations, nil
}

// CountConfirmedByMatch returns the count of confirmed (response = 'yes')
// invitations for the given match.
func (r *PostgresInvitationRepository) CountConfirmedByMatch(ctx context.Context, matchID entities.MatchID) (int, error) {
	count, err := r.queries.CountConfirmedInvitationsByMatchID(ctx, string(matchID))
	if err != nil {
		return 0, fmt.Errorf(invitationRepoOp+": count confirmed for %q: %w", matchID, err)
	}
	return int(count), nil
}

// ListConfirmedParticipants returns the player + confirmation timestamp
// for every confirmed invitation (response = 'yes') of the given match,
// ordered by responded_at asc.
func (r *PostgresInvitationRepository) ListConfirmedParticipants(ctx context.Context, matchID entities.MatchID) ([]entities.MatchParticipant, error) {
	rows, err := r.queries.ListConfirmedParticipantsByMatchID(ctx, string(matchID))
	if err != nil {
		return nil, fmt.Errorf(invitationRepoOp+": list confirmed for %q: %w", matchID, err)
	}
	participants := make([]entities.MatchParticipant, 0, len(rows))
	for _, row := range rows {
		participants = append(participants, entities.MatchParticipant{
			PlayerID:    entities.PlayerID(row.PlayerID),
			PlayerName:  row.PlayerName,
			ConfirmedAt: row.RespondedAt.Time,
		})
	}
	return participants, nil
}

// RespondWithCapacityGuard persists the invitation's new response inside
// a transaction. When the response is "yes", it serializes on the match
// row (SELECT ... FOR UPDATE) and re-counts confirmed invitations before
// committing, so two concurrent confirmations cannot both slip past the
// MaxParticipantsPerMatch cap. Changing to "no" skips the capacity
// check entirely (it only frees a slot).
//
// The capacity check is also skipped when the invitation was already
// "yes" before this call: re-confirming an existing participant must not
// be rejected just because the match is at capacity.
func (r *PostgresInvitationRepository) RespondWithCapacityGuard(ctx context.Context, inv *entities.Invitation, maxConfirmed int) error {
	previouslyConfirmed, err := r.previousResponseWasYes(ctx, inv.ID())
	if err != nil {
		return err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf(invitationRepoOp+": begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := r.queries.WithTx(tx)

	if inv.IsConfirmed() && !previouslyConfirmed {
		if guardErr := enforceCapacity(ctx, qtx, inv.MatchID(), maxConfirmed); guardErr != nil {
			return guardErr
		}
	}

	if _, updErr := qtx.UpdateInvitationResponse(ctx, mappers.InvitationToUpdateResponseParams(inv)); updErr != nil {
		if errors.Is(updErr, pgx.ErrNoRows) {
			return fmt.Errorf(invitationRepoOp+": respond %q: %w",
				inv.ID(), domainerrors.ErrInvitationNotFound)
		}
		return fmt.Errorf(invitationRepoOp+": respond %q: %w", inv.ID(), updErr)
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return fmt.Errorf(invitationRepoOp+": commit: %w", commitErr)
	}
	return nil
}

// previousResponseWasYes reports whether the stored invitation already
// had response='yes' before the caller mutated the in-memory entity.
// Read outside the transaction: it only decides whether the capacity
// guard applies; the FOR UPDATE inside the tx is what actually
// serializes concurrent confirmations.
func (r *PostgresInvitationRepository) previousResponseWasYes(ctx context.Context, id entities.InvitationID) (bool, error) {
	row, err := r.queries.GetInvitationByID(ctx, string(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, fmt.Errorf(invitationRepoOp+": respond %q: %w",
				id, domainerrors.ErrInvitationNotFound)
		}
		return false, fmt.Errorf(invitationRepoOp+": load previous %q: %w", id, err)
	}
	return entities.InvitationResponse(row.Response) == entities.InvitationResponseYes, nil
}

// enforceCapacity locks the match row and rejects the confirmation when
// the match already holds maxConfirmed "yes" invitations.
func enforceCapacity(ctx context.Context, qtx *generated.Queries, matchID entities.MatchID, maxConfirmed int) error {
	if _, err := qtx.LockMatchRow(ctx, string(matchID)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf(invitationRepoOp+": lock match %q: %w",
				matchID, domainerrors.ErrReferencedEntityNotFound)
		}
		return fmt.Errorf(invitationRepoOp+": lock match %q: %w", matchID, err)
	}

	confirmed, err := qtx.CountConfirmedInvitationsByMatchID(ctx, string(matchID))
	if err != nil {
		return fmt.Errorf(invitationRepoOp+": count confirmed for %q: %w", matchID, err)
	}
	if int(confirmed) >= maxConfirmed {
		return fmt.Errorf(invitationRepoOp+": %w", domainerrors.ErrMatchFull)
	}
	return nil
}

// Claim persists the invitation's rotated token hash and claim
// timestamp. The UPDATE is conditional (claimed_at IS NULL AND
// response = 'pending'), so two concurrent claims cannot both win: the
// loser's update matches zero rows and is mapped to
// ErrInvitationAlreadyClaimed. The caller loaded the invitation just
// before, so zero rows means the row settled in the meantime, not that
// it disappeared.
func (r *PostgresInvitationRepository) Claim(ctx context.Context, inv *entities.Invitation) error {
	if _, err := r.queries.ClaimInvitation(ctx, mappers.InvitationToClaimParams(inv)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf(invitationRepoOp+": claim %q: %w",
				inv.ID(), domainerrors.ErrInvitationAlreadyClaimed)
		}
		return fmt.Errorf(invitationRepoOp+": claim %q: %w", inv.ID(), err)
	}
	return nil
}

// Delete removes an invitation by identifier.
func (r *PostgresInvitationRepository) Delete(ctx context.Context, id entities.InvitationID) error {
	if err := r.queries.DeleteInvitation(ctx, string(id)); err != nil {
		return fmt.Errorf(invitationRepoOp+": delete %q: %w", id, err)
	}
	return nil
}

// isInvitationForeignKeyViolation checks if the error is a pg FK violation.
func isInvitationForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == pgerrcode.ForeignKeyViolation
}
