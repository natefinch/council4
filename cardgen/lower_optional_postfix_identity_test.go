package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

func TestOptionalPostfixPlannerExactTypedScope(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*compiler.AbilityContent)
		refuse bool
	}{
		{"opaque zero metadata", func(*compiler.AbilityContent) {}, false},
		{"reindexed clause and condition identities", func(c *compiler.AbilityContent) {
			for i := range c.Effects {
				e := &c.Effects[i]
				e.ClauseID += 10
				for j := range e.OptionalActionClauseIDs {
					e.OptionalActionClauseIDs[j] += 10
				}
				if e.OptionalPostfixElse.ActionClauseID != 0 {
					e.OptionalPostfixElse.ActionClauseID += 10
					e.OptionalPostfixElse.ConditionNodeID += 20
				}
			}
			for i := range c.Conditions {
				c.Conditions[i].NodeID += 20
				for j := range c.Conditions[i].Ownership.ClauseIDs {
					c.Conditions[i].Ownership.ClauseIDs[j] += 10
				}
			}
		}, false},
		{"missing action", func(c *compiler.AbilityContent) { c.Effects[2].OptionalPostfixElse.ActionClauseID = 10000 }, true},
		{"missing predicate", func(c *compiler.AbilityContent) { c.Effects[2].OptionalPostfixElse.ConditionNodeID = 10000 }, true},
		{"duplicate predicate identity", func(c *compiler.AbilityContent) { c.Conditions = append(c.Conditions, c.Conditions[0]) }, true},
		{"wrong predicate owner", func(c *compiler.AbilityContent) { c.Conditions[0].Ownership.ClauseIDs = []int{c.Effects[0].ClauseID} }, true},
		{"compound action", func(c *compiler.AbilityContent) {
			c.Effects[1].OptionalActionClauseIDs = append(c.Effects[1].OptionalActionClauseIDs, c.Effects[2].ClauseID)
		}, true},
		{"optional fallback", func(c *compiler.AbilityContent) { c.Effects[2].Optional = true }, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			document, diagnostics := parser.Parse(
				"Reveal the top card of your library. You may gain 2 life if it's a permanent card with mana value 3 or less. Otherwise, draw a card.",
				parser.Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			compilation, diagnostics := compiler.Compile(document, compiler.Context{})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			content := compilation.Abilities[0].Content
			for i := range content.Effects {
				e := &content.Effects[i]
				e.Text, e.Span, e.ClauseSpan, e.VerbSpan = "opaque", shared.Span{}, shared.Span{}, shared.Span{}
				e.Order, e.VerbOrder = shared.SourceOrder{}, shared.SourceOrder{}
			}
			for i := range content.Conditions {
				c := &content.Conditions[i]
				c.Text, c.Span, c.SubjectSpan = "opaque", shared.Span{}, shared.Span{}
			}
			tc.mutate(&content)
			plan, ok := planOptionalFlow(content)
			if ok == tc.refuse {
				t.Fatalf("planned=%v want refusal=%v reason=%s", ok, tc.refuse, plan.failureCategory)
			}
			if ok && (plan.scoped.postfixElse[2] != 1 || plan.scoped.publishers[1] == "") {
				t.Fatal("typed postfix scope was reconstructed from indices/text instead of identities")
			}
		})
	}
}
