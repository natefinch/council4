package c

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

// CastleLocthwain is the card definition for Castle Locthwain.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless you control a Swamp.
//	{T}: Add {B}.
//	{1}{B}{B}, {T}: Draw a card, then you lose life equal to the number of cards in your hand.
var CastleLocthwain = newCastleLocthwain

func newCastleLocthwain() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:  "Castle Locthwain",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}{B}{B}, {T}: Draw a card, then you lose life equal to the number of cards in your hand.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
						cost.B,
						cost.B,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
									},
									game.Instruction{
										Primitive: game.LoseLife{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountCountCardsInZone,
												Multiplier: 1,
												Player:     func() *game.PlayerReference { ref := game.ControllerReference(); return &ref }(),
												CardZone:   zone.Hand,
												Selection:  &game.Selection{},
											}),
											Player: game.ControllerReference(),
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
				game.TapManaAbility(mana.B),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedIfReplacement("This land enters tapped unless you control a Swamp.", &game.Condition{
					Negate: true,
					ControlsMatching: opt.Val(game.SelectionCount{
						Selection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Swamp")}},
					}),
				}),
			},
			OracleText: `
			This land enters tapped unless you control a Swamp.
			{T}: Add {B}.
			{1}{B}{B}, {T}: Draw a card, then you lose life equal to the number of cards in your hand.
		`,
		},
	}
}
