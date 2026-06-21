package parser

import "testing"

// modifyOrGainExact parses a single creature-modifying sentence (a power/toughness
// pump or a keyword grant) and reports whether its sole resolving effect
// round-tripped to an exact, lowerable production.
func modifyOrGainExact(t *testing.T, source string) bool {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 {
		t.Fatalf("Parse(%q) effects = %#v", source, effects)
	}
	return effects[0].Exact
}

func TestExactGroupKeywordGrantAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"Permanents you control gain hexproof and indestructible until end of turn.",
		"Creatures you control gain trample until end of turn.",
		"Creatures you control gain first strike and deathtouch until end of turn.",
		"Other creatures you control gain vigilance until end of turn.",
		"Attacking creatures gain first strike until end of turn.",
		"Blocking creatures gain first strike until end of turn.",
		"All creatures gain haste until end of turn.",
		"Creatures you control gain lifelink, indestructible, and protection from each color until end of turn.",
		"Creatures you control gain protection from red until end of turn.",
		"Permanents you control gain protection from red until end of turn.",
		"Creatures you control gain protection from the color of your choice until end of turn.",
		"Permanents your opponents control gain hexproof until end of turn.",
		"Permanents you control lose hexproof until end of turn.",
		"Permanents your opponents control lose hexproof and indestructible until end of turn.",
		"Creatures your opponents control lose hexproof until end of turn.",
		"Target creature loses flying until end of turn.",
	}
	for _, source := range accepted {
		if !modifyOrGainExact(t, source) {
			t.Errorf("modifyOrGainExact(%q) = false, want true", source)
		}
	}
}

func TestExactGroupKeywordGrantFailsClosed(t *testing.T) {
	t.Parallel()
	// Each carries a qualifier or wording the canonical group keyword grant does
	// not reconstruct byte-exactly, so it must not be marked exact.
	rejected := []string{
		"Target permanents you control gain indestructible until end of turn.",
		"Permanents you control gain hexproof until your next turn.",
		"Permanents you control gain ward {1} until end of turn.",
		"Permanents you control gain hexproof and \"This permanent can't be sacrificed\" until end of turn.",
		"Nonland permanents you control gain indestructible until end of turn.",
		"Creatures you control gain trample.",
		"Creatures you control gain flying until your next turn.",
	}
	for _, source := range rejected {
		if modifyOrGainExact(t, source) {
			t.Errorf("modifyOrGainExact(%q) = true, want false", source)
		}
	}
}

func TestExactVariableXPumpAccepts(t *testing.T) {
	t.Parallel()
	// A power/toughness side written as the spell's variable "X" (with no
	// "where X is" formula) now reconstructs byte-exactly as "+X"/"-X", so these
	// X-cost pumps round-trip to an exact production.
	accepted := []string{
		"Target creature gets +X/+0 until end of turn.",
		"Target creature gets -X/-X until end of turn.",
		"Target creature gets +X/+X until end of turn.",
		"Target creature gets -X/+X until end of turn.",
	}
	for _, source := range accepted {
		if !modifyOrGainExact(t, source) {
			t.Errorf("modifyOrGainExact(%q) = false, want true", source)
		}
	}
}

func TestExactAsymmetricDynamicPumpAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"Target creature gets +X/+X until end of turn, where X is the number of creature cards in your graveyard.",
		"Target creature gets +X/+0 until end of turn, where X is the number of creature cards in your graveyard.",
		"Target creature gets -X/-X until end of turn, where X is the number of cards in your graveyard.",
		"Target creature gets +X/-X until end of turn, where X is the number of cards in your hand.",
	}
	for _, source := range accepted {
		if !modifyOrGainExact(t, source) {
			t.Errorf("modifyOrGainExact(%q) = false, want true", source)
		}
	}
}

func TestExactDistributivePumpAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"Two target creatures each get -1/-1 until end of turn.",
		"Up to two target creatures each get +2/+2 until end of turn.",
		"Up to five target creatures each get -1/-1 until end of turn.",
		"Up to two target creatures you control each get +1/+0 until end of turn.",
	}
	for _, source := range accepted {
		if !modifyOrGainExact(t, source) {
			t.Errorf("modifyOrGainExact(%q) = false, want true", source)
		}
	}
}

func TestExactDistributivePumpFailsClosed(t *testing.T) {
	t.Parallel()
	// "One or two" is a divided-style enumeration the multi-target round-trip
	// does not reconstruct, and "another"/"other" distributive subjects carry a
	// qualifier the canonical wording drops, so none may be marked exact.
	rejected := []string{
		"One or two target creatures each get +2/+1 until end of turn.",
	}
	for _, source := range rejected {
		if modifyOrGainExact(t, source) {
			t.Errorf("modifyOrGainExact(%q) = true, want false", source)
		}
	}
}

// distributiveCombinedBuffEffectsExact parses a combined distributive buff
// sentence ("<N target creatures> each get +P/+T and gain <keyword> until end of
// turn."), which splits into a modify effect followed by a prior-subject keyword
// grant, and reports whether both resolving effects round-tripped to exact,
// lowerable productions.
func distributiveCombinedBuffEffectsExact(t *testing.T, source string) bool {
	t.Helper()
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 2 {
		t.Fatalf("Parse(%q) effects = %#v", source, effects)
	}
	return effects[0].Exact && effects[1].Exact
}

func TestExactDistributiveCombinedBuffAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"Up to two target creatures each get +1/+1 and gain lifelink until end of turn.",
		"Two target creatures you control each get +2/+2 and gain flying until end of turn.",
		"Up to two target creatures each get +1/+0 and gain first strike until end of turn.",
		"Up to two target creatures each get +2/+2 and gain trample until end of turn.",
	}
	for _, source := range accepted {
		if !distributiveCombinedBuffEffectsExact(t, source) {
			t.Errorf("distributiveCombinedBuffEffectsExact(%q) = false, want true", source)
		}
	}
}

func TestExactDistributiveCombinedBuffFailsClosed(t *testing.T) {
	t.Parallel()
	// "One or two" is a divided-style enumeration the multi-target round-trip
	// does not reconstruct, so the modify clause must not be marked exact even
	// when a keyword grant is appended.
	rejected := []string{
		"One or two target creatures each get +2/+1 and gain trample until end of turn.",
	}
	for _, source := range rejected {
		if distributiveCombinedBuffEffectsExact(t, source) {
			t.Errorf("distributiveCombinedBuffEffectsExact(%q) = true, want false", source)
		}
	}
}

// selfPumpExact parses a single self-referential sentence on a named permanent
// (a self-pump such as "<Name> gets +X/+X until end of turn, where X is its
// power.") and reports whether its sole resolving effect round-tripped to an
// exact, lowerable production. The card name lets the parser recognize explicit
// self-name references in the amount formula.
func selfPumpExact(t *testing.T, cardName, source string) bool {
	t.Helper()
	document, diagnostics := Parse(source, Context{CardName: cardName})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 {
		t.Fatalf("Parse(%q) effects = %#v", source, effects)
	}
	return effects[0].Exact
}

// TestExactSourcePowerSelfPumpAccepts covers self-pumps whose amount reads the
// source's own power. The "where X is its power" formula adds a second reference
// (the "its"/"this creature's"/"<name>'s" power referent) alongside the subject
// reference; the subject reconstruction drops that referent so the clause
// round-trips exactly.
func TestExactSourcePowerSelfPumpAccepts(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, source string
	}{
		{"Yew Spirit", "Yew Spirit gets +X/+X until end of turn, where X is its power."},
		{"Feral Animist", "This creature gets +X/+0 until end of turn, where X is its power."},
		{"Brawler", "This creature gets +X/+X until end of turn, where X is this creature's power."},
	}
	for _, test := range cases {
		if !selfPumpExact(t, test.name, test.source) {
			t.Errorf("selfPumpExact(%q, %q) = false, want true", test.name, test.source)
		}
	}
}

// TestExactThoseSubjectKeywordGrant verifies the demonstrative back-reference
// grant "Those creatures gain <keyword> until end of turn." (Inspiring Call's
// second clause) round-trips to an exact, lowerable production. The group is
// named by "those" referring to the preceding count clause.
func TestExactThoseSubjectKeywordGrant(t *testing.T) {
	t.Parallel()
	source := "Draw a card for each creature you control with a +1/+1 counter on it. " +
		"Those creatures gain indestructible until end of turn."
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 2 {
		t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
	}
	grant := document.Abilities[0].Sentences[1].Effects
	if len(grant) != 1 {
		t.Fatalf("Parse(%q) grant effects = %#v", source, grant)
	}
	if grant[0].Kind != EffectGain || !grant[0].Exact {
		t.Errorf("grant effect = %+v, want exact EffectGain", grant[0])
	}
}

// TestExactThoseSubjectKeywordGrantFailsClosed verifies wordings outside the
// canonical "those <group> gain <keyword> until end of turn." envelope are not
// marked exact.
func TestExactThoseSubjectKeywordGrantFailsClosed(t *testing.T) {
	t.Parallel()
	rejected := []string{
		"Draw a card for each creature you control with a +1/+1 counter on it. " +
			"Those creatures gain indestructible.",
		"Draw a card for each creature you control with a +1/+1 counter on it. " +
			"Those creatures gain ward {1} until end of turn.",
	}
	for _, source := range rejected {
		document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
		if len(diagnostics) != 0 {
			continue
		}
		if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 2 {
			continue
		}
		grant := document.Abilities[0].Sentences[1].Effects
		if len(grant) == 1 && grant[0].Exact {
			t.Errorf("Parse(%q) grant marked exact, want not exact", source)
		}
	}
}

// TestExactLeadingDurationMassPumpAccepts verifies that a sentence-leading
// "Until end of turn," distributes the duration so both the keyword grant and
// the dynamic power/toughness pump reconstruct byte-exactly (Overwhelming
// Stampede's printed wording).
func TestExactLeadingDurationMassPumpAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"Until end of turn, creatures you control gain trample and get +X/+X, where X is the greatest power among creatures you control.",
		"Until end of turn, creatures you control gain trample and get +X/+X, where X is the number of creatures you control.",
	}
	for _, source := range accepted {
		document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
		if len(diagnostics) != 0 {
			t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
		}
		effects := document.Abilities[0].Sentences[0].Effects
		if len(effects) != 2 {
			t.Fatalf("Parse(%q) effects = %#v", source, effects)
		}
		for i := range effects {
			if !effects[i].Exact {
				t.Errorf("Parse(%q) effect[%d] not exact", source, i)
			}
		}
	}
}
