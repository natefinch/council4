package cardgen

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

const blackMarketConnectionsOracle = "At the beginning of your first main phase, choose one or more —\n" +
	"• Sell Contraband — Create a Treasure token. You lose 1 life.\n" +
	"• Buy Information — Draw a card. You lose 2 life.\n" +
	"• Hire a Mercenary — Create a 3/2 colorless Shapeshifter creature token with changeling. You lose 3 life."

func TestLowerBlackMarketConnectionsModalTrigger(t *testing.T) {
	t.Parallel()

	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Broker",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: blackMarketConnectionsOracle,
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %#v, want one", face.TriggeredAbilities)
	}

	ability := face.TriggeredAbilities[0]
	if ability.Trigger.Pattern.Event != game.EventBeginningOfStep ||
		ability.Trigger.Pattern.Step != game.StepPrecombatMain ||
		ability.Trigger.Pattern.Controller != game.TriggerControllerYou {
		t.Fatalf("trigger = %#v, want controller's first main phase", ability.Trigger.Pattern)
	}
	if ability.Content.MinModes != 1 || ability.Content.MaxModes != 3 ||
		ability.Content.AllowDuplicateModes || len(ability.Content.Modes) != 3 {
		t.Fatalf("modal content = %#v, want one to three distinct modes", ability.Content)
	}

	assertCreateTokenThenLoseLife(t, ability.Content.Modes[0], types.Treasure, 1, nil)
	if len(ability.Content.Modes[1].Sequence) != 2 {
		t.Fatalf("information sequence = %#v, want draw then life loss", ability.Content.Modes[1].Sequence)
	}
	if draw, ok := ability.Content.Modes[1].Sequence[0].Primitive.(game.Draw); !ok || draw.Amount.Value() != 1 {
		t.Fatalf("information first primitive = %#v, want draw one", ability.Content.Modes[1].Sequence[0].Primitive)
	}
	assertLoseLife(t, &ability.Content.Modes[1].Sequence[1], 2)
	assertCreateTokenThenLoseLife(t, ability.Content.Modes[2], types.Shapeshifter, 3, func(t *testing.T, def *game.CardDef) {
		t.Helper()
		if !slices.Equal(def.Types, []types.Card{types.Creature}) ||
			!slices.Equal(def.Subtypes, []types.Sub{types.Shapeshifter}) ||
			!def.Power.Exists || def.Power.Val.Value != 3 ||
			!def.Toughness.Exists || def.Toughness.Val.Value != 2 ||
			!def.HasKeyword(game.Changeling) {
			t.Fatalf("mercenary token = %#v, want 3/2 colorless Shapeshifter with changeling", def)
		}
	})
}

func TestBlackMarketConnectionsUnsupportedLabelFailsClosed(t *testing.T) {
	t.Parallel()

	oracle := strings.Replace(blackMarketConnectionsOracle, "Sell Contraband", "Move Merchandise", 1)
	assertExecutableCardUnsupported(t, oracle)
}

func TestBlackMarketConnectionsUnsupportedModeBodyFailsClosed(t *testing.T) {
	t.Parallel()

	oracle := strings.Replace(blackMarketConnectionsOracle, "You lose 2 life.", "You gain 2 life.", 1)
	assertExecutableCardUnsupported(t, oracle)
}

func TestBlackMarketConnectionsUnsupportedChooseCountFailsClosed(t *testing.T) {
	t.Parallel()

	oracle := strings.Replace(blackMarketConnectionsOracle, "choose one or more", "choose two", 1)
	assertExecutableCardUnsupported(t, oracle)
}

func TestBlackMarketConnectionsUnsupportedTriggerTimingFailsClosed(t *testing.T) {
	t.Parallel()

	oracle := strings.Replace(blackMarketConnectionsOracle, "your first main phase", "your upkeep", 1)
	assertExecutableCardUnsupported(t, oracle)
}

func TestBlackMarketConnectionsExtraTokenAbilityFailsClosed(t *testing.T) {
	t.Parallel()

	oracle := strings.Replace(blackMarketConnectionsOracle, "with changeling.", "with changeling and flying.", 1)
	assertExecutableCardUnsupported(t, oracle)
}

