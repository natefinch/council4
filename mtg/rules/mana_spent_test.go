package rules

import (
	"math/rand/v2"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// castingAgent plays a land when it can, then casts any spell it can, so a
// goldfish run actually spends mana.
type castingAgent struct{}

func (castingAgent) ChooseAction(_ PlayerObservation, legal []action.Action) action.Action {
	for _, a := range legal {
		if a.Kind == action.ActionPlayLand {
			return a
		}
	}
	for _, a := range legal {
		if a.Kind == action.ActionCastSpell {
			return a
		}
	}
	return legal[0]
}

func TestManaSpentMatchesSpellsCast(t *testing.T) {
	commander := &game.CardDef{CardFace: game.CardFace{
		Name:       "Green Commander",
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		ManaCost:   opt.Val(cost.Mana{cost.G}),
		Power:      opt.Val(game.PT{Value: 1}),
		Toughness:  opt.Val(game.PT{Value: 1}),
	}}
	forest := &game.CardDef{CardFace: game.CardFace{
		Name:          "Forest",
		Supertypes:    []types.Super{types.Basic},
		Types:         []types.Card{types.Land},
		ManaAbilities: []game.ManaAbility{game.TapManaAbility(mana.G)},
	}}
	bear := &game.CardDef{CardFace: game.CardFace{
		Name:      "One-Drop Bear",
		Types:     []types.Card{types.Creature},
		ManaCost:  opt.Val(cost.Mana{cost.G}),
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}}

	deck := make([]*game.CardDef, 0, 99)
	for range 50 {
		deck = append(deck, forest)
	}
	for range 49 {
		deck = append(deck, bear)
	}
	config := game.PlayerConfig{Name: "Goldfish", Commander: commander, Deck: deck}

	engine := NewEngine(rand.New(rand.NewPCG(7, 11)))
	g := engine.NewGoldfishGame(config)
	result := engine.RunGoldfish(g, castingAgent{}, 8)

	totalSpent := 0
	totalCasts := 0
	for _, turn := range result.Turns {
		totalSpent += turn.ManaSpent
		for _, a := range turn.Actions {
			if a.Player == game.Player1 && a.Action.Kind == action.ActionCastSpell {
				totalCasts++
			}
		}
	}
	if totalCasts == 0 {
		t.Fatal("agent never cast a spell; cannot validate mana spent")
	}
	// Every spell (the commander and every bear) costs exactly one mana, the
	// commander is only cast once (no removal, so no tax), and nothing else
	// spends mana, so total mana spent equals the number of spells cast.
	if totalSpent != totalCasts {
		t.Fatalf("total mana spent = %d, want %d (one per spell cast)", totalSpent, totalCasts)
	}
}
