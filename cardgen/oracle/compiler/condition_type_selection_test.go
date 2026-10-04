package compiler

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestCompileConditionTypeSelectionsTextBlind(t *testing.T) {
	t.Parallel()
	syntax := parser.ConditionSelection{
		RequiredTypes:    []parser.TriggerCardType{parser.TriggerCardTypeArtifact},
		RequiredTypesAny: []parser.TriggerCardType{parser.TriggerCardTypeInstant, parser.TriggerCardTypeSorcery},
		ExcludedTypes:    []parser.TriggerCardType{parser.TriggerCardTypeCreature},
		AnyOf:            []parser.ConditionSelection{{RequiredTypes: []parser.TriggerCardType{parser.TriggerCardTypeLand}}},
	}
	got, ok := compileConditionSelection(syntax)
	if !ok || !slices.Equal(got.RequiredTypes, []types.Card{types.Artifact}) ||
		!slices.Equal(got.RequiredTypesAny, []types.Card{types.Instant, types.Sorcery}) ||
		!slices.Equal(got.ExcludedTypes, []types.Card{types.Creature}) ||
		len(got.AnyOf) != 1 || !slices.Equal(got.AnyOf[0].RequiredTypes, []types.Card{types.Land}) {
		t.Fatalf("compileConditionSelection = %#v, %v", got, ok)
	}
}

func TestCompileConditionTypeSelectionsInvalidValues(t *testing.T) {
	t.Parallel()
	for _, invalid := range []parser.TriggerCardType{parser.TriggerCardTypeUnknown, "unexpected"} {
		for _, syntax := range []parser.ConditionSelection{
			{RequiredTypes: []parser.TriggerCardType{invalid}},
			{RequiredTypesAny: []parser.TriggerCardType{invalid}},
			{ExcludedTypes: []parser.TriggerCardType{invalid}},
			{AnyOf: []parser.ConditionSelection{{RequiredTypesAny: []parser.TriggerCardType{invalid}}}},
		} {
			if got, ok := compileConditionSelection(syntax); ok {
				t.Fatalf("accepted invalid syntax %#v: %#v", syntax, got)
			}
		}
	}
}
