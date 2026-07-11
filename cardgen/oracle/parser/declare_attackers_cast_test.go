package parser

import "testing"

func TestParseDeclareAttackersCastRestriction(t *testing.T) {
	t.Parallel()
	const exact = "Cast this spell only during the declare attackers step and only if you've been attacked this step."
	document, diagnostics := Parse(exact, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 || len(document.Abilities) != 1 {
		t.Fatalf("parse = %#v, diagnostics = %#v", document, diagnostics)
	}
	if !document.Abilities[0].CastOnlyDuringDeclareAttackersAfterAttacked {
		t.Fatal("exact defensive cast restriction was not recognized")
	}
	near, _ := Parse(
		"Cast this spell only during the declare attackers step and only if you've been attacked this turn.",
		Context{InstantOrSorcery: true},
	)
	if near.Abilities[0].CastOnlyDuringDeclareAttackersAfterAttacked {
		t.Fatal("inexact defensive cast restriction was recognized")
	}
}
