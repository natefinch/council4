package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// IncreasingSavagery is the card definition for Increasing Savagery.
//
// Type: Sorcery
// Cost: {2}{G}{G}
//
// Oracle text:
//
//	Put five +1/+1 counters on target creature. If this spell was cast from a graveyard, put ten +1/+1 counters on that creature instead.
//	Flashback {5}{G}{G} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var IncreasingSavagery = newIncreasingSavagery

func newIncreasingSavagery() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Increasing Savagery",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{Cost: cost.Mana{cost.O(5), cost.G, cost.G}},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.AddCounter{
							Amount:      game.Fixed(5),
							Object:      game.TargetPermanentReference(0),
							CounterKind: counter.PlusOnePlusOne,
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Negate:       true,
								CastFromZone: opt.Val(zone.Graveyard),
							}),
						}),
					},
					{
						Primitive: game.AddCounter{
							Amount:      game.Fixed(10),
							Object:      game.TargetPermanentReference(0),
							CounterKind: counter.PlusOnePlusOne,
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								CastFromZone: opt.Val(zone.Graveyard),
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Put five +1/+1 counters on target creature. If this spell was cast from a graveyard, put ten +1/+1 counters on that creature instead.
			Flashback {5}{G}{G} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}
