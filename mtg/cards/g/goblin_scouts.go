package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GoblinScouts is the card definition for Goblin Scouts.
//
// Type: Sorcery
// Cost: {3}{R}{R}
//
// Oracle text:
//
//	Create three 1/1 red Goblin Scout creature tokens with mountainwalk. (They can't be blocked as long as defending player controls a Mountain.)
var GoblinScouts = newGoblinScouts

func newGoblinScouts() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Goblin Scouts",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(3),
							Source: game.TokenDef(goblinScoutsToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create three 1/1 red Goblin Scout creature tokens with mountainwalk. (They can't be blocked as long as defending player controls a Mountain.)
		`,
		},
	}
}

var goblinScoutsToken = newGoblinScoutsToken()

func newGoblinScoutsToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Goblin Scout",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin, types.Scout},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.MountainwalkStaticBody,
			},
		},
	}
}
