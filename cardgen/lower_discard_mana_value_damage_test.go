package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerDiscardDrawThenManaValueDamage proves Summon: Kujata's chapter III
// lowers its reflexive "Discard a card, then draw two cards. When you discard a
// card this way, this creature deals damage equal to that card's mana value to
// each opponent." into a three-instruction sequence: a Discard that publishes
// the discarded card under a linked key, the two-card Draw, and a Damage whose
// dynamic amount reads the linked card's mana value and hits each opponent.
func TestLowerDiscardDrawThenManaValueDamage(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:      "Summon: Kujata",
		Layout:    "saga",
		TypeLine:  "Enchantment Creature — Saga Ox",
		ManaCost:  "{5}{R}",
		Power:     new("0"),
		Toughness: new("0"),
		OracleText: "(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)\n" +
			"I — Lightning — This creature deals 3 damage to each of up to two target creatures.\n" +
			"II — Ice — Up to three target creatures can't block this turn.\n" +
			"III — Fire — Discard a card, then draw two cards. When you discard a card this way, this creature deals damage equal to that card's mana value to each opponent.\n" +
			"Trample, haste",
	})
	if len(face.ChapterAbilities) != 3 {
		t.Fatalf("chapter abilities = %d, want 3", len(face.ChapterAbilities))
	}
	mode := face.ChapterAbilities[2].Content.Modes[0]
	if len(mode.Sequence) != 3 {
		t.Fatalf("chapter III sequence len = %d, want 3", len(mode.Sequence))
	}

	discard, ok := mode.Sequence[0].Primitive.(game.Discard)
	if !ok {
		t.Fatalf("instruction 0 = %#v, want Discard", mode.Sequence[0].Primitive)
	}
	if discard.PublishLinked == "" {
		t.Fatalf("discard does not publish a linked key: %#v", discard)
	}
	if discard.EntireHand || discard.AtRandom {
		t.Fatalf("discard should be a single chosen card: %#v", discard)
	}

	if _, ok := mode.Sequence[1].Primitive.(game.Draw); !ok {
		t.Fatalf("instruction 1 = %#v, want Draw", mode.Sequence[1].Primitive)
	}

	damage, ok := mode.Sequence[2].Primitive.(game.Damage)
	if !ok {
		t.Fatalf("instruction 2 = %#v, want Damage", mode.Sequence[2].Primitive)
	}
	if !damage.Amount.IsDynamic() {
		t.Fatalf("damage amount is not dynamic: %#v", damage.Amount)
	}
}
