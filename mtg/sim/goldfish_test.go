package sim

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestRunGoldfishIsDeterministic(t *testing.T) {
	commander := &game.CardDef{CardFace: game.CardFace{
		Name:       "Goldfish Commander",
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
	}}
	forest := &game.CardDef{CardFace: game.CardFace{
		Name:       "Forest",
		Supertypes: []types.Super{types.Basic},
		Types:      []types.Card{types.Land},
	}}
	deck := make([]*game.CardDef, 99)
	for index := range deck {
		deck[index] = forest
	}
	cfg := GoldfishConfig{
		Player:   game.PlayerConfig{Name: "Goldfish", Commander: commander, Deck: deck},
		Seed:     42,
		MaxTurns: 10,
	}
	first := RunGoldfish(cfg)
	second := RunGoldfish(cfg)
	if !reflect.DeepEqual(first, second) {
		t.Fatal("same goldfish config produced different results")
	}
	if first.TurnCount != 10 || !first.TurnLimitReached {
		t.Fatalf("result = %d turns, limit=%v", first.TurnCount, first.TurnLimitReached)
	}
}
