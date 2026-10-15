package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FearsomeAwakening is the card definition for Fearsome Awakening.
//
// Type: Sorcery
// Cost: {4}{B}
//
// Oracle text:
//
//	Return target creature card from your graveyard to the battlefield. If it's a Dragon, put two +1/+1 counters on it.
var FearsomeAwakening = newFearsomeAwakening

func newFearsomeAwakening() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Fearsome Awakening",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature card from your graveyard",
						Allow:      game.TargetAllowCard,
						TargetZone: zone.Graveyard,
						Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.PutOnBattlefield{
							Source:        game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
							PublishLinked: game.LinkedKey("sequence-effect-0-product"),
						},
						LocalProducts: game.LocalProducts{
							Links: []game.LinkedKey{game.LinkedKey("sequence-effect-0-product")},
						},
					},
					{
						Primitive: game.AddCounter{
							Amount:      game.Fixed(2),
							Object:      game.LinkedObjectReference("sequence-effect-0-product"),
							CounterKind: counter.PlusOnePlusOne,
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Object:        opt.Val(game.LinkedObjectReference("sequence-effect-0-product")),
								ObjectMatches: opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Sub("Dragon")}}),
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Return target creature card from your graveyard to the battlefield. If it's a Dragon, put two +1/+1 counters on it.
		`,
		},
	}
}
