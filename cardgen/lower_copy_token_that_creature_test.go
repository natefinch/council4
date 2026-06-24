package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceCopyTokenThatTriggerCreature covers a
// single-permanent entry trigger whose body copies the triggering creature
// ("Whenever a nontoken Zombie you control enters, create a token that's a copy
// of that creature." — Necroduality). The "that creature" copy source must bind
// to the triggering permanent rather than the just-created token.
func TestGenerateExecutableCardSourceCopyTokenThatTriggerCreature(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Necroduality",
		Layout:     "normal",
		ManaCost:   "{1}{B}{B}",
		TypeLine:   "Enchantment",
		OracleText: "Whenever a nontoken Zombie you control enters, create a token that's a copy of that creature.",
		Colors:     []string{"B"},
	}, "n")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.CreateToken{",
		"Source: game.TokenCopyOf(game.TokenCopySpec{",
		"Source: game.TokenCopySourceObject,",
		"Object: game.EventPermanentReference(),",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutableCardSourceCopyTokenForEachThatCreature covers a per-each
// copy over a controlled battlefield group whose per-iteration pronoun is "that
// creature" ("For each nontoken creature you control, create a token that's a
// copy of that creature, except it isn't legendary." — Multiversal Incursion).
func TestGenerateExecutableCardSourceCopyTokenForEachThatCreature(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Multiversal Incursion",
		Layout:   "normal",
		ManaCost: "{4}{U}{U}",
		TypeLine: "Sorcery",
		OracleText: "For each nontoken creature you control, create a token that's a copy of " +
			"that creature, except it isn't legendary.",
		Colors: []string{"U"},
	}, "m")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.CreateToken{",
		"Source: game.TokenCopyOf(game.TokenCopySpec{",
		"Source:          game.TokenCopySourceEachInGroup,",
		"SetNotLegendary: true,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutableCardSourceCopyTargetToken covers the bare "target
// token" target noun ("Create a token that's a copy of target token you
// control." — Caretaker's Talent's level-2 ability). The target must lower to a
// permanent target restricted to tokens (TokenOnly), not an unrestricted
// permanent.
func TestGenerateExecutableCardSourceCopyTargetToken(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Token Copier",
		Layout:     "normal",
		ManaCost:   "{2}{W}",
		TypeLine:   "Instant",
		OracleText: "Create a token that's a copy of target token you control.",
		Colors:     []string{"W"},
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Allow:      game.TargetAllowPermanent,",
		"TokenOnly:  true,",
		"Source: game.TokenCopyOf(game.TokenCopySpec{",
		"Object: game.TargetPermanentReference(0),",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutableCardSourceCopyTokenOneOfThem covers a "one or more
// other creatures you control enter" trigger whose body copies one of the
// triggering creatures chosen by the controller ("create a token that's a copy
// of one of them." — Twilight Diviner). The copy source must lower to a
// controller-chosen member of the triggering event batch.
func TestGenerateExecutableCardSourceCopyTokenOneOfThem(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Twilight Diviner",
		Layout:   "normal",
		ManaCost: "{2}{B}",
		TypeLine: "Creature — Elf Cleric",
		OracleText: "When this creature enters, surveil 2.\n" +
			"Whenever one or more other creatures you control enter, if they entered or " +
			"were cast from a graveyard, create a token that's a copy of one of them. " +
			"This ability triggers only once each turn.",
		Colors: []string{"B"},
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.CreateToken{",
		"Source: game.TokenCopyOf(game.TokenCopySpec{",
		"Source: game.TokenCopySourceChosenFromTriggerBatch,",
		"MaxTriggersPerTurn: 1,",
		"Primitive: game.Surveil{",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutableCardSourceCopyTargetArtifactToken covers a typed
// token-qualified target noun ("Create a token that's a copy of target artifact
// token you control." — Worldwalker Helm's activated ability). The "<type>
// token" target must round-trip and lower to a permanent target restricted to
// artifact tokens (PermanentTypes + TokenOnly), copying that target object.
func TestGenerateExecutableCardSourceCopyTargetArtifactToken(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Artifact Token Copier",
		Layout:     "normal",
		ManaCost:   "{1}{U}",
		TypeLine:   "Artifact",
		OracleText: "{1}{U}, {T}: Create a token that's a copy of target artifact token you control.",
	}, "a")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Constraint: \"target artifact token you control\",",
		"Allow:      game.TargetAllowPermanent,",
		"PermanentTypes: []types.Card{types.Artifact},",
		"TokenOnly:      true,",
		"Source: game.TokenCopyOf(game.TokenCopySpec{",
		"Source: game.TokenCopySourceObject,",
		"Object: game.TargetPermanentReference(0),",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutableCardSourceCopyTokenExceptKeywordRider covers the inline
// copy-token "except <the token> has <keyword> and it isn't legendary" rider
// (Irenicus's Vile Duplication). The keyword-grant rider verb must fold into the
// copy create rather than stranding a separate keyword-grant effect, producing a
// token copy that drops legendary and adds the granted keyword.
func TestGenerateExecutableCardSourceCopyTokenExceptKeywordRider(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Irenicus's Vile Duplication",
		Layout:     "normal",
		ManaCost:   "{2}{U}",
		TypeLine:   "Sorcery",
		OracleText: "Create a token that's a copy of target creature you control, except the token has flying and it isn't legendary.",
		Colors:     []string{"U"},
	}, "i")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Source: game.TokenCopyOf(game.TokenCopySpec{",
		"SetNotLegendary: true,",
		"[]game.Keyword{game.Flying}",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutableCardSourceCopyTokenExceptQuotedAbilityFailsClosed covers
// a copy-token "except it has <keyword> and \"<quoted ability>\"" rider
// (Electroduplicate, Heat Shimmer). The quoted granted ability cannot be
// represented, so the card must fail closed rather than silently dropping the
// ability and keeping only the keyword.
func TestGenerateExecutableCardSourceCopyTokenExceptQuotedAbilityFailsClosed(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Electroduplicate",
		Layout:     "normal",
		ManaCost:   "{2}{R}",
		TypeLine:   "Sorcery",
		OracleText: "Create a token that's a copy of target creature you control, except it has haste and \"At the beginning of the end step, sacrifice this token.\"",
		Colors:     []string{"R"},
	}, "e")
	if err != nil {
		t.Fatal(err)
	}
	if source != "" || len(diagnostics) == 0 {
		t.Fatalf("want fail-closed (empty source, diagnostics); got source=%q diagnostics=%#v", source, diagnostics)
	}
}
