package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RavenClanWarAxe is the card definition for Raven Clan War-Axe.
//
// Type: Artifact — Equipment
// Cost: {1}{R}{W}
//
// Oracle text:
//
//	When this Equipment enters, you may search your library and/or graveyard for a card named Eivor, Battle-Ready, reveal it, and put it into your hand. If you search your library this way, shuffle.
//	Equipped creature gets +2/+0 and has trample.
//	Equip {2}
var RavenClanWarAxe = newRavenClanWarAxe

func newRavenClanWarAxe() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name: "Raven Clan War-Axe",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.W,
			}),
			Colors:   []color.Color{color.Red, color.White},
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:      game.LayerPowerToughnessModify,
							Group:      game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta: 2,
						},
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddKeywords: []game.Keyword{
								game.Trample,
							},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(2)}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Search{
									Player: game.ControllerReference(),
									Spec: game.SearchSpec{
										SourceZone:    zone.Library,
										Destination:   zone.Hand,
										Name:          "Eivor, Battle-Ready",
										Reveal:        true,
										AlsoGraveyard: true,
									},
									Amount: game.Fixed(1),
								},
								Optional:      true,
								OptionalActor: opt.Val(game.ControllerReference()),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this Equipment enters, you may search your library and/or graveyard for a card named Eivor, Battle-Ready, reveal it, and put it into your hand. If you search your library this way, shuffle.
			Equipped creature gets +2/+0 and has trample.
			Equip {2}
		`,
		},
	}
}
