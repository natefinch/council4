package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

func TestScopedResultFlowCompositions(t *testing.T) {
	for _, tt := range []struct {
		name, text string
		paths      []string
		absent     []string
	}{
		{
			name: "active targeted object destruction",
			text: "Destroy target creature. If you destroyed a creature this way, you gain 2 life.",
			paths: []string{
				"Sequence[0].PublishResult",
				"Sequence[1].ResultGate.Val.ObjectSelection.Val.RequiredTypes[0] = types.Creature",
			},
		},
		{
			name: "bare optional independent rider",
			text: "You may discard a card. Draw a card. You may sacrifice a creature. You gain 2 life.",
			paths: []string{
				"Sequence[0].Optional = true",
				"Sequence[2].Optional = true",
			},
			absent: []string{
				"Sequence[1].Optional = true",
				"Sequence[3].Optional = true",
				"ResultGate.Exists = true",
			},
		},
		{
			name: "optional explicit failure without affirmative",
			text: "You may discard a card. If you don't, draw a card.",
			paths: []string{
				"Sequence[0].Optional = true",
				"Sequence[1].ResultGate.Val.Succeeded = game.TriFalse",
			},
		},
		{
			name:   "optional independent trailing scry",
			text:   "You may discard a card. If you do, draw a card. Scry 2.",
			paths:  []string{"Sequence[1].ResultGate.Val.Succeeded = game.TriTrue"},
			absent: []string{"Sequence[2].ResultGate.Exists = true"},
		},
		{
			name: "mandatory independent rider",
			text: "Discard a card. If a land card is discarded this way, draw a card. You gain 2 life.",
			paths: []string{
				"Sequence[0].PublishResult",
				"Sequence[1].ResultGate.Val.ObjectSelection.Val.RequiredTypes[0] = types.Land",
			},
			absent: []string{"Sequence[2].ResultGate.Exists = true"},
		},
		{
			name: "optional intervening rider",
			text: "You may discard a card. You gain 2 life. If a land card is discarded this way, draw a card.",
			paths: []string{
				"Sequence[0].Optional = true",
				"Sequence[0].PublishResult",
				"Sequence[2].ResultGate.Val.ObjectSelection.Val.RequiredTypes[0] = types.Land",
			},
			absent: []string{"Sequence[1].ResultGate.Exists = true"},
		},
		{
			name: "multiple filters one actual publication",
			text: "Discard a card. If a land card is discarded this way, draw a card. If a creature card is discarded this way, you gain 2 life.",
			paths: []string{
				"Sequence[1].ResultGate.Val.ObjectSelection.Val.RequiredTypes[0] = types.Land",
				"Sequence[2].ResultGate.Val.ObjectSelection.Val.RequiredTypes[0] = types.Creature",
			},
		},
		{
			name: "mandatory filtered complement",
			text: "Discard a card. If a land card is discarded this way, draw two cards. Otherwise, draw a card. You gain 2 life.",
			paths: []string{
				"Sequence[2].ResultGate.Val.ObjectSelection.Val.RequiredTypes[0] = types.Land",
				"Sequence[2].ResultGate.Val.Negate = true",
			},
			absent: []string{"Sequence[3].ResultGate.Exists = true"},
		},
		{
			name: "independent optional producers",
			text: "You may discard a card. If you do, draw a card. You may sacrifice a creature. If you do, you gain 2 life.",
			paths: []string{
				"Sequence[0].Optional = true",
				"Sequence[2].Optional = true",
				"Sequence[0].PublishResult = \"result-clause-1\"",
				"Sequence[1].ResultGate.Val.Key = \"result-clause-1\"",
				"Sequence[2].PublishResult = \"result-clause-3\"",
				"Sequence[3].ResultGate.Val.Key = \"result-clause-3\"",
			},
		},
		{
			name: "nested optional failure",
			text: "You may discard a card. If you do, you may sacrifice a creature. If you don't, you gain 2 life.",
			paths: []string{
				"Sequence[0].PublishResult = \"result-clause-1\"",
				"Sequence[1].PublishResult = \"result-clause-2\"",
				"Sequence[1].ResultGate.Val.Key = \"result-clause-1\"",
				"Sequence[2].ResultGate.Val.Key = \"result-clause-2\"",
				"Sequence[2].ResultGate.Val.Succeeded = game.TriFalse",
			},
		},
		{
			name: "filtered complement continuation",
			text: "Discard a card. If a land card was discarded this way, you gain 2 life. Otherwise, put a +1/+1 counter on target creature. It gains trample until end of turn. You gain 1 life.",
			paths: []string{
				"Sequence[2].ResultGate.Val.Negate = true",
				"Sequence[3].ResultGate.Val.Negate = true",
			},
			absent: []string{"Sequence[4].ResultGate.Exists = true"},
		},
		{
			name: "expanded consumer",
			text: "Discard a card. If a land card is discarded this way, put a +1/+1 counter on each of up to two target creatures. You gain 2 life.",
			paths: []string{
				"Sequence[1].ResultGate.Val.Key",
				"Sequence[2].ResultGate.Val.Key",
			},
			absent: []string{"Sequence[3].ResultGate.Exists = true"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			card := &ScryfallCard{Name: "Scoped Result Probe", Layout: "normal", TypeLine: "Sorcery", OracleText: tt.text}
			assertCardPaths(t, card, tt.paths...)
			assertCardPathsAbsent(t, card, tt.absent...)
		})
	}
}

