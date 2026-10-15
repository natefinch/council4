package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GoldmawChampion is the card definition for Goldmaw Champion.
//
// Type: Creature — Dwarf Warrior
// Cost: {2}{W}
//
// Oracle text:
//
//	Boast — {1}{W}: Tap target creature. (Activate only if this creature attacked this turn and only once each turn.)
var GoldmawChampion = newGoldmawChampion

func newGoldmawChampion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Goldmaw Champion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dwarf, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Boast — {1}{W}: Tap target creature. (Activate only if this creature attacked this turn and only once each turn.)",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
						cost.W,
					}),
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.OncePerTurn,
					ActivationCondition: opt.Val(game.Condition{
						SourceCounterKind: counter.PlusOnePlusOne,
						EventHistory: opt.Val(game.EventHistoryCondition{
							Pattern: game.TriggerPattern{
								Event:           game.EventAttackerDeclared,
								Source:          game.TriggerSourceSelf,
								StackObjectKind: game.StackSpell,
								CounterKind:     counter.PlusOnePlusOne,
							},
							Window: game.EventHistoryCurrentTurn,
						}),
					}),
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target creature",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Tap{
											Object: game.TargetPermanentReference(0),
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
			Boast — {1}{W}: Tap target creature. (Activate only if this creature attacked this turn and only once each turn.)
		`,
		},
	}
}
