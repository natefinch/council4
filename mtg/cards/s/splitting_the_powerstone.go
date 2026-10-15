package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SplittingThePowerstone is the card definition for Splitting the Powerstone.
//
// Type: Sorcery
// Cost: {2}{U}
//
// Oracle text:
//
//	As an additional cost to cast this spell, sacrifice an artifact.
//	Create two tapped Powerstone tokens. If the sacrificed artifact was legendary, draw a card. (The tokens are artifacts with "{T}: Add {C}. This mana can't be spent to cast a nonartifact spell.")
var SplittingThePowerstone = newSplittingThePowerstone

func newSplittingThePowerstone() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Splitting the Powerstone",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			AdditionalCosts: []cost.Additional{
				{
					Kind:               cost.AdditionalSacrifice,
					Text:               "sacrifice an artifact",
					Amount:             1,
					MatchPermanentType: true,
					PermanentType:      types.Artifact,
					SubjectKey:         "paid-cost-1-1",
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount:      game.Fixed(2),
							Source:      game.TokenDef(splittingThePowerstoneToken),
							EntryTapped: true,
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Object:        opt.Val(game.PaidCostReference("paid-cost-1-1", game.PaidCostSacrifice, types.Artifact)),
								ObjectMatches: opt.Val(game.Selection{Supertypes: []types.Super{types.Legendary}}),
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			As an additional cost to cast this spell, sacrifice an artifact.
			Create two tapped Powerstone tokens. If the sacrificed artifact was legendary, draw a card. (The tokens are artifacts with "{T}: Add {C}. This mana can't be spent to cast a nonartifact spell.")
		`,
		},
	}
}

var splittingThePowerstoneToken = newSplittingThePowerstoneToken()

func newSplittingThePowerstoneToken() *game.CardDef {
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
