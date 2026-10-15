package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EnigmaDrake is the card definition for Enigma Drake.
//
// Type: Creature — Drake
// Cost: {1}{U}{R}
//
// Oracle text:
//
//	Flying
//	Enigma Drake's power is equal to the number of instant and sorcery cards in your graveyard.
var EnigmaDrake = newEnigmaDrake

func newEnigmaDrake() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Enigma Drake",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.R,
			}),
			Colors:       []color.Color{color.Red, color.Blue},
			Types:        []types.Card{types.Creature},
			Subtypes:     []types.Sub{types.Drake},
			Power:        opt.Val(game.PT{IsStar: true}),
			Toughness:    opt.Val(game.PT{Value: 4}),
			DynamicPower: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerInstantOrSorceryCardsInGraveyard}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			OracleText: `
			Flying
			Enigma Drake's power is equal to the number of instant and sorcery cards in your graveyard.
		`,
		},
	}
}
