package rules

import (
	"math/rand/v2"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestResolutionLogsEnterAndStartEntry plays a deck of creatures and asserts that
// when a creature spell resolves, the entering permanent is logged within the
// resolution's [StartEntry, resolve) span, so a report can nest it under the
// resolution that caused it.
func TestResolutionLogsEnterAndStartEntry(t *testing.T) {
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

	sawEnterInsideResolution := false
	for _, turn := range result.Turns {
		for i, entry := range turn.Entries {
			if entry.Kind != TurnLogEntryResolve {
				continue
			}
			resolve := entry.Resolve
			if resolve.StartEntry < 0 || resolve.StartEntry > i {
				t.Fatalf("resolve StartEntry %d out of range for resolve at %d", resolve.StartEntry, i)
			}
			for j := resolve.StartEntry; j < i; j++ {
				if turn.Entries[j].Kind == TurnLogEntryEnter {
					sawEnterInsideResolution = true
				}
			}
		}
	}
	if !sawEnterInsideResolution {
		t.Fatal("no permanent-entered entry was logged inside a resolution span")
	}
}
