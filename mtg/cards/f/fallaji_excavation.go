package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FallajiExcavation is the card definition for Fallaji Excavation.
//
// Type: Sorcery
// Cost: {3}{G}{G}
//
// Oracle text:
//
//	Create three tapped Powerstone tokens. You gain 3 life. (The tokens are artifacts with "{T}: Add {C}. This mana can't be spent to cast a nonartifact spell.")
var FallajiExcavation = newFallajiExcavation

func newFallajiExcavation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Fallaji Excavation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount:      game.Fixed(3),
							Source:      game.TokenDef(fallajiExcavationToken),
							EntryTapped: true,
						},
					},
					{
						Primitive: game.GainLife{
							Amount: game.Fixed(3),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create three tapped Powerstone tokens. You gain 3 life. (The tokens are artifacts with "{T}: Add {C}. This mana can't be spent to cast a nonartifact spell.")
		`,
		},
	}
}

var fallajiExcavationToken = newFallajiExcavationToken()

func newFallajiExcavationToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Powerstone",
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Powerstone},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					Text: "{T}: Add {C}. This mana can't be spent to cast a nonartifact spell.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
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
											SpendRider: opt.Val(game.ManaSpendRider{
												Condition:   game.ManaSpendCastArtifactSpell,
												Restriction: game.ManaSpendRestrictedToCondition,
											}),
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
