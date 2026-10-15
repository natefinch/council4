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

// Gloomshrieker is the card definition for Gloomshrieker.
//
// Type: Enchantment Creature — Cat Beast
// Cost: {1}{B}{G}
//
// Oracle text:
//
//	Menace (This creature can't be blocked except by two or more creatures.)
//	When this creature enters, return target permanent card from your graveyard to your hand.
//	If this creature would die, exile it instead.
var Gloomshrieker = newGloomshrieker

func newGloomshrieker() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Gloomshrieker",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
				cost.G,
			}),
			Colors:    []color.Color{color.Black, color.Green},
			Types:     []types.Card{types.Enchantment, types.Creature},
			Subtypes:  []types.Sub{types.Cat, types.Beast},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.MenaceStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature enters, return target permanent card from your graveyard to your hand.",
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
										Constraint: "target permanent card from your graveyard",
										Allow:      game.TargetAllowCard,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Land, types.Planeswalker, types.Battle}, Controller: game.ControllerYou}),
										TargetZone: zone.Graveyard,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveCard{
											Card: game.CardReference{
												Kind: game.CardReferenceTarget,
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
			ReplacementAbilities: []game.ReplacementAbility{
				game.ReplacementAbility{Text: "If this creature would die, exile it instead.", Replacement: game.ReplacementEffect{
					MatchEvent:    game.EventZoneChanged,
					MatchToZone:   true,
					ToZone:        zone.Graveyard,
					ReplaceToZone: zone.Exile,
					Duration:      game.DurationPermanent,
					MatchFromZone: true,
					FromZone:      zone.Battlefield,
				}},
			},
			OracleText: `
			Menace (This creature can't be blocked except by two or more creatures.)
			When this creature enters, return target permanent card from your graveyard to your hand.
			If this creature would die, exile it instead.
		`,
		},
	}
}
