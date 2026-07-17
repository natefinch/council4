package cardgen

import (
	"strings"
	"testing"
)

// ojerTaqCard returns the transform Ojer Taq, Deepest Foundation // Temple of
// Civilization Scryfall record. The front face composes vigilance, the
// creature-token tripling replacement, and the dies-return-tapped-transformed
// trigger; the back face composes the {W} mana ability and the sorcery-speed
// transform activation gated on having attacked with three or more creatures
// this turn.
func ojerTaqCard() *ScryfallCard {
	frontPower := "6"
	frontToughness := "6"
	return &ScryfallCard{
		Name:     "Ojer Taq, Deepest Foundation // Temple of Civilization",
		Layout:   "transform",
		TypeLine: "Legendary Creature — God // Land",
		ManaCost: "{4}{W}{W}",
		Colors:   []string{"W"},
		CardFaces: []ScryfallCardFace{
			{
				Name:     "Ojer Taq, Deepest Foundation",
				TypeLine: "Legendary Creature — God",
				ManaCost: "{4}{W}{W}",
				OracleText: "Vigilance\n" +
					"If one or more creature tokens would be created under your control, three times that many of those tokens are created instead.\n" +
					"When Ojer Taq dies, return it to the battlefield tapped and transformed under its owner's control.",
				Colors:    []string{"W"},
				Power:     &frontPower,
				Toughness: &frontToughness,
			},
			{
				Name:     "Temple of Civilization",
				TypeLine: "Land",
				OracleText: "(Transforms from Ojer Taq, Deepest Foundation.)\n" +
					"{T}: Add {W}.\n" +
					"{2}{W}, {T}: Transform this land. Activate only if you attacked with three or more creatures this turn and only as a sorcery.",
			},
		},
	}
}

// TestGenerateExecutableOjerTaqBothFaces proves the full transform Ojer Taq
// compiles with no diagnostics and both faces lower to the expected primitives:
// the front-face vigilance keyword, the creature-only token tripling
// replacement (a generic factor-3 token multiplier restricted to creature
// tokens under the controller), and the dies trigger returning the permanent
// to the battlefield tapped and transformed; and the back-face {W} mana ability
// plus the sorcery-speed transform activation whose gate is an event-history
// condition requiring three or more attacker-declared events this turn.
func TestGenerateExecutableOjerTaqBothFaces(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(ojerTaqCard(), "o")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Layout: game.LayoutTransform,",
		// Front: vigilance static ability.
		"game.VigilanceStaticBody,",
		// Front: creature-only token tripling via the generic factor-3 multiplier.
		"game.TokenCreationReplacementFiltered(\"If one or more creature tokens would be created under your control, three times that many of those tokens are created instead.\", &game.TokenCreationReplacementSpec{Multiplier: 3, Types: []types.Card{types.Creature}, Filter: game.TriggerControllerYou}),",
		// Front: dies trigger returning to the battlefield tapped and transformed.
		"Event:            game.EventPermanentDied,",
		"Primitive: game.PutOnBattlefield{",
		"EntryTapped:      true,",
		"EntryTransformed: true,",
		// Back: {W} mana ability.
		"game.TapManaAbility(mana.W),",
		// Back: sorcery-speed transform activation gated on the attack event history.
		"Timing:          game.SorceryOnly,",
		"EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{",
		"Event:      game.EventAttackerDeclared,",
		"Window: game.EventHistoryCurrentTurn, MinCount: 3}),",
		"Primitive: game.Transform{",
	} {
		if !strings.Contains(source, want) {
			t.Errorf("generated source missing %q\n---\n%s", want, source)
		}
	}
}
