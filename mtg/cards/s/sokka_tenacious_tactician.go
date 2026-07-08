package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SokkaTenaciousTactician is the card definition for Sokka, Tenacious Tactician.
//
// Type: Legendary Creature — Human Warrior Ally
// Cost: {1}{U}{R}{W}
//
// Oracle text:
//
//	Menace, prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
//	Other Allies you control have menace and prowess.
//	Whenever you cast a noncreature spell, create a 1/1 white Ally creature token.
var SokkaTenaciousTactician = newSokkaTenaciousTactician

func newSokkaTenaciousTactician() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Sokka, Tenacious Tactician",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.R,
				cost.W,
			}),
			Colors:     []color.Color{color.Red, color.Blue, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Warrior, types.Ally},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.MenaceStaticBody,
				game.ProwessStaticBody,
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Ally")}}, game.SourcePermanentReference()),
							AddKeywords: []game.Keyword{
								game.Menace,
								game.Prowess,
							},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:         game.EventSpellCast,
							Controller:    game.TriggerControllerYou,
							CardSelection: game.Selection{ExcludedTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(sokkaTenaciousTacticianToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Menace, prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
			Other Allies you control have menace and prowess.
			Whenever you cast a noncreature spell, create a 1/1 white Ally creature token.
		`,
		},
	}
}

var sokkaTenaciousTacticianToken = newSokkaTenaciousTacticianToken()

func newSokkaTenaciousTacticianToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Ally",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Ally},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
