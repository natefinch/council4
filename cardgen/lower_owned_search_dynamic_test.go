package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

const ownedDynamicLandSearch = "Search your library for up to X basic land cards, where X is the number of lands you control, put them onto the battlefield tapped, then shuffle."

func TestOwnedSearchRulesDerivedQuantity(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		ownedDynamicLandSearch,
		"You gain 1 life. " + ownedDynamicLandSearch,
		ownedDynamicLandSearch + " You gain 1 life.",
		"Put a +1/+1 counter on target creature. " + ownedDynamicLandSearch + " Scry 1.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			defs, diagnostics, err := CompileCardDefs(&ScryfallCard{
				Name: "Rules Derived Search", Layout: "normal", TypeLine: "Sorcery", OracleText: text,
			})
			if err != nil || len(diagnostics) != 0 || len(defs) != 1 {
				t.Fatalf("compile: %v %v, definitions=%d", err, diagnostics, len(defs))
			}
			count := 0
			for _, instruction := range defs[0].SpellAbility.Val.Modes[0].Sequence {
				if search, ok := instruction.Primitive.(game.Search); ok {
					count++
					dynamic := search.Amount.DynamicAmount()
					content := compiledSearchUnitContent(t, ownedDynamicLandSearch)
					group, ok := searchGroupSpec(content.Effects)
					if !ok || !dynamic.Exists || dynamic.Val.Kind != game.DynamicAmountCountSelector ||
						!reflect.DeepEqual(search.Amount, group.Quantity) {
						t.Fatal("the existing rules-derived quantity was replaced or reinterpreted")
					}
				}
			}
			if count != 1 {
				t.Fatal("missing single owned Search instruction")
			}
		})
	}
}

func TestOwnedSearchDynamicCannotPublishSingularArrival(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		ownedDynamicLandSearch + " Untap that land.",
		"You gain 1 life. " + ownedDynamicLandSearch + " Untap that land.",
		"You gain 1 life. Search your library for X basic land cards, put them onto the battlefield, then shuffle.",
		"You gain 1 life. Search your library for zero basic land cards, put them onto the battlefield, then shuffle.",
		"You gain 1 life. Search your library for up to X basic land cards, where X is the number of imaginary lands, put them onto the battlefield, then shuffle.",
	} {
		assertCardUnsupported(t, &ScryfallCard{Name: "Dynamic Search Refusal", Layout: "normal", TypeLine: "Sorcery", OracleText: text})
	}
}
