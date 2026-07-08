package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Aetherling is the card definition for Aetherling.
//
// Type: Creature — Shapeshifter
// Cost: {4}{U}{U}
//
// Oracle text:
//
//	{U}: Exile this creature. Return it to the battlefield under its owner's control at the beginning of the next end step.
//	{U}: This creature can't be blocked this turn.
//	{1}: This creature gets +1/-1 until end of turn.
//	{1}: This creature gets -1/+1 until end of turn.
var Aetherling = newAetherling

func newAetherling() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Aetherling",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Shapeshifter},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 5}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{U}: Exile this creature. Return it to the battlefield under its owner's control at the beginning of the next end step.",
					ManaCost:       opt.Val(cost.Mana{cost.U}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Exile{
									Object:         game.SourcePermanentReference(),
									ExileLinkedKey: game.LinkedKey("delayed-self-blink"),
								},
							},
							{
								Primitive: game.CreateDelayedTrigger{
									Trigger: game.DelayedTriggerDef{
										Timing: game.DelayedAtBeginningOfNextEndStep,
										Content: game.Mode{
											Sequence: []game.Instruction{
												{
													Primitive: game.PutOnBattlefield{
														Source: game.LinkedBattlefieldSource(game.LinkedKey("delayed-self-blink")),
													},
												},
											},
										}.Ability(),
									},
								},
							},
						},
					}.Ability(),
				},
				game.ActivatedAbility{
					Text:           "{U}: This creature can't be blocked this turn.",
					ManaCost:       opt.Val(cost.Mana{cost.U}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyRule{
									Object: opt.Val(game.SourcePermanentReference()),
									RuleEffects: []game.RuleEffect{
										game.RuleEffect{
											Kind: game.RuleEffectCantBeBlocked,
										},
									},
									Duration: game.DurationThisTurn,
								},
							},
						},
					}.Ability(),
				},
				game.ActivatedAbility{
					Text:           "{1}: This creature gets +1/-1 until end of turn.",
					ManaCost:       opt.Val(cost.Mana{cost.O(1)}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ModifyPT{
									Object:         game.SourcePermanentReference(),
									PowerDelta:     game.Fixed(1),
									ToughnessDelta: game.Fixed(-1),
									Duration:       game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
				game.ActivatedAbility{
					Text:           "{1}: This creature gets -1/+1 until end of turn.",
					ManaCost:       opt.Val(cost.Mana{cost.O(1)}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ModifyPT{
									Object:         game.SourcePermanentReference(),
									PowerDelta:     game.Fixed(-1),
									ToughnessDelta: game.Fixed(1),
									Duration:       game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{U}: Exile this creature. Return it to the battlefield under its owner's control at the beginning of the next end step.
			{U}: This creature can't be blocked this turn.
			{1}: This creature gets +1/-1 until end of turn.
			{1}: This creature gets -1/+1 until end of turn.
		`,
		},
	}
}
