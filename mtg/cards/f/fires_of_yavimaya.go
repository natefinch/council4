package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FiresOfYavimaya is the card definition for Fires of Yavimaya.
//
// Type: Enchantment
// Cost: {1}{R}{G}
//
// Oracle text:
//
//	Creatures you control have haste.
//	Sacrifice this enchantment: Target creature gets +2/+2 until end of turn.
var FiresOfYavimaya = func() *game.CardDef {
	card := &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green, color.Red),
		CardFace: game.CardFace{
			Name: "Fires of Yavimaya",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.G,
			}),
			Colors: []color.Color{color.Green, color.Red},
			Types:  []types.Card{types.Enchantment},
			OracleText: `
				Creatures you control have haste.
				Sacrifice this enchantment: Target creature gets +2/+2 until end of turn.
			`,
		},
	}

	card.StaticAbilities = append(card.StaticAbilities, game.StaticAbility{
		Text: `
				Creatures you control have haste.
			`,
		ContinuousEffects: []game.ContinuousEffect{
			{
				Layer:    game.LayerAbility,
				Selector: game.EffectSelectorCreaturesYouControl,
				AddKeywords: []game.Keyword{
					game.Haste,
				},
			},
		},
	},
	)

	card.ActivatedAbilities = append(card.ActivatedAbilities,
		game.ActivatedAbility{
			Text: `
				Sacrifice this enchantment: Target creature gets +2/+2 until end of turn.
			`,
			AdditionalCosts: []cost.Additional{
				{
					Kind: cost.AdditionalSacrificeSource,
				},
			},
			Content: game.Mode{
				Targets: []game.TargetSpec{
					{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "creature",
						Allow:      game.TargetAllowPermanent,
						Predicate: game.TargetPredicate{
							PermanentTypes: []types.Card{
								types.Creature,
							},
						},
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ModifyPT{
							TargetIndex:    0,
							PowerDelta:     game.Fixed(2),
							ToughnessDelta: game.Fixed(2),
							Duration:       game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability(),
		},
	)
	return card
}()
