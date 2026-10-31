package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: Éowyn, Shieldmaiden (LTC 1 printing
// dd4a00ff-2206-4e12-a0ab-61ed82c9e6c5, Oracle
// 98a2a642-a30a-43e7-b47f-4776c22d1cb0) "At the beginning of combat on your
// turn, if another Human entered the battlefield under your control this turn,
// create two 2/2 red Human Knight creature tokens ..."; rulings retrieved from
// https://api.scryfall.com/cards/dd4a00ff-2206-4e12-a0ab-61ed82c9e6c5/rulings
// returned one 2023-06-16 ruling (the Human must have entered before the
// beginning of combat). Official Comprehensive Rules effective 2026-09-25
// (https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt),
// read: 109.5 ("you" on a triggered ability is the controller of the object
// when the ability triggered), 113.8 and 603.3a (that player controls the
// ability on the stack), 113.7a and 608.2h (the ability exists independently of
// its source and uses the source's last-known information once it leaves),
// 400.7 (a card that changes zones becomes a new object) and 603.4 (the
// intervening "if" condition is rechecked on resolution). So after the trigger
// is on the stack, a control change of Éowyn keeps "your control" bound to the
// trigger's controller, and Éowyn leaving keeps "another" relative to the
// object that triggered; a reentered Éowyn is a different, "another" Human.

func eowynResolutionGame(t *testing.T) (*game.Game, *Engine, *game.Permanent) {
	t.Helper()
	g, engine, _, eowyn := entryHistoryGame(t)
	return g, engine, eowyn
}

func putEowynTriggerOnStack(t *testing.T, g *game.Game, engine *Engine, eowyn *game.Permanent) *game.StackObject {
	t.Helper()
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepBeginningOfCombat
	if !eowynTriggersAtCombat(t, g, eowyn) {
		t.Fatal("Éowyn did not trigger after another Human entered under Player1's control")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want Éowyn's trigger", g.Stack.Size())
	}
	obj := g.Stack.Objects()[0]
	if obj.Kind != game.StackTriggeredAbility || obj.SourceID != eowyn.ObjectID || obj.Controller != game.Player1 {
		t.Fatalf("stack object = %+v, want Player1's Éowyn trigger", obj)
	}
	return obj
}

// pushEowynTrigger puts Éowyn's printed trigger on the stack as the trigger
// pipeline would, without its trigger-time check, so a resolution-only
// condition that was never true can be exercised.
func pushEowynTrigger(g *game.Game, eowyn *game.Permanent) {
	g.Stack.Push(&game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility, Controller: game.Player1,
		SourceID: eowyn.ObjectID, SourceCardID: eowyn.CardInstanceID, Face: eowyn.Face,
		AbilityIndex: entryHistoryEowynTriggerIndex,
	})
}

func resolveEowynTrigger(t *testing.T, g *game.Game, engine *Engine) {
	t.Helper()
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Stack.IsEmpty() {
		t.Fatalf("stack size = %d after resolving Éowyn's trigger", g.Stack.Size())
	}
}

func eowynKnightTokens(g *game.Game, controller game.PlayerID) int {
	count := 0
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanentTokenName(permanent) == "Human Knight" &&
			effectiveController(g, permanent) == controller {
			count++
		}
	}
	return count
}

func reenterEowyn(t *testing.T, g *game.Game, eowyn *game.Permanent) *game.Permanent {
	t.Helper()
	cardID := eowyn.CardInstanceID
	if !movePermanentToZone(g, eowyn, zone.Exile) {
		t.Fatal("Éowyn was not exiled")
	}
	card, ok := g.GetCardInstance(cardID)
	if !ok || !g.Players[game.Player1].Exile.Remove(cardID) {
		t.Fatal("exiled Éowyn card is missing")
	}
	returned, ok := createCardPermanent(g, card, game.Player1, zone.Exile)
	if !ok || returned.ObjectID == eowyn.ObjectID {
		t.Fatal("Éowyn did not return as a new object")
	}
	return returned
}

