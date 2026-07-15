package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// conduitLikeInstructions returns the two-instruction resolution sequence that
// mirrors Conduit of Worlds's activated ability: an optional paid cast of the
// targeted graveyard card, gated on the resolving player not having cast a spell
// this turn and publishing whether the cast happened, then a self-scoped
// "can't cast additional spells this turn" rule effect gated on that cast
// actually happening.
func conduitLikeInstructions() []game.Instruction {
	return []game.Instruction{
		{
			Optional: true,
			Primitive: game.CastForFree{
				Player:      game.ControllerReference(),
				Card:        game.CardReference{Kind: game.CardReferenceTarget},
				Zone:        zone.Graveyard,
				PayManaCost: true,
			},
			Condition: opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{
				Negate: true,
				EventHistory: opt.Val(game.EventHistoryCondition{
					Pattern: game.TriggerPattern{Event: game.EventSpellCast, Controller: game.TriggerControllerYou},
					Window:  game.EventHistoryCurrentTurn,
				}),
			})}),
			PublishResult: "cast",
		},
		{
			Primitive: game.ApplyRule{
				RuleEffects: []game.RuleEffect{{
					Kind:           game.RuleEffectCantCastSpells,
					AffectedPlayer: game.PlayerYou,
				}},
				Duration: game.DurationThisTurn,
			},
			ResultGate: opt.Val(game.InstructionResultGate{Key: "cast", Succeeded: game.TriTrue}),
		},
	}
}

func graveBearCard() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Grave Bear",
		ManaCost: opt.Val(cost.Mana{cost.O(1)}),
		Types:    []types.Card{types.Creature},
	}}
}

// addConduitLikeSource pushes a resolving ability whose sequence is the
// Conduit-like paid cast-and-lock, targeting cardID in Player1's graveyard.
func addConduitLikeSource(g *game.Game, targetID game.PlayerID, cardID game.ObjectID, t *testing.T) game.ObjectID {
	sourceID := addInstructionSpellToStackForController(g, game.Player1, conduitLikeInstructions(), []game.Target{currentCardTarget(t, g, cardID)})
	card, ok := g.GetCardInstance(sourceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	card.Def.SpellAbility.Val.Modes[0].Targets = []game.TargetSpec{{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowCard,
		TargetZone: zone.Graveyard,
		Selection:  opt.Val(game.Selection{ExcludedTypes: []types.Card{types.Land}, Controller: game.ControllerYou}),
	}}
	return sourceID
}

// TestCastPaidResolutionCastsAndLocks proves the successful path: the resolving
// player pays for the targeted graveyard card, it is put on the stack under
// their control, its mana cost is spent, and the self "can't cast additional
// spells this turn" restriction is applied because a spell was cast.
func TestCastPaidResolutionCastsAndLocks(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	bearID := addCardToGraveyard(g, game.Player1, graveBearCard())
	addConduitLikeSource(g, game.Player1, bearID, t)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)

	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: true}}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if g.Players[game.Player1].Graveyard.Contains(bearID) {
		t.Fatal("targeted card still in graveyard after a paid cast")
	}
	top, ok := g.Stack.Peek()
	if !ok || top.SourceID != bearID {
		t.Fatalf("stack top = %#v, want the cast graveyard card %v", top, bearID)
	}
	if top.Controller != game.Player1 {
		t.Fatalf("cast spell controller = %v, want %v", top.Controller, game.Player1)
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("mana pool total = %d after paying, want 0", got)
	}
	if !spellCastProhibited(g, game.Player1, vanillaCreatureDef()) {
		t.Fatal("controller can still cast spells; the can't-cast-additional lock was not applied")
	}
	if spellCastProhibited(g, game.Player2, vanillaCreatureDef()) {
		t.Fatal("the self lock wrongly restricts the opponent")
	}
}

// TestCastPaidResolutionDeclineNoCastNoLock proves declining the optional cast
// casts nothing and, because no spell was cast, applies no restriction.
func TestCastPaidResolutionDeclineNoCastNoLock(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	bearID := addCardToGraveyard(g, game.Player1, graveBearCard())
	addConduitLikeSource(g, game.Player1, bearID, t)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)

	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: false}}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Graveyard.Contains(bearID) {
		t.Fatal("targeted card left the graveyard despite declining the cast")
	}
	if _, ok := g.Stack.Peek(); ok {
		t.Fatal("a spell was cast despite declining")
	}
	if spellCastProhibited(g, game.Player1, vanillaCreatureDef()) {
		t.Fatal("the can't-cast lock was applied without a cast")
	}
}

// TestCastPaidResolutionInsufficientManaNoCastNoLock proves that accepting the
// cast but being unable to pay casts nothing, restores the card to the
// graveyard, and applies no restriction (the "If you do" gate fails).
func TestCastPaidResolutionInsufficientManaNoCastNoLock(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	bearID := addCardToGraveyard(g, game.Player1, graveBearCard())
	addConduitLikeSource(g, game.Player1, bearID, t)
	// No mana in the pool: the {1} cost cannot be paid.

	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: true}}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Graveyard.Contains(bearID) {
		t.Fatal("targeted card did not return to the graveyard after an unpaid cast")
	}
	if _, ok := g.Stack.Peek(); ok {
		t.Fatal("a spell was cast despite being unable to pay")
	}
	if spellCastProhibited(g, game.Player1, vanillaCreatureDef()) {
		t.Fatal("the can't-cast lock was applied without a successful cast")
	}
}

// TestCastPaidResolutionCondFailsAfterPriorSpell proves the resolution-time
// condition: if the player already cast a spell this turn, the paid cast is
// skipped entirely and no restriction is applied, even though activation itself
// remains legal.
func TestCastPaidResolutionCondFailsAfterPriorSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	bearID := addCardToGraveyard(g, game.Player1, graveBearCard())
	addConduitLikeSource(g, game.Player1, bearID, t)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	// Record that Player1 has already cast a spell this turn.
	g.AppendEvent(game.Event{
		Kind:       game.EventSpellCast,
		Controller: game.Player1,
	})

	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: true}}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Graveyard.Contains(bearID) {
		t.Fatal("targeted card was cast despite a prior spell this turn")
	}
	if spellCastProhibited(g, game.Player1, vanillaCreatureDef()) {
		t.Fatal("the can't-cast lock was applied despite the condition failing")
	}
}

// TestCastPaidResolutionProhibitedNoCast proves the paid cast obeys existing
// cast prohibitions: with a standing "can't cast creature spells" restriction on
// the controller, accepting the cast still casts nothing.
func TestCastPaidResolutionProhibitedNoCast(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	bearID := addCardToGraveyard(g, game.Player1, graveBearCard())
	addConduitLikeSource(g, game.Player1, bearID, t)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		Kind:           game.RuleEffectCantCastSpells,
		Controller:     game.Player1,
		AffectedPlayer: game.PlayerYou,
	})

	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: true}}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Graveyard.Contains(bearID) {
		t.Fatal("targeted card was cast despite a standing cast prohibition")
	}
}
