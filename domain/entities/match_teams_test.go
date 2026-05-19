package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainerrors "github.com/ianadou/smo/domain/errors"
)

func openMatch(t *testing.T) *Match {
	t.Helper()
	m, err := NewMatch("m1", "g1", "Match", "Hall", time.Now(), time.Now())
	require.NoError(t, err)
	require.NoError(t, m.Open())
	return m
}

func TestAssignTeams_StoresTeams_WhenValidAndOpen(t *testing.T) {
	m := openMatch(t)

	err := m.AssignTeams([]PlayerID{"a", "b"}, []PlayerID{"c", "d"})

	require.NoError(t, err)
	assert.Equal(t, []PlayerID{"a", "b"}, m.TeamA())
	assert.Equal(t, []PlayerID{"c", "d"}, m.TeamB())
}

func TestAssignTeams_RejectsImbalance_WhenDiffGreaterThanOne(t *testing.T) {
	m := openMatch(t)

	err := m.AssignTeams([]PlayerID{"a", "b", "c"}, []PlayerID{"d"})

	assert.ErrorIs(t, err, domainerrors.ErrInvalidAssignment)
}

func TestAssignTeams_AllowsDiffOfOne_WhenOddTotal(t *testing.T) {
	m := openMatch(t)

	err := m.AssignTeams([]PlayerID{"a", "b", "c"}, []PlayerID{"d", "e"})

	require.NoError(t, err)
}

func TestAssignTeams_RejectsOverlap_WhenPlayerInBothTeams(t *testing.T) {
	m := openMatch(t)

	err := m.AssignTeams([]PlayerID{"a", "b"}, []PlayerID{"b", "c"})

	assert.ErrorIs(t, err, domainerrors.ErrInvalidAssignment)
}

func TestAssignTeams_RejectsEmptyTeam(t *testing.T) {
	m := openMatch(t)

	err := m.AssignTeams([]PlayerID{"a"}, nil)

	assert.ErrorIs(t, err, domainerrors.ErrInvalidAssignment)
}

func TestAssignTeams_RejectsWhenNotOpen(t *testing.T) {
	m, err := NewMatch("m1", "g1", "Match", "Hall", time.Now(), time.Now())
	require.NoError(t, err)

	err = m.AssignTeams([]PlayerID{"a"}, []PlayerID{"b"})

	assert.ErrorIs(t, err, domainerrors.ErrTeamsNotEditable)
}

func TestAssignTeams_RejectsIntraTeamDuplicate(t *testing.T) {
	m := openMatch(t)

	err := m.AssignTeams([]PlayerID{"a", "a"}, []PlayerID{"b", "c"})

	assert.ErrorIs(t, err, domainerrors.ErrInvalidAssignment)
}

func TestAssignTeams_RejectsImbalance_WhenTeamBLargerByTwo(t *testing.T) {
	m := openMatch(t)

	err := m.AssignTeams([]PlayerID{"a"}, []PlayerID{"b", "c", "d"})

	assert.ErrorIs(t, err, domainerrors.ErrInvalidAssignment)
}

func TestAssignTeams_AllowsEqualSizeTeams(t *testing.T) {
	m := openMatch(t)

	err := m.AssignTeams([]PlayerID{"a", "b"}, []PlayerID{"c", "d"})

	require.NoError(t, err)
}
