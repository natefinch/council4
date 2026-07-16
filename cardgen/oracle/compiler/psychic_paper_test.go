package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestCompilePsychicPaperUsesTypedSyntaxOnly(t *testing.T) {
	document, diagnostics := parser.Parse(
		"As this Equipment becomes attached to a creature, choose a creature card name and a creature type.\n"+
			"Equipped creature has ward {1}, it can't be blocked, and its name and creature type are the last chosen name and creature type.",
		parser.Context{CardName: "Psychic Paper"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("parser diagnostics = %#v", diagnostics)
	}
	for i := range document.Abilities {
		document.Abilities[i].Text = "compiler must not inspect this text"
	}
	compilation, diagnostics := Compile(document, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("compiler diagnostics = %#v", diagnostics)
	}
	if choices := compilation.Abilities[0].AttachmentChoices; choices == nil ||
		choices.CardNameType != types.Creature ||
		choices.SubtypeOfType != types.Creature {
		t.Fatalf("compiled attachment choices = %#v", choices)
	}
	static := compilation.Abilities[1].Static
	if static == nil || len(static.Declarations) != 4 {
		t.Fatalf("compiled static declarations = %#v", static)
	}
	want := []StaticContinuousOperation{
		StaticContinuousGrantKeywords,
		StaticContinuousUnknown,
		StaticContinuousSetNameFromAttachmentChoice,
		StaticContinuousSetSubtypeFromAttachmentChoice,
	}
	for i, declaration := range static.Declarations {
		if declaration.Group.Domain != StaticGroupAttachedObject {
			t.Fatalf("declaration %d group = %#v", i, declaration.Group)
		}
		if declaration.Continuous != nil && declaration.Continuous.Operation != want[i] {
			t.Fatalf("declaration %d operation = %v, want %v", i, declaration.Continuous.Operation, want[i])
		}
	}
	if static.Declarations[1].Rule == nil || static.Declarations[1].Rule.Kind != StaticRuleCantBeBlocked {
		t.Fatalf("unblockable declaration = %#v", static.Declarations[1])
	}
}
