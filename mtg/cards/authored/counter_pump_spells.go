package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

var CommonBond = cardset.Entry{
	OracleID: "d52e9255-62e7-479f-8def-e92b96efdb89", Name: "Common Bond", New: newCommonBond,
	Verification: "mtg/rules/counter_pump_spells_test.go: TestCommonBondScenario",
}

var SeedsOfStrength = cardset.Entry{
	OracleID: "dd3d4d96-1e88-49e1-8276-7f62c22ad91a", Name: "Seeds of Strength", New: newSeedsOfStrength,
	Verification: "mtg/rules/counter_pump_spells_test.go: TestSeedsOfStrengthScenario",
}

var MartialGlory = cardset.Entry{
	OracleID: "e8aa12f8-8922-41ed-8cdb-219c9e4b9e36", Name: "Martial Glory", New: newMartialGlory,
	Verification: "mtg/rules/counter_pump_spells_test.go: TestMartialGloryScenario",
}

func newCommonBond() *game.CardDef {
	const text = "Put a +1/+1 counter on target creature.\nPut a +1/+1 counter on target creature."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green, color.White),
		CardFace: game.CardFace{
			Name: "Common Bond", OracleText: text,
			ManaCost: opt.Val(cost.Mana{cost.O(1), cost.G, cost.W}),
			Types:    []types.Card{types.Instant},
			Colors:   []color.Color{color.Green, color.White},
			SpellAbility: opt.Val(game.Mode{
				Text:    text,
				Targets: []game.TargetSpec{creatureTarget(), creatureTarget()},
				Sequence: []game.Instruction{
					{Primitive: game.AddCounter{
						Object: game.TargetPermanentReference(0), CounterKind: counter.PlusOnePlusOne, Amount: game.Fixed(1),
					}},
					{Primitive: game.AddCounter{
						Object: game.TargetPermanentReference(1), CounterKind: counter.PlusOnePlusOne, Amount: game.Fixed(1),
					}},
				},
			}.Ability()),
		},
	}
}

func newSeedsOfStrength() *game.CardDef {
	const text = "Target creature gets +1/+1 until end of turn.\nTarget creature gets +1/+1 until end of turn.\nTarget creature gets +1/+1 until end of turn."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green, color.White),
		CardFace: game.CardFace{
			Name: "Seeds of Strength", OracleText: text,
			ManaCost: opt.Val(cost.Mana{cost.G, cost.W}),
			Types:    []types.Card{types.Instant},
			Colors:   []color.Color{color.Green, color.White},
			SpellAbility: opt.Val(game.Mode{
				Text: text,
				Targets: []game.TargetSpec{
					creatureTarget(), creatureTarget(), creatureTarget(),
				},
				Sequence: []game.Instruction{
					{Primitive: game.ModifyPT{
						Object: game.TargetPermanentReference(0), PowerDelta: game.Fixed(1),
						ToughnessDelta: game.Fixed(1), Duration: game.DurationUntilEndOfTurn,
					}},
					{Primitive: game.ModifyPT{
						Object: game.TargetPermanentReference(1), PowerDelta: game.Fixed(1),
						ToughnessDelta: game.Fixed(1), Duration: game.DurationUntilEndOfTurn,
					}},
					{Primitive: game.ModifyPT{
						Object: game.TargetPermanentReference(2), PowerDelta: game.Fixed(1),
						ToughnessDelta: game.Fixed(1), Duration: game.DurationUntilEndOfTurn,
					}},
				},
			}.Ability()),
		},
	}
}

func newMartialGlory() *game.CardDef {
	const text = "Target creature gets +3/+0 until end of turn.\nTarget creature gets +0/+3 until end of turn."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.White),
		CardFace: game.CardFace{
			Name: "Martial Glory", OracleText: text,
			ManaCost: opt.Val(cost.Mana{cost.R, cost.W}),
			Types:    []types.Card{types.Instant},
			Colors:   []color.Color{color.Red, color.White},
			SpellAbility: opt.Val(game.Mode{
				Text:    text,
				Targets: []game.TargetSpec{creatureTarget(), creatureTarget()},
				Sequence: []game.Instruction{
					{Primitive: game.ModifyPT{
						Object: game.TargetPermanentReference(0), PowerDelta: game.Fixed(3),
						Duration: game.DurationUntilEndOfTurn,
					}},
					{Primitive: game.ModifyPT{
						Object: game.TargetPermanentReference(1), ToughnessDelta: game.Fixed(3),
						Duration: game.DurationUntilEndOfTurn,
					}},
				},
			}.Ability()),
		},
	}
}
