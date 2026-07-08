package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ShipwreckSinger is the card definition for Shipwreck Singer.
//
// Type: Creature — Siren
// Cost: {U}{B}
//
// Oracle text:
//
//	Flying
//	{1}{U}: Target creature an opponent controls attacks this turn if able.
//	{1}{B}, {T}: Attacking creatures get -1/-1 until end of turn.
var ShipwreckSinger = newShipwreckSinger

func newShipwreckSinger() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Shipwreck Singer",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
				cost.B,
			}),
			Colors:    []color.Color{color.Black, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Siren},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{1}{U}: Target creature an opponent controls attacks this turn if able.",
					ManaCost:       opt.Val(cost.Mana{cost.O(1), cost.U}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature an opponent controls",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerOpponent}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyRule{
									Object: opt.Val(game.TargetPermanentReference(0)),
									RuleEffects: []game.RuleEffect{
										game.RuleEffect{
											Kind: game.RuleEffectMustAttack,
										},
									},
									Duration: game.DurationThisTurn,
								},
							},
						},
					}.Ability(),
				},
				game.ActivatedAbility{
					Text:            "{1}{B}, {T}: Attacking creatures get -1/-1 until end of turn.",
					ManaCost:        opt.Val(cost.Mana{cost.O(1), cost.B}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:          game.LayerPowerToughnessModify,
											Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, CombatState: game.CombatStateAttacking}),
											PowerDelta:     -1,
											ToughnessDelta: -1,
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
			Flying
			{1}{U}: Target creature an opponent controls attacks this turn if able.
			{1}{B}, {T}: Attacking creatures get -1/-1 until end of turn.
		`,
		},
	}
}
