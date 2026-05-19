package mappers

import (
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ianadou/smo/domain/entities"
	"github.com/ianadou/smo/infrastructure/persistence/postgres/generated"
)

// MatchToDomain converts a sqlc-generated Matches row into a domain
// Match entity. Returns an error if the row contains data that fails
// the domain validation (e.g., empty title, unknown status), which
// would indicate a corrupted database row.
func MatchToDomain(row generated.Matches) (*entities.Match, error) {
	status, err := entities.ParseMatchStatus(row.Status)
	if err != nil {
		return nil, err
	}

	var mvp *entities.PlayerID
	if row.MvpPlayerID != nil {
		id := entities.PlayerID(*row.MvpPlayerID)
		mvp = &id
	}

	return entities.RehydrateMatch(entities.MatchSnapshot{
		ID:          entities.MatchID(row.ID),
		GroupID:     entities.GroupID(row.GroupID),
		Title:       row.Title,
		Venue:       row.Venue,
		ScheduledAt: row.ScheduledAt.Time,
		Status:      status,
		MVPPlayerID: mvp,
		CreatedAt:   row.CreatedAt.Time,
	})
}

// MatchToCreateParams converts a domain Match entity into the parameter
// struct expected by the generated CreateMatch function.
func MatchToCreateParams(match *entities.Match) generated.CreateMatchParams {
	return generated.CreateMatchParams{
		ID:          string(match.ID()),
		GroupID:     string(match.GroupID()),
		Title:       match.Title(),
		Venue:       match.Venue(),
		ScheduledAt: pgtype.Timestamptz{Time: match.ScheduledAt(), Valid: true},
		Status:      string(match.Status()),
		CreatedAt:   pgtype.Timestamptz{Time: match.CreatedAt(), Valid: true},
	}
}

// MatchToUpdateStatusParams converts a domain Match entity into the
// parameter struct expected by the generated UpdateMatchStatus function.
// Only the ID (for WHERE) and the status (for SET) are populated.
func MatchToUpdateStatusParams(match *entities.Match) generated.UpdateMatchStatusParams {
	return generated.UpdateMatchStatusParams{
		ID:     string(match.ID()),
		Status: string(match.Status()),
	}
}

// MatchToFinalizeParams converts a domain Match into the parameter
// struct expected by the generated FinalizeMatch function. The MVP
// is encoded as a nullable string for sqlc; nil means no MVP elected.
func MatchToFinalizeParams(match *entities.Match) generated.FinalizeMatchParams {
	var mvp *string
	if match.MVP() != nil {
		s := string(*match.MVP())
		mvp = &s
	}
	return generated.FinalizeMatchParams{
		ID:          string(match.ID()),
		MvpPlayerID: mvp,
		Status:      string(match.Status()),
	}
}

// MatchTeamMemberInsertParams flattens a match's teamA/teamB into the
// per-row insert params for match_team_members, slot = index within team.
func MatchTeamMemberInsertParams(match *entities.Match) []generated.InsertMatchTeamMemberParams {
	rows := make([]generated.InsertMatchTeamMemberParams, 0, len(match.TeamA())+len(match.TeamB()))
	for slot, id := range match.TeamA() {
		rows = append(rows, generated.InsertMatchTeamMemberParams{
			MatchID: string(match.ID()), PlayerID: string(id), Team: "A", Slot: int32(slot),
		})
	}
	for slot, id := range match.TeamB() {
		rows = append(rows, generated.InsertMatchTeamMemberParams{
			MatchID: string(match.ID()), PlayerID: string(id), Team: "B", Slot: int32(slot),
		})
	}
	return rows
}

// TeamsFromMemberRows splits ordered membership rows into (teamA, teamB).
func TeamsFromMemberRows(rows []generated.MatchTeamMembers) (teamA, teamB []entities.PlayerID) {
	for _, r := range rows {
		if r.Team == "A" {
			teamA = append(teamA, entities.PlayerID(r.PlayerID))
		} else {
			teamB = append(teamB, entities.PlayerID(r.PlayerID))
		}
	}
	return teamA, teamB
}
