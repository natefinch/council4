package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GuardianAugmenter is the card definition for Guardian Augmenter.
//
// Type: Creature — Troll Wizard
// Cost: {2}{G}
//
// Oracle text:
//
//	Flash
//	Commander creatures you control get +2/+2.
//	Commanders you control have hexproof.
var GuardianAugmenter = newGuardianAugmenter

func newGuardianAugmenter() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Guardian Augmenter",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Troll, types.Wizard},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.StaticAbility{
					Text: "Commander creatures you control get +2/+2.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, MatchCommander: true}),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     2,
							ToughnessDelta: 2,
						},
					},
				},
				game.StaticAbility{
					Text: "Commanders you control have hexproof.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{MatchCommander: true}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Hexproof},
						},
					},
				},
			},
			OracleText: `
			Flash
			Commander creatures you control get +2/+2.
			Commanders you control have hexproof.
		`,
		},
	}
}