func TestBlackMarketConnectionsReorderedLabelsFailClosed(t *testing.T) {
	t.Parallel()

	oracle := strings.Replace(blackMarketConnectionsOracle, "Sell Contraband", "Temporary Label", 1)
	oracle = strings.Replace(oracle, "Buy Information", "Sell Contraband", 1)
	oracle = strings.Replace(oracle, "Temporary Label", "Buy Information", 1)
	assertExecutableCardUnsupported(t, oracle)
}

func TestBlackMarketConnectionsMissingLabelsFailClosed(t *testing.T) {
	t.Parallel()

	oracle := blackMarketConnectionsOracle
	for _, label := range []string{"Sell Contraband — ", "Buy Information — ", "Hire a Mercenary — "} {
		oracle = strings.Replace(oracle, label, "", 1)
	}
	assertExecutableCardUnsupported(t, oracle)
}

// TestUnlabeledChooseOneOrMoreModalIsSupported confirms a generic unlabeled
// "Choose one or more" spell lowers when every mode body lowers on its own,
// rather than being restricted to the labeled connection vocabulary. The modal
// range carries MinModes 1 and MaxModes equal to the number of modes.
func TestUnlabeledChooseOneOrMoreModalIsSupported(t *testing.T) {
	t.Parallel()

	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Test Broker",
		Layout:   "normal",
		TypeLine: "Sorcery",
		OracleText: "Choose one or more —\n" +
			"• Destroy target artifact.\n" +
			"• Destroy target enchantment.\n" +
			"• Destroy target land.",
	})
	content := face.SpellAbility.Val
	if !content.IsModal() {
		t.Fatalf("content not modal: %#v", content)
	}
	if content.MinModes != 1 || content.MaxModes != 3 || len(content.Modes) != 3 {
		t.Fatalf("modal range = min %d max %d modes %d, want 1/3/3",
			content.MinModes, content.MaxModes, len(content.Modes))
	}
}

func TestBlackMarketConnectionsExecutableSourcePreservesModeLabels(t *testing.T) {
	t.Parallel()

	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Broker",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: blackMarketConnectionsOracle,
	}, "t")
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("generate: err=%v diagnostics=%#v", err, diagnostics)
	}
	for _, label := range []string{"Sell Contraband", "Buy Information", "Hire a Mercenary"} {
		if !strings.Contains(source, `Text: "`+label) {
			t.Fatalf("generated source does not preserve mode label %q:\n%s", label, source)
		}
	}
}

func assertExecutableCardUnsupported(t *testing.T, oracle string) {
	t.Helper()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Broker",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: oracle,
	}, "t")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if source != "" || len(diagnostics) == 0 {
		t.Fatalf("source = %q diagnostics = %#v, want fail-closed diagnostics and no source", source, diagnostics)
	}
}

func assertCreateTokenThenLoseLife(
	t *testing.T,
	mode game.Mode,
	subtype types.Sub,
	life int,
	assertDef func(*testing.T, *game.CardDef),
) {
	t.Helper()
	if len(mode.Sequence) != 2 {
		t.Fatalf("mode sequence = %#v, want token creation then life loss", mode.Sequence)
	}
	create, ok := mode.Sequence[0].Primitive.(game.CreateToken)
	if !ok || create.Amount.Value() != 1 {
		t.Fatalf("first primitive = %#v, want create one token", mode.Sequence[0].Primitive)
	}
	def, ok := create.Source.TokenDefRef()
	if !ok || !def.HasSubtype(subtype) {
		t.Fatalf("token definition = %#v, want subtype %s", def, subtype)
	}
	if assertDef != nil {
		assertDef(t, def)
	}
	assertLoseLife(t, &mode.Sequence[1], life)
}

func assertLoseLife(t *testing.T, instruction *game.Instruction, amount int) {
	t.Helper()
	lose, ok := instruction.Primitive.(game.LoseLife)
	if !ok || lose.Amount.Value() != amount || lose.Player != game.ControllerReference() {
		t.Fatalf("primitive = %#v, want controller lose %d life", instruction.Primitive, amount)
	}
}
