package cardgen

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestLowerEventHistoryInterveningConditions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracle     string
		wantEvent  game.EventKind
		wantNegate bool
		wantWindow game.EventHistoryWindow
	}{
		{
			name:       "attacked this turn",
			oracle:     "When this creature enters, if you attacked this turn, draw a card.",
			wantEvent:  game.EventAttackerDeclared,
			wantWindow: game.EventHistoryCurrentTurn,
		},
		{
			name:       "creature died this turn",
			oracle:     "At the beginning of your end step, if a creature died this turn, draw a card.",
			wantEvent:  game.EventPermanentDied,
			wantWindow: game.EventHistoryCurrentTurn,
		},
		{
			name:       "you gained life this turn",
			oracle:     "At the beginning of each end step, if you gained life this turn, draw a card.",
			wantEvent:  game.EventLifeGained,
			wantWindow: game.EventHistoryCurrentTurn,
		},
		{
			name:       "an opponent lost life this turn",
			oracle:     "At the beginning of your end step, if an opponent lost life this turn, draw a card.",
			wantEvent:  game.EventLifeLost,
			wantWindow: game.EventHistoryCurrentTurn,
		},
		{
			name:       "you lost life this turn",
			oracle:     "At the beginning of your end step, if you lost life this turn, draw a card.",
			wantEvent:  game.EventLifeLost,
			wantWindow: game.EventHistoryCurrentTurn,
		},
		{
			name:       "an opponent lost life last turn",
			oracle:     "At the beginning of each upkeep, if an opponent lost life last turn, draw a card.",
			wantEvent:  game.EventLifeLost,
			wantWindow: game.EventHistoryPreviousTurn,
		},
		{
			name:       "you lost life last turn",
			oracle:     "At the beginning of each upkeep, if you lost life last turn, draw a card.",
			wantEvent:  game.EventLifeLost,
			wantWindow: game.EventHistoryPreviousTurn,
		},
		{
			name:       "no spells cast last turn",
			oracle:     "At the beginning of your upkeep, if no spells were cast last turn, draw a card.",
			wantEvent:  game.EventSpellCast,
			wantNegate: true,
			wantWindow: game.EventHistoryPreviousTurn,
		},
		{
			name:       "you descended this turn",
			oracle:     "At the beginning of your end step, if you descended this turn, draw a card.",
			wantEvent:  game.EventZoneChanged,
			wantWindow: game.EventHistoryCurrentTurn,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: tc.oracle,
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			trigger := face.TriggeredAbilities[0].Trigger
			if trigger.InterveningIf == "" || !trigger.InterveningCondition.Exists {
				t.Fatalf("trigger = %+v, want intervening condition", trigger)
			}
			cond := trigger.InterveningCondition.Val
			if !cond.EventHistory.Exists {
				t.Fatalf("condition = %+v, want EventHistory", cond)
			}
			hist := cond.EventHistory.Val
			if hist.Pattern.Event != tc.wantEvent {
				t.Errorf("EventHistory.Pattern.Event = %v, want %v", hist.Pattern.Event, tc.wantEvent)
			}
			if hist.Window != tc.wantWindow {
				t.Errorf("EventHistory.Window = %v, want %v", hist.Window, tc.wantWindow)
			}
			if cond.Negate != tc.wantNegate {
				t.Errorf("Condition.Negate = %v, want %v", cond.Negate, tc.wantNegate)
			}
		})
	}
}

