package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BeanstalkWurm is the card definition for Beanstalk Wurm // Plant Beans.
//
// Type: Creature — Plant Wurm // Sorcery — Adventure
// Cost: {4}{G} // {1}{G}
// Face: Plant Beans — Sorcery — Adventure ({1}{G})
//
// Oracle text:
//
//	Reach
var BeanstalkWurm = newBeanstalkWurm

func newBeanstalkWurm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Beanstalk Wurm",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Plant, types.Wurm},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
			},
			OracleText: `
			Reach
		`,
		},
		Layout: game.LayoutAdventure,
		Alternate: opt.Val(game.CardFace{
			Name: "Plant Beans",
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
