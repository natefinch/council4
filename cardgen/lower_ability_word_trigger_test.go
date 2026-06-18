package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestRulesFreeAbilityWordTriggerBodies verifies that purely cosmetic
// (rules-free) ability words that label triggered abilities — Enrage,
// Inspired, Flurry, Opus, Parley — no longer block lowering of an otherwise
// supported trigger body. The ability-word label carries no rules meaning, so
// the body must lower exactly as it would without the label.
func TestRulesFreeAbilityWordTriggerBodies(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		oracle    string
		wantEvent game.EventKind
	}{
		{
			name:      "enrage draw",
			oracle:    "Enrage — Whenever this creature is dealt damage, draw a card.",
			wantEvent: game.EventDamageDealt,
		},
		{
			name:      "enrage gain life",
			oracle:    "Enrage — Whenever this creature is dealt damage, you gain 2 life.",
			wantEvent: game.EventDamageDealt,
		},
		{
			name:      "inspired draw",
			oracle:    "Inspired — Whenever this creature becomes untapped, draw a card.",
			wantEvent: game.EventPermanentUntapped,
		},
		{
			name:      "flurry counter on self",
			oracle:    "Flurry — Whenever you cast your second spell each turn, put a +1/+1 counter on this creature.",
			wantEvent: game.EventSpellCast,
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
			if got := face.TriggeredAbilities[0].Trigger.Pattern.Event; got != tc.wantEvent {
				t.Errorf("event = %v, want %v", got, tc.wantEvent)
			}
			content := face.TriggeredAbilities[0].Content
			if len(content.Modes) != 1 || len(content.Modes[0].Sequence) == 0 {
				t.Fatalf("expected a lowered body sequence, got %#v", content)
			}
		})
	}
}

// TestRulesFreeAbilityWordMatchesUnlabeledBody verifies that adding a
// rules-free ability-word label produces the same lowered triggered ability as
// the identical body with no label, confirming the label is cosmetic.
func TestRulesFreeAbilityWordMatchesUnlabeledBody(t *testing.T) {
	t.Parallel()
	labeled := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "Enrage — Whenever this creature is dealt damage, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	unlabeled := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "Whenever this creature is dealt damage, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(labeled.TriggeredAbilities) != 1 || len(unlabeled.TriggeredAbilities) != 1 {
		t.Fatalf("labeled=%d unlabeled=%d triggered abilities, want 1 each",
			len(labeled.TriggeredAbilities), len(unlabeled.TriggeredAbilities))
	}
	labeled.TriggeredAbilities[0].Text = ""
	unlabeled.TriggeredAbilities[0].Text = ""
	if !reflect.DeepEqual(labeled.TriggeredAbilities[0].Trigger, unlabeled.TriggeredAbilities[0].Trigger) {
		t.Errorf("trigger condition differs:\nlabeled=%#v\nunlabeled=%#v",
			labeled.TriggeredAbilities[0].Trigger, unlabeled.TriggeredAbilities[0].Trigger)
	}
}
