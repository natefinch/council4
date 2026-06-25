package payment

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestIsAutomaticManaAbility(t *testing.T) {
	tests := []struct {
		name string
		body game.ManaAbility
		want bool
	}{
		{
			name: "fixed tap output",
			body: game.TapManaAbility(mana.G),
			want: true,
		},
		{
			name: "mana choice",
			body: game.TapManaChoiceAbility(mana.G, mana.U),
		},
		{
			name: "additional rider",
			body: game.ManaAbility{
				AdditionalCosts: cost.Tap,
				Content: game.Mode{Sequence: []game.Instruction{
					{Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.G}},
					{Primitive: game.GainLife{Amount: game.Fixed(1)}},
				}}.Ability(),
			},
		},
		{
			name: "entry choice output",
			body: game.ManaAbility{
				AdditionalCosts: cost.Tap,
				Content: game.Mode{Sequence: []game.Instruction{{
					Primitive: game.AddMana{
						Amount:          game.Fixed(1),
						EntryChoiceFrom: game.EntryColorChoiceKey,
					},
				}}}.Ability(),
			},
		},
		{
			name: "sacrifice source",
			body: game.ManaAbility{
				AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalSacrificeSource}},
				Content: game.Mode{Sequence: []game.Instruction{{
					Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.C},
				}}}.Ability(),
			},
		},
		{
			// A fixed-output mana ability that tags its mana with a spend rider
			// must not be auto-activated: the automatic path adds untagged pool
			// mana and would silently drop the rider, so it stays a manual choice.
			name: "fixed output with spend rider",
			body: game.ManaAbility{
				AdditionalCosts: cost.Tap,
				Content: game.Mode{Sequence: []game.Instruction{{
					Primitive: game.AddMana{
						Amount:    game.Fixed(1),
						ManaColor: mana.G,
						SpendRider: opt.Val(game.ManaSpendRider{
							Condition: game.ManaSpendCastCommanderCreatureType,
							Effect: game.Mode{Sequence: []game.Instruction{
								{Primitive: game.Scry{Amount: game.Fixed(1), Player: game.ControllerReference()}},
							}},
						}),
					},
				}}}.Ability(),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := IsAutomaticManaAbility(&test.body); got != test.want {
				t.Fatalf("IsAutomaticManaAbility() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestPaymentManaAbilityOutputRequiresCanonicalTapAnyColorAbility(t *testing.T) {
	canonical := game.TapAnyColorManaAbility()
	if output, ok := paymentManaAbilityOutput(&canonical); !ok ||
		output.amount != 1 ||
		len(output.colors) != 5 {
		t.Fatalf("paymentManaAbilityOutput(canonical) = (%+v, %v), want WUBRG one-mana output", output, ok)
	}

	tests := []struct {
		name   string
		mutate func(*game.ManaAbility)
	}{
		{
			name: "mutated tap cost",
			mutate: func(ability *game.ManaAbility) {
				ability.AdditionalCosts[0].Kind = cost.AdditionalUntap
			},
		},
		{
			name: "nonbattlefield zone",
			mutate: func(ability *game.ManaAbility) {
				ability.ZoneOfFunction = zone.Hand
			},
		},
		{
			name: "activation condition",
			mutate: func(ability *game.ManaAbility) {
				ability.ActivationCondition = opt.Val(game.Condition{Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerLife, Op: compare.GreaterOrEqual, Value: 1}}})
			},
		},
		{
			name: "commander identity color source with WUBRG colors",
			mutate: func(ability *game.ManaAbility) {
				choose, ok := ability.Content.Modes[0].Sequence[0].Primitive.(game.Choose)
				if !ok {
					panic("TapAnyColorManaAbility choice instruction is not Choose")
				}
				choose.Choice.ColorSource = game.ResolutionChoiceColorSourceCommanderIdentity
				ability.Content.Modes[0].Sequence[0].Primitive = choose
			},
		},
		{
			name: "condition-gated choice",
			mutate: func(ability *game.ManaAbility) {
				ability.Content.Modes[0].Sequence[0].Condition = opt.Val(game.EffectCondition{Text: "condition"})
			},
		},
		{
			name: "result-gated AddMana",
			mutate: func(ability *game.ManaAbility) {
				ability.Content.Modes[0].Sequence[1].ResultGate = opt.Val(game.InstructionResultGate{
					Key:       "prior",
					Succeeded: game.TriTrue,
				})
			},
		},
		{
			name: "optional AddMana",
			mutate: func(ability *game.ManaAbility) {
				ability.Content.Modes[0].Sequence[1].Optional = true
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ability := game.TapAnyColorManaAbility()
			test.mutate(&ability)
			if output, ok := paymentManaAbilityOutput(&ability); ok {
				t.Fatalf("paymentManaAbilityOutput() = (%+v, true), want rejected", output)
			}
		})
	}
}

func TestSacrificeManaChoiceOutput(t *testing.T) {
	treasure := game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G)
	treasure.AdditionalCosts = append(treasure.AdditionalCosts, cost.Additional{
		Kind:               cost.AdditionalSacrificeSource,
		Amount:             1,
		MatchPermanentType: true,
		PermanentType:      types.Artifact,
	})
	if colors, amount, ok := sacrificeManaChoiceOutput(&treasure); !ok || amount != 1 || len(colors) != 5 {
		t.Fatalf("sacrificeManaChoiceOutput(Treasure) = (%v, %d, %v), want five colors, one mana, true", colors, amount, ok)
	}

	withNonManaEffect := treasure
	withNonManaEffect.Content.Modes[0].Sequence = append(
		withNonManaEffect.Content.Modes[0].Sequence,
		game.Instruction{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}},
	)
	if _, _, ok := sacrificeManaChoiceOutput(&withNonManaEffect); ok {
		t.Fatal("sacrificeManaChoiceOutput() accepted a mana ability with a non-mana effect")
	}

	withoutSacrifice := game.TapManaChoiceAbility(mana.W, mana.U)
	if _, _, ok := sacrificeManaChoiceOutput(&withoutSacrifice); ok {
		t.Fatal("sacrificeManaChoiceOutput() accepted a choice mana ability without sacrificing its source")
	}
}

