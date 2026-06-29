package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BarbedSpike is the card definition for Barbed Spike.
//
// Type: Artifact — Equipment
// Cost: {1}{W}
//
// Oracle text:
//
//	When this Equipment enters, create a 1/1 colorless Thopter artifact creature token with flying, then attach this Equipment to it.
//	Equipped creature gets +1/+0.
//	Equip {2}
var BarbedSpike = newBarbedSpike()

func newBarbedSpike() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Barbed Spike",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:   []color.Color{color.White},
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:      game.LayerPowerToughnessModify,
							Group:      game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta: 1,
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
								Primitive: game.CreateToken{
									Amount:        game.Fixed(1),
									Source:        game.TokenDef(barbedSpikeToken),
									PublishLinked: game.LinkedKey("created-token"),
								},
							},
							{
								Primitive: game.Attach{
									Attachment: game.SourcePermanentReference(),
									Target:     game.LinkedObjectReference("created-token"),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this Equipment enters, create a 1/1 colorless Thopter artifact creature token with flying, then attach this Equipment to it.
			Equipped creature gets +1/+0.
			Equip {2}
		`,
		},
	}
}

var barbedSpikeToken = newBarbedSpikeToken()

func newBarbedSpikeToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Thopter",
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Thopter},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
