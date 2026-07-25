package cardgen

import (
	"testing"
)

// TestGenerateEmryGraveyardTargetCastSource covers the split-sentence graveyard
// cast-permission body "Choose target <card> in your graveyard. You may cast
// that card this turn." (Emry, Lurker of the Loch). The leading "Choose target
// artifact card in your graveyard." sentence selects the target and the
// optional "You may cast that card this turn." sentence refers back to it, which
// lowers to a GrantCastPermission that lets the controller cast the chosen
// graveyard card until end of turn.
func TestGenerateEmryGraveyardTargetCastSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Emry, Lurker of the Loch",
		Layout:   "normal",
		TypeLine: "Legendary Creature — Merfolk Wizard",
		ManaCost: "{2}{U}",
		OracleText: "Affinity for artifacts (This spell costs {1} less to cast for each artifact you control.)\n" +
			"When Emry enters, mill four cards.\n" +
			"{T}: Choose target artifact card in your graveyard. You may cast that card this turn. (You still pay its costs. Timing rules still apply.)",
	}, "e")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Constraint: \"target artifact card in your graveyard\",",
		"Allow:      game.TargetAllowCard,",
		"TargetZone: zone.Graveyard,",
		"Primitive: game.GrantCastPermission{",
		// Card is now formatted multi-line; containsNormalized handles whitespace.
		"Card:     game.CardReference{Kind: game.CardReferenceTarget},",
		"FromZone: zone.Graveyard,",
		// Face: game.FaceFront is omitted (zero value — front face is the default).
		// Assert the full GrantCastPermission shape to verify Face is not set.
		"game.GrantCastPermission{Card: game.CardReference{Kind: game.CardReferenceTarget}, FromZone: zone.Graveyard, Duration: game.DurationUntilEndOfTurn}",
		"Duration: game.DurationUntilEndOfTurn,",
	} {
		if !containsNormalized(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}

// TestGraveyardTargetCastThatCardRequiresGraveyardTarget verifies the recognizer
// fails closed when the chosen target is not a graveyard card: a battlefield
// permanent "that card" cast permission is outside the recognized envelope and
// must leave the body unsupported rather than granting a graveyard cast.
func TestGraveyardTargetCastThatCardRequiresGraveyardTarget(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Battlefield Cast",
		Layout:     "normal",
		TypeLine:   "Creature — Wizard",
		ManaCost:   "{2}{U}",
		OracleText: "{T}: Choose target artifact you control. You may cast that card this turn.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected a diagnostic for a non-graveyard cast-permission body")
	}
}