func TestSacrificeManaChoiceOutputRejectsInstructionGating(t *testing.T) {
	newTreasure := func() game.ManaAbility {
		treasure := game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G)
		treasure.AdditionalCosts = append(treasure.AdditionalCosts, cost.Additional{
			Kind:               cost.AdditionalSacrificeSource,
			Amount:             1,
			MatchPermanentType: true,
			PermanentType:      types.Artifact,
		})
		return treasure
	}
	tests := []struct {
		name   string
		mutate func([]game.Instruction)
	}{
		{
			name: "false-gated choice",
			mutate: func(sequence []game.Instruction) {
				sequence[0].ResultGate = opt.Val(game.InstructionResultGate{
					Key:       "prior",
					Succeeded: game.TriFalse,
				})
			},
		},
		{
			name: "condition-gated choice",
			mutate: func(sequence []game.Instruction) {
				sequence[0].Condition = opt.Val(game.EffectCondition{Text: "condition"})
			},
		},
		{
			name: "result-gated AddMana",
			mutate: func(sequence []game.Instruction) {
				sequence[1].ResultGate = opt.Val(game.InstructionResultGate{
					Key:       "prior",
					Succeeded: game.TriTrue,
				})
			},
		},
		{
			name: "optional AddMana",
			mutate: func(sequence []game.Instruction) {
				sequence[1].Optional = true
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			treasure := newTreasure()
			test.mutate(treasure.Content.Modes[0].Sequence)
			if _, _, ok := sacrificeManaChoiceOutput(&treasure); ok {
				t.Fatal("sacrificeManaChoiceOutput() accepted behaviorally gated instructions")
			}
		})
	}
}
