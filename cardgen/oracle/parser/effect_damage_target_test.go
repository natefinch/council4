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
		{"Seismic Shudder", "Seismic Shudder deals 1 damage to each creature without flying."},
		{"Roast", "Roast deals 5 damage to target creature without flying."},
		{"Savage Twister", "Savage Twister deals X damage to each creature."},
		{"Windstorm", "Windstorm deals X damage to each creature with flying."},
		{"Earthquake", "Earthquake deals X damage to each creature without flying and each player."},
		{"Hurricane", "Hurricane deals X damage to each creature with flying and each player."},
	}
	for _, test := range tests {
		if !damageEffectExact(t, test.name, test.source) {
			t.Errorf("damageEffectExact(%q) = false, want true", test.source)
		}
	}
}

func TestExactDamageTargetFailsClosed(t *testing.T) {
	t.Parallel()
	// A recipient that carries both a required and an excluded keyword cannot be
	// reconstructed by the canonical group-damage phrasing, so it stays
	// fail-closed rather than lowering to an approximate filter.
	tests := []struct{ name, source string }{
		{"Antiflyer", "Antiflyer deals 1 damage to each creature with flying without trample."},
		// "where X is ..." group damage is a dynamic amount form the group path
		// does not reconstruct, so it must stay fail-closed rather than lower as
		// a bare X amount that drops the count clause.
		{"Chain Reaction", "Chain Reaction deals X damage to each creature, where X is the number of creatures on the battlefield."},
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
		{"Burn You", "Burn You deals 2 damage to you.", DamageRecipientReferenceYou},
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
	// A possessive recipient that is neither a single controller nor owner (here a
	// repeated controller phrase) must not be read as a referenced-player
	// recipient, so it stays fail-closed rather than lowering to an approximation.
	tests := []struct{ name, source string }{
		{"Burn Color", "Burn Color deals 2 damage to that creature's controller and that creature's controller."},
	}
	for _, test := range tests {
		got, _ := damageRecipientReferenceOf(t, test.name, test.source)
		if got != DamageRecipientReferenceNone {
			t.Errorf("DamageRecipientReference(%q) = %v, want None", test.source, got)
		}
	}
}

func TestExactSourcePowerDamageAccepts(t *testing.T) {
	t.Parallel()
	tests := []struct{ name, source string }{
		{"Justice Strike", "Target creature deals damage to itself equal to its power."},
		{"Rabid Bite", "Target creature you control deals damage equal to its power to target creature you don't control."},
		{"Bite Down", "Target creature you control deals damage equal to its power to target creature or planeswalker you don't control."},
		{"Soul's Fire", "Target creature you control deals damage equal to its power to any target."},
		{"Fall of the Hammer", "Target creature you control deals damage equal to its power to another target creature."},
	}
	for _, test := range tests {
		if !damageEffectExact(t, test.name, test.source) {
			t.Errorf("damageEffectExact(%q) = false, want true", test.source)
		}
	}
}

func TestExactSourcePowerDamageFailsClosed(t *testing.T) {
	t.Parallel()
	// The mass "each creature deals damage to itself" form has no single target
	// source, so it stays fail-closed rather than lowering to an approximation.
	tests := []struct{ name, source string }{
		{"Wave of Reckoning", "Each creature deals damage to itself equal to its power."},
	}
	for _, test := range tests {
		if damageEffectExact(t, test.name, test.source) {
			t.Errorf("damageEffectExact(%q) = true, want false", test.source)
		}
	}
}

func TestExactEachOfTargetsDamageAccepts(t *testing.T) {
	t.Parallel()
	tests := []struct{ name, source string }{
		{"Furious Reprisal", "Furious Reprisal deals 2 damage to each of two targets."},
		{"Jagged Lightning", "Jagged Lightning deals 3 damage to each of two target creatures."},
		{"Dual Shot", "Dual Shot deals 1 damage to each of up to two target creatures."},
	}
	for _, test := range tests {
		if !damageEffectExact(t, test.name, test.source) {
			t.Errorf("damageEffectExact(%q) = false, want true", test.source)
		}
	}
}

// selfDamageRiderOf parses a single self-name damage sentence and returns its
// resolving effect's self-damage rider fields together with its exactness, so a
// "and N damage to you" rider can be asserted without inspecting the full effect.
func selfDamageRiderOf(t *testing.T, name, source string) (hasRider bool, riderValue int, exact bool) {
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
	return effects[0].HasSelfDamageRider, effects[0].SelfDamageRiderValue, effects[0].Exact
}

func TestSelfDamageRiderAccepts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, source string
		wantValue    int
	}{
		{"Char", "Char deals 4 damage to any target and 2 damage to you.", 2},
		{"Psionic Blast", "Psionic Blast deals 4 damage to any target and 2 damage to you.", 2},
		{"Forge Devil", "Forge Devil deals 1 damage to target creature and 1 damage to you.", 1},
	}
	for _, test := range tests {
		has, value, exact := selfDamageRiderOf(t, test.name, test.source)
		if !has {
			t.Errorf("HasSelfDamageRider(%q) = false, want true", test.source)
		}
		if value != test.wantValue {
			t.Errorf("SelfDamageRiderValue(%q) = %d, want %d", test.source, value, test.wantValue)
		}
		if !exact {
			t.Errorf("damageEffectExact(%q) = false, want true", test.source)
		}
	}
}

