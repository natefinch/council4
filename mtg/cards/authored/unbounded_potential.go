package authored

import (
	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const unboundedPotentialOracleID = "d9ae08cf-eef0-407a-ad2c-08b4c89f99f0"

var (
	unboundedPotentialCounters    = game.ProgramReference{ID: unboundedPotentialOracleID + "/spell/0", Version: "1"}
	unboundedPotentialProliferate = game.ProgramReference{ID: unboundedPotentialOracleID + "/spell/1", Version: "1"}
)

var UnboundedPotential = cardset.Entry{
	OracleID: unboundedPotentialOracleID,
	Name:     "Unbounded Potential",
	New:      newUnboundedPotential,
	Programs: []cardprogram.Program{
		{Reference: unboundedPotentialCounters, Resolve: resolveUnboundedPotentialCounters},
		{Reference: unboundedPotentialProliferate, Resolve: resolveUnboundedPotentialProliferate},
	},
	Verification: "mtg/rules/unbounded_potential_test.go: TestUnboundedPotentialModesAndEntwine",
}

func newUnboundedPotential() *game.CardDef {
	const counters = "Put a +1/+1 counter on each of up to two target creatures."
	const proliferate = "Proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)"
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:       "Unbounded Potential",
			OracleText: "Choose one —\n• " + counters + "\n• " + proliferate + "\nEntwine {3}{W} (Choose both if you pay the entwine cost.)",
			ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.W}),
			Types:      []types.Card{types.Instant},
			Colors:     []color.Color{color.White},
			SpellAbility: opt.Val(game.AbilityContent{
				MinModes: 1, MaxModes: 2,
				EscalateCost: opt.Val(cost.Mana{cost.O(3), cost.W}),
				Modes: []game.Mode{
					{Text: counters, Program: unboundedPotentialCounters, Targets: []game.TargetSpec{{
						MinTargets: 0, MaxTargets: 2, Constraint: "up to two target creatures",
						Allow:     game.TargetAllowPermanent,
						Selection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
					}}},
					{Text: proliferate, Program: unboundedPotentialProliferate},
				},
			}),
		},
	}
}

func resolveUnboundedPotentialCounters(ctx cardprogram.Context) error {
	_, err := ctx.Execute(game.AddCounter{
		Object: game.AllTargetPermanentsReference(0), Amount: game.Fixed(1), CounterKind: counter.PlusOnePlusOne,
	})
	return err
}

func resolveUnboundedPotentialProliferate(ctx cardprogram.Context) error {
	_, err := ctx.Execute(game.Proliferate{})
	return err
}
