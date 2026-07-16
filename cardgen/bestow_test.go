package cardgen

import (
	"reflect"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
)

// nyxbornRollickerCard is the real, fixed-mana Bestow card used to prove the
// Bestow keyword lowers generically: a plain enchantment creature whose only
// extra text is a fixed-cost Bestow and a static "Enchanted creature gets"
// grant.
func nyxbornRollickerCard() *ScryfallCard {
	return &ScryfallCard{
		Name:      "Nyxborn Rollicker",
		Layout:    "normal",
		TypeLine:  "Enchantment Creature — Snake",
		ManaCost:  "{R}",
		Power:     new("1"),
		Toughness: new("1"),
		OracleText: "Bestow {1}{R} (If you cast this card for its bestow cost, it's an Aura spell with enchant creature. It becomes a creature if it's not attached to a creature.)\n" +
			"Enchanted creature gets +1/+1.",
	}
}

// TestLowerBestowStaticAbility proves the Bestow keyword lowers to a single
// BestowStaticAbility carrying the fixed bestow mana cost, alongside the
// separately lowered "Enchanted creature gets +1/+1" grant, on an ordinary
// enchantment creature card that is neither an Aura subtype nor an Enchant
// keyword card.
func TestLowerBestowStaticAbility(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, nyxbornRollickerCard())
	if len(face.StaticAbilities) != 2 {
		t.Fatalf("static abilities = %d, want 2", len(face.StaticAbilities))
	}
	bestow, ok := game.StaticBodyBestow(&face.StaticAbilities[0].Body)
	if !ok {
		t.Fatalf("first static ability is not a Bestow ability: %#v", face.StaticAbilities[0].Body)
	}
	if want := (cost.Mana{cost.O(1), cost.R}); !reflect.DeepEqual(bestow.Cost, want) {
		t.Fatalf("bestow cost = %v, want %v", bestow.Cost, want)
	}
	// The grant must not carry the Bestow keyword; it is a plain attached-group
	// pump so it only applies while the permanent is attached.
	if _, ok := game.StaticBodyBestow(&face.StaticAbilities[1].Body); ok {
		t.Fatal("the enchanted-creature grant unexpectedly carries the Bestow keyword")
	}
}

// TestGenerateExecutableCardSourceBestow proves the generated Go source for a
// Bestow card calls the game.BestowStaticAbility factory with the fixed bestow
// mana cost, so the committed card reconstructs the keyword, target, and gated
// type-change deterministically.
func TestGenerateExecutableCardSourceBestow(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(nyxbornRollickerCard(), "n")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.BestowStaticAbility(cost.Mana{cost.O(1), cost.R}, &game.TargetSpec{",
		`Constraint: "creature",`,
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestBestowVariableCostFailsClosed proves a Bestow card with a variable ({X})
// bestow cost is rejected: only fixed-mana Bestow is supported, and the extra
// counters/dynamic-P/T body is unsupported too, so the whole card fails closed.
func TestBestowVariableCostFailsClosed(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:      "Nyxborn Hydra",
		Layout:    "normal",
		TypeLine:  "Enchantment Creature — Hydra",
		ManaCost:  "{X}{G}",
		Power:     new("0"),
		Toughness: new("0"),
		OracleText: "Bestow {X}{G}{G} (If you cast this card for its bestow cost, it's an Aura spell with enchant creature. It becomes a creature again if it's not attached.)\n" +
			"Reach, trample\n" +
			"This permanent enters with X +1/+1 counters on it.\n" +
			"Enchanted creature gets +1/+1 for each +1/+1 counter on this Aura and has reach and trample.",
	})
}

// TestBestowNonManaCostFailsClosed proves the em-dash non-mana Bestow form
// ("Bestow—{R}, Collect evidence 6.") is not recognized as a Bestow keyword and
// fails closed rather than being misread as a fixed-mana Bestow.
func TestBestowNonManaCostFailsClosed(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:      "Detective's Phoenix",
		Layout:    "normal",
		TypeLine:  "Enchantment Creature — Phoenix",
		ManaCost:  "{2}{R}",
		Power:     new("2"),
		Toughness: new("2"),
		OracleText: "Bestow—{R}, Collect evidence 6. (To pay this bestow cost, pay {R} and exile cards with total mana value 6 or greater from your graveyard.)\n" +
			"Flying, haste\n" +
			"Enchanted creature gets +2/+2 and has flying and haste.\n" +
			"You may cast this card from your graveyard using its bestow ability.",
	})
}
