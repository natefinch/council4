package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func millPutAmongSingleCard() *ScryfallCard {
	power, toughness := "2", "2"
	return &ScryfallCard{
		Name:     "Test Burrower",
		Layout:   "normal",
		ManaCost: "{3}",
		TypeLine: "Creature — Construct",
		OracleText: "Whenever Test Burrower deals combat damage to a player, mill four cards. " +
			"You may put a permanent card from among them onto the battlefield.",
		Power:     &power,
		Toughness: &toughness,
	}
}

func millPutAmongAnyNumberCard() *ScryfallCard {
	power, toughness := "2", "2"
	return &ScryfallCard{
		Name:     "Test Leaper",
		Layout:   "normal",
		ManaCost: "{3}",
		TypeLine: "Creature — Frog",
		OracleText: "Whenever Test Leaper deals combat damage to a player, you may mill that many cards. " +
			"Put any number of land cards from among them onto the battlefield tapped.",
		Power:     &power,
		Toughness: &toughness,
	}
}

// TestLowerMillThenOptionalPutPermanentAmongToBattlefield asserts the broadened
// single-card form ("mill four cards. You may put a permanent card from among
// them onto the battlefield.") lowers to a mandatory fixed mill that publishes
// the milled cards followed by one optional put returning a single permanent
// card from among them onto the battlefield.
func TestLowerMillThenOptionalPutPermanentAmongToBattlefield(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, millPutAmongSingleCard())
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	sequence := face.TriggeredAbilities[0].Content.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("sequence length = %d, want 2", len(sequence))
	}
	mill, ok := sequence[0].Primitive.(game.Mill)
	if !ok {
		t.Fatalf("sequence[0] = %#v, want Mill", sequence[0].Primitive)
	}
	if sequence[0].Optional {
		t.Fatal("mill must be mandatory")
	}
	if mill.PublishLinked != milledCardsLinkKey || mill.Amount.Value() != 4 {
		t.Fatalf("mill = %#v, want fixed 4 publishing milled cards", mill)
	}
	put, ok := sequence[1].Primitive.(game.ReturnFromGraveyard)
	if !ok {
		t.Fatalf("sequence[1] = %#v, want ReturnFromGraveyard", sequence[1].Primitive)
	}
	if !sequence[1].Optional {
		t.Fatal("single-card put must be optional (\"you may\")")
	}
	if put.AnyNumber {
		t.Fatal("single-card put must not be any-number")
	}
	if put.Amount.Value() != 1 || put.Destination != zone.Battlefield ||
		put.FromLinked != milledCardsLinkKey || put.EntryTapped {
		t.Fatalf("put = %#v, want one card onto battlefield from milled cards, untapped", put)
	}
	wantTypes := []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Land, types.Planeswalker, types.Battle}
	if len(put.Selection.RequiredTypesAny) != len(wantTypes) {
		t.Fatalf("put selection = %#v, want permanent type union", put.Selection)
	}
}

// TestLowerMillThenPutAnyNumberLandsAmongToBattlefield asserts the broadened
// any-number form ("you may mill that many cards. Put any number of land cards
// from among them onto the battlefield tapped.") lowers to an optional dynamic
// mill that publishes the milled cards followed by a mandatory any-number put of
// land cards from among them entering tapped.
func TestLowerMillThenPutAnyNumberLandsAmongToBattlefield(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, millPutAmongAnyNumberCard())
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	sequence := face.TriggeredAbilities[0].Content.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("sequence length = %d, want 2", len(sequence))
	}
	mill, ok := sequence[0].Primitive.(game.Mill)
	if !ok {
		t.Fatalf("sequence[0] = %#v, want Mill", sequence[0].Primitive)
	}
	if !sequence[0].Optional {
		t.Fatal("\"you may mill\" must lower to an optional mill")
	}
	dynamic := mill.Amount.DynamicAmount()
	if !dynamic.Exists || dynamic.Val.Kind != game.DynamicAmountEventDamage {
		t.Fatalf("mill amount = %#v, want dynamic event damage", mill.Amount)
	}
	if mill.PublishLinked != milledCardsLinkKey {
		t.Fatalf("mill PublishLinked = %q", mill.PublishLinked)
	}
	put, ok := sequence[1].Primitive.(game.ReturnFromGraveyard)
	if !ok {
		t.Fatalf("sequence[1] = %#v, want ReturnFromGraveyard", sequence[1].Primitive)
	}
	if sequence[1].Optional {
		t.Fatal("mandatory \"Put any number\" must not be an optional instruction")
	}
	if !put.AnyNumber || !put.EntryTapped || put.Destination != zone.Battlefield ||
		put.FromLinked != milledCardsLinkKey {
		t.Fatalf("put = %#v, want any-number tapped lands onto battlefield from milled cards", put)
	}
	if len(put.Selection.RequiredTypes) != 1 || put.Selection.RequiredTypes[0] != types.Land {
		t.Fatalf("put selection = %#v, want land cards", put.Selection)
	}
}
