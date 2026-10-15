package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MetalworkColossus is the card definition for Metalwork Colossus.
//
// Type: Artifact Creature — Construct
// Cost: {11}
//
// Oracle text:
//
//	This spell costs {X} less to cast, where X is the total mana value of noncreature artifacts you control.
//	Sacrifice two artifacts: Return this card from your graveyard to your hand.
var MetalworkColossus = newMetalworkColossus

func newMetalworkColossus() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Metalwork Colossus",
			ManaCost: opt.Val(cost.Mana{
				cost.O(11),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Construct},
			Power:     opt.Val(game.PT{Value: 10}),
			Toughness: opt.Val(game.PT{Value: 10}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This spell costs {X} less to cast, where X is the total mana value of noncreature artifacts you control.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							CostModifier: game.CostModifier{
								Kind: game.CostModifierSpell,
								DynamicReduction: &game.DynamicAmount{
									Kind:        game.DynamicAmountTotalManaValueInGroup,
									Multiplier:  1,
									CounterKind: counter.PlusOnePlusOne,
									Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Artifact}, ExcludedTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
								},
							},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Sacrifice two artifacts: Return this card from your graveyard to your hand.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:               cost.AdditionalSacrifice,
							Text:               "Sacrifice two artifacts",
							Amount:             2,
							MatchPermanentType: true,
							PermanentType:      types.Artifact,
							CounterKind:        counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Graveyard,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveCard{
											Card: game.CardReference{
												Kind: game.CardReferenceSource,
											},
											FromZone:    zone.Graveyard,
											Destination: zone.Hand,
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
			OracleText: `
			This spell costs {X} less to cast, where X is the total mana value of noncreature artifacts you control.
			Sacrifice two artifacts: Return this card from your graveyard to your hand.
		`,
		},
	}
}
