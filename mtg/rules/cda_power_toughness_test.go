package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestDynamicStarPowerOnlyKeepsPrintedToughness covers the power-only
// characteristic-defining ability ("Adeline's power is equal to the number of
// creatures you control."): the printed toughness stands while the power tracks
// the live count.
func TestDynamicStarPowerOnlyKeepsPrintedToughness(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:         "Adeline",
		Types:        []types.Card{types.Creature},
		Power:        opt.Val(game.PT{IsStar: true}),
		Toughness:    opt.Val(game.PT{Value: 4}),
		DynamicPower: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerCreatureCount}),
	}})
	for range 2 {
		addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name: "Soldier", Types: []types.Card{types.Creature}}})
	}

	if got := effectivePower(g, creature); got != 3 {
		t.Fatalf("effective power = %d, want creature count 3 (self plus two)", got)
	}
	toughness, ok := effectiveToughness(g, creature)
	if !ok || toughness != 4 {
		t.Fatalf("effective toughness = %d (ok=%v), want printed 4", toughness, ok)
	}

	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Soldier", Types: []types.Card{types.Creature}}})
	if got := effectivePower(g, creature); got != 4 {
		t.Fatalf("effective power after extra creature = %d, want 4", got)
	}
	if toughness, ok := effectiveToughness(g, creature); !ok || toughness != 4 {
		t.Fatalf("effective toughness after extra creature = %d (ok=%v), want printed 4", toughness, ok)
	}
}

// TestDynamicStarToughnessOffsetTracksGraveyards covers the Tarmogoyf/Lhurgoyf
// form ("power is equal to the number of creature cards in all graveyards and
// its toughness is equal to that number plus 1."): power tracks the live count
// and toughness is that count plus one, updating as the count changes.
func TestDynamicStarToughnessOffsetTracksGraveyards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	count := game.DynamicValue{Kind: game.DynamicValueCreatureCardsInAllGraveyards}
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:             "Lhurgoyf",
		Types:            []types.Card{types.Creature},
		Power:            opt.Val(game.PT{IsStar: true}),
		Toughness:        opt.Val(game.PT{IsStar: true}),
		DynamicPower:     opt.Val(count),
		DynamicToughness: opt.Val(game.DynamicValue{Kind: count.Kind, Offset: 1}),
	}})
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Bear", Types: []types.Card{types.Creature}}})
	addCardToGraveyard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Elf", Types: []types.Card{types.Creature}}})
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Bolt", Types: []types.Card{types.Instant}}})

	if got := effectivePower(g, creature); got != 2 {
		t.Fatalf("effective power = %d, want 2 creature cards in all graveyards", got)
	}
	toughness, ok := effectiveToughness(g, creature)
	if !ok || toughness != 3 {
		t.Fatalf("effective toughness = %d (ok=%v), want count plus one (3)", toughness, ok)
	}

	addCardToGraveyard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Wolf", Types: []types.Card{types.Creature}}})
	if got := effectivePower(g, creature); got != 3 {
		t.Fatalf("effective power after extra creature card = %d, want 3", got)
	}
	if toughness, ok := effectiveToughness(g, creature); !ok || toughness != 4 {
		t.Fatalf("effective toughness after extra creature card = %d (ok=%v), want 4", toughness, ok)
	}
}
