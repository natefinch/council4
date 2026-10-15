package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SentinelSEyes is the card definition for Sentinel's Eyes.
//
// Type: Enchantment — Aura
// Cost: {W}
//
// Oracle text:
//
//	Enchant creature
//	Enchanted creature gets +1/+1 and has vigilance.
//	Escape—{W}, Exile two other cards from your graveyard. (You may cast this card from your graveyard for its escape cost.)
var SentinelSEyes = newSentinelSEyes

func newSentinelSEyes() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Sentinel's Eyes",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors:   []color.Color{color.White},
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
					Text: "Enchanted creature gets +1/+1 and has vigilance.",
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
							AddKeywords: []game.Keyword{game.Vigilance},
						},
					},
				},
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.SimpleKeyword{
							Kind: game.Escape,
						},
					},
				},
			},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:    "Escape",
					ManaCost: opt.Val(cost.Mana{cost.W}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:          cost.AdditionalExile,
							Text:          "Exile two other cards from your graveyard",
							Amount:        2,
							Source:        zone.Graveyard,
							ExcludeSource: true,
						},
					},
					Mechanic: cost.AlternativeMechanicEscape,
				},
			},
			OracleText: `
			Enchant creature
			Enchanted creature gets +1/+1 and has vigilance.
			Escape—{W}, Exile two other cards from your graveyard. (You may cast this card from your graveyard for its escape cost.)
		`,
		},
	}
}
