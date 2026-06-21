package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestEntersAsCopyOverlaysChosenPermanentValues(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	dragonPT := game.PT{Value: 4}
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:            "Shivan Dragon",
		Types:           []types.Card{types.Creature},
		Subtypes:        []types.Sub{types.Dragon},
		Power:           opt.Val(dragonPT),
		Toughness:       opt.Val(dragonPT),
		StaticAbilities: []game.StaticAbility{{Text: "Flying", KeywordAbilities: game.SimpleKeywords(game.Flying)}},
	}})
	clone := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Clone",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 0}),
	}})

	replacement := game.EntersAsCopyReplacement(
		"You may have Clone enter the battlefield as a copy of any creature on the battlefield.",
		&game.Selection{RequiredTypes: []types.Card{types.Creature}},
		true, false,
	)
	applyEntersAsCopy(enterBattlefieldContext{}, g, clone, &replacement.Replacement)

	if got := permanentEffectiveName(g, clone); got != "Shivan Dragon" {
		t.Fatalf("effective name = %q, want Shivan Dragon", got)
	}
	if got := effectivePower(g, clone); got != 4 {
		t.Fatalf("effective power = %d, want copied 4", got)
	}
	if !hasKeyword(g, clone, game.Flying) {
		t.Fatal("copy did not grant copied Flying keyword")
	}
}

func TestEntersAsCopyNotLegendaryRiderDropsLegendary(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	pt := game.PT{Value: 3}
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:       "Legendary Bear",
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Power:      opt.Val(pt),
		Toughness:  opt.Val(pt),
	}})
	clone := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Spark Double",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 0}),
	}})

	replacement := game.EntersAsCopyReplacement(
		"copy text",
		&game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou},
		false, true,
	)
	applyEntersAsCopy(enterBattlefieldContext{}, g, clone, &replacement.Replacement)

	values := effectivePermanentValues(g, clone)
	for _, super := range values.supertypes {
		if super == types.Legendary {
			t.Fatal("not-legendary rider failed to drop the legendary supertype")
		}
	}
	if got := effectivePower(g, clone); got != 3 {
		t.Fatalf("effective power = %d, want copied 3", got)
	}
}
