package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SkarrganHellkite is the card definition for Skarrgan Hellkite.
//
// Type: Creature — Dragon
// Cost: {3}{R}{R}
//
// Oracle text:
//
//	Riot (This creature enters with your choice of a +1/+1 counter or haste.)
//	Flying
//	{3}{R}: This creature deals 2 damage divided as you choose among one or two targets. Activate only if this creature has a +1/+1 counter on it.
var SkarrganHellkite = newSkarrganHellkite

func newSkarrganHellkite() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Skarrgan Hellkite",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dragon},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.RiotStaticBody,
				game.FlyingStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{3}{R}: This creature deals 2 damage divided as you choose among one or two targets. Activate only if this creature has a +1/+1 counter on it.",
					ManaCost:       opt.Val(cost.Mana{cost.O(3), cost.R}),
					ZoneOfFunction: zone.Battlefield,
					ActivationCondition: opt.Val(game.Condition{
						Object:        opt.Val(game.SourcePermanentReference()),
						ObjectMatches: opt.Val(game.Selection{RequiredCounterCount: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 1}), RequiredCounter: counter.PlusOnePlusOne}),
					}),
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 2,
								Constraint: "one or two targets",
								Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:       game.Fixed(2),
									Recipient:    game.AnyTargetDamageRecipient(0),
									Divided:      true,
									DamageSource: opt.Val(game.SourcePermanentReference()),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Riot (This creature enters with your choice of a +1/+1 counter or haste.)
			Flying
			{3}{R}: This creature deals 2 damage divided as you choose among one or two targets. Activate only if this creature has a +1/+1 counter on it.
		`,
		},
	}
}
