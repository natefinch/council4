package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// VarragothBloodskySire is the card definition for Varragoth, Bloodsky Sire.
//
// Type: Legendary Creature — Demon Rogue
// Cost: {2}{B}
//
// Oracle text:
//
//	Deathtouch
//	Boast — {1}{B}: Target player searches their library for a card, then shuffles and puts that card on top. (Activate only if this creature attacked this turn and only once each turn.)
var VarragothBloodskySire = newVarragothBloodskySire

func newVarragothBloodskySire() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Varragoth, Bloodsky Sire",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Demon, types.Rogue},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.DeathtouchStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Boast — {1}{B}: Target player searches their library for a card, then shuffles and puts that card on top. (Activate only if this creature attacked this turn and only once each turn.)",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
						cost.B,
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
										Constraint: "Target player",
										Allow:      game.TargetAllowPlayer,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Search{
											Player: game.TargetPlayerReference(0),
											Spec: game.SearchSpec{
												SourceZone:          zone.Library,
												Destination:         zone.Library,
												DestinationPosition: game.SearchPositionTop,
												FailToFindPolicy:    game.SearchMustFindIfAvailable,
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
			Deathtouch
			Boast — {1}{B}: Target player searches their library for a card, then shuffles and puts that card on top. (Activate only if this creature attacked this turn and only once each turn.)
		`,
		},
	}
}
