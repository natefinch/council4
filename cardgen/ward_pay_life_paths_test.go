package cardgen

import (
	"strings"
	"testing"
)

// Ward—Pay N life followed by parenthetical reminder text ("Ward—Pay 3 life.
// (Whenever this creature becomes the target ...)"): the reminder must not turn
// the recognized Ward keyword into an "unsupported mixed keyword ability". The
// keyword lowers to a pay-life Ward static ability and the reminder is dropped.
func TestGenerateWardPayLifeWithReminderText(t *testing.T) {
	t.Parallel()
	power, toughness := "3", "2"
	card := &ScryfallCard{
		Name:      "Test Ward Pay Life Reminder",
		Layout:    "normal",
		ManaCost:  "{1}{B}{B}",
		TypeLine:  "Creature — Human Warlock",
		Power:     &power,
		Toughness: &toughness,
		OracleText: "Ward—Pay 3 life. (Whenever this creature becomes the target of a spell or " +
			"ability an opponent controls, counter it unless that player pays 3 life.)",
	}
	generatedSourceContains(t, card, []string{
		"game.WardStaticAbilityWithCosts(cost.Mana{}, []cost.Additional{",
		"cost.AdditionalPayLife",
	})
}

// A static grant of a parameterized Ward keyword body ("Other creatures you
// control have \"Ward—Pay 2 life.\"", Hexing Squelcher): the quoted keyword
// grant lowers to a reusable continuous ability grant that carries the Ward
// static ability, not a card-specific special case.
func TestGenerateGrantedQuotedWardPayLife(t *testing.T) {
	t.Parallel()
	power, toughness := "2", "2"
	card := &ScryfallCard{
		Name:       "Test Granted Quoted Ward",
		Layout:     "normal",
		ManaCost:   "{1}{R}",
		TypeLine:   "Creature — Goblin Sorcerer",
		Power:      &power,
		Toughness:  &toughness,
		OracleText: `Other creatures you control have "Ward—Pay 2 life."`,
	}
	generatedSourceContains(t, card, []string{
		"ContinuousEffects:",
		"game.ObjectControlledGroupExcluding(",
		"new(game.WardStaticAbilityWithCosts(cost.Mana{}, []cost.Additional{",
		"cost.AdditionalPayLife",
	})
}

// The reminder-text Ward form generates without diagnostics: a direct guard on
// the fail-closed path that previously rejected the card wholesale.
func TestGenerateWardPayLifeReminderHasNoDiagnostics(t *testing.T) {
	t.Parallel()
	power, toughness := "3", "3"
	card := &ScryfallCard{
		Name:      "Test Ward Reminder Clean",
		Layout:    "normal",
		ManaCost:  "{2}{G}",
		TypeLine:  "Creature — Beast",
		Power:     &power,
		Toughness: &toughness,
		OracleText: "Ward—Pay 2 life. (Whenever this creature becomes the target of a spell or " +
			"ability an opponent controls, counter it unless that player pays 2 life.)",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "cards")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if strings.Contains(source, "unsupported") {
		t.Fatalf("generated source unexpectedly mentions unsupported:\n%s", source)
	}
}
