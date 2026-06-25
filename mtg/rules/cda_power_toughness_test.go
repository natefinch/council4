package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
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

// TestDynamicStarLandSubtypeCount covers "the number of <BasicLand> you control"
// (Korlash, Dungrove Elder): the power/toughness equals the count of controlled
// lands with the named subtype and updates as lands change.
func TestDynamicStarLandSubtypeCount(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	count := game.DynamicValue{Kind: game.DynamicValueControllerSubtypeCount, Subtype: types.Swamp}
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:             "Korlash",
		Types:            []types.Card{types.Creature},
		Power:            opt.Val(game.PT{IsStar: true}),
		Toughness:        opt.Val(game.PT{IsStar: true}),
		DynamicPower:     opt.Val(count),
		DynamicToughness: opt.Val(count),
	}})
	addLandPermanent(g, game.Player1, "Swamp", types.Swamp)
	addLandPermanent(g, game.Player1, "Overgrown Tomb", types.Swamp, types.Forest)
	addLandPermanent(g, game.Player1, "Forest", types.Forest)
	addLandPermanent(g, game.Player2, "Swamp", types.Swamp)

	if got := effectivePower(g, creature); got != 2 {
		t.Fatalf("effective power = %d, want 2 Swamps you control", got)
	}
	if toughness, ok := effectiveToughness(g, creature); !ok || toughness != 2 {
		t.Fatalf("effective toughness = %d (ok=%v), want 2", toughness, ok)
	}

	addLandPermanent(g, game.Player1, "Swamp", types.Swamp)
	if got := effectivePower(g, creature); got != 3 {
		t.Fatalf("effective power after extra Swamp = %d, want 3", got)
	}
}

// TestDynamicStarCreatureSubtypeCount covers "the number of <Subtype> you
// control" for a non-land creature subtype: the power/toughness equals the count
// of controlled permanents with the named subtype and updates as they change.
func TestDynamicStarCreatureSubtypeCount(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	count := game.DynamicValue{Kind: game.DynamicValueControllerSubtypeCount, Subtype: types.Goblin}
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:             "Goblin Anthemist",
		Types:            []types.Card{types.Creature},
		Subtypes:         []types.Sub{types.Goblin},
		Power:            opt.Val(game.PT{IsStar: true}),
		Toughness:        opt.Val(game.PT{IsStar: true}),
		DynamicPower:     opt.Val(count),
		DynamicToughness: opt.Val(count),
	}})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Goblin Token",
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Goblin},
	}})
	addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:     "Enemy Goblin",
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Goblin},
	}})

	if got := effectivePower(g, creature); got != 2 {
		t.Fatalf("effective power = %d, want 2 Goblins you control", got)
	}
	if toughness, ok := effectiveToughness(g, creature); !ok || toughness != 2 {
		t.Fatalf("effective toughness = %d (ok=%v), want 2", toughness, ok)
	}

	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Another Goblin",
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Goblin},
	}})
	if got := effectivePower(g, creature); got != 3 {
		t.Fatalf("effective power after extra Goblin = %d, want 3", got)
	}
}

// TestDynamicStarColorPermanentCount covers "the number of <color> permanents
// you control": the power/toughness equals the count of controlled permanents
// whose printed colors include the named color and updates as they change.
func TestDynamicStarColorPermanentCount(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	count := game.DynamicValue{Kind: game.DynamicValueControllerColorPermanentCount, Color: color.Red}
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:             "Crimson Avatar",
		Types:            []types.Card{types.Creature},
		Colors:           []color.Color{color.Red},
		Power:            opt.Val(game.PT{IsStar: true}),
		Toughness:        opt.Val(game.PT{IsStar: true}),
		DynamicPower:     opt.Val(count),
		DynamicToughness: opt.Val(count),
	}})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:   "Red Artifact",
		Types:  []types.Card{types.Artifact},
		Colors: []color.Color{color.Red},
	}})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:   "Blue Creature",
		Types:  []types.Card{types.Creature},
		Colors: []color.Color{color.Blue},
	}})
	addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:   "Enemy Red",
		Types:  []types.Card{types.Creature},
		Colors: []color.Color{color.Red},
	}})

	if got := effectivePower(g, creature); got != 2 {
		t.Fatalf("effective power = %d, want 2 red permanents you control", got)
	}
	if toughness, ok := effectiveToughness(g, creature); !ok || toughness != 2 {
		t.Fatalf("effective toughness = %d (ok=%v), want 2", toughness, ok)
	}

	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:   "Red Enchantment",
		Types:  []types.Card{types.Enchantment},
		Colors: []color.Color{color.Red},
	}})
	if got := effectivePower(g, creature); got != 3 {
		t.Fatalf("effective power after extra red permanent = %d, want 3", got)
	}
}

