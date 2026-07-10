package cardgen

import (
	"strings"
	"testing"
)

// flamewarCard returns the transform Flamewar // Flamewar Scryfall record whose
// back face exercises the face-down top-of-library exile with a named counter
// fed by the combat-damage "that many" dynamic amount, and whose front face
// exercises the mass return-from-exile-by-counter.
func flamewarCard() *ScryfallCard {
	power := "3"
	toughness := "2"
	backPower := "2"
	backToughness := "1"
	return &ScryfallCard{
		Name:     "Flamewar, Brash Veteran // Flamewar, Streetwise Operative",
		Layout:   "transform",
		TypeLine: "Legendary Artifact Creature — Robot // Legendary Artifact — Vehicle",
		ManaCost: "{1}{B}{R}",
		CardFaces: []ScryfallCardFace{
			{
				Name:     "Flamewar, Brash Veteran",
				TypeLine: "Legendary Artifact Creature — Robot",
				ManaCost: "{1}{B}{R}",
				OracleText: "More Than Meets the Eye {B}{R} (You may cast this card converted for {B}{R}.)\n" +
					"Sacrifice another artifact: Put a +1/+1 counter on Flamewar and convert it. Activate only as a sorcery.\n" +
					"{1}, Discard your hand: Put all exiled cards you own with intel counters on them into your hand.",
				Power:     &power,
				Toughness: &toughness,
			},
			{
				Name:     "Flamewar, Streetwise Operative",
				TypeLine: "Legendary Artifact — Vehicle",
				OracleText: "Living metal (During your turn, this Vehicle is also a creature.)\n" +
					"Menace, deathtouch\n" +
					"Whenever Flamewar deals combat damage to a player, exile that many cards from the top of your library face down. Put an intel counter on each of them. Convert Flamewar.",
				Power:     &backPower,
				Toughness: &backToughness,
			},
		},
	}
}

// TestGenerateExecutableFlamewarBothFaces proves the full transform Flamewar
// compiles with no diagnostics and both faces lower to the expected primitives:
// the front-face mass return-from-exile-by-counter and sacrifice-cost
// counter-then-convert, and the back-face combat-damage-scaled face-down
// top-of-library exile that stamps the card-defined named counter before
// converting. The named counter (intel) flows entirely from the parser; nothing
// downstream names it, so any named-counter-exile card lowers the same way.
func TestGenerateExecutableFlamewarBothFaces(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(flamewarCard(), "f")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		// Front: sacrifice-cost, sorcery-speed counter-then-convert.
		"Kind:               cost.AdditionalSacrifice,",
		"Timing:         game.SorceryOnly,",
		"Primitive: game.AddCounter{",
		// Front: discard-hand mass return filtered by the named counter.
		"Kind:          cost.AdditionalDiscard,",
		"Primitive: game.ReturnExiledCardsWithCounter{",
		"Counter: counter.Intel,",
		// Back: combat-damage-scaled face-down exile with the named counter.
		"Event:               game.EventDamageDealt,",
		"Primitive: game.ExileTopOfLibrary{",
		"Kind:       game.DynamicAmountEventDamage,",
		"Counter:  opt.Val(counter.Intel),",
		"FaceDown: true,",
		"Primitive: game.Transform{",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
