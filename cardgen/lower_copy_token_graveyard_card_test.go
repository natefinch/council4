package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceCopyTokenGraveyardCard covers the
// copy-a-graveyard-card token family ("{2}{R}, {T}: Create a token that's a copy
// of target creature card in your graveyard, except it's an artifact in addition
// to its other types. It gains haste. Sacrifice it at the beginning of the next
// end step." — Feldon of the Third Path). Unlike the battlefield-permanent copy
// (Cogwork Assembler), the blueprint is a card chosen in a graveyard, so the
// copy source is a TargetCardReference and the target is a card-in-graveyard
// spec. The "It gains haste." rider folds into the copy's granted keywords and
// the delayed "Sacrifice it" clause binds to the freshly created token through a
// linked key the CreateToken publishes.
func TestGenerateExecutableCardSourceCopyTokenGraveyardCard(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Feldon of the Third Path",
		Layout:     "normal",
		ManaCost:   "{1}{R}{R}",
		TypeLine:   "Legendary Creature — Human Artificer",
		OracleText: "{2}{R}, {T}: Create a token that's a copy of target creature card in your graveyard, except it's an artifact in addition to its other types. It gains haste. Sacrifice it at the beginning of the next end step.",
		Colors:     []string{"R"},
	}, "f")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Allow:      game.TargetAllowCard,",
		"TargetZone: zone.Graveyard,",
		"Primitive: game.CreateToken{",
		"Source: game.TokenCopyOf(game.TokenCopySpec{",
		"Object:      game.TargetCardReference(0),",
		"AddTypes:    []types.Card{types.Artifact},",
		"AddKeywords: []game.Keyword{game.Haste},",
		"PublishLinked: game.LinkedKey(\"delayed-sacrifice-1\"),",
		"Primitive: game.CreateDelayedTrigger{",
		"Timing: game.DelayedAtBeginningOfNextEndStep,",
		"Primitive: game.Sacrifice{",
		"Object: game.LinkedObjectReference(\"delayed-sacrifice-1\"),",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutableCardSourceCopyTokenBattlefieldPermanentUnchanged guards
// that the battlefield-permanent copy path is untouched by the graveyard-card
// copy support: the copy source stays a TargetPermanentReference and the token
// is not published under a linked key. Only a graveyard-card copy source takes
// the new publish path, so Molten Duplication's "It gains haste ... Sacrifice
// it" riders continue to bind directly to the target permanent.
func TestGenerateExecutableCardSourceCopyTokenBattlefieldPermanentUnchanged(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Molten Duplication",
		Layout:     "normal",
		ManaCost:   "{1}{R}",
		TypeLine:   "Sorcery",
		OracleText: "Create a token that's a copy of target artifact or creature you control, except it's an artifact in addition to its other types. It gains haste until end of turn. Sacrifice it at the beginning of the next end step.",
		Colors:     []string{"R"},
	}, "m")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if strings.Contains(source, "game.TargetCardReference(") {
		t.Fatalf("battlefield-permanent copy unexpectedly used a card reference:\n%s", source)
	}
	if !strings.Contains(source, "Object:   game.TargetPermanentReference(0),") {
		t.Fatalf("battlefield-permanent copy source missing TargetPermanentReference:\n%s", source)
	}
}