// TestDynamicStarInstantOrSorceryCardsInGraveyard covers "the number of instant
// and sorcery cards in your graveyard" (Haughty Djinn): only the controller's
// instant and sorcery cards count.
func TestDynamicStarInstantOrSorceryCardsInGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	count := game.DynamicValue{Kind: game.DynamicValueControllerInstantOrSorceryCardsInGraveyard}
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:         "Haughty Djinn",
		Types:        []types.Card{types.Creature},
		Power:        opt.Val(game.PT{IsStar: true}),
		Toughness:    opt.Val(game.PT{Value: 4}),
		DynamicPower: opt.Val(count),
	}})
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Bolt", Types: []types.Card{types.Instant}}})
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Divination", Types: []types.Card{types.Sorcery}}})
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Bear", Types: []types.Card{types.Creature}}})
	addCardToGraveyard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Shock", Types: []types.Card{types.Instant}}})

	if got := effectivePower(g, creature); got != 2 {
		t.Fatalf("effective power = %d, want 2 instant/sorcery in your graveyard", got)
	}
}

// TestDynamicStarControllerLifeTotal covers "your life total" (Soul of
// Eternity): the power/toughness equals the controller's current life.
func TestDynamicStarControllerLifeTotal(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	count := game.DynamicValue{Kind: game.DynamicValueControllerLifeTotal}
	g.Players[game.Player1].Life = 17
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:             "Soul of Eternity",
		Types:            []types.Card{types.Creature},
		Power:            opt.Val(game.PT{IsStar: true}),
		Toughness:        opt.Val(game.PT{IsStar: true}),
		DynamicPower:     opt.Val(count),
		DynamicToughness: opt.Val(count),
	}})

	if got := effectivePower(g, creature); got != 17 {
		t.Fatalf("effective power = %d, want life total 17", got)
	}
	g.Players[game.Player1].Life = 25
	if got := effectivePower(g, creature); got != 25 {
		t.Fatalf("effective power after life change = %d, want 25", got)
	}
}

// TestDynamicStarBasicLandTypeCount covers "the number of basic land types among
// lands you control" (Territorial Kavu): the count is the number of distinct
// basic land subtypes present among the controller's lands.
func TestDynamicStarBasicLandTypeCount(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	count := game.DynamicValue{Kind: game.DynamicValueControllerBasicLandTypeCount}
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:             "Territorial Kavu",
		Types:            []types.Card{types.Creature},
		Power:            opt.Val(game.PT{IsStar: true}),
		Toughness:        opt.Val(game.PT{IsStar: true}),
		DynamicPower:     opt.Val(count),
		DynamicToughness: opt.Val(count),
	}})
	addLandPermanent(g, game.Player1, "Swamp", types.Swamp)
	addLandPermanent(g, game.Player1, "Forest", types.Forest)
	addLandPermanent(g, game.Player1, "Overgrown Tomb", types.Swamp, types.Forest)

	if got := effectivePower(g, creature); got != 2 {
		t.Fatalf("effective power = %d, want 2 distinct basic land types", got)
	}
	addLandPermanent(g, game.Player1, "Island", types.Island)
	if got := effectivePower(g, creature); got != 3 {
		t.Fatalf("effective power after adding Island = %d, want 3", got)
	}
}

// TestDynamicStarAllPlayersHandSize covers "the total number of cards in all
// players' hands" (Multani): the count sums every player's hand size.
func TestDynamicStarAllPlayersHandSize(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	count := game.DynamicValue{Kind: game.DynamicValueAllPlayersHandSize}
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:             "Multani",
		Types:            []types.Card{types.Creature},
		Power:            opt.Val(game.PT{IsStar: true}),
		Toughness:        opt.Val(game.PT{IsStar: true}),
		DynamicPower:     opt.Val(count),
		DynamicToughness: opt.Val(count),
	}})
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "A"}})
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "B"}})
	addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "C"}})

	if got := effectivePower(g, creature); got != 3 {
		t.Fatalf("effective power = %d, want 3 cards in all players' hands", got)
	}
}
