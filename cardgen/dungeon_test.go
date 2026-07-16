package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// singleTriggeredPrimitive returns the single primitive of a face's only
// triggered ability, failing if the face is not exactly one triggered ability of
// one instruction.
func singleTriggeredPrimitive(t *testing.T, face loweredFaceAbilities) game.Primitive {
	t.Helper()
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	content := face.TriggeredAbilities[0].Content
	if len(content.Modes) != 1 || len(content.Modes[0].Sequence) != 1 {
		t.Fatalf("content = %#v, want one instruction", content)
	}
	return content.Modes[0].Sequence[0].Primitive
}

// TestVentureIntoDungeonLowers proves "venture into the dungeon." lowers to the
// controller-scoped VentureIntoDungeon primitive from an enters trigger, and the
// parenthetical reminder text is ignored.
func TestVentureIntoDungeonLowers(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Veteran Dungeoneer",
		Layout:     "normal",
		TypeLine:   "Creature — Human Warrior",
		OracleText: "When this creature enters, venture into the dungeon. (Enter the first room or advance to the next room.)",
	}
	prim, ok := singleTriggeredPrimitive(t, lowerSingleFace(t, card)).(game.VentureIntoDungeon)
	if !ok {
		t.Fatalf("primitive = %#v, want game.VentureIntoDungeon", prim)
	}
	if prim.Player.Kind() != game.PlayerReferenceController {
		t.Fatalf("player = %v, want controller", prim.Player.Kind())
	}
}

// TestTakeInitiativeLowers proves "you take the initiative." lowers to the
// controller-scoped TakeInitiative primitive from an enters trigger.
func TestTakeInitiativeLowers(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Goliath Paladin",
		Layout:     "normal",
		TypeLine:   "Creature — Giant Knight",
		OracleText: "Vigilance\nWhen this creature enters, you take the initiative.",
	}
	prim, ok := singleTriggeredPrimitive(t, lowerSingleFace(t, card)).(game.TakeInitiative)
	if !ok {
		t.Fatalf("primitive = %#v, want game.TakeInitiative", prim)
	}
	if prim.Player.Kind() != game.PlayerReferenceController {
		t.Fatalf("player = %v, want controller", prim.Player.Kind())
	}
}

// TestVentureIntoUndercityLowers proves "venture into Undercity." lowers to the
// controller-scoped VentureIntoUndercity primitive from a sorcery.
func TestVentureIntoUndercityLowers(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Undercity Delve",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Venture into Undercity.",
	}
	face := lowerSingleFace(t, card)
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability was not lowered")
	}
	content := face.SpellAbility.Val
	if len(content.Modes) != 1 || len(content.Modes[0].Sequence) != 1 {
		t.Fatalf("content = %#v, want one instruction", content)
	}
	prim, ok := content.Modes[0].Sequence[0].Primitive.(game.VentureIntoUndercity)
	if !ok {
		t.Fatalf("primitive = %#v, want game.VentureIntoUndercity", content.Modes[0].Sequence[0].Primitive)
	}
	if prim.Player.Kind() != game.PlayerReferenceController {
		t.Fatalf("player = %v, want controller", prim.Player.Kind())
	}
}

// TestVentureIntoDungeonGeneratesSource proves the full generate pipeline
// (recognize -> lower -> validate -> render) emits a compilable VentureIntoDungeon
// primitive for a real venture card.
func TestVentureIntoDungeonGeneratesSource(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Veteran Dungeoneer",
		Layout:     "normal",
		TypeLine:   "Creature — Human Warrior",
		ManaCost:   "{2}{W}",
		OracleText: "When this creature enters, venture into the dungeon. (Enter the first room or advance to the next room.)",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "v")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %+v", diagnostics)
	}
	if !contains(source, "game.VentureIntoDungeon{") || !contains(source, "Player: game.ControllerReference(),") {
		t.Fatalf("generated source missing venture primitive:\n%s", source)
	}
}

// TestTakeInitiativeGeneratesSource proves the full generate pipeline emits a
// compilable TakeInitiative primitive for a real initiative card.
func TestTakeInitiativeGeneratesSource(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Goliath Paladin",
		Layout:     "normal",
		TypeLine:   "Creature — Giant Knight",
		ManaCost:   "{4}{W}{W}",
		OracleText: "Vigilance\nWhen this creature enters, you take the initiative.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "g")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %+v", diagnostics)
	}
	if !contains(source, "game.TakeInitiative{") {
		t.Fatalf("generated source missing take-initiative primitive:\n%s", source)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})()
}

// TestCompletedDungeonConditionGatesSecondDraw proves the "Draw another card if
// you've completed a dungeon." resolution gate (Imoen, Mystic Trickster) lowers
// as a per-effect ControllerCompletedADungeon condition on the second draw, while
// the intervening "if you have the initiative" gates the trigger.
func TestCompletedDungeonConditionGatesSecondDraw(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Imoen Fragment",
		Layout:     "normal",
		TypeLine:   "Creature — Human Rogue Wizard",
		OracleText: "At the beginning of your end step, if you have the initiative, draw a card. Draw another card if you've completed a dungeon.",
	}
	face := lowerSingleFace(t, card)
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	seq := face.TriggeredAbilities[0].Content.Modes[0].Sequence
	if len(seq) != 2 {
		t.Fatalf("body instructions = %d, want 2", len(seq))
	}
	if _, ok := seq[0].Primitive.(game.Draw); !ok {
		t.Fatalf("first primitive = %#v, want game.Draw", seq[0].Primitive)
	}
	second := seq[1]
	if _, ok := second.Primitive.(game.Draw); !ok {
		t.Fatalf("second primitive = %#v, want game.Draw", second.Primitive)
	}
	if !second.Condition.Exists || !second.Condition.Val.Condition.Exists ||
		!second.Condition.Val.Condition.Val.ControllerCompletedADungeon {
		t.Fatalf("second draw condition = %#v, want ControllerCompletedADungeon", second.Condition)
	}
}