func TestSelfDamageRiderFailsClosed(t *testing.T) {
	t.Parallel()
	// A second recipient that is not the controller ("its controller") is not a
	// self-damage rider at all, so the rider field stays unset.
	has, _, _ := selfDamageRiderOf(t, "Spit Flame",
		"Spit Flame deals 4 damage to target creature and 2 damage to its controller.")
	if has {
		t.Error("HasSelfDamageRider(its controller) = true, want false")
	}
	// A variable primary amount paired with a rider is outside the bounded exact
	// form, so even though the rider is recognized the effect stays non-exact
	// rather than lowering to an approximation.
	_, _, exact := selfDamageRiderOf(t, "Variable Char",
		"Variable Char deals X damage to any target and 2 damage to you.")
	if exact {
		t.Error("damageEffectExact(variable primary with rider) = true, want false")
	}
}

// TestExactDamageShortNameSubjectAccepts covers legendary cards whose Oracle
// text refers to the permanent by the short name preceding the comma in the
// full card name (CR 201.3 lets the pre-comma portion stand in for the whole
// name). The effect subject reference must resolve to the self permanent so the
// damage clause round-trips exact, even though the subject word differs from the
// full card name.
func TestExactDamageShortNameSubjectAccepts(t *testing.T) {
	t.Parallel()
	tests := []struct{ name, source string }{
		{"Kamahl, Pit Fighter", "Kamahl deals 3 damage to any target."},
		{"Jeska, Warrior Adept", "Jeska deals 1 damage to any target."},
	}
	for _, test := range tests {
		if !damageEffectExact(t, test.name, test.source) {
			t.Errorf("damageEffectExact(%q, %q) = false, want true", test.name, test.source)
		}
	}
}

// TestExactDamageNameKeywordCollisionAccepts covers cards whose name ends in a
// word that is also a keyword ability ("Storm"). The keyword scanner must treat
// that word as part of the self name rather than a granted keyword, so the
// single fixed-amount damage effect carries no spurious keyword and lowers.
func TestExactDamageNameKeywordCollisionAccepts(t *testing.T) {
	t.Parallel()
	tests := []struct{ name, source string }{
		{"Command the Storm", "Command the Storm deals 5 damage to target creature."},
		{"Cinder Storm", "Cinder Storm deals 5 damage to any target."},
	}
	for _, test := range tests {
		if !damageEffectExact(t, test.name, test.source) {
			t.Errorf("damageEffectExact(%q, %q) = false, want true", test.name, test.source)
		}
	}
}

// TestExactDamageShortNameFailsClosed guards against over-broadening the
// short-name aliasing: only the pre-comma legend name (and DFC front name) may
// alias the self permanent, never the bare first word of an ordinary multi-word
// name. A subject that is merely the first word of a non-legendary name must not
// resolve to the self permanent, so the clause stays fail-closed.
func TestExactDamageShortNameFailsClosed(t *testing.T) {
	t.Parallel()
	tests := []struct{ name, source string }{
		{"Fire Whip", "Fire deals 2 damage to any target."},
		{"Goblin Welder", "Goblin deals 2 damage to any target."},
	}
	for _, test := range tests {
		if damageEffectExact(t, test.name, test.source) {
			t.Errorf("damageEffectExact(%q, %q) = true, want false", test.name, test.source)
		}
	}
}
