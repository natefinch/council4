package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestOptionalObservationCharacteristicGuardExactOwnership(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*compiler.CompiledEffect, *compiler.CompiledEffect)
		refuse bool
	}{
		{"power", func(*compiler.CompiledEffect, *compiler.CompiledEffect) {}, false},
		{"toughness", func(e *compiler.CompiledEffect, _ *compiler.CompiledEffect) {
			e.Amount.DynamicKind = compiler.DynamicAmountSourceToughness
		}, false},
		{"mana value remains unmodeled", func(e *compiler.CompiledEffect, _ *compiler.CompiledEffect) {
			e.Amount.DynamicKind = compiler.DynamicAmountSourceManaValue
		}, true},
		{"wrong amount node", func(e *compiler.CompiledEffect, _ *compiler.CompiledEffect) {
			e.Amount.ReferenceNodeID++
		}, true},
		{"duplicate amount node", func(e *compiler.CompiledEffect, _ *compiler.CompiledEffect) {
			e.References = append(e.References, e.References[0])
		}, true},
		{"wrong producer clause", func(e *compiler.CompiledEffect, _ *compiler.CompiledEffect) {
			e.References[0].ProducerClauseID++
		}, true},
		{"source rather than observation", func(e *compiler.CompiledEffect, _ *compiler.CompiledEffect) {
			e.References[0].Binding = compiler.ReferenceBindingSource
		}, true},
		{"missing reference", func(e *compiler.CompiledEffect, _ *compiler.CompiledEffect) {
			e.References = nil
		}, true},
		{"plural producer", func(_ *compiler.CompiledEffect, p *compiler.CompiledEffect) {
			p.Amount.Value = 2
		}, true},
		{"modified amount", func(e *compiler.CompiledEffect, _ *compiler.CompiledEffect) {
			e.Amount.Addend = 1
		}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			producer := compiler.CompiledEffect{
				Kind: compiler.EffectReveal, Exact: true, Optional: true, ClauseID: 3,
				Context: parser.EffectContextController, CardSource: parser.EffectCardSourceTopOfPlayerLibrary,
				Amount: compiler.CompiledAmount{Known: true, Value: 1},
			}

			effect := compiler.CompiledEffect{
				Exact: true, ClauseID: 7,
				Amount: compiler.CompiledAmount{DynamicKind: compiler.DynamicAmountSourcePower, ReferenceNodeID: 19, Multiplier: 1},
				References: []compiler.CompiledReference{{NodeID: 19, ProducerClauseID: 3,
					Binding: compiler.ReferenceBindingPriorInstructionResult, PriorInstruction: 0}},
			}
			tc.mutate(&effect, &producer)
			content := compiler.AbilityContent{Effects: []compiler.CompiledEffect{producer, effect}}
			if refused := optionalAntecedentUnmodeled(content, 1); refused != tc.refuse {
				t.Errorf("unmodeled=%v, want %v for exact typed amount node/producer (opaque zero-span data)", refused, tc.refuse)
			}
		})
	}
}

func TestOptionalObservationCharacteristicRefusals(t *testing.T) {
	for _, text := range []string{
		"You may reveal the top card of your library. You gain life equal to its mana value.",
		"You may reveal the top two cards of your library. You gain life equal to their power.",
		"You may reveal the top card of your library. You gain life equal to the looked-at card's power.",
		"You may look at the top card of your library. You gain life equal to the revealed card's toughness.",
		"You may create a 2/2 green Bear creature token. You gain life equal to its power.",
	} {
		t.Run(text, func(t *testing.T) {
			assertCardUnsupported(t, &ScryfallCard{
				Name: "Unmodeled Optional Characteristic", Layout: "normal", TypeLine: "Sorcery", OracleText: text,
			})
		})
	}
}

func TestAdjacentOptionalObservationCharacteristicCardPaths(t *testing.T) {
	for _, tc := range []struct{ verb, property, primitive string }{
		{"reveal", "toughness", "Reveal"},
		{"look at", "power", "LookAtLibraryTop"},
	} {
		for _, intervening := range []string{"", "Scry 1. "} {
			t.Run(tc.verb+"/"+intervening, func(t *testing.T) {
				card := &ScryfallCard{
					Name: "Optional Characteristic Card", Layout: "normal", TypeLine: "Sorcery",
					OracleText: "You may " + tc.verb + " the top card of your library. " + intervening +
						"If it's a creature card, you gain life equal to its " + tc.property + ".",
				}
				key, field := "sequence-effect-0-"+tc.property, "Power"
				if tc.property == "toughness" {
					field = "Toughness"
				}
				assertCardPaths(t, card,
					`Sequence[0].Primitive.(game.`+tc.primitive+`).PublishCharacteristics.`+field+` = "`+key+`"`,
					`Primitive.(game.GainLife)`,
					`ResultGate.Val.Key = "`+key+`"`,
					`ResultGate.Val.AmountAvailable = true`,
				)
			})
		}
	}
}
