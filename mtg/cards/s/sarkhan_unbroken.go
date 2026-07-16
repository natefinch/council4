package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SarkhanUnbroken is the card definition for Sarkhan Unbroken.
//
// Type: Legendary Planeswalker — Sarkhan
// Cost: {2}{G}{U}{R}
//
// Oracle text:
//
//	+1: Draw a card, then add one mana of any color.
//	−2: Create a 4/4 red Dragon creature token with flying.
//	−8: Search your library for any number of Dragon creature cards, put them onto the battlefield, then shuffle.
var SarkhanUnbroken = newSarkhanUnbroken

func newSarkhanUnbroken() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Sarkhan Unbroken",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.U,
				cost.R,
			}),
			Colors:     []color.Color{color.Green, color.Red, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Planeswalker},
			Subtypes:   []types.Sub{types.Sarkhan},
			Loyalty:    opt.Val(4),
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					LoyaltyCost: 1,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
							{
								Primitive: game.Choose{
									Choice: game.ResolutionChoice{
										Kind:   game.ResolutionChoiceMana,
										Prompt: "Choose a color",
										Colors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
									},
									PublishChoice: game.ChoiceKey("oracle-mana-color"),
								},
							},
							{
								Primitive: game.AddMana{
									Amount:     game.Fixed(1),
									ChoiceFrom: game.ChoiceKey("oracle-mana-color"),
								},
							},
						},
					}.Ability(),
				},
				game.LoyaltyAbility{
					LoyaltyCost: -2,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(sarkhanUnbrokenToken),
								},
							},
						},
					}.Ability(),
				},
				game.LoyaltyAbility{
					LoyaltyCost: -8,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Search{
									Player: game.ControllerReference(),
									Spec: game.SearchSpec{
										SourceZone:  zone.Library,
										Destination: zone.Battlefield,
										Filter:      game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypesAny: []types.Sub{types.Sub("Dragon")}},
										AnyNumber:   true,
									},
									Amount: game.Fixed(0),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			+1: Draw a card, then add one mana of any color.
			−2: Create a 4/4 red Dragon creature token with flying.
			−8: Search your library for any number of Dragon creature cards, put them onto the battlefield, then shuffle.
		`,
		},
	}
}

var sarkhanUnbrokenToken = newSarkhanUnbrokenToken()

func newSarkhanUnbrokenToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Dragon",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dragon},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
