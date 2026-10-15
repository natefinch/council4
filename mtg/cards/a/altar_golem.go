package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AltarGolem is the card definition for Altar Golem.
//
// Type: Artifact Creature — Golem
// Cost: {7}
//
// Oracle text:
//
//	Trample
//	Altar Golem's power and toughness are each equal to the number of creatures on the battlefield.
//	This creature doesn't untap during your untap step.
//	Tap five untapped creatures you control: Untap this creature.
var AltarGolem = newAltarGolem

func newAltarGolem() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Altar Golem",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
			}),
			Types:            []types.Card{types.Artifact, types.Creature},
			Subtypes:         []types.Sub{types.Golem},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueAllBattlefieldCreatureCount}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueAllBattlefieldCreatureCount}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
				game.StaticAbility{
					Text: "This creature doesn't untap during your untap step.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectDoesntUntap,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Tap five untapped creatures you control: Untap this creature.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:               cost.AdditionalTapPermanents,
							Text:               "Tap five untapped creatures you control",
							Amount:             5,
							MatchPermanentType: true,
							PermanentType:      types.Creature,
							CounterKind:        counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Untap{
											Object: game.SourcePermanentReference(),
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
			Trample
			Altar Golem's power and toughness are each equal to the number of creatures on the battlefield.
			This creature doesn't untap during your untap step.
			Tap five untapped creatures you control: Untap this creature.
		`,
		},
	}
}
