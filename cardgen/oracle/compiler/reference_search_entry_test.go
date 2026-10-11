package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestSearchUnitSubjectProof(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		text  string
		valid bool
	}{
		{"Search your library for a basic land card, put it onto the battlefield tapped, then shuffle. You gain 2 life. Untap that land.", true},
		{"Search your library for a basic land card, put it onto the battlefield tapped, then shuffle. You gain 2 life. Untap that creature.", false},
		{"Search your library for two basic land cards, put them onto the battlefield tapped, then shuffle. You gain 2 life. Untap that land.", false},
		{"Search your library for a basic land card, put it into your hand, then shuffle. You gain 2 life. Untap that land.", false},
	} {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()
			doc, diagnostics := parser.Parse(tt.text, parser.Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			comp, diagnostics := Compile(doc, Context{})
			if len(diagnostics) != 0 || len(comp.Abilities) != 1 {
				t.Fatal(diagnostics)
			}
			content := comp.Abilities[0].Content
			reference := content.References[len(content.References)-1]
			if content.OwnsSubject(reference) != tt.valid {
				t.Fatalf("subject supported=%v, want %v", reference.SubjectSupported(), tt.valid)
			}
			if tt.valid && (!reference.SubjectProducerMatches(content.Effects[0]) ||
				reference.SubjectDomain() != ReferenceSubjectPermanent ||
				reference.SubjectLifetime() != ReferenceLifetimeActualProduct) {
				t.Fatal("arrived permanent did not retain the exact Search producer")
			}
			if reference.ProducerClauseID != content.Effects[0].ClauseID {
				t.Fatal("unsupported arrived subject fell back to a source, event, or later target")
			}
		})
	}
}
