package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GimliSAxe is the card definition for Gimli's Axe.
//
// Type: Artifact — Equipment
// Cost: {2}{R}
//
// Oracle text:
//
//	Equipped creature gets +3/+0.
//	As long as equipped creature is legendary, it has menace. (It can't be blocked except by two or more creatures.)
//	Equip {2} ({2}: Attach to target creature you control. Equip only as a sorcery.)
var GimliSAxe = newGimliSAxe

func newGimliSAxe() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Gimli's Axe",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors:   []color.Color{color.Red},
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Equipped creature gets +3/+0.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:   game.DurationPermanent,
							Group:      game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:      game.LayerPowerToughnessModify,
							PowerDelta: 3,
						},
					},
				},
				game.StaticAbility{
					Text: "As long as equipped creature is legendary, it has menace. (It can't be blocked except by two or more creatures.)",
					Condition: opt.Val(game.Condition{
						Text:              "As long as equipped creature is legendary",
						Object:            opt.Val(game.SourceAttachedPermanentReference()),
						ObjectMatches:     opt.Val(game.Selection{Supertypes: []types.Super{types.Legendary}}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Menace},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(2)}),
			},
			OracleText: `
			Equipped creature gets +3/+0.
			As long as equipped creature is legendary, it has menace. (It can't be blocked except by two or more creatures.)
			Equip {2} ({2}: Attach to target creature you control. Equip only as a sorcery.)
		`,
		},
	}
}
