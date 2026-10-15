package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VexilusPraetor is the card definition for Vexilus Praetor.
//
// Type: Creature — Custodes Warrior
// Cost: {3}{W}
//
// Oracle text:
//
//	Flash
//	Vigilance
//	Aegis of the Emperor — Commanders you control have protection from everything.
var VexilusPraetor = newVexilusPraetor

func newVexilusPraetor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Vexilus Praetor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Custodes, types.Warrior},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.VigilanceStaticBody,
				game.StaticAbility{
					Text: "Aegis of the Emperor — Commanders you control have protection from everything.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{MatchCommander: true}),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								new(game.ProtectionFromEverythingStaticAbility()),
							},
						},
					},
				},
			},
			OracleText: `
			Flash
			Vigilance
			Aegis of the Emperor — Commanders you control have protection from everything.
		`,
		},
	}
}
