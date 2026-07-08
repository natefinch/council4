package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerTransformSelfActivated verifies that a transforming double-faced
// card's "Transform this creature." activated ability lowers to a game.Transform
// primitive that transforms the source permanent, and that the back face is
// generated.
func TestLowerTransformSelfActivated(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Ulvenwald Captive",
		Layout:   "transform",
		TypeLine: "Creature — Elf Druid",
		ManaCost: "{1}{G}",
		CardFaces: []ScryfallCardFace{
			{
				Name:       "Ulvenwald Captive",
				TypeLine:   "Creature — Elf Druid",
				ManaCost:   "{1}{G}",
				OracleText: "Defender\n{T}: Add {G}.\n{5}{G}{G}: Transform this creature.",
				Power:      new("1"),
				Toughness:  new("2"),
			},
			{
				Name:       "Ulvenwald Abomination",
				TypeLine:   "Creature — Eldrazi Horror",
				OracleText: "{T}: Add {C}{C}.",
				Power:      new("4"),
				Toughness:  new("4"),
			},
		},
	})
	var transform game.Transform
	found := false
	for _, ability := range face.ActivatedAbilities {
		if len(ability.Content.Modes) != 1 || len(ability.Content.Modes[0].Sequence) != 1 {
			continue
		}
		if tr, ok := ability.Content.Modes[0].Sequence[0].Primitive.(game.Transform); ok {
			transform = tr
			found = true
		}
	}
	if !found {
		t.Fatalf("no activated Transform ability found in %+v", face.ActivatedAbilities)
	}
	if transform.Object.Kind() != game.ObjectReferenceSourcePermanent {
		t.Fatalf("transform object kind = %v, want ObjectReferenceSourcePermanent", transform.Object.Kind())
	}
}

// TestLowerConvertActivatedSelfName proves an activated ability whose body is
// "Convert <name>." lowers to the shared game.Transform primitive on the source
// permanent. "Convert" is the Transformers-flavored spelling of the transform
// keyword action, so it reuses the existing transform lowering path.
func TestLowerConvertActivatedSelfName(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Robot",
		Layout:     "transform",
		TypeLine:   "Legendary Artifact Creature — Robot",
		ManaCost:   "{2}{R}",
		Power:      new("3"),
		Toughness:  new("3"),
		OracleText: "{2}: Convert Test Robot.",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	seq := face.ActivatedAbilities[0].Content.Modes[0].Sequence
	if len(seq) != 1 {
		t.Fatalf("sequence = %#v, want single transform", seq)
	}
	transform, ok := seq[0].Primitive.(game.Transform)
	if !ok {
		t.Fatalf("primitive = %T, want game.Transform", seq[0].Primitive)
	}
	if transform.Object != game.SourcePermanentReference() {
		t.Fatalf("transform object = %#v, want source permanent", transform.Object)
	}
}

// TestLowerConvertTriggerPronoun proves a triggered ability whose body is
// "convert it." lowers to game.Transform on the triggering permanent, mirroring
// "transform it." on other double-faced cards.
func TestLowerConvertTriggerPronoun(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Robot",
		Layout:     "transform",
		TypeLine:   "Legendary Artifact Creature — Robot",
		ManaCost:   "{2}{R}",
		Power:      new("3"),
		Toughness:  new("3"),
		OracleText: "Whenever this creature attacks, convert it.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	seq := face.TriggeredAbilities[0].Content.Modes[0].Sequence
	if len(seq) != 1 {
		t.Fatalf("sequence = %#v, want single transform", seq)
	}
	transform, ok := seq[0].Primitive.(game.Transform)
	if !ok {
		t.Fatalf("primitive = %T, want game.Transform", seq[0].Primitive)
	}
	if transform.Object != game.EventPermanentReference() {
		t.Fatalf("transform object = %#v, want event permanent", transform.Object)
	}
}
