package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestRecurringSelfCardOwnsGraveyardFunctionZone(t *testing.T) {
	t.Parallel()
	for _, subject := range []string{"Dormant Source", "this card"} {
		t.Run(subject, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := parser.Parse("At the beginning of your upkeep, you may return "+subject+" from your graveyard to your hand.", parser.Context{CardName: "Dormant Source"})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			compilation, diagnostics := Compile(document, Context{})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			content := compilation.Abilities[0].Content
			got, ok := content.TriggerFunctionZone()
			if !ok || got != zone.Graveyard {
				t.Fatalf("zone=%v ok=%v", got, ok)
			}
		})
	}
}

func TestTriggerFunctionZoneRefusesUnownedProof(t *testing.T) {
	compile := func() AbilityContent {
		t.Helper()
		document, diagnostics := parser.Parse("At the beginning of your upkeep, return this card from your graveyard to your hand.", parser.Context{})
		if len(diagnostics) != 0 {
			t.Fatal(diagnostics)
		}
		compilation, diagnostics := Compile(document, Context{})
		if len(diagnostics) != 0 {
			t.Fatal(diagnostics)
		}
		return compilation.Abilities[0].Content
	}
	for _, tc := range []struct {
		name   string
		mutate func(*AbilityContent)
	}{
		{"missing source", func(c *AbilityContent) { c.Source = ReferenceSourceContext{} }},
		{"foreign source", func(c *AbilityContent) { c.Source = compile().Source }},
		{"missing subject", func(c *AbilityContent) { c.References[0].Subject = ReferenceSubjectProof{} }},
		{"foreign subject", func(c *AbilityContent) { c.References[0] = compile().References[0] }},
		{"incompatible event", func(c *AbilityContent) { c.Source.event = TriggerEventPermanentDied }},
		{"unavailable zone", func(c *AbilityContent) { c.Source.zone = zone.Exile }},
		{"wrong reference identity", func(c *AbilityContent) { c.References[0].NodeID++ }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			content := compile()
			tc.mutate(&content)
			if got, ok := content.TriggerFunctionZone(); ok || got != zone.None {
				t.Fatalf("unowned function-zone proof admitted: zone=%v ok=%v", got, ok)
			}
		})
	}
}
