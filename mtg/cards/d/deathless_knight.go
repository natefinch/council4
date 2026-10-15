package d

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

// DeathlessKnight is the card definition for Deathless Knight.
//
// Type: Creature — Skeleton Knight
// Cost: {B/G}{B/G}{B/G}{B/G}
//
// Oracle text:
//
//	Haste
//	When you gain life for the first time each turn, return this card from your graveyard to your hand.
var DeathlessKnight = newDeathlessKnight

func newDeathlessKnight() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Deathless Knight",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.B, mana.G),
				cost.HybridMana(mana.B, mana.G),
				cost.HybridMana(mana.B, mana.G),
				cost.HybridMana(mana.B, mana.G),
			}),
			Colors:    []color.Color{color.Black, color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Skeleton, types.Knight},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When you gain life for the first time each turn, return this card from your graveyard to your hand.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:                      game.EventLifeGained,
							Player:                     game.TriggerPlayerYou,
							StackObjectKind:            game.StackSpell,
							PlayerEventOrdinalThisTurn: 1,
							CounterKind:                counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveCard{
											Card: game.CardReference{
												Kind: game.CardReferenceSource,
											},
											FromZone:    zone.Graveyard,
											Destination: zone.Hand,
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
			Haste
			When you gain life for the first time each turn, return this card from your graveyard to your hand.
		`,
		},
	}
}
