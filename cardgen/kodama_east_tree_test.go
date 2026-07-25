package cardgen

import (
	"testing"
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

// TestKodamaOfTheEastTreeCompilesBothNewCapabilities proves the whole card
// compiles with no diagnostics and threads both new capabilities into the
// CardDef: the choose-from-hand filter carries the event-relative
// ManaValueLessOrEqualEventPermanent bound, and the enter trigger carries both
// ExcludeSelf ("another") and the provenance guard, wrapped in an optional
// resolution ("you may").
func TestKodamaOfTheEastTreeCompilesBothNewCapabilities(t *testing.T) {
	t.Parallel()
	card := kodamaEastTreeCard()
	assertCardPaths(t, card,
		// Reach and Partner survive alongside the new trigger.
		"StaticAbilities[0].KeywordAbilities[0].(game.SimpleKeyword).Kind = game.Reach",
		"StaticAbilities[1].KeywordAbilities[0].(game.SimpleKeyword).Kind = game.Partner",
		// "Whenever another permanent you control enters".
		"TriggeredAbilities[0].Trigger.Pattern.Event = game.EventPermanentEnteredBattlefield",
		"TriggeredAbilities[0].Trigger.Pattern.Controller = game.TriggerControllerYou",
		"TriggeredAbilities[0].Trigger.Pattern.ExcludeSelf = true",
		// "if it wasn't put onto the battlefield with this ability": the
		// anti-recursion guard that stops one instance re-triggering on the
		// permanent it put onto the battlefield.
		"TriggeredAbilities[0].Trigger.InterveningIfEventPermanentWasNotPutByThisAbilitySource = true",
		// "you may put a permanent card with equal or lesser mana value from
		// your hand onto the battlefield", carrying the event-relative bound.
		"TriggeredAbilities[0].Content.Modes[0].Sequence[0].Optional = true",
		"Sequence[0].Primitive.(game.ChooseFromZone).SourceZone = zone.Hand",
		"Sequence[0].Primitive.(game.ChooseFromZone).Destination.Zone = zone.Battlefield",
		"Sequence[0].Primitive.(game.ChooseFromZone).Filter.ManaValueLessOrEqualEventPermanent = true",
	)
	// The shape must be exactly one trigger running exactly one instruction.
	// An extra ability or instruction would mean the text lowered twice or
	// partially, which the assertions above cannot detect on their own.
	assertCardPathsAbsent(t, card, "TriggeredAbilities[1]", "Modes[0].Sequence[1]", "Content.Modes[1]")
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
