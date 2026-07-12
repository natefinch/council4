package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BringerOfTheWhiteDawn is the card definition for Bringer of the White Dawn.
//
// Type: Creature — Bringer
// Cost: {7}{W}{W}
//
// Oracle text:
//
//	You may pay {W}{U}{B}{R}{G} rather than pay this spell's mana cost.
//	Trample
//	At the beginning of your upkeep, you may return target artifact card from your graveyard to the battlefield.
var BringerOfTheWhiteDawn = newBringerOfTheWhiteDawn

func newBringerOfTheWhiteDawn() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Black, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Bringer of the White Dawn",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bringer},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepUpkeep,
						},
					},
					Optional: true,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target artifact card from your graveyard",
								Allow:      game.TargetAllowCard,
								TargetZone: zone.Graveyard,
								Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Artifact}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.PutOnBattlefield{
									Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
								},
							},
						},
					}.Ability(),
				},
			},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:    "Pay {W}{U}{B}{R}{G}",
					ManaCost: opt.Val(cost.Mana{cost.W, cost.U, cost.B, cost.R, cost.G}),
				},
			},
			OracleText: `
			You may pay {W}{U}{B}{R}{G} rather than pay this spell's mana cost.
			Trample
			At the beginning of your upkeep, you may return target artifact card from your graveyard to the battlefield.
		`,
		},
	}
}
