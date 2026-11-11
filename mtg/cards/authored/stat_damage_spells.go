package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

var AgonyWarp = cardset.Entry{
	OracleID: "377c217b-ddad-4e1a-9122-f347113d3dfa",
	Name:     "Agony Warp",
	New:      newAgonyWarp,
	Verification: "mtg/rules/authored_stat_damage_spells_test.go: " +
		"TestAgonyWarpScenario",
}

var HeatedDebate = cardset.Entry{
	OracleID: "06c3e85c-17d8-44e1-a0f8-07ef3f1624cb",
	Name:     "Heated Debate",
	New:      newHeatedDebate,
	Verification: "mtg/rules/authored_stat_damage_spells_test.go: " +
		"TestHeatedDebateScenario",
}

var SliceFromTheShadows = cardset.Entry{
	OracleID: "4d31d3f4-4748-4713-808f-24699fcd7b10",
	Name:     "Slice from the Shadows",
	New:      newSliceFromTheShadows,
	Verification: "mtg/rules/authored_stat_damage_spells_test.go: " +
		"TestSliceFromTheShadowsScenario",
}

func creatureTarget() game.TargetSpec {
	return game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Constraint: "target creature",
		Allow:      game.TargetAllowPermanent,
		Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
	}
}

func creatureOrPlaneswalkerTarget() game.TargetSpec {
	return game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Constraint: "target creature or planeswalker",
		Allow:      game.TargetAllowPermanent,
		Selection: opt.Val(game.Selection{
			RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker},
		}),
	}
}

func newAgonyWarp() *game.CardDef {
	const text = "Target creature gets -3/-0 until end of turn.\nTarget creature gets -0/-3 until end of turn."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Agony Warp", OracleText: text,
			ManaCost: opt.Val(cost.Mana{cost.U, cost.B}),
			Types:    []types.Card{types.Instant},
			Colors:   []color.Color{color.Blue, color.Black},
			SpellAbility: opt.Val(game.Mode{
				Text: text,
				Targets: []game.TargetSpec{
					creatureTarget(),
					creatureTarget(),
				},
				Sequence: []game.Instruction{
					{Primitive: game.ModifyPT{
						Object: game.TargetPermanentReference(0), PowerDelta: game.Fixed(-3),
						Duration: game.DurationUntilEndOfTurn,
					}},
					{Primitive: game.ModifyPT{
						Object: game.TargetPermanentReference(1), ToughnessDelta: game.Fixed(-3),
						Duration: game.DurationUntilEndOfTurn,
					}},
				},
			}.Ability()),
		},
	}
}

func newHeatedDebate() *game.CardDef {
	const text = "This spell can't be countered. (This includes by the ward ability.)\nHeated Debate deals 4 damage to target creature or planeswalker."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Heated Debate", OracleText: text,
			ManaCost: opt.Val(cost.Mana{cost.O(2), cost.R}),
			Types:    []types.Card{types.Instant},
			Colors:   []color.Color{color.Red},
			StaticAbilities: []game.StaticAbility{
				game.CantBeCounteredStaticBody,
			},
			SpellAbility: opt.Val(game.Mode{
				Text: text,
				Targets: []game.TargetSpec{
					creatureOrPlaneswalkerTarget(),
				},
				Sequence: []game.Instruction{{
					Primitive: game.Damage{
						Amount:    game.Fixed(4),
						Recipient: game.ObjectDamageRecipient(game.TargetPermanentReference(0)),
					},
				}},
			}.Ability()),
		},
	}
}

func newSliceFromTheShadows() *game.CardDef {
	const text = "This spell can't be countered. (This includes by the ward ability.)\nTarget creature gets -X/-X until end of turn."
	minusX := game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX, Multiplier: -1})
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Slice from the Shadows", OracleText: text,
			ManaCost: opt.Val(cost.Mana{cost.X, cost.B}),
			Types:    []types.Card{types.Instant},
			Colors:   []color.Color{color.Black},
			StaticAbilities: []game.StaticAbility{
				game.CantBeCounteredStaticBody,
			},
			SpellAbility: opt.Val(game.Mode{
				Text:    text,
				Targets: []game.TargetSpec{creatureTarget()},
				Sequence: []game.Instruction{{
					Primitive: game.ModifyPT{
						Object: game.TargetPermanentReference(0), PowerDelta: minusX,
						ToughnessDelta: minusX, Duration: game.DurationUntilEndOfTurn,
					},
				}},
			}.Ability()),
		},
	}
}
