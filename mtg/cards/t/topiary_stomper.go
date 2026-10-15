package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TopiaryStomper is the card definition for Topiary Stomper.
//
// Type: Creature — Plant Dinosaur
// Cost: {1}{G}{G}
//
// Oracle text:
//
//	Vigilance (Attacking doesn't cause this creature to tap.)
//	When this creature enters, search your library for a basic land card, put it onto the battlefield tapped, then shuffle.
//	This creature can't attack or block unless you control seven or more lands.
var TopiaryStomper = newTopiaryStomper

func newTopiaryStomper() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Topiary Stomper",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Plant, types.Dinosaur},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
				game.StaticAbility{
					Text: "This creature can't attack or block unless you control seven or more lands.",
					Condition: opt.Val(game.Condition{
						Text:   "unless you control seven or more lands",
						Negate: true,
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Land}},
							MinCount:  7,
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantAttack,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
						},
						game.RuleEffect{
							Kind:           game.RuleEffectCantBlock,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature enters, search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Search{
											Player: game.ControllerReference(),
											Spec: game.SearchSpec{
												SourceZone:   zone.Library,
												Destination:  zone.Battlefield,
												Filter:       game.Selection{RequiredTypes: []types.Card{types.Land}, Supertypes: []types.Super{types.Basic}},
												EntersTapped: true,
											},
											Amount: game.Fixed(1),
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
			Vigilance (Attacking doesn't cause this creature to tap.)
			When this creature enters, search your library for a basic land card, put it onto the battlefield tapped, then shuffle.
			This creature can't attack or block unless you control seven or more lands.
		`,
		},
	}
}