// TestLowerDescendInterveningCondition verifies the descend event-history
// intervening-if "if you descended this turn" lowers to a current-turn
// zone-change condition matching any nontoken permanent card put into the
// controller's graveyard from anywhere (CR 701.51).
func TestLowerDescendInterveningCondition(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Descender",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "At the beginning of your end step, if you descended this turn, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	cond := face.TriggeredAbilities[0].Trigger.InterveningCondition
	if !cond.Exists || !cond.Val.EventHistory.Exists {
		t.Fatalf("condition = %+v, want EventHistory", cond)
	}
	hist := cond.Val.EventHistory.Val
	if hist.Window != game.EventHistoryCurrentTurn {
		t.Errorf("Window = %v, want EventHistoryCurrentTurn", hist.Window)
	}
	if cond.Val.Negate {
		t.Error("Negate = true, want false")
	}
	want := game.TriggerPattern{
		Event:       game.EventZoneChanged,
		Player:      game.TriggerPlayerYou,
		MatchToZone: true,
		ToZone:      zone.Graveyard,
		SubjectSelection: game.Selection{
			RequiredTypesAny: []types.Card{
				types.Artifact,
				types.Battle,
				types.Creature,
				types.Enchantment,
				types.Land,
				types.Planeswalker,
			},
			NonToken: true,
		},
	}
	if !reflect.DeepEqual(hist.Pattern, want) {
		t.Errorf("EventHistory.Pattern = %+v, want %+v", hist.Pattern, want)
	}
}

// TestLowerEnteredBattlefieldInterveningCondition verifies the
// enters-the-battlefield event-history intervening-if "if <count> <Selection>
// entered the battlefield under your control this turn" lowers to a current-turn
// zone-change condition matching permanents of the selection entering the
// battlefield under the ability's controller, carrying the counted minimum and
// the self-excluding "another" qualifier.
func TestLowerEnteredBattlefieldInterveningCondition(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		oracle       string
		wantMinCount int
		wantPattern  game.TriggerPattern
	}{
		{
			name:         "two or more nonland permanents",
			oracle:       "At the beginning of your end step, if two or more nonland permanents entered the battlefield under your control this turn, draw a card.",
			wantMinCount: 2,
			wantPattern: game.TriggerPattern{
				Event:            game.EventPermanentEnteredBattlefield,
				Controller:       game.TriggerControllerYou,
				SubjectSelection: game.Selection{ExcludedTypes: []types.Card{types.Land}},
			},
		},
		{
			name:   "another creature excludes self",
			oracle: "At the beginning of your end step, if another creature entered the battlefield under your control this turn, draw a card.",
			wantPattern: game.TriggerPattern{
				Event:            game.EventPermanentEnteredBattlefield,
				Controller:       game.TriggerControllerYou,
				ExcludeSelf:      true,
				SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
			},
		},
		{
			name:   "a planeswalker",
			oracle: "At the beginning of each end step, if a planeswalker entered the battlefield under your control this turn, draw a card.",
			wantPattern: game.TriggerPattern{
				Event:            game.EventPermanentEnteredBattlefield,
				Controller:       game.TriggerControllerYou,
				SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Planeswalker}},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: tc.oracle,
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			cond := face.TriggeredAbilities[0].Trigger.InterveningCondition
			if !cond.Exists || !cond.Val.EventHistory.Exists {
				t.Fatalf("condition = %+v, want EventHistory", cond)
			}
			hist := cond.Val.EventHistory.Val
			if hist.Window != game.EventHistoryCurrentTurn {
				t.Errorf("Window = %v, want EventHistoryCurrentTurn", hist.Window)
			}
			if cond.Val.Negate {
				t.Error("Negate = true, want false")
			}
			if hist.MinCount != tc.wantMinCount {
				t.Errorf("MinCount = %d, want %d", hist.MinCount, tc.wantMinCount)
			}
			if !reflect.DeepEqual(hist.Pattern, tc.wantPattern) {
				t.Errorf("EventHistory.Pattern = %+v, want %+v", hist.Pattern, tc.wantPattern)
			}
		})
	}
}

