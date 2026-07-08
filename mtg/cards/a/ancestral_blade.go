package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AncestralBlade is the card definition for Ancestral Blade.
//
// Type: Artifact — Equipment
// Cost: {1}{W}
//
// Oracle text:
//
//	When this Equipment enters, create a 1/1 white Soldier creature token, then attach this Equipment to it.
//	Equipped creature gets +1/+1.
//	Equip {1} ({1}: Attach to target creature you control. Equip only as a sorcery.)
var AncestralBlade = newAncestralBlade

func newAncestralBlade() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Ancestral Blade",
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
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(1)}),
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
									Source:        game.TokenDef(ancestralBladeToken),
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
			When this Equipment enters, create a 1/1 white Soldier creature token, then attach this Equipment to it.
			Equipped creature gets +1/+1.
			Equip {1} ({1}: Attach to target creature you control. Equip only as a sorcery.)
		`,
		},
	}
}

var ancestralBladeToken = newAncestralBladeToken()

func newAncestralBladeToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Soldier",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
