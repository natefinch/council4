// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RTReturnedSubject is the card definition for RT Returned Subject.
//
// Type: Sorcery
//
// Oracle text:
//
//	Return target creature card from your graveyard to the battlefield. If it's an Elf, put a +1/+1 counter on it.
var RTReturnedSubject = newRTReturnedSubject

func newRTReturnedSubject() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "RT Returned Subject",
			Types: []types.Card{types.Sorcery},
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
							Amount:      game.Fixed(1),
							Object:      game.LinkedObjectReference("sequence-effect-0-product"),
							CounterKind: counter.PlusOnePlusOne,
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Object:        opt.Val(game.LinkedObjectReference("sequence-effect-0-product")),
								ObjectMatches: opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Sub("Elf")}}),
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Return target creature card from your graveyard to the battlefield. If it's an Elf, put a +1/+1 counter on it.
		`,
		},
	}
}
