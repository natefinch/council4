package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// WitchEngine is the card definition for Witch Engine.
//
// Type: Creature — Horror
// Cost: {5}{B}
//
// Oracle text:
//
//	Swampwalk (This creature can't be blocked as long as defending player controls a Swamp.)
//	{T}: Add {B}{B}{B}{B}. Target opponent gains control of this creature. (Activate only as an instant.)
var WitchEngine = newWitchEngine

func newWitchEngine() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Witch Engine",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Horror},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.SwampwalkStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{T}: Add {B}{B}{B}{B}. Target opponent gains control of this creature. (Activate only as an instant.)",
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "Target opponent",
								Allow:      game.TargetAllowPlayer,
								Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.B,
								},
							},
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.B,
								},
							},
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.B,
								},
							},
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.B,
								},
							},
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.SourcePermanentReference()),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:            game.LayerControl,
											NewControllerRef: opt.Val(game.TargetPlayerReference(0)),
										},
									},
									Duration: game.DurationPermanent,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Swampwalk (This creature can't be blocked as long as defending player controls a Swamp.)
			{T}: Add {B}{B}{B}{B}. Target opponent gains control of this creature. (Activate only as an instant.)
		`,
		},
	}
}
