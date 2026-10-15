package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WreckingBallArm is the card definition for Wrecking Ball Arm.
//
// Type: Legendary Artifact — Equipment
// Cost: {2}
//
// Oracle text:
//
//	Equipped creature has base power and toughness 7/7 and can't be blocked by creatures with power 2 or less.
//	Equip legendary creature {3}
//	Equip {7}
var WreckingBallArm = newWreckingBallArm

func newWreckingBallArm() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Wrecking Ball Arm",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			Subtypes:   []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature has base power and toughness 7/7 and can't be blocked by creatures with power 2 or less.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerPowerToughnessSet,
							SetPower: opt.Val(game.PT{
								Value: 7,
							}),
							SetToughness: opt.Val(game.PT{
								Value: 7,
							}),
						},
					},
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:             game.RuleEffectCantBeBlockedByCreaturesWith,
							AffectedAttached: true,
							Duration:         game.DurationPermanent,
							BlockerRestriction: game.BlockerRestriction{
								Kind:  game.BlockerRestrictionPowerLessOrEqual,
								Power: 2,
							},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipRestrictedActivatedAbility(cost.Mana{cost.O(3)}, []types.Super{types.Legendary}, nil),
				game.EquipActivatedAbility(cost.Mana{cost.O(7)}),
			},
			OracleText: `
			Equipped creature has base power and toughness 7/7 and can't be blocked by creatures with power 2 or less.
			Equip legendary creature {3}
			Equip {7}
		`,
		},
	}
}