func TestEowynTriggerKeepsControllerAfterSourceChangesControl(t *testing.T) {
	t.Parallel()
	g, engine, eowyn := eowynResolutionGame(t)
	enterEntryHistoryCard(t, g, game.Player1, entryHistoryCreature("Resolution Squire", types.Human))
	putEowynTriggerOnStack(t, g, engine, eowyn)

	gainEntryHistoryControl(g, eowyn, game.Player2)
	if effectiveController(g, eowyn) != game.Player2 {
		t.Fatal("Player2 did not gain control of Éowyn")
	}
	resolveEowynTrigger(t, g, engine)
	if got := eowynKnightTokens(g, game.Player1); got != 2 {
		t.Fatalf("Player1 Human Knight tokens = %d, want 2 after Éowyn changed control", got)
	}
	if got := eowynKnightTokens(g, game.Player2); got != 0 {
		t.Fatalf("Player2 Human Knight tokens = %d, want 0", got)
	}
}

func TestEowynTriggerKeepsSourceIdentityAfterSourceLeaves(t *testing.T) {
	t.Parallel()
	g, engine, eowyn := eowynResolutionGame(t)
	enterEntryHistoryCard(t, g, game.Player1, entryHistoryCreature("Resolution Page", types.Human))
	putEowynTriggerOnStack(t, g, engine, eowyn)

	if !movePermanentToZone(g, eowyn, zone.Graveyard) {
		t.Fatal("Éowyn was not put into the graveyard")
	}
	resolveEowynTrigger(t, g, engine)
	if got := eowynKnightTokens(g, game.Player1); got != 2 {
		t.Fatalf("Player1 Human Knight tokens = %d, want 2 after Éowyn left", got)
	}
}

func TestEowynResolutionDoesNotRebindYouToNewSourceController(t *testing.T) {
	t.Parallel()
	g, engine, eowyn := eowynResolutionGame(t)
	enterEntryHistoryCard(t, g, game.Player2, entryHistoryCreature("Rival Squire", types.Human))
	pushEowynTrigger(g, eowyn)
	gainEntryHistoryControl(g, eowyn, game.Player2)

	resolveEowynTrigger(t, g, engine)
	if got := eowynKnightTokens(g, game.Player1) + eowynKnightTokens(g, game.Player2); got != 0 {
		t.Fatalf("Human Knight tokens = %d; a Human entering under Player2 satisfied Player1's trigger", got)
	}
}

func TestEowynResolutionExcludesOwnEntryAfterChangesOfControlOrZone(t *testing.T) {
	t.Parallel()
	for _, transition := range []string{"control change", "graveyard"} {
		t.Run(transition, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			g.Turn.ActivePlayer = game.Player1
			g.Turn.PriorityPlayer = game.Player1
			g.Turn.Phase = game.PhaseCombat
			g.Turn.Step = game.StepBeginningOfCombat
			engine := NewEngine(nil)
			eowyn := enterEntryHistoryCard(t, g, game.Player1, entryHistoryRegistryCard(t, "Éowyn, Shieldmaiden"))
			pushEowynTrigger(g, eowyn)
			switch transition {
			case "control change":
				gainEntryHistoryControl(g, eowyn, game.Player2)
			default:
				if !movePermanentToZone(g, eowyn, zone.Graveyard) {
					t.Fatal("Éowyn was not put into the graveyard")
				}
			}

			resolveEowynTrigger(t, g, engine)
			if got := eowynKnightTokens(g, game.Player1) + eowynKnightTokens(g, game.Player2); got != 0 {
				t.Fatalf("Human Knight tokens = %d; Éowyn's own entry counted as another Human", got)
			}
		})
	}
}

func TestEowynResolutionCountsReenteredSourceAsAnotherObject(t *testing.T) {
	t.Parallel()
	g, engine, eowyn := eowynResolutionGame(t)
	pushEowynTrigger(g, eowyn)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepBeginningOfCombat
	reenterEowyn(t, g, eowyn)

	resolveEowynTrigger(t, g, engine)
	if got := eowynKnightTokens(g, game.Player1); got != 2 {
		t.Fatalf("Player1 Human Knight tokens = %d, want 2: the returned Éowyn is a new Human object", got)
	}
}
