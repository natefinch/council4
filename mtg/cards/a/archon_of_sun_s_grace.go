package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ArchonOfSunSGrace is the card definition for Archon of Sun's Grace.
//
// Type: Creature — Archon
// Cost: {2}{W}{W}
//
// Oracle text:
//
//	Flying
//	Lifelink (Damage dealt by this creature also causes you to gain that much life.)
//	Pegasus creatures you control have lifelink.
//	Constellation — Whenever an enchantment you control enters, create a 2/2 white Pegasus creature token with flying.
var ArchonOfSunSGrace = newArchonOfSunSGrace

func newArchonOfSunSGrace() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Archon of Sun's Grace",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Archon},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.LifelinkStaticBody,
				game.StaticAbility{
					Text: "Pegasus creatures you control have lifelink.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Pegasus")}}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Lifelink},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Constellation — Whenever an enchantment you control enters, create a 2/2 white Pegasus creature token with flying.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							Controller:       game.TriggerControllerYou,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Enchantment}},
							StackObjectKind:  game.StackSpell,
							CounterKind:      counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(archonOfSunSGraceToken),
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
			Flying
			Lifelink (Damage dealt by this creature also causes you to gain that much life.)
			Pegasus creatures you control have lifelink.
			Constellation — Whenever an enchantment you control enters, create a 2/2 white Pegasus creature token with flying.
		`,
		},
	}
}

var archonOfSunSGraceToken = newArchonOfSunSGraceToken()

func newArchonOfSunSGraceToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Pegasus",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Pegasus},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
