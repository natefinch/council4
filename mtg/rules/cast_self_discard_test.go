package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// selfDiscardSpell builds a free instant whose only additional cost is
// "discard a card" from hand, with no cheaper alternative. It exists to prove
// the card being cast can never be the card discarded to pay its own cost.
func selfDiscardSpell() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Self Discard",
		ManaCost: opt.Val(cost.Mana{}),
		Types:    []types.Card{types.Instant},
		AdditionalCosts: []cost.Additional{
			{Kind: cost.AdditionalDiscard, Text: "discard a card", Amount: 1, Source: zone.Hand},
		},
		SpellAbility: opt.Val(game.AbilityContent{}),
	}}
}

// TestCastSpellCannotDiscardItselfForOwnCost proves CR 601.2a: a spell moves to
// the stack before its costs are paid, so a "discard a card" cost paid from
// hand can't select the very spell being cast. When that spell is the only card
// in hand, the cost is unpayable and the proposal is reversed (CR 728), leaving
// the card safely in hand rather than discarding — and losing — itself. Before
// the fix, the spell was still in hand during payment, discarded itself, and the
// engine then panicked finding it gone from its source zone.
func TestCastSpellCannotDiscardItselfForOwnCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, selfDiscardSpell())
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	// Must not panic; the unpayable cast is simply reversed.
	engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil))

	if !g.Players[game.Player1].Hand.Contains(spellID) {
		t.Fatal("spell left hand even though its own discard cost was unpayable")
	}
	if g.Players[game.Player1].Graveyard.Contains(spellID) {
		t.Fatal("spell discarded itself to pay its own cost")
	}
	if !g.Stack.IsEmpty() {
		t.Fatal("an unpayable cast left a spell on the stack")
	}
}

// TestCastSpellDiscardsAnotherCardNotItself proves the positive case: with one
// other card in hand, the discard cost is paid with that other card while the
// spell goes to the stack.
func TestCastSpellDiscardsAnotherCardNotItself(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, selfDiscardSpell())
	otherID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Spare Card", Types: []types.Card{types.Sorcery},
	}})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("cast with a discardable other card failed")
	}
	if !g.Players[game.Player1].Graveyard.Contains(otherID) {
		t.Fatal("the other card was not discarded to pay the cost")
	}
	if g.Players[game.Player1].Graveyard.Contains(spellID) {
		t.Fatal("the spell discarded itself instead of the other card")
	}
	if g.Players[game.Player1].Hand.Contains(spellID) {
		t.Fatal("cast spell remained in hand instead of going to the stack")
	}
}
