package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SpellheartChimera is the card definition for Spellheart Chimera.
//
// Type: Creature — Chimera
// Cost: {1}{U}{R}
//
// Oracle text:
//
//	Flying, trample
//	Spellheart Chimera's power is equal to the number of instant and sorcery cards in your graveyard.
var SpellheartChimera = newSpellheartChimera

func newSpellheartChimera() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Spellheart Chimera",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.R,
			}),
			Colors:       []color.Color{color.Red, color.Blue},
			Types:        []types.Card{types.Creature},
			Subtypes:     []types.Sub{types.Chimera},
			Power:        opt.Val(game.PT{IsStar: true}),
			Toughness:    opt.Val(game.PT{Value: 3}),
			DynamicPower: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerInstantOrSorceryCardsInGraveyard}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.TrampleStaticBody,
			},
			OracleText: `
			Flying, trample
			Spellheart Chimera's power is equal to the number of instant and sorcery cards in your graveyard.
		`,
		},
	}
}