// TestLowerLandfallDistinctNamesCondition verifies Field of the Dead's landfall
// trigger with the "if you control seven or more lands with different names"
// intervening-if lowers to a ControlsMatching condition carrying the
// distinct-name threshold.
func TestLowerLandfallDistinctNamesCondition(t *testing.T) {
	t.Parallel()
	power := "0"
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Field of the Dead Test",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "Whenever a land you control enters, if you control seven or more lands with different names, create a 2/2 black Zombie creature token.",
		Power:      &power,
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	cond := face.TriggeredAbilities[0].Trigger.InterveningCondition
	if !cond.Exists || !cond.Val.ControlsMatching.Exists {
		t.Fatalf("trigger condition = %+v, want ControlsMatching", cond)
	}
	count := cond.Val.ControlsMatching.Val
	if !slices.Contains(count.Selection.RequiredTypes, types.Land) {
		t.Fatalf("selection = %+v, want land type", count.Selection)
	}
	if !count.DistinctNames.Exists ||
		count.DistinctNames.Val.Op != compare.GreaterOrEqual ||
		count.DistinctNames.Val.Value != 7 {
		t.Fatalf("DistinctNames = %+v, want >= 7", count.DistinctNames)
	}
}

func TestLowerControlsCommanderInterveningCondition(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Loyal Apprentice Test",
		Layout:     "normal",
		TypeLine:   "Creature — Human Soldier",
		OracleText: "Lieutenant — At the beginning of combat on your turn, if you control your commander, create a 1/1 colorless Thopter artifact creature token with flying.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	trigger := face.TriggeredAbilities[0].Trigger
	if trigger.InterveningIf == "" || !trigger.InterveningCondition.Exists {
		t.Fatalf("trigger = %+v, want intervening condition", trigger)
	}
	cond := trigger.InterveningCondition.Val
	if !cond.ControllerControlsCommander {
		t.Fatalf("condition = %+v, want ControllerControlsCommander", cond)
	}
}

// TestLowerOpponentLandfallControlComparison proves Archaeomancer's Map's second
// ability — "Whenever a land an opponent controls enters, if that player
// controls more lands than you, ..." — lowers to a triggered ability whose
// intervening condition is a triggering-player control-count comparison.
func TestLowerOpponentLandfallControlComparison(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Archaeomancer's Map Test",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "Whenever a land an opponent controls enters, if that player controls more lands than you, you may put a land card from your hand onto the battlefield.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	cond := face.TriggeredAbilities[0].Trigger.InterveningCondition
	if !cond.Exists || !cond.Val.ControlComparison.Exists {
		t.Fatalf("trigger condition = %+v, want ControlComparison", cond)
	}
	cmp := cond.Val.ControlComparison.Val
	if cmp.Left != game.ControlPlayerTriggeringPlayer {
		t.Fatalf("comparison left = %v, want ControlPlayerTriggeringPlayer", cmp.Left)
	}
	if cmp.Right != game.ControlPlayerController {
		t.Fatalf("comparison right = %v, want ControlPlayerController", cmp.Right)
	}
	if cmp.Op != compare.GreaterThan {
		t.Fatalf("comparison op = %v, want GreaterThan", cmp.Op)
	}
	if !slices.Contains(cmp.Selection.RequiredTypes, types.Land) {
		t.Fatalf("comparison selection = %+v, want land type", cmp.Selection)
	}
}

func TestLowerEventHistoryInterveningConditionFailsClosed(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"At the beginning of your upkeep, if no creatures attacked last turn, draw a card.",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: oracleText,
				Power:      new("2"),
				Toughness:  new("2"),
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) == 0 {
				t.Fatalf("expected diagnostic for unsupported event history condition %q", oracleText)
			}
		})
	}
}

func TestRenderGeneratedEventHistoryInterveningCondition(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "At the beginning of your upkeep, if no spells were cast last turn, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Negate: true",
		"EventHistory: opt.Val(game.EventHistoryCondition{",
		"Event: game.EventSpellCast",
		"Window: game.EventHistoryPreviousTurn",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}

func TestLowerEventHistoryCreatureDiedHasCreatureFilter(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "At the beginning of your end step, if a creature died this turn, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	trigger := face.TriggeredAbilities[0].Trigger
	if !trigger.InterveningCondition.Exists {
		t.Fatal("no intervening condition")
	}
	hist := trigger.InterveningCondition.Val.EventHistory
	if !hist.Exists {
		t.Fatal("no EventHistory")
	}
	if !slices.Equal(hist.Val.Pattern.SubjectSelection.RequiredTypes, []types.Card{types.Creature}) {
		t.Fatalf("SubjectSelection = %#v, want creature", hist.Val.Pattern.SubjectSelection)
	}
}

