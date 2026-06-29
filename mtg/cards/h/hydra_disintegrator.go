package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HYDRADisintegrator is the card definition for HYDRA Disintegrator.
//
// Type: Artifact — Equipment
// Cost: {4}{B}
//
// Oracle text:
//
//	When this Equipment enters, create a 2/1 black Villain creature token with menace, then attach this Equipment to it. (A creature with menace can't be blocked except by two or more creatures.)
//	Equipped creature gets +3/+3.
//	Equip {4} ({4}: Attach to target creature you control. Equip only as a sorcery.)
var HYDRADisintegrator = newHYDRADisintegrator()

func newHYDRADisintegrator() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "HYDRA Disintegrator",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
			}),
			Colors:   []color.Color{color.Black},
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta:     3,
							ToughnessDelta: 3,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(4)}),
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
									Source:        game.TokenDef(hYDRADisintegratorToken),
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
			When this Equipment enters, create a 2/1 black Villain creature token with menace, then attach this Equipment to it. (A creature with menace can't be blocked except by two or more creatures.)
			Equipped creature gets +3/+3.
			Equip {4} ({4}: Attach to target creature you control. Equip only as a sorcery.)
		`,
		},
	}
}

var hYDRADisintegratorToken = newHYDRADisintegratorToken()

func newHYDRADisintegratorToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Villain",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Villain},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.MenaceStaticBody,
			},
		},
	}
}
