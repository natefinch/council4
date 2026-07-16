package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

const pollywogProdigyOracle = "Evolve (Whenever a creature you control enters, if that creature has greater power or toughness than this creature, put a +1/+1 counter on this creature.)\nWhenever an opponent casts a noncreature spell with mana value less than this creature's power, draw a card."

func pollywogProdigyCard() *ScryfallCard {
	power, toughness := "1", "3"
	return &ScryfallCard{
		Name:       "Pollywog Prodigy",
		Layout:     "normal",
		ManaCost:   "{1}{U}",
		TypeLine:   "Creature — Frog Wizard",
		OracleText: pollywogProdigyOracle,
		Power:      &power,
		Toughness:  &toughness,
	}
}

func TestGenerateExecutableCardSourcePollywogProdigy(t *testing.T) {
	t.Parallel()
	card := pollywogProdigyCard()
	source, diagnostics, err := GenerateExecutableCardSource(card, "p")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.EvolveStaticBody",
		"Event:         game.EventSpellCast",
		"game.TriggerControllerOpponent",
		"ExcludedTypes: []types.Card{types.Creature}",
		"ManaValueLessThanSourcePower: true",
		"game.Draw{",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}

	face := lowerSingleFace(t, card)
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	pattern := face.TriggeredAbilities[0].Trigger.Pattern
	if pattern.Event != game.EventSpellCast {
		t.Fatalf("event = %v, want EventSpellCast", pattern.Event)
	}
	if pattern.Controller != game.TriggerControllerOpponent {
		t.Fatalf("controller = %v, want opponent", pattern.Controller)
	}
	if !pattern.CardSelection.ManaValueLessThanSourcePower {
		t.Fatalf("CardSelection = %#v, want ManaValueLessThanSourcePower", pattern.CardSelection)
	}
	if pattern.CardSelection.ManaValue.Exists {
		t.Fatalf("CardSelection carries a fixed mana-value bound: %#v", pattern.CardSelection)
	}
	if len(pattern.CardSelection.ExcludedTypes) != 1 ||
		pattern.CardSelection.ExcludedTypes[0] != types.Creature {
		t.Fatalf("excluded types = %#v, want creature", pattern.CardSelection.ExcludedTypes)
	}
	// Evolve is preserved as a static ability alongside the new trigger.
	if len(face.StaticAbilities) == 0 {
		t.Fatal("Evolve static ability was dropped")
	}
}

func TestPollywogProdigyFailClosedNearMisses(t *testing.T) {
	t.Parallel()
	for _, oracle := range []string{
		// Copies are not casts; the runtime spell-copy event must not fire.
		"Whenever an opponent copies a noncreature spell with mana value less than this creature's power, draw a card.",
		// Toughness is not the modeled source characteristic.
		"Whenever an opponent casts a noncreature spell with mana value less than this creature's toughness, draw a card.",
		// A different possessive source (another player's) is not this creature's power.
		"Whenever an opponent casts a noncreature spell with mana value less than that player's power, draw a card.",
	} {
		t.Run(oracle, func(t *testing.T) {
			t.Parallel()
			power, toughness := "1", "3"
			faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Near Miss",
				Layout:     "normal",
				TypeLine:   "Creature — Frog Wizard",
				OracleText: oracle,
				Power:      &power,
				Toughness:  &toughness,
			})
			if len(diagnostics) == 0 {
				t.Fatalf("expected unsupported diagnostic for %q", oracle)
			}
			if len(faces) != 0 && len(faces[0].TriggeredAbilities) != 0 {
				t.Fatalf("unexpected executable trigger for %q", oracle)
			}
		})
	}
}
