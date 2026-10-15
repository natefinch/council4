package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FirecatBlitz is the card definition for Firecat Blitz.
//
// Type: Sorcery
// Cost: {X}{R}{R}
//
// Oracle text:
//
//	Create X 1/1 red Elemental Cat creature tokens with haste. Exile them at the beginning of the next end step.
//	Flashback—{R}{R}, Sacrifice X Mountains. (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var FirecatBlitz = newFirecatBlitz

func newFirecatBlitz() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Firecat Blitz",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.SimpleKeyword{
							Kind: game.Flashback,
						},
					},
				},
			},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:    "Flashback",
					ManaCost: opt.Val(cost.Mana{cost.R, cost.R}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalSacrifice,
							Text:        "Sacrifice X Mountains",
							AmountFromX: true,
							SubtypesAny: cost.SubtypeSet{types.Mountain},
						},
					},
					Mechanic: cost.AlternativeMechanicFlashback,
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							}),
							Source:        game.TokenDef(firecatBlitzToken),
							PublishLinked: game.LinkedKey("sequence-effect-0-product"),
						},
						LocalProducts: game.LocalProducts{
							Links: []game.LinkedKey{game.LinkedKey("sequence-effect-0-product")},
						},
					},
					{
						Primitive: game.CreateDelayedTrigger{
							Trigger: game.DelayedTriggerDef{
								Timing: game.DelayedAtBeginningOfNextEndStep,
								Content: game.AbilityContent{
									Modes: []game.Mode{
										game.Mode{
											Sequence: []game.Instruction{
												game.Instruction{
													Primitive: game.MovePermanent{
														Group:       game.CapturedObjectsGroup(),
														Destination: zone.Exile,
													},
												},
											},
										},
									},
									MinModes: 1,
									MaxModes: 1,
								},
								CapturedObjectGroup: opt.Val(game.LinkedObjectReference("sequence-effect-0-product")),
							},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create X 1/1 red Elemental Cat creature tokens with haste. Exile them at the beginning of the next end step.
			Flashback—{R}{R}, Sacrifice X Mountains. (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}

var firecatBlitzToken = newFirecatBlitzToken()

func newFirecatBlitzToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Elemental Cat",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental, types.Cat},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
			},
		},
	}
}
