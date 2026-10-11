package parser

import (
	"strings"
	"testing"
)

func TestParseTargetCountParameters(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		text   string
		exact  bool
		countX bool
		maxX   bool
	}{
		{"Destroy X target artifacts.", true, true, false},
		{"Destroy X target nonblack creatures.", true, true, false},
		{"Destroy X target snow lands.", true, true, false},
		{"Tap X target artifacts, creatures, and/or lands.", true, true, false},
		{"Tap X target creatures you don't control.", true, true, false},
		{"Destroy up to X target artifacts and/or enchantments.", true, false, true},
		{"Goad up to X target creatures your opponents control.", true, false, true},
		{"Exile X target cards from a single graveyard.", true, true, false},
		{"Exile up to X target cards from a single graveyard.", true, false, true},
		{"Return X target creature cards from your graveyard to your hand.", true, true, false},
		{"Return X target creature cards with mana value 2 or less from your graveyard to the battlefield.", true, true, false},
		{"Return X target nonland permanents to their owners' hands.", true, true, false},
		{"Destroy any number of target artifacts and/or enchantments.", true, false, false},
		{"Destroy X target artifacts that entered this turn.", false, true, false},
		{"Destroy twice X target artifacts.", false, true, false},
		{"Destroy X target artifacts. Draw a card.", true, true, false},
	} {
		t.Run(tc.text, func(t *testing.T) {
			t.Parallel()
			doc, diagnostics := Parse(tc.text, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("parse diagnostics: %v", diagnostics)
			}
			effect := doc.Abilities[0].Sentences[0].Effects[0]
			if len(effect.Targets) != 1 {
				t.Fatalf("targets: %+v", effect.Targets)
			}
			target := effect.Targets[0]
			if target.Cardinality.CountEqualsX != tc.countX || target.Cardinality.MaxX != tc.maxX {
				t.Fatalf("count: %+v", target.Cardinality)
			}
			if effect.Exact != tc.exact {
				t.Fatalf("effect exact=%v target=%+v", effect.Exact, target)
			}
			if effect.ClauseID == 0 || target.Span.Start.Offset < effect.ClauseSpan.Start.Offset ||
				target.Span.End.Offset > effect.ClauseSpan.End.Offset {
				t.Fatalf("target count escaped its owning clause: %+v", target)
			}
			if tc.countX || tc.maxX {
				if !strings.Contains(target.Text, "X target") {
					t.Fatalf("count missing from consumed target: %q", target.Text)
				}
			}
		})
	}
}