func TestLowerEventHistoryAttackedConditionHasControllerYou(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature enters, if you attacked this turn, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	trigger := face.TriggeredAbilities[0].Trigger
	if !trigger.InterveningCondition.Exists {
		t.Fatal("no intervening condition")
	}

	hist := trigger.InterveningCondition.Val.EventHistory
	if !hist.Exists {
		t.Fatal("no EventHistory")
	}
	if hist.Val.Pattern.Controller != game.TriggerControllerYou {
		t.Fatalf("Controller = %v, want TriggerControllerYou", hist.Val.Pattern.Controller)
	}
}

func TestLowerObjectInterveningConditionsUseSharedReferencesAndSelections(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		typeLine    string
		oracle      string
		wantRef     game.ObjectReference
		wantTypes   []types.Card
		wantSubtype types.Sub
		wantTapped  game.TriState
		wantMatches bool
	}{
		{
			name:        "event creature LKI",
			typeLine:    "Creature — Spirit",
			oracle:      "When this creature dies, if it was a creature, draw a card.",
			wantRef:     game.EventPermanentReference(),
			wantTypes:   []types.Card{types.Creature},
			wantMatches: true,
		},
		{
			name:        "event creature contraction",
			typeLine:    "Creature — Spirit",
			oracle:      "When this creature dies, if it's a creature, draw a card.",
			wantRef:     game.EventPermanentReference(),
			wantTypes:   []types.Card{types.Creature},
			wantMatches: true,
		},
		{
			name:        "event Human LKI",
			typeLine:    "Creature — Human",
			oracle:      "When this creature dies, if it was a Human, draw a card.",
			wantRef:     game.EventPermanentReference(),
			wantTypes:   nil,
			wantSubtype: types.Human,
			wantMatches: true,
		},
		{
			name:        "source artifact untapped",
			typeLine:    "Artifact",
			oracle:      "At the beginning of your upkeep, if this artifact is untapped, draw a card.",
			wantRef:     game.SourcePermanentReference(),
			wantTypes:   []types.Card{types.Artifact},
			wantTapped:  game.TriFalse,
			wantMatches: true,
		},
		{
			name:        "source creature untapped",
			typeLine:    "Creature — Bear",
			oracle:      "At the beginning of your upkeep, if this creature is untapped, draw a card.",
			wantRef:     game.SourcePermanentReference(),
			wantTypes:   []types.Card{types.Creature},
			wantTapped:  game.TriFalse,
			wantMatches: true,
		},
		{
			name:        "source enchantment",
			typeLine:    "Enchantment",
			oracle:      "At the beginning of your upkeep, if this permanent is an enchantment, draw a card.",
			wantRef:     game.SourcePermanentReference(),
			wantTypes:   []types.Card{types.Enchantment},
			wantMatches: true,
		},
		{
			name:     "source battlefield existence",
			typeLine: "Creature — Bear",
			oracle:   "At the beginning of your upkeep, if this creature is on the battlefield, draw a card.",
			wantRef:  game.SourcePermanentReference(),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Subject",
				Layout:     "normal",
				TypeLine:   test.typeLine,
				OracleText: test.oracle,
			}
			if strings.Contains(test.typeLine, "Creature") {
				card.Power = new("2")
				card.Toughness = new("2")
			}
			face := lowerSingleFace(t, card)
			trigger := face.TriggeredAbilities[0].Trigger
			if !trigger.InterveningCondition.Exists {
				t.Fatalf("trigger = %+v, want structured condition", trigger)
			}
			condition := trigger.InterveningCondition.Val
			if !condition.Object.Exists || condition.Object.Val != test.wantRef {
				t.Fatalf("condition = %+v, want object %v", condition, test.wantRef)
			}
			if condition.ObjectMatches.Exists != test.wantMatches {
				t.Fatalf("condition = %+v, ObjectMatches.Exists = %v", condition, condition.ObjectMatches.Exists)
			}
			if !test.wantMatches {
				return
			}
			selection := condition.ObjectMatches.Val
			if !slices.Equal(selection.RequiredTypes, test.wantTypes) ||
				selection.Tapped != test.wantTapped {
				t.Fatalf("selection = %+v", selection)
			}
			if test.wantSubtype != "" &&
				!slices.Equal(selection.SubtypesAny, []types.Sub{test.wantSubtype}) {
				t.Fatalf("selection = %+v, want subtype %v", selection, test.wantSubtype)
			}
		})
	}
}

