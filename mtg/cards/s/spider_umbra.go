package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SpiderUmbra is the card definition for Spider Umbra.
//
// Type: Enchantment — Aura
// Cost: {G}
//
// Oracle text:
//
//	Enchant creature
//	Enchanted creature gets +1/+1 and has reach. (It can block creatures with flying.)
//	Umbra armor (If enchanted creature would be destroyed, instead remove all damage from it and destroy this Aura.)
var SpiderUmbra = newSpiderUmbra

func newSpiderUmbra() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Spider Umbra",
			ManaCost: opt.Val(cost.Mana{
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
					Text: "Enchanted creature gets +1/+1 and has reach. (It can block creatures with flying.)",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Reach},
						},
					},
				},
				game.UmbraArmorStaticBody,
			},
			OracleText: `
			Enchant creature
			Enchanted creature gets +1/+1 and has reach. (It can block creatures with flying.)
			Umbra armor (If enchanted creature would be destroyed, instead remove all damage from it and destroy this Aura.)
		`,
		},
	}
}
