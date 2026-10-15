package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GreatUglyLookingGoblin is the card definition for Great Ugly-Looking Goblin // Clap! Snap!.
//
// Type: Creature — Goblin Soldier // Sorcery — Adventure
// Cost: {5}{B} // {1}{B}
// Face: Clap! Snap! — Sorcery — Adventure ({1}{B})
//
// Oracle text:
//
//	Each creature you control with a +1/+1 counter on it has menace. (It can't be blocked except by two or more creatures.)
var GreatUglyLookingGoblin = newGreatUglyLookingGoblin

func newGreatUglyLookingGoblin() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Great Ugly-Looking Goblin",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin, types.Soldier},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Each creature you control with a +1/+1 counter on it has menace. (It can't be blocked except by two or more creatures.)",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, MatchCounter: true, RequiredCounter: counter.PlusOnePlusOne}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Menace},
						},
					},
				},
			},
			OracleText: `
			Each creature you control with a +1/+1 counter on it has menace. (It can't be blocked except by two or more creatures.)
		`,
		},
		Layout: game.LayoutAdventure,
		Alternate: opt.Val(game.CardFace{
			Name: "Clap! Snap!",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:   []color.Color{color.Black},
			Types:    []types.Card{types.Sorcery},
			Subtypes: []types.Sub{types.Adventure},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Amass{
							Amount:  game.Fixed(2),
							Subtype: types.Goblin,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Amass Goblins 2. (Then exile this card. You may cast the creature later from exile.)
		`,
		}),
	}
}
