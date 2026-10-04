package parser

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

func TestResultThisWayRetainsSelection(t *testing.T) {
	for _, test := range []struct {
		noun, outcome string
		kind          SelectionKind
		subtype       string
		union         int
		colors        []Color
		colored       bool
	}{
		{"land card", "discarded", SelectionLand, "", 0, nil, false},
		{"creature", "destroyed", SelectionCreature, "", 0, nil, false},
		{"Goblin", "sacrificed", SelectionUnknown, "Goblin", 0, nil, false},
		{"Pirate", "exiled", SelectionUnknown, "Pirate", 0, nil, false},
		{"Lesson card", "milled", SelectionCard, "Lesson", 0, nil, false},
		{"instant or sorcery card", "exiled", SelectionCard, "", 2, nil, false},
		{"red card", "discarded", SelectionCard, "", 0, []Color{ColorRed}, false},
		{"creature with one or more colors", "sacrificed", SelectionCreature, "", 0, nil, true},
	} {
		t.Run(test.noun, func(t *testing.T) {
			document, diagnostics := Parse("Draw a card. If a "+test.noun+" was "+test.outcome+" this way, draw a card.", Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %v", diagnostics)
			}
			condition := document.Abilities[0].ConditionClauses[0]
			if condition.ThisWayCardNoun != strings.HasSuffix(test.noun, " card") {
				t.Fatal("result noun lost its card/permanent domain")
			}
			selection := condition.ThisWaySelection
			if selection == nil {
				t.Fatal("result noun selection was discarded")
			}
			if selection.Kind != test.kind || (test.subtype != "" && !slices.Contains(selection.SubtypesAny, types.Sub(test.subtype))) ||
				(test.union != 0 && len(selection.RequiredTypesAny) != test.union) ||
				!slices.Equal(selection.ColorsAny, test.colors) || selection.Colored != test.colored {
				t.Fatalf("selection = %#v", selection)
			}
		})
	}
}
