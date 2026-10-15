package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LeechGauntlet is the card definition for Leech Gauntlet.
//
// Type: Artifact Creature — Equipment Leech
// Cost: {1}{B}
//
// Oracle text:
//
//	Lifelink
//	Equipped creature has lifelink.
//	Reconfigure {4} ({4}: Attach to target creature you control; or unattach from a creature. Reconfigure only as a sorcery. While attached, this isn't a creature.)
var LeechGauntlet = newLeechGauntlet

func newLeechGauntlet() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Leech Gauntlet",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Equipment, types.Leech},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.LifelinkStaticBody,
				game.StaticAbility{
					Text: "Equipped creature has lifelink.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Lifelink},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ReconfigureActivatedAbility(cost.Mana{cost.O(4)}),
			},
			OracleText: `
			Lifelink
			Equipped creature has lifelink.
			Reconfigure {4} ({4}: Attach to target creature you control; or unattach from a creature. Reconfigure only as a sorcery. While attached, this isn't a creature.)
		`,
		},
	}
}
