package discord

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ianadou/smo/domain/entities"
	domainerrors "github.com/ianadou/smo/domain/errors"
	"github.com/ianadou/smo/domain/events"
	"github.com/ianadou/smo/domain/ports"
)

// Subscriber turns a MatchTeamsReady domain event into a Discord
// webhook notification. It implements the ports.EventSubscriber
// interface and is registered with the publisher at boot time.
//
// A group with no webhook configured is a no-op (returns nil): the
// organizer has explicitly opted out of Discord notifications for
// that group, which is a valid configuration. See ADR 0003.
//
// A failure to fetch the group, fetch the match, or send to Discord
// returns the error to the publisher, which logs it at WARN and
// continues with the next subscriber. The use case that emitted the
// event is never aware of a notification failure — Discord is
// best-effort, never authoritative.
type Subscriber struct {
	notifier  Notifier
	groupRepo ports.GroupRepository
	matchRepo ports.MatchRepository
	logger    *slog.Logger
}

// NewSubscriber returns a Discord subscriber. Every dependency is
// required.
func NewSubscriber(notifier Notifier, groupRepo ports.GroupRepository, matchRepo ports.MatchRepository, logger *slog.Logger) *Subscriber {
	return &Subscriber{
		notifier:  notifier,
		groupRepo: groupRepo,
		matchRepo: matchRepo,
		logger:    logger,
	}
}

// Handle reacts to a MatchTeamsReady event by posting a Discord
// notification on the group's configured webhook. Other event types
// are ignored.
func (s *Subscriber) Handle(ctx context.Context, event events.Event) error {
	teamsReady, ok := event.(events.MatchTeamsReady)
	if !ok {
		return nil
	}

	group, err := s.groupRepo.FindByID(ctx, teamsReady.GroupID)
	if err != nil {
		return fmt.Errorf("discord subscriber: load group %q: %w", teamsReady.GroupID, err)
	}
	if group.WebhookURL() == "" {
		return nil
	}

	match, err := s.matchRepo.FindByID(ctx, teamsReady.MatchID)
	if err != nil {
		// A missing match for a just-emitted teams_ready event is a
		// data-consistency anomaly. Log and bail without leaking it
		// further: nothing useful to notify about.
		if errors.Is(err, domainerrors.ErrMatchNotFound) {
			s.logger.WarnContext(ctx, "discord subscriber: match disappeared between emit and handle",
				"match_id", string(teamsReady.MatchID))
			return nil
		}
		return fmt.Errorf("discord subscriber: load match %q: %w", teamsReady.MatchID, err)
	}

	return s.notifier.Send(ctx, group.WebhookURL(), buildPayload(group, match))
}

// buildPayload assembles the Discord embed for a teams-ready
// notification. Kept private so the wire format remains an internal
// detail of the package.
func buildPayload(group *entities.Group, match *entities.Match) Payload {
	return Payload{
		Title:       fmt.Sprintf("Teams ready — %s", match.Title()),
		Description: fmt.Sprintf("Group: %s", group.Name()),
		Fields: []Field{
			{Name: "Venue", Value: match.Venue(), Inline: true},
			{Name: "Scheduled at", Value: match.ScheduledAt().Format("2006-01-02 15:04 MST"), Inline: true},
		},
	}
}
