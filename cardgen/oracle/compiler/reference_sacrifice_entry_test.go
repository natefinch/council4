package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestSacrificeEntryCounterAnaphorKeepsCardDomain(t *testing.T) {
	for _, verb := range []string{"put that card onto", "return that card to"} {
		for _, event := range []string{
			"Whenever an opponent sacrifices a nontoken permanent, ",
			"Whenever a creature dies, ",
		} {
			t.Run(event+verb, func(t *testing.T) {
				document, diagnostics := parser.Parse(event+verb+
					" the battlefield under your control with a +1/+1 counter on it.", parser.Context{})
				if len(diagnostics) != 0 {
					t.Fatal(diagnostics)
				}

				compilation, diagnostics := Compile(document, Context{})
				if len(diagnostics) != 0 {
					t.Fatal(diagnostics)
				}
				references := compilation.Abilities[0].Content.References
				if len(references) != 2 {
					t.Fatalf("references=%+v", references)
				}
				for _, reference := range references {
					if reference.Binding != ReferenceBindingEventCard ||
						!reference.SupportsUse(ReferenceUseCardAction) ||
						reference.SupportsUse(ReferenceUseMutation) {
						t.Fatalf("entry rider must act on exact card, not departed permanent: %+v", reference)
					}
				}
			})
		}
	}
}

func TestSacrificeCounterWithoutEntryDoesNotAcquireCardAction(t *testing.T) {
	document, diagnostics := parser.Parse(
		"Whenever an opponent sacrifices a nontoken permanent, put a +1/+1 counter on it.", parser.Context{})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	compilation, diagnostics := Compile(document, Context{})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	for _, reference := range compilation.Abilities[0].Content.References {
		if reference.Binding == ReferenceBindingEventCard || reference.SupportsUse(ReferenceUseCardAction) {
			t.Fatal("ordinary counter mutation was reinterpreted as card entry")
		}
	}
}
