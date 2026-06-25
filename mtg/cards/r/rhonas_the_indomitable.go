package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RhonasTheIndomitable is the card definition for Rhonas the Indomitable.
//
// Type: Legendary Creature — God
// Cost: {2}{G}
//
// Oracle text:
//
//	Deathtouch, indestructible
//	Rhonas can't attack or block unless you control another creature with power 4 or greater.
//	{2}{G}: Another target creature gets +2/+0 and gains trample until end of turn.
var RhonasTheIndomitable = func() *game.CardDef {
	card := &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Rhonas the Indomitable",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.God},
			Power:      opt.Val(game.PT{Value: 5}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			OracleText: `
				Deathtouch, indestructible
				Rhonas can't attack or block unless you control another creature with power 4 or greater.
				{2}{G}: Another target creature gets +2/+0 and gains trample until end of turn.
			`,
		},
	}

	card.StaticAbilities = append(card.StaticAbilities,
		game.DeathtouchStaticBody,
	)

	card.StaticAbilities = append(card.StaticAbilities,
		game.IndestructibleStaticBody,
	)

	card.StaticAbilities = append(card.StaticAbilities, game.StaticAbility{
		Text: `
				Rhonas can't attack or block unless you control another creature with power 4 or greater.
			`,
		Condition: opt.Val(game.Condition{
			Text:   "unless you control another creature with power 4 or greater",
			Negate: true,
			ControlsMatching: opt.Val(game.SelectionCount{
				Selection: game.Selection{
					RequiredTypes: []types.Card{
						types.Creature,
					},
					Power: opt.Val(compare.Int{
						Op:    compare.GreaterOrEqual,
						Value: 4,
					}),
					ExcludeSource: true,
				},
			}),
		}),
		RuleEffects: []game.RuleEffect{
			{
				Kind:           game.RuleEffectCantAttack,
				AffectedSource: true,
			},
			{
				Kind:           game.RuleEffectCantBlock,
				AffectedSource: true,
			},
		},
	},
	)

	card.ActivatedAbilities = append(card.ActivatedAbilities,
		game.ActivatedAbility{
			Text: `
				{2}{G}: Another target creature gets +2/+0 and gains trample until end of turn.
			`,
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Content: game.Mode{
				Targets: []game.TargetSpec{
					{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "another target creature",
						Allow:      game.TargetAllowPermanent,
						Selection: opt.Val(game.Selection{
							RequiredTypesAny: []types.Card{
								types.Creature,
							},
							ExcludeSource: true,
						}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ModifyPT{
							Object:     game.TargetPermanentReference(0),
							PowerDelta: game.Fixed(2),
							Duration:   game.DurationUntilEndOfTurn,
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
