package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SeraphicSteed is the card definition for Seraphic Steed.
//
// Type: Creature — Unicorn Mount
// Cost: {G}{W}
//
// Oracle text:
//
//	First strike, lifelink
//	Whenever this creature attacks while saddled, create a 3/3 white Angel creature token with flying.
//	Saddle 4 (Tap any number of other creatures you control with total power 4 or more: This Mount becomes saddled until end of turn. Saddle only as a sorcery.)
var SeraphicSteed = newSeraphicSteed

func newSeraphicSteed() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Seraphic Steed",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
				cost.W,
			}),
			Colors:    []color.Color{color.Green, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Unicorn, types.Mount},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
				game.LifelinkStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.SaddleActivatedAbility(4),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever this creature attacks while saddled, create a 3/3 white Angel creature token with flying.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:              game.EventAttackerDeclared,
							Source:             game.TriggerSourceSelf,
							StackObjectKind:    game.StackSpell,
							CounterKind:        counter.PlusOnePlusOne,
							AttackWhileSaddled: true,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(seraphicSteedToken),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			First strike, lifelink
			Whenever this creature attacks while saddled, create a 3/3 white Angel creature token with flying.
			Saddle 4 (Tap any number of other creatures you control with total power 4 or more: This Mount becomes saddled until end of turn. Saddle only as a sorcery.)
		`,
		},
	}
}

var seraphicSteedToken = newSeraphicSteedToken()

func newSeraphicSteedToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Angel",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Angel},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
