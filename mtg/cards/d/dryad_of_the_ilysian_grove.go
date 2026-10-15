package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DryadOfTheIlysianGrove is the card definition for Dryad of the Ilysian Grove.
//
// Type: Enchantment Creature — Nymph Dryad
// Cost: {2}{G}
//
// Oracle text:
//
//	You may play an additional land on each of your turns.
//	Lands you control are every basic land type in addition to their other types.
var DryadOfTheIlysianGrove = newDryadOfTheIlysianGrove

func newDryadOfTheIlysianGrove() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Dryad of the Ilysian Grove",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Enchantment, types.Creature},
			Subtypes:  []types.Sub{types.Nymph, types.Dryad},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "You may play an additional land on each of your turns.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:                game.RuleEffectAdditionalLandPlays,
							Duration:            game.DurationPermanent,
							AffectedPlayer:      game.PlayerYou,
							AdditionalLandPlays: 1,
						},
					},
				},
				game.StaticAbility{
					Text: "Lands you control are every basic land type in addition to their other types.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:              game.DurationPermanent,
							Group:                 game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Land}}),
							Layer:                 game.LayerType,
							AddEveryBasicLandType: true,
						},
					},
				},
			},
			OracleText: `
			You may play an additional land on each of your turns.
			Lands you control are every basic land type in addition to their other types.
		`,
		},
	}
}
