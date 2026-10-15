package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ReleaseToMemory is the card definition for Release to Memory.
//
// Type: Instant
// Cost: {3}{W}
//
// Oracle text:
//
//	Exile target opponent's graveyard. For each creature card exiled this way, create a 1/1 colorless Spirit creature token.
var ReleaseToMemory = newReleaseToMemory

func newReleaseToMemory() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Release to Memory",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target opponent's graveyard",
						Allow:      game.TargetAllowPlayer,
						Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MoveCard{
							Player:      game.TargetPlayerReference(0),
							FromZone:    zone.Graveyard,
							Destination: zone.Exile,
						},
						PublishResult: game.ResultKey("actual-card-quantity-1"),
						LocalProducts: game.LocalProducts{
							Results: []game.ResultKey{game.ResultKey("actual-card-quantity-1")},
						},
					},
					{
						Primitive: game.CreateToken{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountPreviousEffectResult,
								Multiplier: 1,
								Selection:  &game.Selection{RequiredTypes: []types.Card{types.Creature}},
								ResultKey:  game.ResultKey("actual-card-quantity-1"),
							}),
							Source: game.TokenDef(releaseToMemoryToken),
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:      "actual-card-quantity-1",
							Accepted: game.TriTrue,
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Exile target opponent's graveyard. For each creature card exiled this way, create a 1/1 colorless Spirit creature token.
		`,
		},
	}
}

var releaseToMemoryToken = newReleaseToMemoryToken()

func newReleaseToMemoryToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Spirit",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