func TestLowerHadCountersInterveningConditionUsesSharedTriggerSlot(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Artificer",
		Layout:     "normal",
		TypeLine:   "Creature — Artificer",
		OracleText: "Whenever an artifact is put into a graveyard from the battlefield, if it had counters on it, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	trigger := face.TriggeredAbilities[0].Trigger
	if trigger.InterveningIf != "if it had counters on it" ||
		!trigger.InterveningIfEventPermanentHadCounters {
		t.Fatalf("trigger = %+v, want had-counters intervening-if", trigger)
	}
}

func TestRenderGuardianProjectNameUniqueInterveningCondition(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Guardian Project",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Whenever a nontoken creature you control enters, if it doesn't have the same name as another creature you control or a creature card in your graveyard, draw a card.",
	}, "g")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "EventPermanentNameUniqueAmongControlledAndGraveyardCreatures: true") {
		t.Fatalf("source missing name-unique condition field:\n%s", source)
	}
}

func TestLowerProvenControllerSelectionConditions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		phrase        string
		threshold     int
		negated       bool
		requiredTypes []types.Card
		subtypes      []types.Sub
		tapped        game.TriState
		power         int
		excludeSource bool
	}{
		{"if you control two or more Gates", 2, false, nil, []types.Sub{types.Gate}, game.TriAny, 0, false},
		{"if you control two or more tapped creatures", 2, false, []types.Card{types.Creature}, nil, game.TriTrue, 0, false},
		{"if you control a creature with power 5 or greater", 0, false, []types.Card{types.Creature}, nil, game.TriAny, 5, false},
		{"if you control another creature with power 4 or greater", 0, false, []types.Card{types.Creature}, nil, game.TriAny, 4, true},
		{"if you control an Equipment", 0, false, nil, []types.Sub{types.Equipment}, game.TriAny, 0, false},
		{"if you control no creatures", 1, true, []types.Card{types.Creature}, nil, game.TriAny, 0, false},
		{"if you control three or more creatures", 3, false, []types.Card{types.Creature}, nil, game.TriAny, 0, false},
		{"if you control a tapped creature", 0, false, []types.Card{types.Creature}, nil, game.TriTrue, 0, false},
	}
	for _, test := range tests {
		t.Run(test.phrase, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Watcher",
				Layout:     "normal",
				TypeLine:   "Creature — Human",
				OracleText: "At the beginning of your upkeep, " + test.phrase + ", draw a card.",
				Power:      new("2"),
				Toughness:  new("2"),
			})
			condition := face.TriggeredAbilities[0].Trigger.InterveningCondition
			if !condition.Exists || !condition.Val.ControlsMatching.Exists {
				t.Fatalf("condition = %+v, want ControlsMatching", condition)
			}
			count := condition.Val.ControlsMatching.Val
			selection := count.Selection
			if count.MinCount != test.threshold ||
				condition.Val.Negate != test.negated ||
				!slices.Equal(selection.RequiredTypes, test.requiredTypes) ||
				!slices.Equal(selection.SubtypesAny, test.subtypes) ||
				selection.Tapped != test.tapped ||
				selection.ExcludeSource != test.excludeSource {
				t.Fatalf("condition = %+v", condition.Val)
			}
			if test.power == 0 {
				if selection.Power.Exists {
					t.Fatalf("selection = %+v, want no power predicate", selection)
				}
			} else if !selection.Power.Exists ||
				selection.Power.Val.Op != compare.GreaterOrEqual ||
				selection.Power.Val.Value != test.power {
				t.Fatalf("selection = %+v, want power >= %d", selection, test.power)
			}
		})
	}
}

