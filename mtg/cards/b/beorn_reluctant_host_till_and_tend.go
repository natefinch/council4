package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BeornReluctantHost is the card definition for Beorn, Reluctant Host // Till and Tend.
//
// Type: Legendary Creature — Human Bear Shapeshifter // Sorcery — Adventure
// Cost: {4}{G} // {1}{G}
// Face: Till and Tend — Sorcery — Adventure ({1}{G})
//
// Oracle text:
//
//	Trample
var BeornReluctantHost = newBeornReluctantHost

func newBeornReluctantHost() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Beorn, Reluctant Host",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Bear, types.Shapeshifter},
			Power:      opt.Val(game.PT{Value: 5}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			OracleText: `
			Trample
		`,
		},
		Layout: game.LayoutAdventure,
		Alternate: opt.Val(game.CardFace{
			Name: "Till and Tend",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:   []color.Color{color.Green},
			Types:    []types.Card{types.Sorcery},
			Subtypes: []types.Sub{types.Adventure},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyRule{
							RuleEffects: []game.RuleEffect{
								game.RuleEffect{
									Kind:                game.RuleEffectAdditionalLandPlays,
									Duration:            game.DurationPermanent,
									AffectedPlayer:      game.PlayerYou,
									AdditionalLandPlays: 1,
								},
							},
							Duration: game.DurationThisTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			You may play an additional land this turn. (Then exile this card. You may cast the creature later from exile.)
		`,
		}),
	}
}
