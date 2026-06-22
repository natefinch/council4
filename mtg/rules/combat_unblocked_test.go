package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

// TestEmitUnblockedAttackerEventsFiresOnlyForUnblockedAttackers covers the
// unblocked-attacker combat trigger (CR 509.1h): once every defending player
// has declared blockers, each attacker that no creature blocked emits a single
// EventAttackerBecameUnblocked keyed to itself, while blocked attackers emit
// nothing. This is the runtime event behind "whenever this creature attacks and
// isn't blocked" and the Frenzy keyword.
func TestEmitUnblockedAttackerEventsFiresOnlyForUnblockedAttackers(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	unblocked := addCombatPermanent(g, game.Player1, vanillaCreature("Lone Raider", 2, 2))
	blocked := addCombatPermanent(g, game.Player1, vanillaCreature("Stopped Raider", 2, 2))
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: unblocked.ObjectID},
			{Attacker: blocked.ObjectID},
		},
		BlockedAttackers: map[id.ID]bool{blocked.ObjectID: true},
	}

	before := len(g.Events)
	emitUnblockedAttackerEvents(g)

	var unblockedEvents, blockedEvents int
	for _, event := range g.Events[before:] {
		if event.Kind != game.EventAttackerBecameUnblocked {
			continue
		}
		switch event.PermanentID {
		case unblocked.ObjectID:
			unblockedEvents++
			if event.Controller != game.Player1 {
				t.Fatalf("unblocked event controller = %v, want Player1", event.Controller)
			}
		case blocked.ObjectID:
			blockedEvents++
		default:
		}
	}
	if unblockedEvents != 1 {
		t.Fatalf("EventAttackerBecameUnblocked for unblocked attacker fired %d times, want 1", unblockedEvents)
	}
	if blockedEvents != 0 {
		t.Fatalf("EventAttackerBecameUnblocked wrongly fired for the blocked attacker %d times", blockedEvents)
	}
}

// TestUnblockedAttackerTriggerMatchesSelf confirms a Source=Self pattern keyed
// to EventAttackerBecameUnblocked matches the emitting attacker and ignores an
// unrelated creature's event, so "whenever this creature attacks and isn't
// blocked" triggers on the right permanent.
func TestUnblockedAttackerTriggerMatchesSelf(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatPermanent(g, game.Player1, vanillaCreature("Lone Raider", 2, 2))
	other := addCombatPermanent(g, game.Player1, vanillaCreature("Bystander", 2, 2))
	pattern := &game.TriggerPattern{
		Event:  game.EventAttackerBecameUnblocked,
		Source: game.TriggerSourceSelf,
	}

	event := game.Event{
		Kind:           game.EventAttackerBecameUnblocked,
		Controller:     game.Player1,
		SourceObjectID: attacker.ObjectID,
		PermanentID:    attacker.ObjectID,
	}
	if !triggerMatchesEvent(g, attacker, pattern, event) {
		t.Fatal("self trigger did not match the attacker's own unblocked event")
	}
	if triggerMatchesEvent(g, other, pattern, event) {
		t.Fatal("self trigger wrongly matched an unrelated creature's unblocked event")
	}
}
