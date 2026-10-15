package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// UlvenwaldCaptive is the card definition for Ulvenwald Captive // Ulvenwald Abomination.
//
// Type: Creature — Werewolf Horror // Creature — Eldrazi Werewolf
// Face: Ulvenwald Abomination — Creature — Eldrazi Werewolf
//
// Oracle text:
//
//	Defender
//	{T}: Add {G}.
//	{5}{G}{G}: Transform this creature.
var UlvenwaldCaptive = newUlvenwaldCaptive

func newUlvenwaldCaptive() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Ulvenwald Captive",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Werewolf, types.Horror},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{5}{G}{G}: Transform this creature.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(5),
						cost.G,
						cost.G,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Transform{
											Object: game.SourcePermanentReference(),
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
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.G),
			},
			OracleText: `
			Defender
			{T}: Add {G}.
			{5}{G}{G}: Transform this creature.
		`,
		},
		Layout: game.LayoutTransform,
		Back: opt.Val(game.CardFace{
			Name:      "Ulvenwald Abomination",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi, types.Werewolf},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 6}),
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					Text: "{T}: Add {C}{C}.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddMana{
											Amount:    game.Fixed(1),
											ManaColor: mana.C,
										},
									},
									game.Instruction{
										Primitive: game.AddMana{
											Amount:    game.Fixed(1),
											ManaColor: mana.C,
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
			{T}: Add {C}{C}.
		`,
		}),
	}
}
