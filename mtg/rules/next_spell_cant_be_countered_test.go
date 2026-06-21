package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// nextSpellUncounterableEffect builds the global, turn-scoped, controller-scoped
// "next spell can't be countered" rule effect produced by Mistrise Village.
func nextSpellUncounterableEffect(g *game.Game) game.RuleEffect {
	return game.RuleEffect{
		ID:                     g.IDGen.Next(),
		Kind:                   game.RuleEffectCantBeCountered,
		Controller:             game.Player1,
		AffectedController:     game.ControllerYou,
		AppliesToNextSpellOnly: true,
		Duration:               game.DurationThisTurn,
		CreatedTurn:            g.Turn.TurnNumber,
	}
}

// TestNextSpellCantBeCounteredConsumedByFirstSpell proves the next-spell-only
// uncounterable buff (Mistrise Village) protects exactly the first spell its
// controller casts and is then consumed, leaving later spells counterable.
func TestNextSpellCantBeCounteredConsumedByFirstSpell(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.RuleEffects = append(g.RuleEffects, nextSpellUncounterableEffect(g))
	def := elfCreatureDef()

	first := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, Controller: game.Player1, SourceTokenDef: def}
	g.Stack.Push(first)
	consumeNextSpellCantBeCounteredEffects(g, first)
	if stackSpellCanBeCountered(g, first) {
		t.Fatal("the next spell cast should be uncounterable")
	}
	if len(g.RuleEffects) != 0 {
		t.Fatalf("global rule effects = %#v, want the one-shot buff consumed", g.RuleEffects)
	}

	second := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, Controller: game.Player1, SourceTokenDef: def}
	g.Stack.Push(second)
	consumeNextSpellCantBeCounteredEffects(g, second)
	if !stackSpellCanBeCountered(g, second) {
		t.Fatal("a later spell should be counterable after the one-shot buff is consumed")
	}
}

// TestNextSpellCantBeCounteredIgnoresOpponentSpell proves the controller-scoped
// buff neither protects nor is consumed by an opponent's spell.
func TestNextSpellCantBeCounteredIgnoresOpponentSpell(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.RuleEffects = append(g.RuleEffects, nextSpellUncounterableEffect(g))
	def := elfCreatureDef()

	opponentSpell := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, Controller: game.Player2, SourceTokenDef: def}
	g.Stack.Push(opponentSpell)
	consumeNextSpellCantBeCounteredEffects(g, opponentSpell)
	if !stackSpellCanBeCountered(g, opponentSpell) {
		t.Fatal("an opponent's spell must not be protected by the controller-scoped buff")
	}
	if len(g.RuleEffects) != 1 {
		t.Fatal("the buff must not be consumed by an opponent's spell")
	}
}
