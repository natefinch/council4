package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestScopedResultMultiModeUsesLexicalNamespaces(t *testing.T) {
	body := "\n\u2022 Discard a card. If a land card was discarded this way, you gain 1 life.\n" +
		"\u2022 If you control a creature, discard a card. If a land card was discarded this way, you gain 5 life."
	for _, test := range []struct {
		header    string
		countPath string
	}{
		{"Choose two \u2014", "SpellAbility.Val.MaxModes = 2"},
		{"Choose one or both \u2014", "SpellAbility.Val.MaxModes = 2"},
		{"Choose one. If you control a commander as you cast this spell, you may choose both instead.",
			"SpellAbility.Val.ModeChoiceBonus.AdditionalMaxModes = 1"},
	} {
		t.Run(test.header, func(t *testing.T) {
			assertCardPaths(t, &ScryfallCard{
				Name: "Mode Result Namespaces", Layout: "normal", TypeLine: "Sorcery", OracleText: test.header + body,
			}, test.countPath,
				"Modes[0].Sequence[0].LocalProducts.Results[0]",
				"Modes[1].Sequence[0].LocalProducts.Results[0]",
				"Modes[0].Sequence[1].ResultGate.Val.Key",
				"Modes[1].Sequence[1].ResultGate.Val.Key")
		})
	}
	assertCardPaths(t, &ScryfallCard{
		Name: "One Mode Scope", Layout: "normal", TypeLine: "Sorcery", OracleText: "Choose one \u2014" + body,
	}, "SpellAbility.Val.MaxModes = 1",
		"Modes[1].Sequence[0].PublishResult = \"if-you-do\"",
		"Modes[1].Sequence[1].ResultGate.Val.Key = \"if-you-do\"")
}

func TestModeResultScopeKeysAndNestedPublications(t *testing.T) {
	for _, tt := range []struct {
		name      string
		modes     []game.Mode
		supported bool
	}{
		{
			"unconditional overwrite",
			[]game.Mode{
				{Sequence: []game.Instruction{{PublishResult: "same"}}},
				{Sequence: []game.Instruction{{PublishResult: "same"}}},
			}, true,
		},
		{
			"skipped later publisher",
			[]game.Mode{
				{Sequence: []game.Instruction{{PublishResult: "same"}}},
				{Sequence: []game.Instruction{{PublishResult: "same", ConditionGate: "condition"}}},
			}, false,
		},
		{
			"skipped earlier publisher",
			[]game.Mode{
				{Sequence: []game.Instruction{{PublishResult: "same", ConditionGate: "condition"}}},
				{Sequence: []game.Instruction{{PublishResult: "same"}}},
			}, false,
		},
		{
			"distinct scoped keys",
			[]game.Mode{
				{Sequence: []game.Instruction{{PublishResult: "result-clause-1", ConditionGate: "condition"}}},
				{Sequence: []game.Instruction{{PublishResult: "result-clause-2", ConditionGate: "condition"}}},
			}, true,
		},
		{
			"repeat body shares results and may not execute",
			[]game.Mode{
				{Sequence: []game.Instruction{{Primitive: game.RepeatProcess{
					Body: game.Mode{Sequence: []game.Instruction{{PublishResult: "same"}}}.Ability(),
				}}}},
				{Sequence: []game.Instruction{{PublishResult: "same"}}},
			}, false,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := modeResultScopesCompatible(tt.modes); got != tt.supported {
				t.Fatalf("supported=%t, want %t", got, tt.supported)
			}
		})
	}
}
