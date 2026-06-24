package cardgen

import (
	"strings"
	"testing"
)

func TestGenerateExecutableCardSourceRadicalIdeaJumpStart(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Radical Idea",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{1}{U}",
		OracleText: "Draw a card.\nJump-start (You may cast this card from your graveyard by discarding a card in addition to paying its other costs. Then exile this card.)",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "r")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "game.JumpStartStaticBody") {
		t.Fatalf("generated source missing Jump-start keyword:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceBareJumpStartKeyword(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Bare Jump Spell",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{R}",
		OracleText: "Draw a card.\nJump-start",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "b")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "game.JumpStartStaticBody") {
		t.Fatalf("generated source missing Jump-start keyword:\n%s", source)
	}
}
