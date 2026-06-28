package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RevivalOfTheAncestors is the card definition for Revival of the Ancestors.
//
// Type: Enchantment — Saga
// Cost: {1}{W}{B}{G}
//
// Oracle text:
//
//	(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)
//	I — Create three 1/1 white Spirit creature tokens.
//	II — Distribute three +1/+1 counters among one, two, or three target creatures you control.
//	III — Creatures you control gain trample and lifelink until end of turn.
var RevivalOfTheAncestors = newRevivalOfTheAncestors()

func newRevivalOfTheAncestors() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Revival of the Ancestors",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
				cost.B,
				cost.G,
			}),
			Colors:   []color.Color{color.Black, color.Green, color.White},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Saga},
			ChapterAbilities: []game.ChapterAbility{
				game.ChapterAbility{
					Text:     "I — Create three 1/1 white Spirit creature tokens.",
					Chapters: []int{1},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(3),
									Source: game.TokenDef(revivalOfTheAncestorsToken),
								},
							},
						},
					}.Ability(),
				},
				game.ChapterAbility{
					Text:     "II — Distribute three +1/+1 counters among one, two, or three target creatures you control.",
					Chapters: []int{2},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 3,
								Constraint: "one, two, or three target creatures you control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(3),
									Object:      game.AllTargetPermanentsReference(0),
									CounterKind: counter.PlusOnePlusOne,
									Distribute:  true,
								},
							},
						},
					}.Ability(),
				},
				game.ChapterAbility{
					Text:     "III — Creatures you control gain trample and lifelink until end of turn.",
					Chapters: []int{3},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											AddKeywords: []game.Keyword{
												game.Trample,
												game.Lifelink,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)
			I — Create three 1/1 white Spirit creature tokens.
			II — Distribute three +1/+1 counters among one, two, or three target creatures you control.
			III — Creatures you control gain trample and lifelink until end of turn.
		`,
		},
	}
}

var revivalOfTheAncestorsToken = newRevivalOfTheAncestorsToken()

func newRevivalOfTheAncestorsToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Spirit",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
