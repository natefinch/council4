package cardgen

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerAtTriggerYourUpkeepDrawCard(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "At the beginning of your upkeep, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	ta := face.TriggeredAbilities[0]
	if ta.Trigger.Type != game.TriggerAt {
		t.Fatalf("trigger type = %v, want TriggerAt", ta.Trigger.Type)
	}
	if ta.Trigger.Pattern.Event != game.EventBeginningOfStep {
		t.Fatalf("event = %v, want EventBeginningOfStep", ta.Trigger.Pattern.Event)
	}
	if ta.Trigger.Pattern.Step != game.StepUpkeep {
		t.Fatalf("step = %v, want StepUpkeep", ta.Trigger.Pattern.Step)
	}
	if ta.Trigger.Pattern.Controller != game.TriggerControllerYou {
		t.Fatalf("controller = %v, want TriggerControllerYou", ta.Trigger.Pattern.Controller)
	}
	draw, ok := ta.Content.Modes[0].Sequence[0].Primitive.(game.Draw)
	if !ok || draw.Amount != game.Fixed(1) {
		t.Fatalf("primitive = %+v, want Draw{Amount: Fixed(1)}", ta.Content.Modes[0].Sequence[0].Primitive)
	}
}

func TestLowerAtTriggerEachOpponentUpkeepDamage(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Pinger",
		Layout:     "normal",
		TypeLine:   "Creature — Goblin",
		OracleText: "At the beginning of each opponent's upkeep, draw a card.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	ta := face.TriggeredAbilities[0]
	if ta.Trigger.Type != game.TriggerAt {
		t.Fatalf("trigger type = %v, want TriggerAt", ta.Trigger.Type)
	}
	if ta.Trigger.Pattern.Event != game.EventBeginningOfStep {
		t.Fatalf("event = %v, want EventBeginningOfStep", ta.Trigger.Pattern.Event)
	}
	if ta.Trigger.Pattern.Step != game.StepUpkeep {
		t.Fatalf("step = %v, want StepUpkeep", ta.Trigger.Pattern.Step)
	}
	if ta.Trigger.Pattern.Controller != game.TriggerControllerOpponent {
		t.Fatalf("controller = %v, want TriggerControllerOpponent", ta.Trigger.Pattern.Controller)
	}
}

func TestLowerAtTriggerEachUpkeepAny(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Watcher",
		Layout:     "normal",
		TypeLine:   "Creature — Human",
		OracleText: "At the beginning of each upkeep, draw a card.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	ta := face.TriggeredAbilities[0]
	if ta.Trigger.Pattern.Controller != game.TriggerControllerAny {
		t.Fatalf("controller = %v, want TriggerControllerAny", ta.Trigger.Pattern.Controller)
	}
}

func TestLowerAtTriggerYourEndStep(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Mystic",
		Layout:     "normal",
		TypeLine:   "Creature — Human Wizard",
		OracleText: "At the beginning of your end step, draw a card.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	ta := face.TriggeredAbilities[0]
	if ta.Trigger.Pattern.Step != game.StepEnd {
		t.Fatalf("step = %v, want StepEnd", ta.Trigger.Pattern.Step)
	}
	if ta.Trigger.Pattern.Controller != game.TriggerControllerYou {
		t.Fatalf("controller = %v, want TriggerControllerYou", ta.Trigger.Pattern.Controller)
	}
}

func TestLowerAtTriggerBeginningOfCombatYourTurn(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Fighter",
		Layout:     "normal",
		TypeLine:   "Creature — Human Warrior",
		OracleText: "At the beginning of combat on your turn, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	ta := face.TriggeredAbilities[0]
	if ta.Trigger.Pattern.Step != game.StepBeginningOfCombat {
		t.Fatalf("step = %v, want StepBeginningOfCombat", ta.Trigger.Pattern.Step)
	}
	if ta.Trigger.Pattern.Controller != game.TriggerControllerYou {
		t.Fatalf("controller = %v, want TriggerControllerYou", ta.Trigger.Pattern.Controller)
	}
}

func TestLowerAtTriggerYourDrawStep(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Scholar",
		Layout:     "normal",
		TypeLine:   "Creature — Human Wizard",
		OracleText: "At the beginning of your draw step, draw a card.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	ta := face.TriggeredAbilities[0]
	if ta.Trigger.Pattern.Step != game.StepDraw {
		t.Fatalf("step = %v, want StepDraw", ta.Trigger.Pattern.Step)
	}
	if ta.Trigger.Pattern.Controller != game.TriggerControllerYou {
		t.Fatalf("controller = %v, want TriggerControllerYou", ta.Trigger.Pattern.Controller)
	}
}

func TestLowerAtTriggerEachCombat(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Battler",
		Layout:     "normal",
		TypeLine:   "Creature — Human Soldier",
		OracleText: "At the beginning of each combat, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	ta := face.TriggeredAbilities[0]
	if ta.Trigger.Pattern.Step != game.StepBeginningOfCombat {
		t.Fatalf("step = %v, want StepBeginningOfCombat", ta.Trigger.Pattern.Step)
	}
	if ta.Trigger.Pattern.Controller != game.TriggerControllerAny {
		t.Fatalf("controller = %v, want TriggerControllerAny", ta.Trigger.Pattern.Controller)
	}
}

