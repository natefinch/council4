package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// kodamaEastTreeCard builds the authoritative Kodama of the East Tree printing.
// Its ETB trigger "Whenever another permanent you control enters, if it wasn't
// put onto the battlefield with this ability, you may put a permanent card with
// equal or lesser mana value from your hand onto the battlefield." exercises two
// capabilities that no earlier card needed together: an event-relative "equal or
// lesser mana value" bound comparing each hand card to the permanent that just
// entered, and an anti-recursion intervening-if that stops one ability instance
// from re-triggering on the permanent it put onto the battlefield.
func kodamaEastTreeCard() *ScryfallCard {
	return &ScryfallCard{
		Name:      "Kodama of the East Tree",
		Layout:    "normal",
		TypeLine:  "Legendary Creature — Spirit",
		ManaCost:  "{4}{G}{G}",
		Power:     new("6"),
		Toughness: new("6"),
		OracleText: "Reach\n" +
			"Whenever another permanent you control enters, if it wasn't put onto the battlefield with this ability, you may put a permanent card with equal or lesser mana value from your hand onto the battlefield.\n" +
			"Partner (You can have two commanders if both have partner.)",
	}
}

// TestGenerateExecutableKodamaOfTheEastTreeSource proves the whole card lowers
// with no diagnostics and threads both new capabilities to the runtime: the
// choose-from-hand filter carries the event-relative ManaValueLessOrEqualEvent-
// Permanent bound, and the enter trigger carries both ExcludeSelf ("another")
// and the InterveningIfEventPermanentWasNotPutByThisAbilitySource provenance
// guard, wrapped in an optional resolution ("you may").
func TestGenerateExecutableKodamaOfTheEastTreeSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(kodamaEastTreeCard(), "k")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.ReachStaticBody",
		"game.PartnerStaticBody",
		"Event:       game.EventPermanentEnteredBattlefield",
		"ExcludeSelf: true",
		"InterveningIfEventPermanentWasNotPutByThisAbilitySource: true",
		"Primitive: game.ChooseFromZone{",
		"ManaValueLessOrEqualEventPermanent: true",
		"Optional: true",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}

// TestKodamaEastTreeLowersPutFromHandChoice confirms the resolving effect lowers
// to a single optional put-from-hand choice of exactly one card carrying the
// event-relative mana-value bound, rather than any partial or over-permissive
// shape.
func TestKodamaEastTreeLowersPutFromHandChoice(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, kodamaEastTreeCard())
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d; want the single ETB put trigger", len(face.TriggeredAbilities))
	}
	trigger := face.TriggeredAbilities[0]
	if !trigger.Trigger.Pattern.ExcludeSelf {
		t.Fatalf("trigger pattern = %#v; want ExcludeSelf for \"another permanent\"", trigger.Trigger.Pattern)
	}
	if !trigger.Trigger.InterveningIfEventPermanentWasNotPutByThisAbilitySource {
		t.Fatalf("trigger = %#v; want provenance intervening-if", trigger.Trigger)
	}
	mode := trigger.Content.Modes[0]
	if len(mode.Sequence) != 1 || !mode.Sequence[0].Optional {
		t.Fatalf("sequence = %#v; want one optional instruction", mode.Sequence)
	}
	choose, ok := mode.Sequence[0].Primitive.(game.ChooseFromZone)
	if !ok {
		t.Fatalf("primitive = %#v; want game.ChooseFromZone", mode.Sequence[0].Primitive)
	}
	if choose.SourceZone != zone.Hand || choose.Destination.Zone != zone.Battlefield {
		t.Fatalf("choose zones = %v -> %v; want hand -> battlefield", choose.SourceZone, choose.Destination.Zone)
	}
	if !choose.Filter.ManaValueLessOrEqualEventPermanent {
		t.Fatalf("choose filter = %#v; want ManaValueLessOrEqualEventPermanent", choose.Filter)
	}
}

// TestEventRelativeManaValueCastForFreeFailsClosed guards the anti-fail-open
// invariant behind Kodama of the East Tree's mana-value bound. The parser now
// recognizes "equal or lesser mana value", so a cast-for-free spell that reuses
// that wording ("Counter target spell. You may cast a spell with equal or lesser
// mana value from your hand without paying its mana cost.", Reinterpret) would
// silently drop the bound and let any nonland spell be cast for free. The bound
// is only expressible on a put-from-hand choice that threads the triggering
// event, so every other context — here a free cast with no event permanent —
// must fail closed rather than generate a more permissive card.
func TestEventRelativeManaValueCastForFreeFailsClosed(t *testing.T) {
	t.Parallel()
	reinterpret := &ScryfallCard{
		Name:       "Reinterpret",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{2}{U}{R}",
		OracleText: "Counter target spell. You may cast a spell with equal or lesser mana value from your hand without paying its mana cost.",
	}
	lowerSingleFaceExpectingUnsupported(t, reinterpret)
}
