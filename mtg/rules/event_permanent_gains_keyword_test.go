package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// TestEventPermanentGainsKeywordUntilEndOfTurn proves that a creature-ETB trigger
// granting a temporary keyword to the entering creature ("Whenever a [filter]
// creature you control enters, it gains haste until end of turn." — Dragon
// Tempest) resolves the event-permanent reference, grants the keyword to that
// creature, and removes it during cleanup.
func TestEventPermanentGainsKeywordUntilEndOfTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Haste Granter",
		Types: nil,
	}})
	entered := addCombatCreaturePermanent(g, game.Player1)

	if hasKeyword(g, entered, game.Haste) {
		t.Fatal("entering creature unexpectedly already had haste")
	}

	obj := &game.StackObject{
		Kind:            game.StackTriggeredAbility,
		SourceID:        source.ObjectID,
		SourceCardID:    source.CardInstanceID,
		Controller:      game.Player1,
		HasTriggerEvent: true,
		TriggerEvent:    game.Event{Kind: game.EventPermanentEnteredBattlefield, PermanentID: entered.ObjectID},
	}
	resolveInstruction(engine, g, obj, game.ApplyContinuous{
		Object: opt.Val(game.EventPermanentReference()),
		ContinuousEffects: []game.ContinuousEffect{{
			Layer:       game.LayerAbility,
			AddKeywords: []game.Keyword{game.Haste},
		}},
		Duration: game.DurationUntilEndOfTurn,
	}, &TurnLog{})

	if !hasKeyword(g, entered, game.Haste) {
		t.Fatal("entering creature did not gain haste this turn")
	}

	expireCleanupDurations(g)
	if hasKeyword(g, entered, game.Haste) {
		t.Fatal("entering creature still has haste after cleanup (next turn)")
	}
}
