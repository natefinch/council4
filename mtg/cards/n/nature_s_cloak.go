package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NatureSCloak is the card definition for Nature's Cloak.
//
// Type: Sorcery
// Cost: {2}{G}
//
// Oracle text:
//
//	Green creatures you control gain forestwalk until end of turn. (They can't be blocked as long as defending player controls a Forest.)
var NatureSCloak = newNatureSCloak

func newNatureSCloak() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Nature's Cloak",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer: game.LayerAbility,
									Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ColorsAny: []color.Color{color.Green}, Controller: game.ControllerYou}),
									AddAbilities: []game.Ability{
										new(game.StaticAbility{
											KeywordAbilities: []game.KeywordAbility{
												game.LandwalkKeyword{Subtype: types.Forest},
											},
										}),
									},
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Green creatures you control gain forestwalk until end of turn. (They can't be blocked as long as defending player controls a Forest.)
		`,
		},
	}
}
