package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestFilterLandManaAbilityAddsTwoPairColorMana verifies the runtime behavior of
// the filter-land template "{W/U}, {T}: Add {W}{W}, {W}{U}, or {U}{U}.": paying
// one hybrid {W/U} mana (here floated as {W}) and tapping the source adds two
// mana, each one of the {W, U} pair, with no stack object.
func TestFilterLandManaAbilityAddsTwoPairColorMana(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// Float one {W} to pay the hybrid {W/U} activation cost.
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	body := game.TwoColorFilterManaAbility(mana.W, mana.U)
	source := addComplexManaAbilityPermanent(g, game.Player1,
		&game.CardDef{CardFace: game.CardFace{Name: "Mystic Gate", Types: []types.Card{types.Land}}},
		&body,
	)
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)

	// Illegal while tapped.
	source.Tapped = true
	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("filter mana ability was legal while source was tapped")
	}
	source.Tapped = false

	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(filter mana ability) = false, want true")
	}
	if !source.Tapped {
		t.Fatal("source was not tapped after filter activation")
	}
	pool := g.Players[game.Player1].ManaPool
	// Started with 1 mana, paid 1 for the cost, added 2: net 2 mana, all W or U.
	if got := pool.Amount(mana.W) + pool.Amount(mana.U); got != 2 {
		t.Fatalf("W+U mana = %d, want 2", got)
	}
	if got := pool.Total(); got != 2 {
		t.Fatalf("total mana = %d, want 2 (only W/U produced)", got)
	}
	if got := g.Stack.Size(); got != 0 {
		t.Fatalf("stack size = %d, want 0 for mana ability", got)
	}
}

// TestFilterLandManaAbilityRequiresPairMana verifies the filter ability cannot be
// activated without a {W} or {U} to pay its hybrid cost.
func TestFilterLandManaAbilityRequiresPairMana(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	body := game.TwoColorFilterManaAbility(mana.W, mana.U)
	source := addComplexManaAbilityPermanent(g, game.Player1,
		&game.CardDef{CardFace: game.CardFace{Name: "Mystic Gate", Types: []types.Card{types.Land}}},
		&body,
	)
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)
	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("filter mana ability was legal with no {W}/{U} to pay its hybrid cost")
	}
	if engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(filter ability without mana) = true, want false")
	}
	if source.Tapped {
		t.Fatal("source was tapped by failed payment")
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("mana pool = %d, want 0 after failed payment", g.Players[game.Player1].ManaPool.Total())
	}
}
