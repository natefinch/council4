package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// HoofprintsOfTheStag is the card definition for Hoofprints of the Stag.
//
// Type: Kindred Enchantment — Elemental
// Cost: {1}{W}
//
// Oracle text:
//
//	Whenever you draw a card, you may put a hoofprint counter on this enchantment.
//	{2}{W}, Remove four hoofprint counters from this enchantment: Create a 4/4 white Elemental creature token with flying. Activate only during your turn.
var HoofprintsOfTheStag = newHoofprintsOfTheStag

func newHoofprintsOfTheStag() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Hoofprints of the Stag",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:   []color.Color{color.White},
			Types:    []types.Card{types.Kindred, types.Enchantment},
			Subtypes: []types.Sub{types.Elemental},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{2}{W}, Remove four hoofprint counters from this enchantment: Create a 4/4 white Elemental creature token with flying. Activate only during your turn.",
					ManaCost: opt.Val(cost.Mana{cost.O(2), cost.W}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove four hoofprint counters from this enchantment",
							Amount:      4,
							CounterKind: counter.Hoofprint,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.DuringYourTurn,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(hoofprintsOfTheStagToken),
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventCardDrawn,
							Player: game.TriggerPlayerYou,
						},
					},
					Optional: true,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.Hoofprint,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever you draw a card, you may put a hoofprint counter on this enchantment.
			{2}{W}, Remove four hoofprint counters from this enchantment: Create a 4/4 white Elemental creature token with flying. Activate only during your turn.
		`,
		},
	}
}

var hoofprintsOfTheStagToken = newHoofprintsOfTheStagToken()

func newHoofprintsOfTheStagToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Elemental",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