func TestLowerObjectConditionInvalidSemanticShapesFailClosed(t *testing.T) {
	t.Parallel()
	tests := []compiler.CompiledCondition{
		{
			Kind:          compiler.ConditionIf,
			Intervening:   true,
			Predicate:     compiler.ConditionPredicateObjectMatches,
			ObjectBinding: compiler.ReferenceBindingAmbiguous,
			Selection: compiler.ConditionSelection{
				RequiredTypes: []types.Card{types.Creature},
			},
		},
		{
			Kind:          compiler.ConditionIf,
			Intervening:   true,
			Predicate:     compiler.ConditionPredicateObjectExists,
			ObjectBinding: compiler.ReferenceBindingEventPermanent,
		},
		{
			Kind:          compiler.ConditionIf,
			Intervening:   true,
			Predicate:     compiler.ConditionPredicateObjectExists,
			ObjectBinding: compiler.ReferenceBindingSource,
			Selection: compiler.ConditionSelection{
				RequiredTypes: []types.Card{types.Creature},
			},
		},
		{
			Kind:          compiler.ConditionIf,
			Intervening:   true,
			Predicate:     compiler.ConditionPredicateObjectMatches,
			ObjectBinding: compiler.ReferenceBindingSource,
			Selection: compiler.ConditionSelection{
				RequiredTypes: []types.Card{types.Creature},
				Tapped:        compiler.ConditionTriState(99),
			},
		},
	}
	for _, condition := range tests {
		if got, ok := lowerCondition(condition, conditionContextInterveningTrigger); ok {
			t.Fatalf("lowerCondition(%+v) = %+v, true; want fail closed", condition, got)
		}
	}
}

func TestRenderGeneratedObjectInterveningConditions(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		oracleText string
		want       string
	}{
		{"When this creature dies, if it was a Human, draw a card.", "ObjectMatches:"},
		{"At the beginning of your upkeep, if this creature is on the battlefield, draw a card.", "Object:"},
		{"Whenever an artifact is put into a graveyard from the battlefield, if it had counters on it, draw a card.", "InterveningIfEventPermanentHadCounters: true"},
	} {
		source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
			Name:       "Test Human",
			Layout:     "normal",
			TypeLine:   "Creature — Human",
			OracleText: test.oracleText,
			Power:      new("2"),
			Toughness:  new("2"),
		}, "t")
		if err != nil {
			t.Fatal(err)
		}

		if len(diagnostics) != 0 {
			t.Fatalf("diagnostics = %#v", diagnostics)
		}
		if !strings.Contains(source, test.want) {
			t.Fatalf("source missing %q:\n%s", test.want, source)
		}

		rendered, err := (Renderer{}).renderControllerControlsCondition(newRenderCtx(), &game.Condition{
			Object: opt.Val(game.EventPermanentReference()),
			Types:  []types.Card{types.Creature},
		}, "legacy object")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(rendered, "Object:") || !strings.Contains(rendered, "Types:") {
			t.Fatalf("legacy Object+Types condition rendered as %s", rendered)
		}
	}
}

// TestAttackingControllerInterveningConditionLowers verifies the Mangara combat
// intervening-if "if two or more of those creatures are attacking you and/or
// planeswalkers you control" lowers cleanly and renders the typed condition
// field.
func TestAttackingControllerInterveningConditionLowers(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Mangara, the Diplomat",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Human Advisor",
		OracleText: "Lifelink\nWhenever an opponent attacks with creatures, if two or more of those creatures are attacking you and/or planeswalkers you control, draw a card.",
		ManaCost:   "{2}{W}",
		Power:      new("1"),
		Toughness:  new("4"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "Aggregate: game.AggregateAttackersAttackingController, Op: compare.GreaterOrEqual, Value: 2") {
		t.Fatalf("source missing attackers-attacking-controller aggregate:\n%s", source)
	}
}

