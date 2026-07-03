package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourcePerCreatureUntapNongreen covers Dream Tides:
// "At the beginning of each player's upkeep, that player may choose any number
// of tapped nongreen creatures they control and pay {2} for each creature chosen
// this way. If the player does, untap those creatures." The offer lowers to an
// event-player PayRepeatedly whose published count sizes an Untap that lets the
// upkeep player choose that many of their tapped nongreen creatures.
func TestGenerateExecutableCardSourcePerCreatureUntapNongreen(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Dream Tides",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Creatures don't untap during their controllers' untap steps.\nAt the beginning of each player's upkeep, that player may choose any number of tapped nongreen creatures they control and pay {2} for each creature chosen this way. If the player does, untap those creatures.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "d")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.PayRepeatedly",
		"Payer:  opt.Val(game.EventPlayerReference())",
		"PublishCount: \"per-creature-untap-count\"",
		"Primitive: game.Untap",
		"ChooseUpTo: true,",
		"Chooser: game.EventPlayerReference(),",
		"Kind:      game.DynamicAmountChosenNumber,",
		"ExcludedColors: []color.Color{color.Green}",
		"Tapped: game.TriTrue",
		"game.PlayerControlledGroup(game.EventPlayerReference()",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutableCardSourcePerCreatureUntapColored covers Thelon's Curse:
// a colored per-creature cost ({U}) and a "tapped blue creatures" filter map to
// the same PayRepeatedly + Untap shape with a ColorsAny filter.
func TestGenerateExecutableCardSourcePerCreatureUntapColored(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Thelon's Curse",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Blue creatures don't untap during their controllers' untap steps.\nAt the beginning of each player's upkeep, that player may choose any number of tapped blue creatures they control and pay {U} for each creature chosen this way. If the player does, untap those creatures.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.PayRepeatedly",
		"cost.U,",
		"Primitive: game.Untap",
		"Chooser: game.EventPlayerReference(),",
		"ColorsAny: []color.Color{color.Blue}",
		"Tapped: game.TriTrue",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
