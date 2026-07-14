package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GideonSPhalanx is the card definition for Gideon's Phalanx.
//
// Type: Instant
// Cost: {5}{W}{W}
//
// Oracle text:
//
//	Create four 2/2 white Knight creature tokens with vigilance.
//	Spell mastery — If there are two or more instant and/or sorcery cards in your graveyard, creatures you control gain indestructible until end of turn.
var GideonSPhalanx = newGideonSPhalanx

func newGideonSPhalanx() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Gideon's Phalanx",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(4),
							Source: game.TokenDef(gideonSPhalanxToken),
						},
					},
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer: game.LayerAbility,
									Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									AddKeywords: []game.Keyword{
										game.Indestructible,
									},
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								ControllerGraveyardInstantOrSorceryCountAtLeast: 2,
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Create four 2/2 white Knight creature tokens with vigilance.
			Spell mastery — If there are two or more instant and/or sorcery cards in your graveyard, creatures you control gain indestructible until end of turn.
		`,
		},
	}
}

var gideonSPhalanxToken = newGideonSPhalanxToken()

func newGideonSPhalanxToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Knight",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Knight},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
			},
		},
	}
}
