package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DreamcallerSiren is the card definition for Dreamcaller Siren.
//
// Type: Creature — Siren Pirate
// Cost: {2}{U}{U}
//
// Oracle text:
//
//	Flash
//	Flying
//	This creature can block only creatures with flying.
//	When this creature enters, if you control another Pirate, tap up to two target nonland permanents.
var DreamcallerSiren = newDreamcallerSiren

func newDreamcallerSiren() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Dreamcaller Siren",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Siren, types.Pirate},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.FlyingStaticBody,
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCanBlockOnlyCreaturesWith,
							AffectedSource: true,
							BlockerRestriction: game.BlockerRestriction{
								Kind: game.BlockerRestrictionFlying,
							},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
						InterveningIf: "if you control another Pirate",
						InterveningCondition: opt.Val(game.Condition{
							ControlsMatching: opt.Val(game.SelectionCount{
								Selection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Pirate")}, ExcludeSource: true},
							}),
						}),
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 0,
								MaxTargets: 2,
								Constraint: "up to two target nonland permanents",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{ExcludedTypes: []types.Card{types.Land}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Tap{
									Object: game.TargetPermanentReference(0),
								},
							},
							{
								Primitive: game.Tap{
									Object: game.TargetPermanentReference(1),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flash
			Flying
			This creature can block only creatures with flying.
			When this creature enters, if you control another Pirate, tap up to two target nonland permanents.
		`,
		},
	}
}