func TestTargetLoweringFollowsTypedMeaningNotText(t *testing.T) {
	t.Parallel()
	permanent, ok := permanentTargetSpec(compiler.CompiledTarget{
		Text:        "irrelevant",
		Cardinality: compiler.TargetCardinality{Min: 1, Max: 1},
		Exact:       true,
		Selector: compiler.CompiledSelector{
			Kind:       compiler.SelectorCreature,
			Controller: compiler.ControllerOpponent,
		},
	})
	if !ok ||
		permanent.Allow != game.TargetAllowPermanent ||
		permanent.Selection.Val.Controller != game.ControllerOpponent ||
		!slices.Equal(permanent.Selection.Val.RequiredTypesAny, []types.Card{types.Creature}) {
		t.Fatalf("permanent target = %#v, %v", permanent, ok)
	}
	if _, ok := permanentTargetSpec(compiler.CompiledTarget{
		Cardinality: compiler.TargetCardinality{Min: 0, Max: 2},
		Selector:    compiler.CompiledSelector{Kind: compiler.SelectorCreature},
	}); ok {
		t.Fatal("multi-object cardinality lowered through single-object target spec")
	}

	damage, ok := damageTargetSpec(compiler.CompiledTarget{
		Text:        "irrelevant",
		Cardinality: compiler.TargetCardinality{Min: 1, Max: 1},
		Exact:       true,
		Selector:    compiler.CompiledSelector{Kind: compiler.SelectorAny},
	})
	if !ok || damage.Allow != game.TargetAllowPermanent|game.TargetAllowPlayer {
		t.Fatalf("damage target = %#v, %v", damage, ok)
	}

	ability, ok := counterAbilityTargetSpec(compiler.CompiledTarget{
		Text:        "irrelevant",
		Cardinality: compiler.TargetCardinality{Min: 1, Max: 1},
		Selector:    compiler.CompiledSelector{Kind: compiler.SelectorActivatedOrTriggeredAbility},
	})
	if !ok || !slices.Equal(
		ability.Predicate.StackObjectKinds,
		[]game.StackObjectKind{game.StackActivatedAbility, game.StackTriggeredAbility},
	) {
		t.Fatalf("ability target = %#v, %v", ability, ok)
	}

	player, ok := playerTargetSpec(compiler.CompiledTarget{
		Text:        "irrelevant",
		Cardinality: compiler.TargetCardinality{Min: 1, Max: 1},
		Exact:       true,
		Selector:    compiler.CompiledSelector{Kind: compiler.SelectorOpponent},
	})
	if !ok || player.Selection.Val.Player != game.PlayerOpponent {
		t.Fatalf("player target = %#v, %v", player, ok)
	}
}

func TestManifestLoweringFollowsTypedMeaningNotText(t *testing.T) {
	t.Parallel()
	content, diagnostic := lowerManifestSpell(contentCtx{
		text: "irrelevant",
		content: compiler.AbilityContent{
			Effects: []compiler.CompiledEffect{{Kind: compiler.EffectManifestDread, Exact: true}},
		},
	})
	if diagnostic != nil {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
	mode := content.Modes[0]
	manifest, ok := mode.Sequence[0].Primitive.(game.Manifest)
	if !ok {
		t.Fatalf("primitive = %#v, want game.Manifest", mode.Sequence[0].Primitive)
	}
	if !manifest.Dread {
		t.Fatal("typed manifest dread lowered with Dread=false")
	}
}

func TestReplacementLoweringFollowsTypedMeaningNotText(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"If another red source you control would deal damage to a permanent or player, it deals that much damage plus 1 to that permanent or player instead.",
		"If one or more +1/+1 counters would be put on a creature you control, twice that many +1/+1 counters are put on that creature instead.",
		"If an effect would create one or more tokens under your control, it creates twice that many of those tokens instead.",
		"If this creature would die, exile it instead.",
		"This creature enters with three +1/+1 counters on it.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileTestOracle(
				source,
				parser.Context{CardName: "Test Card"},
				compiler.Context{},
			)
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			ability := compilation.Abilities[0]
			ability.Text = "irrelevant"
			for i := range ability.Content.Effects {
				ability.Content.Effects[i].Text = "irrelevant"
				ability.Content.Effects[i].Amount.Text = "irrelevant"
			}
			lowered, diagnostic := lowerReplacementAbility(ability)
			if diagnostic != nil || !lowered.replacementAbility.Exists {
				t.Fatalf("lowering = %#v, diagnostic = %#v", lowered, diagnostic)
			}
		})
	}
}

