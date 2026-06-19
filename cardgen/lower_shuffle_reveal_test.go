package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

const shuffleRevealPermanentText = "The owner of target permanent shuffles it into their library, then reveals the top card of their library. If it's a permanent card, they put it onto the battlefield."

func TestLowerShuffleRevealPermanentSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Warp",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: shuffleRevealPermanentText,
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 || len(mode.Sequence) != 3 {
		t.Fatalf("mode = %#v, want one target and three instructions", mode)
	}
	shuffle, ok := mode.Sequence[0].Primitive.(game.ShufflePermanentIntoLibrary)
	if !ok || shuffle.Object != game.TargetPermanentReference(0) {
		t.Fatalf("shuffle = %#v", mode.Sequence[0].Primitive)
	}
	reveal, ok := mode.Sequence[1].Primitive.(game.Reveal)
	if !ok ||
		reveal.Amount.Value() != 1 ||
		reveal.Player.Kind() != game.PlayerReferenceObjectOwner ||
		reveal.PublishLinked == "" {
		t.Fatalf("reveal = %#v", mode.Sequence[1].Primitive)
	}
	put, ok := mode.Sequence[2].Primitive.(game.PutOnBattlefield)
	if !ok ||
		!put.Recipient.Exists ||
		put.Recipient.Val.Kind() != game.PlayerReferenceObjectOwner {
		t.Fatalf("put = %#v", mode.Sequence[2].Primitive)
	}
	key, linked := put.Source.LinkedKey()
	condition := mode.Sequence[2].CardCondition
	if !linked ||
		key != reveal.PublishLinked ||
		!condition.Exists ||
		condition.Val.Card.Kind != game.CardReferenceLinked ||
		condition.Val.Card.LinkID != string(key) ||
		!condition.Val.RequirePermanentCard {
		t.Fatalf("linked put = %#v, condition = %#v", put, condition)
	}
	if err := game.ValidateInstructionSequence(mode.Sequence, mode.Targets); err != nil {
		t.Fatalf("instruction sequence validation: %v", err)
	}
}

func TestLowerShuffleRevealPermanentSequenceFailsClosed(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"The owner of target permanent shuffles it into their library, then reveals the bottom card of their library. If it's a permanent card, they put it onto the battlefield.",
		"The controller of target permanent shuffles it into their library, then reveals the top card of their library. If it's a permanent card, they put it onto the battlefield.",
		"The owner of target permanent shuffles it into their library, then reveals the top card of their library. They put it onto the battlefield.",
		"The owner of target permanent shuffles it into their library, then reveals the top card of their library. If it's a creature card, they put it onto the battlefield.",
		"The owner of target permanent shuffles it into their library, then reveals the top card of their library. If it's a permanent card, they may put it onto the battlefield.",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
				Name:       "Test Warp",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: oracleText,
			})
			if face.SpellAbility.Exists {
				t.Fatal("near-miss sequence produced a partial spell ability")
			}
		})
	}
}

func TestGenerateChaosWarpEndToEnd(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Chaos Warp",
		Layout:     "normal",
		ManaCost:   "{2}{R}",
		TypeLine:   "Instant",
		OracleText: shuffleRevealPermanentText,
		Colors:     []string{"R"},
	}, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.ShufflePermanentIntoLibrary",
		"game.Reveal",
		"PublishLinked: game.LinkedKey(",
		"game.LinkedBattlefieldSource",
		"game.ObjectOwnerReference(game.TargetPermanentReference(0))",
		"CardCondition: opt.Val(game.CardCondition",
		"RequirePermanentCard: true",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
