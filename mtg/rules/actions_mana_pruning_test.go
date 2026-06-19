package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestLegalActionsPrunePaymentOnlyManaAbilities(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	for range 40 {
		addBasicLandPermanent(g, game.Player1, types.Forest)
	}
	setMainPhasePriority(g, game.Player1)

	actions := engine.legalActions(g, game.Player1)
	if len(actions) != 1 || actions[0].Kind != action.ActionPass {
		t.Fatalf("legal actions = %+v, want only Pass", actions)
	}
}

func TestLegalActionsRetainComplexManaAbility(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	body := painlandColoredManaAbility(mana.B, 1)
	source := addComplexManaAbilityPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Pain Land",
		Types: []types.Card{types.Land},
	}}, &body)
	setMainPhasePriority(g, game.Player1)

	want := action.ActivateAbility(source.ObjectID, 0, nil, 0)
	if !containsAction(engine.legalActions(g, game.Player1), want) {
		t.Fatal("complex mana ability was pruned from legal actions")
	}
}

func TestPrunedManaAbilityStillPaysForSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:         "Green Spell",
		ManaCost:     opt.Val(cost.Mana{cost.G}),
		Types:        []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.AbilityContent{}),
	}})
	setMainPhasePriority(g, game.Player1)

	cast := action.CastSpell(spellID, nil, 0, nil)
	if !containsAction(engine.legalActions(g, game.Player1), cast) {
		t.Fatal("spell payable by a pruned mana ability was not legal")
	}
	if !engine.applyAction(g, game.Player1, cast) {
		t.Fatal("casting with automatic mana payment failed")
	}
	if !forest.Tapped {
		t.Fatal("automatic payment did not tap the Forest")
	}
}

func BenchmarkLegalActionsLandHeavy(b *testing.B) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	for range 99 {
		addBasicLandPermanent(g, game.Player1, types.Forest)
	}
	setMainPhasePriority(g, game.Player1)

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = engine.legalActions(g, game.Player1)
	}
}
