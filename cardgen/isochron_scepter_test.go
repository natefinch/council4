package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

const isochronScepterOracle = "Imprint — When Isochron Scepter enters, you may exile an instant card with mana value 2 or less from your hand.\n" +
	"{2}, {T}: You may copy the exiled card. If you do, you may cast the copy without paying its mana cost."

// TestLowerIsochronScepterImprintETB proves the imprint enters trigger lowers to
// the optional exile-from-hand choice filtered to instant cards with mana value
// 2 or less, publishing the object-scoped imprint link the activated ability
// reads.
func TestLowerIsochronScepterImprintETB(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Isochron Scepter",
		Layout:     "normal",
		TypeLine:   "Artifact",
		ManaCost:   "{2}",
		OracleText: isochronScepterOracle,
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	seq := face.TriggeredAbilities[0].Content.Modes[0].Sequence
	if len(seq) != 1 {
		t.Fatalf("etb sequence length = %d, want 1", len(seq))
	}
	choose, ok := seq[0].Primitive.(game.ChooseFromZone)
	if !ok {
		t.Fatalf("etb primitive = %#v, want ChooseFromZone", seq[0].Primitive)
	}
	if choose.SourceZone != zone.Hand || choose.Destination.Zone != zone.Exile {
		t.Fatalf("etb move = %v -> %v, want hand -> exile", choose.SourceZone, choose.Destination.Zone)
	}
	if string(choose.Riders.PublishLinked) != imprintLinkKey || !choose.Riders.PublishObjectScoped {
		t.Fatalf("etb link riders = %+v, want object-scoped %q", choose.Riders, imprintLinkKey)
	}
	sel := choose.Filter
	if len(sel.RequiredTypesAny) != 1 || sel.RequiredTypesAny[0] != types.Instant {
		t.Fatalf("etb selection types = %v, want [Instant]", sel.RequiredTypesAny)
	}
	if !sel.ManaValue.Exists || sel.ManaValue.Val.Value != 2 {
		t.Fatalf("etb selection mana value = %+v, want <= 2", sel.ManaValue)
	}
}

// TestLowerIsochronScepterCopyCast proves the imprint copy/cast activated ability
// lowers to the two-instruction CopyCard + PlayLinkedExiledCard sequence, gated
// so the free copy cast happens only if a linked exiled card was copied.
func TestLowerIsochronScepterCopyCast(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Isochron Scepter",
		Layout:     "normal",
		TypeLine:   "Artifact",
		ManaCost:   "{2}",
		OracleText: isochronScepterOracle,
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	seq := face.ActivatedAbilities[0].Content.Modes[0].Sequence
	if len(seq) != 2 {
		t.Fatalf("activated sequence length = %d, want 2", len(seq))
	}

	copyInstr := seq[0]
	copyCard, ok := copyInstr.Primitive.(game.CopyCard)
	if !ok {
		t.Fatalf("instruction[0] primitive = %#v, want CopyCard", copyInstr.Primitive)
	}
	if copyCard.LinkID != imprintLinkKey {
		t.Fatalf("CopyCard.LinkID = %q, want %q", copyCard.LinkID, imprintLinkKey)
	}
	if !copyInstr.Optional {
		t.Fatal("CopyCard instruction must be optional (\"you may copy\")")
	}
	if copyInstr.PublishResult == "" {
		t.Fatal("CopyCard instruction must publish its result to gate the cast")
	}

	castInstr := seq[1]
	play, ok := castInstr.Primitive.(game.PlayLinkedExiledCard)
	if !ok {
		t.Fatalf("instruction[1] primitive = %#v, want PlayLinkedExiledCard", castInstr.Primitive)
	}
	if play.LinkID != imprintLinkKey || !play.Copy || !play.WithoutPayingManaCost {
		t.Fatalf("PlayLinkedExiledCard = %+v, want imprint copy free-cast", play)
	}
	if !castInstr.Optional {
		t.Fatal("PlayLinkedExiledCard instruction must be optional (\"you may cast\")")
	}
	if !castInstr.ResultGate.Exists ||
		castInstr.ResultGate.Val.Key != copyInstr.PublishResult ||
		castInstr.ResultGate.Val.Succeeded != game.TriTrue {
		t.Fatalf("cast gate = %+v, want success gate on %q", castInstr.ResultGate, copyInstr.PublishResult)
	}
}

// TestLowerImprintCopyCastFailsClosed proves a body variant that is not the exact
// imprint copy/cast idiom fails closed rather than lowering a partial ability.
func TestLowerImprintCopyCastFailsClosed(t *testing.T) {
	t.Parallel()
	face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:     "Faux Scepter",
		Layout:   "normal",
		TypeLine: "Artifact",
		ManaCost: "{2}",
		OracleText: "Imprint — When Faux Scepter enters, you may exile an instant card with mana value 2 or less from your hand.\n" +
			"{2}, {T}: You may copy the exiled card. If you do, you may cast the copy.",
	})
	if len(face.ActivatedAbilities) != 0 {
		t.Fatalf("activated abilities = %d, want 0 (fail closed)", len(face.ActivatedAbilities))
	}
}
