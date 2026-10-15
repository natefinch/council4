package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// UnderrealmLich is the card definition for Underrealm Lich.
//
// Type: Creature — Zombie Elf Shaman
// Cost: {3}{B}{G}
//
// Oracle text:
//
//	If you would draw a card, instead look at the top three cards of your library, then put one into your hand and the rest into your graveyard.
//	Pay 4 life: This creature gains indestructible until end of turn. Tap it.
var UnderrealmLich = newUnderrealmLich

func newUnderrealmLich() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Underrealm Lich",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.G,
			}),
			Colors:    []color.Color{color.Black, color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Elf, types.Shaman},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 3}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Pay 4 life: This creature gains indestructible until end of turn. Tap it.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalPayLife,
							Text:        "Pay 4 life",
							Amount:      4,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.SourcePermanentReference()),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration:    game.DurationPermanent,
													Layer:       game.LayerAbility,
													AddKeywords: []game.Keyword{game.Indestructible},
												},
											},
											Duration: game.DurationUntilEndOfTurn,
										},
									},
									game.Instruction{
										Primitive: game.Tap{
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
			ReplacementAbilities: []game.ReplacementAbility{
				game.DrawCardDigReplacement("If you would draw a card, instead look at the top three cards of your library, then put one into your hand and the rest into your graveyard.", 3, 1, game.DigRemainderGraveyard),
			},
			OracleText: `
			If you would draw a card, instead look at the top three cards of your library, then put one into your hand and the rest into your graveyard.
			Pay 4 life: This creature gains indestructible until end of turn. Tap it.
		`,
		},
	}
}
