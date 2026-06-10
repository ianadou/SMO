package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainerrors "github.com/ianadou/smo/domain/errors"
)

var (
	teamLockKickoff = time.Date(2026, 6, 1, 20, 0, 0, 0, time.UTC)
	beforeTeamLock  = teamLockKickoff.Add(-time.Hour)        // 60 min out → editable
	atTeamLock      = teamLockKickoff.Add(-TeamLockLeadTime) // exactly at the lock → frozen
	afterTeamLock   = teamLockKickoff.Add(-5 * time.Minute)  // 5 min out → frozen
)

func openMatch(t *testing.T) *Match {
	t.Helper()
	m, err := NewMatch("m1", "g1", "Match", "Hall", teamLockKickoff, teamLockKickoff.Add(-24*time.Hour))
	require.NoError(t, err)
	require.NoError(t, m.Open())
	return m
}

func teamsReadyMatch(t *testing.T) *Match {
	t.Helper()
	m := openMatch(t)
	require.NoError(t, m.AssignTeams([]PlayerID{"a", "b"}, []PlayerID{"c", "d"}, beforeTeamLock))
	require.NoError(t, m.MarkTeamsReady())
	return m
}

func TestAssignTeams_StoresTeams_WhenValidAndOpen(t *testing.T) {
	m := openMatch(t)

	err := m.AssignTeams([]PlayerID{"a", "b"}, []PlayerID{"c", "d"}, beforeTeamLock)

	require.NoError(t, err)
	assert.Equal(t, []PlayerID{"a", "b"}, m.TeamA())
	assert.Equal(t, []PlayerID{"c", "d"}, m.TeamB())
}

func TestAssignTeams_StoresTeams_WhenTeamsReadyAndBeforeLock(t *testing.T) {
	m := teamsReadyMatch(t)

	err := m.AssignTeams([]PlayerID{"a", "c"}, []PlayerID{"b", "d"}, beforeTeamLock)

	require.NoError(t, err)
	assert.Equal(t, []PlayerID{"a", "c"}, m.TeamA())
	assert.Equal(t, []PlayerID{"b", "d"}, m.TeamB())
}

func TestAssignTeams_RejectsLocked_WhenWithinLeadTimeOfKickoff(t *testing.T) {
	m := openMatch(t)

	err := m.AssignTeams([]PlayerID{"a", "b"}, []PlayerID{"c", "d"}, afterTeamLock)

	assert.ErrorIs(t, err, domainerrors.ErrTeamsLocked)
}

func TestAssignTeams_RejectsLocked_AtExactlyLeadTimeBeforeKickoff(t *testing.T) {
	m := openMatch(t)

	err := m.AssignTeams([]PlayerID{"a", "b"}, []PlayerID{"c", "d"}, atTeamLock)

	assert.ErrorIs(t, err, domainerrors.ErrTeamsLocked)
}

func TestAssignTeams_RejectsLocked_WhenTeamsReadyAndPastLock(t *testing.T) {
	m := teamsReadyMatch(t)

	err := m.AssignTeams([]PlayerID{"a", "c"}, []PlayerID{"b", "d"}, afterTeamLock)

	assert.ErrorIs(t, err, domainerrors.ErrTeamsLocked)
}

func TestAssignTeams_RejectsImbalance_WhenDiffGreaterThanOne(t *testing.T) {
	m := openMatch(t)

	err := m.AssignTeams([]PlayerID{"a", "b", "c"}, []PlayerID{"d"}, beforeTeamLock)

	assert.ErrorIs(t, err, domainerrors.ErrInvalidAssignment)
}

func TestAssignTeams_AllowsDiffOfOne_WhenOddTotal(t *testing.T) {
	m := openMatch(t)

	err := m.AssignTeams([]PlayerID{"a", "b", "c"}, []PlayerID{"d", "e"}, beforeTeamLock)

	require.NoError(t, err)
}

func TestAssignTeams_RejectsOverlap_WhenPlayerInBothTeams(t *testing.T) {
	m := openMatch(t)

	err := m.AssignTeams([]PlayerID{"a", "b"}, []PlayerID{"b", "c"}, beforeTeamLock)

	assert.ErrorIs(t, err, domainerrors.ErrInvalidAssignment)
}

func TestAssignTeams_RejectsEmptyTeam(t *testing.T) {
	m := openMatch(t)

	err := m.AssignTeams([]PlayerID{"a"}, nil, beforeTeamLock)

	assert.ErrorIs(t, err, domainerrors.ErrInvalidAssignment)
}

func TestAssignTeams_RejectsWhenDraft(t *testing.T) {
	m, err := NewMatch("m1", "g1", "Match", "Hall", teamLockKickoff, teamLockKickoff.Add(-24*time.Hour))
	require.NoError(t, err)

	err = m.AssignTeams([]PlayerID{"a"}, []PlayerID{"b"}, beforeTeamLock)

	assert.ErrorIs(t, err, domainerrors.ErrTeamsNotEditable)
}

func TestAssignTeams_RejectsWhenInProgress(t *testing.T) {
	m := teamsReadyMatch(t)
	require.NoError(t, m.Start())

	err := m.AssignTeams([]PlayerID{"a", "c"}, []PlayerID{"b", "d"}, beforeTeamLock)

	assert.ErrorIs(t, err, domainerrors.ErrTeamsNotEditable)
}

func TestAssignTeams_RejectsIntraTeamDuplicate(t *testing.T) {
	m := openMatch(t)

	err := m.AssignTeams([]PlayerID{"a", "a"}, []PlayerID{"b", "c"}, beforeTeamLock)

	assert.ErrorIs(t, err, domainerrors.ErrInvalidAssignment)
}

func TestAssignTeams_RejectsImbalance_WhenTeamBLargerByTwo(t *testing.T) {
	m := openMatch(t)

	err := m.AssignTeams([]PlayerID{"a"}, []PlayerID{"b", "c", "d"}, beforeTeamLock)

	assert.ErrorIs(t, err, domainerrors.ErrInvalidAssignment)
}

func TestAssignTeams_AllowsEqualSizeTeams(t *testing.T) {
	m := openMatch(t)

	err := m.AssignTeams([]PlayerID{"a", "b"}, []PlayerID{"c", "d"}, beforeTeamLock)

	require.NoError(t, err)
}
