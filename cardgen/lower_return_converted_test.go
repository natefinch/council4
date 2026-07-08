package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerReturnConvertedDiesTrigger proves the Transformers "return it to the
// battlefield converted" rider lowers to a PutOnBattlefield that enters the
// returned card transformed (as its back face). Optimus Prime's front face uses
// this on its dies trigger, the return-half of the convert subsystem.
func TestLowerReturnConvertedDiesTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bot",
		Layout:     "transform",
		TypeLine:   "Legendary Artifact Creature — Robot",
		ManaCost:   "{2}{W}",
		Power:      new("2"),
		Toughness:  new("2"),
		OracleText: "When Test Bot dies, return it to the battlefield converted under its owner's control.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	seq := face.TriggeredAbilities[0].Content.Modes[0].Sequence
	if len(seq) != 1 {
		t.Fatalf("sequence = %#v, want single put-on-battlefield", seq)
	}
	put, ok := seq[0].Primitive.(game.PutOnBattlefield)
	if !ok {
		t.Fatalf("primitive = %T, want game.PutOnBattlefield", seq[0].Primitive)
	}
	if !put.EntryTransformed {
		t.Fatal("PutOnBattlefield.EntryTransformed = false, want true")
	}
	if put.Source != game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceEvent}) {
		t.Fatalf("PutOnBattlefield.Source = %#v, want event card reference", put.Source)
	}
}

// TestLowerReturnPlainNotTransformed proves an ordinary "return it to the
// battlefield" (no "converted" rider) leaves EntryTransformed unset, so only the
// Transformers convert wording enters the card as its back face.
func TestLowerReturnPlainNotTransformed(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Plain Bot",
		Layout:     "transform",
		TypeLine:   "Legendary Artifact Creature — Robot",
		ManaCost:   "{2}{W}",
		Power:      new("2"),
		Toughness:  new("2"),
		OracleText: "When Plain Bot dies, return it to the battlefield under its owner's control.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	put, ok := face.TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.PutOnBattlefield)
	if !ok {
		t.Fatalf("primitive = %T, want game.PutOnBattlefield", face.TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive)
	}
	if put.EntryTransformed {
		t.Fatal("PutOnBattlefield.EntryTransformed = true, want false for plain return")
	}
}
