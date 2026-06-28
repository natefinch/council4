package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DeathknellBerserker is the card definition for Deathknell Berserker.
//
// Type: Creature — Elf Berserker
// Cost: {1}{B}
//
// Oracle text:
//
//	When this creature dies, if its power was 3 or greater, create a 2/2 black Zombie Berserker creature token.
var DeathknellBerserker = newDeathknellBerserker()

func newDeathknellBerserker() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Deathknell Berserker",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Berserker},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Source:           game.TriggerSourceSelf,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
						InterveningIf: "if its power was 3 or greater",
						InterveningCondition: opt.Val(game.Condition{
							Object:        opt.Val(game.EventPermanentReference()),
							ObjectMatches: opt.Val(game.Selection{Power: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 3})}),
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(deathknellBerserkerToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this creature dies, if its power was 3 or greater, create a 2/2 black Zombie Berserker creature token.
		`,
		},
	}
}

var deathknellBerserkerToken = newDeathknellBerserkerToken()

func newDeathknellBerserkerToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Zombie Berserker",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Berserker},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
