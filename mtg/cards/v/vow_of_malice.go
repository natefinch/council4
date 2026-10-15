package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VowOfMalice is the card definition for Vow of Malice.
//
// Type: Enchantment — Aura
// Cost: {2}{B}
//
// Oracle text:
//
//	Enchant creature
//	Enchanted creature gets +2/+2, has intimidate, and can't attack you or planeswalkers you control. (A creature with intimidate can't be blocked except by artifact creatures and/or creatures that share a color with it.)
var VowOfMalice = newVowOfMalice

func newVowOfMalice() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Vow of Malice",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:   []color.Color{color.Black},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}),
				game.StaticAbility{
					Text: "Enchanted creature gets +2/+2, has intimidate, and can't attack you or planeswalkers you control. (A creature with intimidate can't be blocked except by artifact creatures and/or creatures that share a color with it.)",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     2,
							ToughnessDelta: 2,
						},
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Intimidate},
						},
					},
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:             game.RuleEffectCantAttack,
							AffectedAttached: true,
							Duration:         game.DurationPermanent,
							DefendingPlayer:  game.PlayerYou,
						},
					},
				},
			},
			OracleText: `
			Enchant creature
			Enchanted creature gets +2/+2, has intimidate, and can't attack you or planeswalkers you control. (A creature with intimidate can't be blocked except by artifact creatures and/or creatures that share a color with it.)
		`,
		},
	}
}
