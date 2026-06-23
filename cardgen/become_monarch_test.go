package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// becomeMonarchPrimitive returns the single BecomeMonarch primitive of an
// ability body, failing the test if the body is not exactly that primitive.
func becomeMonarchPrimitive(t *testing.T, content game.AbilityContent) game.BecomeMonarch {
	t.Helper()
	if len(content.Modes) != 1 || len(content.Modes[0].Sequence) != 1 {
		t.Fatalf("content = %#v, want one instruction", content)
	}
	prim, ok := content.Modes[0].Sequence[0].Primitive.(game.BecomeMonarch)
	if !ok {
		t.Fatalf("primitive = %#v, want game.BecomeMonarch", content.Modes[0].Sequence[0].Primitive)
	}
	return prim
}

// TestGenerateThornOfTheBlackRoseControllerMonarch proves the controller form
// "you become the monarch" lowers to a BecomeMonarch primitive targeting the
// resolving controller from an enters trigger.
func TestGenerateThornOfTheBlackRoseControllerMonarch(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Thorn of the Black Rose",
		Layout:     "normal",
		TypeLine:   "Creature — Human Assassin",
		OracleText: "Deathtouch\nWhen this creature enters, you become the monarch.",
	}
	face := lowerSingleFace(t, card)
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	prim := becomeMonarchPrimitive(t, face.TriggeredAbilities[0].Content)
	if prim.Player.Kind() != game.PlayerReferenceController {
		t.Fatalf("player reference = %v, want controller", prim.Player.Kind())
	}
}

// TestGenerateTargetPlayerBecomesMonarch proves the single player-target form
// "target player becomes the monarch" lowers to a BecomeMonarch primitive that
// reads the spell's player target.
func TestGenerateTargetPlayerBecomesMonarch(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Crown Bestowal",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Target player becomes the monarch.",
	}
	face := lowerSingleFace(t, card)
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability was not lowered")
	}
	content := face.SpellAbility.Val
	prim := becomeMonarchPrimitive(t, content)
	if prim.Player.Kind() != game.PlayerReferenceTargetPlayer {
		t.Fatalf("player reference = %v, want target player", prim.Player.Kind())
	}
	targets := content.Modes[0].Targets
	if len(targets) != 1 || targets[0].Allow != game.TargetAllowPlayer {
		t.Fatalf("targets = %#v, want one player target", targets)
	}
}

// TestGenerateOathOfEorlChapterMonarch proves the Oath of Eorl target card
// generates and its chapter III lowers the "You become the monarch." sub-effect
// to a BecomeMonarch primitive after the indestructible-counter placement.
func TestGenerateOathOfEorlChapterMonarch(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Oath of Eorl",
		Layout:   "normal",
		TypeLine: "Enchantment — Saga",
		OracleText: "(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)\n" +
			"I — Create two 1/1 white Human Soldier creature tokens.\n" +
			"II — Create two 2/2 red Human Knight creature tokens with trample and haste.\n" +
			"III — Put an indestructible counter on up to one target Human. You become the monarch.",
	}
	face := lowerSingleFace(t, card)
	if len(face.ChapterAbilities) != 3 {
		t.Fatalf("chapter abilities = %d, want 3", len(face.ChapterAbilities))
	}
	chapterIII := face.ChapterAbilities[2].Content
	if len(chapterIII.Modes) != 1 || len(chapterIII.Modes[0].Sequence) != 2 {
		t.Fatalf("chapter III = %#v, want two instructions", chapterIII)
	}
	prim, ok := chapterIII.Modes[0].Sequence[1].Primitive.(game.BecomeMonarch)
	if !ok {
		t.Fatalf("chapter III second instruction = %#v, want BecomeMonarch", chapterIII.Modes[0].Sequence[1].Primitive)
	}
	if prim.Player.Kind() != game.PlayerReferenceController {
		t.Fatalf("player reference = %v, want controller", prim.Player.Kind())
	}
}
