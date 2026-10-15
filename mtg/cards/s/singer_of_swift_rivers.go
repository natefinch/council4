package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SingerOfSwiftRivers is the card definition for Singer of Swift Rivers.
//
// Type: Creature — Merfolk Shaman
// Cost: {1}{G}{U}
//
// Oracle text:
//
//	Flash
//	When this creature enters, put a shield counter on another target creature you control. (If it would be dealt damage or destroyed, remove a shield counter from it instead.)
//	You may cast Merfolk spells as though they had flash.
var SingerOfSwiftRivers = newSingerOfSwiftRivers

func newSingerOfSwiftRivers() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Singer of Swift Rivers",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.U,
			}),
			Colors:    []color.Color{color.Green, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Merfolk, types.Shaman},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.StaticAbility{
					Text: "You may cast Merfolk spells as though they had flash.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCastSpellsAsThoughFlash,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							SpellSubtypes:  []types.Sub{types.Merfolk},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature enters, put a shield counter on another target creature you control. (If it would be dealt damage or destroyed, remove a shield counter from it instead.)",
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
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "another target creature you control",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, ExcludeSource: true, Controller: game.ControllerYou}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.TargetPermanentReference(0),
											CounterKind: counter.Shield,
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
			Flash
			When this creature enters, put a shield counter on another target creature you control. (If it would be dealt damage or destroyed, remove a shield counter from it instead.)
			You may cast Merfolk spells as though they had flash.
		`,
		},
	}
}
