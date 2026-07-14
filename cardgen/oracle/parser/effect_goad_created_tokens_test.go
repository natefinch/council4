package parser

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// TestCreditGoadCreatedTokensRider proves the trailing "The tokens are goaded for
// the rest of the game." sentence is folded onto the preceding create-token
// effect (Life of the Party). The create effect records the rider span, the rider
// sentence is emptied of effects and marked so coverage credits it, and the
// group-recipient copy create is recognized as a copy-of-reference.
func TestCreditGoadCreatedTokensRider(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"When this creature enters, if it's not a token, each opponent creates a token that's a copy of it. The tokens are goaded for the rest of the game.",
		Context{CardName: "Life of the Party"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %d, want 1", len(document.Abilities))
	}
	ability := document.Abilities[0]

	var create *EffectSyntax
	riderSentences := 0
	for i := range ability.Sentences {
		if ability.Sentences[i].GoadCreatedTokensRider {
			riderSentences++
			if len(ability.Sentences[i].Effects) != 0 {
				t.Fatalf("rider sentence still carries %d effects", len(ability.Sentences[i].Effects))
			}
		}
		for j := range ability.Sentences[i].Effects {
			if ability.Sentences[i].Effects[j].Kind == EffectCreate {
				create = &ability.Sentences[i].Effects[j]
			}
		}
	}
	if riderSentences != 1 {
		t.Fatalf("goad-created-tokens rider sentences = %d, want 1", riderSentences)
	}
	if create == nil {
		t.Fatal("no create-token effect found")
	}
	if create.GoadCreatedTokensRiderSpan == (shared.Span{}) {
		t.Fatal("create effect did not record the goad-created-tokens rider span")
	}
	if !create.TokenCopyOfReference {
		t.Fatal("group-recipient copy create was not recognized as copy-of-reference")
	}
	if create.Context != EffectContextEachOpponent {
		t.Fatalf("create context = %v, want EffectContextEachOpponent", create.Context)
	}
}
