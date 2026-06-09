package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PrimalMight is the card definition for Primal Might.
//
// Type: Sorcery
// Cost: {X}{G}
//
// Oracle text:
//
//	Target creature you control gets +X/+X until end of turn. Then it fights up to one target creature you don't control. (Each deals damage equal to its power to the other.)
var PrimalMight = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name: "Primal Might",
		ManaCost: opt.Val(cost.Mana{
			cost.X,
			cost.G,
		}),
		Colors: []color.Color{color.Green},
		Types:  []types.Card{types.Sorcery},
		OracleText: `
			Target creature you control gets +X/+X until end of turn. Then it fights up to one target creature you don't control. (Each deals damage equal to its power to the other.)
		`,
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{
				{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature you control",
					Allow:      game.TargetAllowPermanent,
					Predicate: game.TargetPredicate{
						PermanentTypes: []types.Card{
							types.Creature,
						},
						Controller: game.ControllerYou,
					},
				},
				{
					MinTargets: 0,
					MaxTargets: 1,
					Constraint: "creature you don't control",
					Allow:      game.TargetAllowPermanent,
					Predicate: game.TargetPredicate{
						PermanentTypes: []types.Card{
							types.Creature,
						},
						Controller: game.ControllerNotYou,
					},
				},
			},
			Sequence: []game.Instruction{
				{
					Primitive: game.ModifyPT{
						Object: game.TargetPermanentReference(0),
						PowerDelta: game.Dynamic(game.DynamicAmount{
							Kind: game.DynamicAmountX,
						}),
						ToughnessDelta: game.Dynamic(game.DynamicAmount{
							Kind: game.DynamicAmountX,
						}),
						Duration: game.DurationUntilEndOfTurn,
					},
				},
				{
					Primitive: game.Fight{
						Object:        game.TargetPermanentReference(0),
						RelatedObject: game.TargetPermanentReference(1),
					},
					Description: "target creature you control fights up to one target creature you don't control",
				},
			},
		}.Ability()),
	},
}
