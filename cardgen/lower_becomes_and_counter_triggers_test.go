package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

func TestLowerBecomesTabTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Tapper",
		Layout:     "normal",
		TypeLine:   "Creature — Human",
		OracleText: "Whenever this creature becomes tapped, draw a card.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	trigger := face.TriggeredAbilities[0]
	if trigger.Trigger.Type != game.TriggerWhenever {
		t.Fatalf("trigger type = %v, want TriggerWhenever", trigger.Trigger.Type)
	}
	if trigger.Trigger.Pattern.Event != game.EventPermanentTapped {
		t.Fatalf("trigger event = %v, want EventPermanentTapped", trigger.Trigger.Pattern.Event)
	}
	if trigger.Trigger.Pattern.Source != game.TriggerSourceSelf {
		t.Fatalf("trigger source = %v, want TriggerSourceSelf", trigger.Trigger.Pattern.Source)
	}
}

func TestLowerLandBecomesTabTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Land",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "Whenever this land becomes tapped, draw a card.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	trigger := face.TriggeredAbilities[0]
	if trigger.Trigger.Pattern.Event != game.EventPermanentTapped {
		t.Fatalf("trigger event = %v, want EventPermanentTapped", trigger.Trigger.Pattern.Event)
	}
	if trigger.Trigger.Pattern.Source != game.TriggerSourceSelf {
		t.Fatalf("trigger source = %v, want TriggerSourceSelf", trigger.Trigger.Pattern.Source)
	}
}

func TestLowerBecomesUntappedTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Untapper",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "Whenever this artifact becomes untapped, draw a card.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	trigger := face.TriggeredAbilities[0]
	if trigger.Trigger.Type != game.TriggerWhenever {
		t.Fatalf("trigger type = %v, want TriggerWhenever", trigger.Trigger.Type)
	}
	if trigger.Trigger.Pattern.Event != game.EventPermanentUntapped {
		t.Fatalf("trigger event = %v, want EventPermanentUntapped", trigger.Trigger.Pattern.Event)
	}
	if trigger.Trigger.Pattern.Source != game.TriggerSourceSelf {
		t.Fatalf("trigger source = %v, want TriggerSourceSelf", trigger.Trigger.Pattern.Source)
	}
}

func TestLowerNamedCardBecomesTabTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Gran-Gran",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Human",
		OracleText: "Whenever Gran-Gran becomes tapped, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	trigger := face.TriggeredAbilities[0]
	if trigger.Trigger.Pattern.Event != game.EventPermanentTapped {
		t.Fatalf("trigger event = %v, want EventPermanentTapped", trigger.Trigger.Pattern.Event)
	}
	if trigger.Trigger.Pattern.Source != game.TriggerSourceSelf {
		t.Fatalf("trigger source = %v, want TriggerSourceSelf", trigger.Trigger.Pattern.Source)
	}
}

func TestLowerAuraDiesTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Aura",
		Layout:     "normal",
		TypeLine:   "Enchantment — Aura",
		OracleText: "When this aura is put into a graveyard from the battlefield, draw a card.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	trigger := face.TriggeredAbilities[0]
	if trigger.Trigger.Type != game.TriggerWhen {
		t.Fatalf("trigger type = %v, want TriggerWhen", trigger.Trigger.Type)
	}
	if trigger.Trigger.Pattern.Event != game.EventZoneChanged {
		t.Fatalf("trigger event = %v, want EventZoneChanged", trigger.Trigger.Pattern.Event)
	}
	if trigger.Trigger.Pattern.Source != game.TriggerSourceSelf {
		t.Fatalf("trigger source = %v, want TriggerSourceSelf", trigger.Trigger.Pattern.Source)
	}
}

func TestLowerArtifactDiesTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Artifact",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "When this artifact is put into a graveyard from the battlefield, draw a card.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	trigger := face.TriggeredAbilities[0]
	if trigger.Trigger.Type != game.TriggerWhen {
		t.Fatalf("trigger type = %v, want TriggerWhen", trigger.Trigger.Type)
	}
	if trigger.Trigger.Pattern.Event != game.EventZoneChanged {
		t.Fatalf("trigger event = %v, want EventZoneChanged", trigger.Trigger.Pattern.Event)
	}
	if trigger.Trigger.Pattern.Source != game.TriggerSourceSelf {
		t.Fatalf("trigger source = %v, want TriggerSourceSelf", trigger.Trigger.Pattern.Source)
	}
}

func TestLowerEnchantmentDiesTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Enchantment",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "When this enchantment is put into a graveyard from the battlefield, draw a card.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	trigger := face.TriggeredAbilities[0]
	if trigger.Trigger.Pattern.Event != game.EventZoneChanged {
		t.Fatalf("trigger event = %v, want EventZoneChanged", trigger.Trigger.Pattern.Event)
	}
	if trigger.Trigger.Pattern.Source != game.TriggerSourceSelf {
		t.Fatalf("trigger source = %v, want TriggerSourceSelf", trigger.Trigger.Pattern.Source)
	}
}

func TestLowerCounterAddedOneOrMoreTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Counter Watcher",
		Layout:     "normal",
		TypeLine:   "Creature — Human",
		OracleText: "Whenever one or more +1/+1 counters are put on this creature, draw a card.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	trigger := face.TriggeredAbilities[0]
	pat := trigger.Trigger.Pattern
	if trigger.Trigger.Type != game.TriggerWhenever {
		t.Fatalf("trigger type = %v, want TriggerWhenever", trigger.Trigger.Type)
	}
	if pat.Event != game.EventCountersAdded {
		t.Fatalf("trigger event = %v, want EventCountersAdded", pat.Event)
	}
	if pat.Source != game.TriggerSourceSelf {
		t.Fatalf("trigger source = %v, want TriggerSourceSelf", pat.Source)
	}
	if !pat.MatchCounterKind {
		t.Fatal("MatchCounterKind = false, want true")
	}
	if pat.CounterKind != counter.PlusOnePlusOne {
		t.Fatalf("CounterKind = %v, want PlusOnePlusOne", pat.CounterKind)
	}
	if !pat.OneOrMore {
		t.Fatal("OneOrMore = false, want true")
	}
}

func TestLowerCounterAddedSingleTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Counter Watcher",
		Layout:     "normal",
		TypeLine:   "Creature — Human",
		OracleText: "Whenever a +1/+1 counter is put on this creature, draw a card.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	pat := face.TriggeredAbilities[0].Trigger.Pattern
	if !pat.MatchCounterKind {
		t.Fatal("MatchCounterKind = false, want true")
	}
	if pat.CounterKind != counter.PlusOnePlusOne {
		t.Fatalf("CounterKind = %v, want PlusOnePlusOne", pat.CounterKind)
	}
	if pat.OneOrMore {
		t.Fatal("OneOrMore = true, want false for singular counter trigger")
	}
}

func TestLowerCounterAddedMinusOneTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Minus Counter Watcher",
		Layout:     "normal",
		TypeLine:   "Creature — Human",
		OracleText: "Whenever one or more -1/-1 counters are put on this creature, draw a card.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	pat := face.TriggeredAbilities[0].Trigger.Pattern
	if pat.CounterKind != counter.MinusOneMinusOne {
		t.Fatalf("CounterKind = %v, want MinusOneMinusOne", pat.CounterKind)
	}
	if !pat.OneOrMore {
		t.Fatal("OneOrMore = false, want true")
	}
}

func TestLowerCounterAddedGenericKindIsSupported(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Charge Watcher",
		Layout:     "normal",
		TypeLine:   "Creature — Human",
		OracleText: "Whenever one or more charge counters are put on this creature, draw a card.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	pat := face.TriggeredAbilities[0].Trigger.Pattern
	if !pat.MatchCounterKind || pat.CounterKind != counter.Charge {
		t.Fatalf("counter kind = (%v, %v), want (true, Charge)", pat.MatchCounterKind, pat.CounterKind)
	}
	if !pat.OneOrMore {
		t.Fatal("OneOrMore = false, want true")
	}
}

func TestLowerGenericPatternTriggerSupportedInterveningCondition(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		oracle    string
		wantEvent game.EventKind
	}{
		{
			name:      "another creature tapped if you control an artifact",
			oracle:    "Whenever another creature you control becomes tapped, if you control an artifact, draw a card.",
			wantEvent: game.EventPermanentTapped,
		},
		{
			name:      "another creature tapped if opponent controls land",
			oracle:    "Whenever another creature you control becomes tapped, if an opponent controls two or more lands, draw a card.",
			wantEvent: game.EventPermanentTapped,
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
			if trigger.Pattern.Event != tc.wantEvent {
				t.Errorf("event = %v, want %v", trigger.Pattern.Event, tc.wantEvent)
			}
			if trigger.InterveningIf == "" || !trigger.InterveningCondition.Exists {
				t.Fatalf("trigger = %+v, want intervening condition", trigger)
			}
		})
	}
}

func TestLowerGenericPatternTriggerInterveningIfFailsClosedOnUnsupportedCondition(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "Whenever another creature you control becomes tapped, if you have seven or more cards in hand, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	}, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("generic trigger with unsupported intervening condition unexpectedly lowered")
	}
	if !strings.Contains(diagnostics[0].Detail, "condition") {
		t.Fatalf("diagnostic = %#v, want condition detail", diagnostics[0])
	}
}
