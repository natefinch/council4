package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VowOfWildness is the card definition for Vow of Wildness.
//
// Type: Enchantment — Aura
// Cost: {2}{G}
//
// Oracle text:
//
//	Enchant creature
//	Enchanted creature gets +3/+3, has trample, and can't attack you or planeswalkers you control.
var VowOfWildness = newVowOfWildness

func newVowOfWildness() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Vow of Wildness",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:   []color.Color{color.Green},
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
					Text: "Enchanted creature gets +3/+3, has trample, and can't attack you or planeswalkers you control.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     3,
							ToughnessDelta: 3,
						},
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Trample},
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
			Enchanted creature gets +3/+3, has trample, and can't attack you or planeswalkers you control.
		`,
		},
	}
}
