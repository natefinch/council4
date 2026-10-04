package compiler

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestCompileResultSelectionIsTextBlind(t *testing.T) {
	for _, test := range []struct {
		noun    string
		kind    SelectorKind
		subtype string
		union   []types.Card
		colors  []color.Color
		colored bool
	}{
		{"land card", SelectorLand, "", nil, nil, false},
		{"Pirate", SelectorUnknown, "Pirate", nil, nil, false},
		{"instant or sorcery card", SelectorCard, "", []types.Card{types.Instant, types.Sorcery}, nil, false},
		{"red card", SelectorCard, "", nil, []color.Color{color.Red}, false},
		{"creature with one or more colors", SelectorCreature, "", nil, nil, true},
	} {
		t.Run(test.noun, func(t *testing.T) {
			lead, participle, outcome := "Exile the top card of your library.", "exiled", EffectExile
			if test.colored {
				lead, participle, outcome = "You may sacrifice a creature.", "sacrificed", EffectSacrifice
			}
			document, diagnostics := parser.Parse(lead+" If a "+test.noun+" was "+participle+" this way, draw a card.", parser.Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("parse = %v", diagnostics)
			}
			for i := range document.Abilities {
				ability := &document.Abilities[i]
				ability.Text = "opaque"
				ability.Tokens = nil
				for ci := range ability.ConditionClauses {
					ability.ConditionClauses[ci].ThisWaySelection.Text = "creature"
				}
				for si := range ability.Sentences {
					ability.Sentences[si].Text = "opaque"
					for ei := range ability.Sentences[si].Effects {
						ability.Sentences[si].Effects[ei].Text = "opaque"
						ability.Sentences[si].Effects[ei].Tokens = nil
					}
				}
			}
			compilation, diagnostics := Compile(document, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("compile = %v", diagnostics)
			}
			condition := compilation.Abilities[0].Content.Conditions[0]
			if condition.ThisWayCardNoun != strings.HasSuffix(test.noun, " card") {
				t.Fatal("compiler lost typed result card/permanent domain")
			}
			selection := condition.ThisWaySelection
			if condition.Predicate != ConditionPredicateResultThisWay || condition.ThisWayOutcome != outcome || selection == nil {
				t.Fatalf("condition = %#v", condition)
			}
			if selection.Kind != test.kind || (test.subtype != "" && !slices.Contains(selection.SubtypesAny(), types.Sub(test.subtype))) ||
				(test.union != nil && !slices.Equal(selection.RequiredTypesAny(), test.union)) ||
				!slices.Equal(selection.ColorsAny(), test.colors) || selection.Colored != test.colored {
				t.Fatalf("typed selection = %#v", selection)
			}
		})
	}
}
