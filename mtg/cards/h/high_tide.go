package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HighTide is the card definition for High Tide.
//
// Type: Instant
// Cost: {U}
//
// Oracle text:
//
//	Until end of turn, whenever a player taps an Island for mana, that player
//	adds an additional {U}.
var HighTide = newHighTide

func newHighTide() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "High Tide",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateDelayedTrigger{
							Trigger: game.DelayedTriggerDef{
								EventPattern: opt.Val(game.TriggerPattern{
									Event:                game.EventManaProduced,
									Controller:           game.TriggerControllerAny,
									RequireTappedForMana: true,
									SubjectSelection:     game.Selection{SubtypesAny: []types.Sub{types.Island}},
								}),
								Window: game.DelayedWindowThisTurn,
								Content: game.Mode{
									Sequence: []game.Instruction{
										{
											Primitive: game.AddMana{
												Amount:    game.Fixed(1),
												ManaColor: mana.U,
												Player:    opt.Val(game.EventPlayerReference()),
											},
										},
									},
								}.Ability(),
							},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Until end of turn, whenever a player taps an Island for mana, that player adds an additional {U}.
		`,
		},
	}
}
