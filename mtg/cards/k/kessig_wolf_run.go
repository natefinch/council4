package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KessigWolfRun is the card definition for Kessig Wolf Run.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add {C}.
//	{X}{R}{G}, {T}: Target creature gets +X/+0 and gains trample until end of turn.
var KessigWolfRun = func() *game.CardDef {
	card := &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green, color.Red),
		CardFace: game.CardFace{
			Name:  "Kessig Wolf Run",
			Types: []types.Card{types.Land},
			OracleText: `
				{T}: Add {C}.
				{X}{R}{G}, {T}: Target creature gets +X/+0 and gains trample until end of turn.
			`,
		},
	}

	card.ManaAbilities = append(card.ManaAbilities, game.TapManaAbility(mana.C))

	card.ActivatedAbilities = append(card.ActivatedAbilities,
		game.ActivatedAbility{
			Text: `
				{X}{R}{G}, {T}: Target creature gets +X/+0 and gains trample until end of turn.
			`,
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.R,
				cost.G,
			}),
			AdditionalCosts: cost.Tap,
			Content: game.Mode{
				Targets: []game.TargetSpec{
					{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "creature",
						Allow:      game.TargetAllowPermanent,
						Selection: opt.Val(game.Selection{
							RequiredTypesAny: []types.Card{
								types.Creature,
							},
						}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ModifyPT{
							Object: game.TargetPermanentReference(0),
							PowerDelta: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							}),
							Duration: game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.ApplyContinuous{
							Object: opt.Val(game.TargetPermanentReference(0)),
							ContinuousEffects: []game.ContinuousEffect{
								{
									Layer: game.LayerAbility,
									AddKeywords: []game.Keyword{
										game.Trample,
									},
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability(),
		},
	)
	return card
}()
