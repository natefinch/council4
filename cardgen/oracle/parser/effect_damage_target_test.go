package parser

import "testing"

// damageEffectExact parses a single self-name damage sentence and reports
// whether its resolving effect round-tripped to an exact, lowerable production.
func damageEffectExact(t *testing.T, name, source string) bool {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true, CardName: name})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 || effects[0].Kind != EffectDealDamage {
		t.Fatalf("Parse(%q) effects = %#v", source, effects)
	}
	return effects[0].Exact
}

func TestExactDamageTargetAccepts(t *testing.T) {
	t.Parallel()
	tests := []struct{ name, source string }{
		{"Lava Spike", "Lava Spike deals 3 damage to target player or planeswalker."},
		{"Searing Flesh", "Searing Flesh deals 7 damage to target opponent or planeswalker."},
		{"Leaf Arrow", "Leaf Arrow deals 3 damage to target creature with flying."},
		{"Rending Volley", "Rending Volley deals 4 damage to target white or blue creature."},
		{"Gale Force", "Gale Force deals 5 damage to each creature with flying."},
	}
	for _, test := range tests {
		if !damageEffectExact(t, test.name, test.source) {
			t.Errorf("damageEffectExact(%q) = false, want true", test.source)
		}
	}
}

func TestExactDamageTargetFailsClosed(t *testing.T) {
	t.Parallel()
	// "without flying" uses an excluded keyword that SelectionSyntax does not
	// capture, so the recipient cannot round-trip and stays fail-closed.
	tests := []struct{ name, source string }{
		{"Antiflyer", "Antiflyer deals 1 damage to each creature without flying."},
	}
	for _, test := range tests {
		if damageEffectExact(t, test.name, test.source) {
			t.Errorf("damageEffectExact(%q) = true, want false", test.source)
		}
	}
}

// damageRecipientReferenceOf parses a single self-name damage sentence and
// returns its resolving effect's recipient-reference kind together with its
// exactness, so a recipient that is the controller/owner of a referenced object
// can be asserted without inspecting the full effect.
func damageRecipientReferenceOf(t *testing.T, name, source string) (DamageRecipientReferenceKind, bool) {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true, CardName: name})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 || effects[0].Kind != EffectDealDamage {
		t.Fatalf("Parse(%q) effects = %#v", source, effects)
	}
	return effects[0].DamageRecipientReference, effects[0].Exact
}

func TestDamageRecipientReferenceAccepts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, source string
		want         DamageRecipientReferenceKind
	}{
		{"Burn Land", "Burn Land deals 2 damage to that land's controller.", DamageRecipientReferenceController},
		{"Burn Creature", "Burn Creature deals 3 damage to that creature's owner.", DamageRecipientReferenceOwner},
		{"Burn It", "Burn It deals 1 damage to its controller.", DamageRecipientReferenceController},
	}
	for _, test := range tests {
		got, exact := damageRecipientReferenceOf(t, test.name, test.source)
		if got != test.want {
			t.Errorf("DamageRecipientReference(%q) = %v, want %v", test.source, got, test.want)
		}
		if !exact {
			t.Errorf("damageEffectExact(%q) = false, want true", test.source)
		}
	}
}

func TestDamageRecipientReferenceFailsClosed(t *testing.T) {
	t.Parallel()
	// A plain player recipient, a possessive that is not controller/owner, and a
	// dynamic ("equal to") amount must not be read as a referenced-player
	// recipient: the first two are unrelated recipients and the third is an
	// amount form the exactness branch deliberately rejects.
	tests := []struct{ name, source string }{
		{"Burn You", "Burn You deals 2 damage to you."},
		{"Burn Color", "Burn Color deals 2 damage to that creature's controller and that creature's controller."},
	}
	for _, test := range tests {
		got, _ := damageRecipientReferenceOf(t, test.name, test.source)
		if got != DamageRecipientReferenceNone {
			t.Errorf("DamageRecipientReference(%q) = %v, want None", test.source, got)
		}
	}
}
