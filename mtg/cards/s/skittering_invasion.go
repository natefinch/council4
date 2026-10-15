package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SkitteringInvasion is the card definition for Skittering Invasion.
//
// Type: Kindred Sorcery — Eldrazi
// Cost: {7}
//
// Oracle text:
//
//	Create five 0/1 colorless Eldrazi Spawn creature tokens. They have "Sacrifice this token: Add {C}."
var SkitteringInvasion = newSkitteringInvasion

func newSkitteringInvasion() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Skittering Invasion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
			}),
			Types:    []types.Card{types.Kindred, types.Sorcery},
			Subtypes: []types.Sub{types.Eldrazi},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(5),
							Source: game.TokenDef(skitteringInvasionToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create five 0/1 colorless Eldrazi Spawn creature tokens. They have "Sacrifice this token: Add {C}."
		`,
		},
	}
}

var skitteringInvasionToken = newSkitteringInvasionToken()

func newSkitteringInvasionToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Eldrazi Spawn",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi, types.Spawn},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					Text: "Sacrifice this token: Add {C}.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalSacrificeSource,
							Text:        "Sacrifice this token",
							Amount:      1,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddMana{
											Amount:    game.Fixed(1),
											ManaColor: mana.C,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
		},
	}
}