func TestLowerAtTriggerOptional(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Sage",
		Layout:     "normal",
		TypeLine:   "Creature — Human Wizard",
		OracleText: "At the beginning of your upkeep, you may draw a card.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	ta := face.TriggeredAbilities[0]
	if !ta.Optional {
		t.Fatal("expected Optional = true for 'you may' trigger")
	}
	if ta.Trigger.Pattern.Step != game.StepUpkeep {
		t.Fatalf("step = %v, want StepUpkeep", ta.Trigger.Pattern.Step)
	}
}

func TestLowerAtTriggerMainPhasePhrases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		phrase string
		step   game.Step
	}{
		{"your first main phase", game.StepPrecombatMain},
		{"your precombat main phase", game.StepPrecombatMain},
		{"each of your first main phases", game.StepPrecombatMain},
		{"your second main phase", game.StepPostcombatMain},
		{"your postcombat main phase", game.StepPostcombatMain},
		{"each of your postcombat main phases", game.StepPostcombatMain},
	}
	for _, test := range tests {
		t.Run(test.phrase, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Planner",
				Layout:     "normal",
				TypeLine:   "Creature — Human",
				OracleText: "At the beginning of " + test.phrase + ", draw a card.",
				Power:      new("2"),
				Toughness:  new("2"),
			})
			trigger := face.TriggeredAbilities[0].Trigger
			if trigger.Pattern.Step != test.step || trigger.Pattern.Controller != game.TriggerControllerYou {
				t.Fatalf("trigger pattern = %+v, want step %v controlled by you", trigger.Pattern, test.step)
			}
		})
	}
}

func TestLowerAtTriggerEnchantedPlayerMainPhaseFailsClosed(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Aura",
		Layout:     "normal",
		TypeLine:   "Enchantment — Aura",
		OracleText: "Enchant player\nAt the beginning of each of enchanted player's postcombat main phases, draw a card.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("enchanted-player main-phase trigger unexpectedly lowered")
	}
	if !slices.ContainsFunc(diagnostics, func(d shared.Diagnostic) bool {
		return strings.Contains(d.Summary, "unsupported phase/step trigger phrase")
	}) {
		t.Fatalf("diagnostics = %#v, want unsupported phase/step trigger phrase", diagnostics)
	}
}

func TestLowerAtTriggerInterveningIfConditions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		condition string
		assert    func(*testing.T, game.Condition)
	}{
		{
			name:      "controls creature",
			condition: "if you control a creature",
			assert: func(t *testing.T, condition game.Condition) {
				t.Helper()
				controls := condition.ControlsMatching
				if !controls.Exists || !slices.Equal(controls.Val.Selection.RequiredTypes, []types.Card{types.Creature}) {
					t.Fatalf("condition = %+v, want controls a creature", condition)
				}
			},
		},
		{
			name:      "controller life",
			condition: "if you have 10 or more life",
			assert: func(t *testing.T, condition game.Condition) {
				t.Helper()
				if condition.ControllerLifeAtLeast != 10 {
					t.Fatalf("ControllerLifeAtLeast = %d, want 10", condition.ControllerLifeAtLeast)
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: "At the beginning of your upkeep, " + test.condition + ", draw a card.",
				Power:      new("2"),
				Toughness:  new("2"),
			})
			trigger := face.TriggeredAbilities[0].Trigger
			if trigger.InterveningIf != test.condition || !trigger.InterveningCondition.Exists {
				t.Fatalf("trigger = %+v, want %q intervening-if condition", trigger, test.condition)
			}
			test.assert(t, trigger.InterveningCondition.Val)
		})
	}
}

func TestLowerAtTriggerUnsupportedInterveningIfFailsClosed(t *testing.T) {
	t.Parallel()
	for _, condition := range []string{
		"if you gained 2 or more life this turn",
		"if this creature came under your control since the beginning of your last upkeep",
	} {
		t.Run(condition, func(t *testing.T) {
			t.Parallel()
			_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: "At the beginning of your upkeep, " + condition + ", draw a card.",
				Power:      new("2"),
				Toughness:  new("2"),
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) == 0 {
				t.Fatal("unsupported intervening-if condition unexpectedly lowered")
			}
			if !strings.Contains(diagnostics[0].Detail, "does not support this intervening-if condition") {
				t.Fatalf("diagnostics = %#v, want intervening-if diagnostic", diagnostics)
			}
		})
	}
}

func TestLowerAtTriggerPhraseVariants(t *testing.T) {
	t.Parallel()
	tests := []struct {
		phrase     string
		step       game.Step
		controller game.TriggerControllerFilter
	}{
		{"each upkeep", game.StepUpkeep, game.TriggerControllerAny},
		{"each player's upkeep", game.StepUpkeep, game.TriggerControllerAny},
		{"each opponent's upkeep", game.StepUpkeep, game.TriggerControllerOpponent},
		{"each end step", game.StepEnd, game.TriggerControllerAny},
		{"each player's end step", game.StepEnd, game.TriggerControllerAny},
		{"each combat", game.StepBeginningOfCombat, game.TriggerControllerAny},
	}
	for _, tc := range tests {
		t.Run(tc.phrase, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Card",
				Layout:     "normal",
				TypeLine:   "Creature — Human",
				OracleText: "At the beginning of " + tc.phrase + ", draw a card.",
				Power:      new("1"),
				Toughness:  new("1"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			ta := face.TriggeredAbilities[0]
			if ta.Trigger.Pattern.Step != tc.step {
				t.Errorf("step = %v, want %v", ta.Trigger.Pattern.Step, tc.step)
			}
			if ta.Trigger.Pattern.Controller != tc.controller {
				t.Errorf("controller = %v, want %v", ta.Trigger.Pattern.Controller, tc.controller)
			}
		})
	}
}
