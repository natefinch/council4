package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ClanGuildmage is the card definition for Clan Guildmage.
//
// Type: Creature — Human Shaman
// Cost: {R}{G}
//
// Oracle text:
//
//	{1}{R}, {T}: Target creature can't block this turn.
//	{2}{G}, {T}: Target land you control becomes a 4/4 Elemental creature with haste until end of turn. It's still a land.
var ClanGuildmage = newClanGuildmage

func newClanGuildmage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Clan Guildmage",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
				cost.G,
			}),
			Colors:    []color.Color{color.Green, color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Shaman},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{1}{R}, {T}: Target creature can't block this turn.",
					ManaCost:        opt.Val(cost.Mana{cost.O(1), cost.R}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
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
								Primitive: game.ApplyRule{
									Object: opt.Val(game.TargetPermanentReference(0)),
									RuleEffects: []game.RuleEffect{
										game.RuleEffect{
											Kind: game.RuleEffectCantBlock,
										},
									},
									Duration: game.DurationThisTurn,
								},
							},
						},
					}.Ability(),
				},
				game.ActivatedAbility{
					Text:            "{2}{G}, {T}: Target land you control becomes a 4/4 Elemental creature with haste until end of turn. It's still a land.",
					ManaCost:        opt.Val(cost.Mana{cost.O(2), cost.G}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target land you control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Land}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.TargetPermanentReference(0)),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:       game.LayerType,
											AddTypes:    []types.Card{types.Creature},
											AddSubtypes: []types.Sub{types.Elemental},
										},
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											AddKeywords: []game.Keyword{
												game.Haste,
											},
										},
										game.ContinuousEffect{
											Layer:        game.LayerPowerToughnessSet,
											SetPower:     opt.Val(game.PT{Value: 4}),
											SetToughness: opt.Val(game.PT{Value: 4}),
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
			{1}{R}, {T}: Target creature can't block this turn.
			{2}{G}, {T}: Target land you control becomes a 4/4 Elemental creature with haste until end of turn. It's still a land.
		`,
		},
	}
}