func TestReplacementLoweringRejectsUnrepresentedTypedModifier(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileTestOracle(
		"If an effect would create one or more tokens under your control, it creates twice that many of those tokens instead.",
		parser.Context{CardName: "Test Card"},
		compiler.Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[0]
	ability.Text = "irrelevant"
	ability.Content.Effects[1].Text = "irrelevant"
	ability.Content.Effects[1].Selector.Tapped = true
	_, diagnostic := lowerReplacementAbility(ability)
	if diagnostic == nil {
		t.Fatal("expected unrepresented tapped-token modifier to fail closed")
	}

	compilation, diagnostics = compileTestOracle(
		"If an effect would create one or more tokens under your control, it creates twice that many Treasure tokens instead.",
		parser.Context{CardName: "Test Card"},
		compiler.Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("Treasure parse/compile diagnostics = %#v", diagnostics)
	}
	_, diagnostic = lowerReplacementAbility(compilation.Abilities[0])
	if diagnostic == nil {
		t.Fatal("expected unrepresented Treasure-token modifier to fail closed")
	}
}

func TestManaLoweringFollowsTypedMeaningNotText(t *testing.T) {
	t.Parallel()
	content, ok := typedManaEffectContent(compiler.CompiledEffectMana{
		Symbols:     []string{"{G}", "{W}"},
		Colors:      []mana.Color{mana.G, mana.W},
		ColorsKnown: true,
		Choice:      true,
	})
	if !ok || len(content.Modes) != 1 || len(content.Modes[0].Sequence) == 0 {
		t.Fatalf("content = %#v, ok = %v", content, ok)
	}
}

// TestLowerTriggeringPlayerHandSizeCondition verifies the phase-trigger
// intervening-if "if that player has N or fewer/more cards in hand" lowers to a
// hand-size comparison against the triggering player on each opponent's upkeep,
// including the "no cards in hand" zero-threshold form.
func TestLowerTriggeringPlayerHandSizeCondition(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		oracle string
		op     compare.Op
		value  int
	}{
		{"two or fewer", "At the beginning of each opponent's upkeep, if that player has two or fewer cards in hand, this creature deals 3 damage to that player.", compare.LessOrEqual, 2},
		{"no cards", "At the beginning of each opponent's upkeep, if that player has no cards in hand, they lose 2 life.", compare.LessOrEqual, 0},
		{"five or more", "At the beginning of each opponent's upkeep, if that player has five or more cards in hand, this creature deals 4 damage to that player.", compare.GreaterOrEqual, 5},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Hand Watcher",
				Layout:     "normal",
				TypeLine:   "Creature — Imp",
				OracleText: tc.oracle,
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			cond := face.TriggeredAbilities[0].Trigger.InterveningCondition
			want := game.AggregateComparison{Aggregate: game.AggregateEventPlayerHandSize, Op: tc.op, Value: tc.value}
			if !cond.Exists || len(cond.Val.Aggregates) != 1 || cond.Val.Aggregates[0] != want {
				t.Fatalf("aggregates = %#v, want %#v", cond.Val.Aggregates, want)
			}
		})
	}
}