func TestQuantifiedResultCardPathsAndNearMisses(t *testing.T) {
	for _, text := range []string{
		"Discard three cards. If two or more land cards were discarded this way, draw a card.",
		"Discard three cards. If you discarded two or more land cards this way, draw a card.",
		"Mill three cards. If two or more creature cards were milled this way, draw a card.",
	} {
		assertCardPaths(t, &ScryfallCard{Name: "Count Probe", Layout: "normal", TypeLine: "Sorcery", OracleText: text},
			"Sequence[1].ResultGate.Val.ObjectCountRange.Val.Min = 2",
			"Sequence[1].ResultGate.Val.CardOnly = true",
		)
	}
	for _, tt := range []struct{ text, reason string }{
		{"Repeat the following process two times. You may draw a card and gain 2 life.", "action group acceptance not modeled"},
		{"You may discard a card. When you do, draw a card. You may sacrifice a creature. If you do, you gain 2 life.", "mixed reflexive actual-result flow not modeled"},
		{"You may discard a card. When you do, draw a card. If you do, you gain 2 life.", "mixed reflexive actual-result flow not modeled"},
		{"Target player discards a card. If you discarded a land card this way, draw a card.", "actor ownership not modeled"},
		{"If a land card was discarded this way, draw a card. Discard a card.", "no unique earlier typed producer"},
		{"Put a +1/+1 counter on each of up to two target creatures. If you do, draw a card.", "requires one instruction"},
		{"Draw a card and gain 2 life. If you do, scry 1.", "no unique earlier typed producer"},
		{"You may have target player discard a card and draw a card. If you do, scry 1.", "action group acceptance not modeled"},
		{"You may draw a card and gain 2 life. If you do, scry 1.", "action group acceptance not modeled"},
		{"{T}, Discard a card: If a land card was discarded this way, draw a card.", "unsupported draw spell"},
	} {
		typ := "Sorcery"
		if tt.text[0] == '{' {
			typ = "Artifact"
		}
		assertCardUnsupported(t, &ScryfallCard{Name: "Result Refusal", Layout: "normal", TypeLine: typ, OracleText: tt.text}, tt.reason)
	}
}

func TestSpecializedSacrificePreservesQuantifiedGate(t *testing.T) {
	assertCardPaths(t, &ScryfallCard{
		Name: "Quantified Sacrifice Return", Layout: "normal", TypeLine: "Sorcery",
		OracleText: "You may sacrifice a creature. If two creatures were sacrificed this way, return that card to the battlefield under its owner's control with three +1/+1 counters on it.",
	},
		"Sequence[1].ResultGate.Val.ObjectCountRange.Val.Min = 2",
		"Sequence[1].ResultGate.Val.ObjectCountRange.Val.Max = 2",
	)
}

func TestScopedResultPlannerUsesOnlyTypedIdentity(t *testing.T) {
	document, diagnostics := parser.Parse("Discard a card. You gain 2 life. If a land card was discarded this way, draw a card.", parser.Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	compilation, diagnostics := compiler.Compile(document, compiler.Context{})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	content := compilation.Abilities[0].Content
	for ei := range content.Effects {
		content.Effects[ei].Text = "opaque"
		content.Effects[ei].Order = shared.SourceOrder{}
	}
	for ci := range content.Conditions {
		content.Conditions[ci].Text = "opaque"
		content.Conditions[ci].Order = shared.SourceOrder{}
	}
	plan, ok := planOptionalFlow(content)
	if !ok || plan.scoped.gates[2].Key == "" || plan.scoped.gates[1].Key != "" {
		t.Fatalf("plan = %#v, %t", plan, ok)
	}
	content.Conditions[0].Ownership.ResultProducerClauseID = content.Effects[2].ClauseID
	if _, ok := planOptionalFlow(content); ok {
		t.Fatal("forward producer identity was admitted")
	}
	content.Conditions[0].Ownership.ResultProducerClauseID = 999
	if _, ok := planOptionalFlow(content); ok {
		t.Fatal("missing producer identity was admitted")
	}
	content.Conditions[0].Ownership.ResultProducerClauseID = content.Effects[0].ClauseID
	content.Effects[1].ClauseID = content.Effects[0].ClauseID
	if _, ok := planOptionalFlow(content); ok {
		t.Fatal("ambiguous producer identity was admitted")
	}
}

func TestScopedResultActorOwnership(t *testing.T) {
	for _, tt := range []struct {
		text      string
		supported bool
	}{
		{"Target opponent exiles a card from their hand. If you exiled a land card this way, you gain 2 life.", false},
		{"Target player exiles a card from their hand. If you exiled a land card this way, you gain 2 life.", false},
		{"Target opponent exiles a card from their hand. If a land card was exiled this way, you gain 2 life.", true},
		{"Exile target creature. If you exiled a creature card this way, you gain 2 life.", true},
	} {
		t.Run(tt.text, func(t *testing.T) {
			document, diagnostics := parser.Parse(tt.text, parser.Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			compilation, diagnostics := compiler.Compile(document, compiler.Context{})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			plan, ok := planOptionalFlow(compilation.Abilities[0].Content)
			if ok != tt.supported {
				t.Fatalf("supported=%t, want %t: %s", ok, tt.supported, plan.failureCategory)
			}
		})
	}
}

func TestScopedOptionalMissingGroupIdentityFailsClosed(t *testing.T) {
	document, diagnostics := parser.Parse("You may discard a card. You gain 2 life.", parser.Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	compilation, diagnostics := compiler.Compile(document, compiler.Context{})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	content := compilation.Abilities[0].Content
	content.Effects[0].OptionalActionClauseIDs = nil
	if _, ok := planOptionalFlow(content); ok {
		t.Fatal("optional action with missing group identity was admitted")
	}
}
