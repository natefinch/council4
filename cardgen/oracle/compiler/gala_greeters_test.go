package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestCompileModesUniquePerTurnIsTextBlind(t *testing.T) {
	t.Parallel()

	document := parser.Document{Abilities: []parser.Ability{{
		Kind: parser.AbilityTriggered,
		Text: "unrelated metadata",
		Modal: &parser.Modal{
			MinModes:           1,
			MaxModes:           1,
			ChoiceKnown:        true,
			ModesUniquePerTurn: true,
			Options: []parser.Mode{
				{Text: "first"},
				{Text: "second"},
			},
		},
	}}}
	compilation, _ := Compile(document, Context{})
	modal := compilation.Abilities[0].Content.Modes[0].Modal
	if modal == nil || !modal.ModesUniquePerTurn ||
		modal.MinModes != 1 || modal.MaxModes != 1 {
		t.Fatalf("compiled modal = %#v, want one mode unique per turn", modal)
	}
}
