package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ArastaOfTheEndlessWeb is the card definition for Arasta of the Endless Web.
//
// Type: Legendary Enchantment Creature — Spider
// Cost: {2}{G}{G}
//
// Oracle text:
//
//	Reach
//	Whenever an opponent casts an instant or sorcery spell, create a 1/2 green Spider creature token with reach.
var ArastaOfTheEndlessWeb = newArastaOfTheEndlessWeb

func newArastaOfTheEndlessWeb() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Arasta of the Endless Web",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Enchantment, types.Creature},
			Subtypes:   []types.Sub{types.Spider},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever an opponent casts an instant or sorcery spell, create a 1/2 green Spider creature token with reach.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventSpellCast,
							Controller:      game.TriggerControllerOpponent,
							CardSelection:   game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}},
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(arastaOfTheEndlessWebToken),
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
			Reach
			Whenever an opponent casts an instant or sorcery spell, create a 1/2 green Spider creature token with reach.
		`,
		},
	}
}

var arastaOfTheEndlessWebToken = newArastaOfTheEndlessWebToken()

func newArastaOfTheEndlessWebToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Spider",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spider},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
			},
		},
	}
}
