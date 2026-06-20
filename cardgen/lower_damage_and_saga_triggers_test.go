package cardgen

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerDamageSourceTriggers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		text string
		want game.TriggerPattern
	}{
		{
			name: "self damage",
			text: "Whenever this creature deals damage, draw a card.",
			want: game.TriggerPattern{
				Event:   game.EventDamageDealt,
				Source:  game.TriggerSourceSelf,
				Subject: game.TriggerSubjectDamageSource,
			},
		},
		{
			name: "self damage to player",
			text: "Whenever this creature deals damage to a player, draw a card.",
			want: game.TriggerPattern{
				Event:           game.EventDamageDealt,
				Source:          game.TriggerSourceSelf,
				Subject:         game.TriggerSubjectDamageSource,
				DamageRecipient: game.DamageRecipientPlayer,
			},
		},
		{
			name: "self damage to opponent",
			text: "Whenever this creature deals damage to an opponent, draw a card.",
			want: game.TriggerPattern{
				Event:           game.EventDamageDealt,
				Source:          game.TriggerSourceSelf,
				Subject:         game.TriggerSubjectDamageSource,
				Player:          game.TriggerPlayerOpponent,
				DamageRecipient: game.DamageRecipientPlayer,
			},
		},
		{
			name: "self damage to creature",
			text: "Whenever this creature deals damage to a creature, draw a card.",
			want: game.TriggerPattern{
				Event:                game.EventDamageDealt,
				Source:               game.TriggerSourceSelf,
				Subject:              game.TriggerSubjectDamageSource,
				DamageRecipient:      game.DamageRecipientPermanent,
				DamageRecipientTypes: []types.Card{types.Creature},
			},
		},
		{
			name: "self combat damage",
			text: "Whenever this creature deals combat damage, draw a card.",
			want: game.TriggerPattern{
				Event:               game.EventDamageDealt,
				Source:              game.TriggerSourceSelf,
				Subject:             game.TriggerSubjectDamageSource,
				RequireCombatDamage: true,
			},
		},
		{
			name: "self combat damage to opponent",
			text: "Whenever this creature deals combat damage to an opponent, draw a card.",
			want: game.TriggerPattern{
				Event:               game.EventDamageDealt,
				Source:              game.TriggerSourceSelf,
				Subject:             game.TriggerSubjectDamageSource,
				Player:              game.TriggerPlayerOpponent,
				DamageRecipient:     game.DamageRecipientPlayer,
				RequireCombatDamage: true,
			},
		},
		{
			name: "equipped creature combat damage to player",
			text: "Whenever equipped creature deals combat damage to a player, draw a card.",
			want: game.TriggerPattern{
				Event:               game.EventDamageDealt,
				Source:              game.TriggerSourceAttachedPermanent,
				Subject:             game.TriggerSubjectDamageSource,
				DamageRecipient:     game.DamageRecipientPlayer,
				RequireCombatDamage: true,
				DamageSourceSelection: game.Selection{
					RequiredTypes: []types.Card{types.Creature},
				},
			},
		},
		{
			name: "enchanted creature damage",
			text: "Whenever enchanted creature deals damage, draw a card.",
			want: game.TriggerPattern{
				Event:   game.EventDamageDealt,
				Source:  game.TriggerSourceAttachedPermanent,
				Subject: game.TriggerSubjectDamageSource,
				DamageSourceSelection: game.Selection{
					RequiredTypes: []types.Card{types.Creature},
				},
			},
		},
		{
			name: "enchanted creature damage to opponent",
			text: "Whenever enchanted creature deals damage to an opponent, draw a card.",
			want: game.TriggerPattern{
				Event:           game.EventDamageDealt,
				Source:          game.TriggerSourceAttachedPermanent,
				Subject:         game.TriggerSubjectDamageSource,
				Player:          game.TriggerPlayerOpponent,
				DamageRecipient: game.DamageRecipientPlayer,
				DamageSourceSelection: game.Selection{
					RequiredTypes: []types.Card{types.Creature},
				},
			},
		},
		{
			name: "equipped creature damage to creature",
			text: "Whenever equipped creature deals damage to a creature, draw a card.",
			want: game.TriggerPattern{
				Event:                game.EventDamageDealt,
				Source:               game.TriggerSourceAttachedPermanent,
				Subject:              game.TriggerSubjectDamageSource,
				DamageRecipient:      game.DamageRecipientPermanent,
				DamageRecipientTypes: []types.Card{types.Creature},
				DamageSourceSelection: game.Selection{
					RequiredTypes: []types.Card{types.Creature},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Fighter",
				Layout:     "normal",
				TypeLine:   "Creature — Human Warrior",
				OracleText: tc.text,
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			trigger := face.TriggeredAbilities[0].Trigger
			if trigger.Type != game.TriggerWhenever {
				t.Fatalf("trigger type = %v, want TriggerWhenever", trigger.Type)
			}
			if !reflect.DeepEqual(trigger.Pattern, tc.want) {
				t.Fatalf("pattern = %+v, want %+v", trigger.Pattern, tc.want)
			}
		})
	}
}

func TestLowerDamageSourceTriggersFailClosed(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Whenever this creature deals combat damage to defending player, draw a card.",
		"Whenever equipped creature or this Equipment deals damage, draw a card.",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Fighter",
				Layout:     "normal",
				TypeLine:   "Creature — Human Warrior",
				OracleText: oracleText,
				Power:      new("2"),
				Toughness:  new("2"),
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) == 0 {
				t.Fatal("unsupported damage-source trigger unexpectedly lowered")
			}
		})
	}
}

func TestLowerLifeDamageReceivedTriggers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		text string
		want game.TriggerPattern
	}{
		{
			name: "you gain life",
			text: "Whenever you gain life, draw a card.",
			want: game.TriggerPattern{
				Event:  game.EventLifeGained,
				Player: game.TriggerPlayerYou,
			},
		},
		{
			name: "you lose life",
			text: "Whenever you lose life, draw a card.",
			want: game.TriggerPattern{
				Event:  game.EventLifeLost,
				Player: game.TriggerPlayerYou,
			},
		},
		{
			name: "opponent gains life",
			text: "Whenever an opponent gains life, draw a card.",
			want: game.TriggerPattern{
				Event:  game.EventLifeGained,
				Player: game.TriggerPlayerOpponent,
			},
		},
		{
			name: "opponent loses life",
			text: "Whenever an opponent loses life, you gain 1 life.",
			want: game.TriggerPattern{
				Event:  game.EventLifeLost,
				Player: game.TriggerPlayerOpponent,
			},
		},
		{
			name: "self dealt damage",
			text: "Whenever this creature is dealt damage, draw a card.",
			want: game.TriggerPattern{
				Event:           game.EventDamageDealt,
				Source:          game.TriggerSourceSelf,
				Subject:         game.TriggerSubjectPermanent,
				DamageRecipient: game.DamageRecipientPermanent,
			},
		},
		{
			name: "enchanted creature dealt damage",
			text: "Whenever enchanted creature is dealt damage, draw a card.",
			want: game.TriggerPattern{
				Event:           game.EventDamageDealt,
				Source:          game.TriggerSourceAttachedPermanent,
				DamageRecipient: game.DamageRecipientPermanent,
				SubjectSelection: game.Selection{
					RequiredTypes: []types.Card{types.Creature},
				},
			},
		},
		{
			name: "equipped creature dealt damage",
			text: "Whenever equipped creature is dealt damage, draw a card.",
			want: game.TriggerPattern{
				Event:           game.EventDamageDealt,
				Source:          game.TriggerSourceAttachedPermanent,
				DamageRecipient: game.DamageRecipientPermanent,
				SubjectSelection: game.Selection{
					RequiredTypes: []types.Card{types.Creature},
				},
			},
		},
		{
			name: "you are dealt damage",
			text: "Whenever you're dealt damage, draw a card.",
			want: game.TriggerPattern{
				Event:           game.EventDamageDealt,
				Player:          game.TriggerPlayerYou,
				DamageRecipient: game.DamageRecipientPlayer,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Cleric",
				Layout:     "normal",
				TypeLine:   "Creature — Human Cleric",
				OracleText: tc.text,
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			trigger := face.TriggeredAbilities[0].Trigger
			if trigger.Type != game.TriggerWhenever {
				t.Fatalf("trigger type = %v, want TriggerWhenever", trigger.Type)
			}
			if !reflect.DeepEqual(trigger.Pattern, tc.want) {
				t.Fatalf("pattern = %+v, want %+v", trigger.Pattern, tc.want)
			}
		})
	}
}

func TestLowerLifeDamageReceivedTriggersFailClosed(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Whenever you gain or lose life, draw a card.",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Cleric",
				Layout:     "normal",
				TypeLine:   "Creature — Human Cleric",
				OracleText: oracleText,
				Power:      new("2"),
				Toughness:  new("2"),
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) == 0 {
				t.Fatal("unsupported life/damage trigger unexpectedly lowered")
			}
		})
	}
}

func TestLowerLifeDamageTriggerSupportedInterveningCondition(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		oracle    string
		wantEvent game.EventKind
	}{
		{
			name:      "life gain if you control an artifact",
			oracle:    "Whenever you gain life, if you control an artifact, draw a card.",
			wantEvent: game.EventLifeGained,
		},
		{
			name:      "life loss if you have 5 or more life",
			oracle:    "Whenever you lose life, if you have 5 or more life, you gain 1 life.",
			wantEvent: game.EventLifeLost,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Cleric",
				Layout:     "normal",
				TypeLine:   "Creature — Human Cleric",
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

func TestLowerLifeDamageTriggerInterveningIfFailsClosedOnUnsupportedCondition(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Cleric",
		Layout:     "normal",
		TypeLine:   "Creature — Human Cleric",
		OracleText: "Whenever you gain life, if you have seven or more cards in hand, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("life trigger with unsupported intervening condition unexpectedly lowered")
	}
	if !strings.Contains(diagnostics[0].Detail, "condition") {
		t.Fatalf("diagnostic = %#v, want condition detail", diagnostics[0])
	}
}

func TestLowerKickedEnterTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Layout:     "normal",
		ManaCost:   "{2}{U}",
		TypeLine:   "Creature — Wizard",
		OracleText: "Kicker {1}{U}\nWhen this creature enters, if it was kicked, draw two cards.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	trigger := face.TriggeredAbilities[0].Trigger
	if trigger.InterveningIf != "if it was kicked" ||
		!trigger.InterveningIfEventPermanentWasKicked {
		t.Fatalf("trigger = %+v, want kicked intervening-if", trigger)
	}
	draw, ok := face.TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.Draw)
	if !ok || draw.Amount != game.Fixed(2) {
		t.Fatalf("primitive = %+v, want draw two", face.TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive)
	}
}

func TestLowerWasCastEnterTriggers(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Construct",
		Layout:     "normal",
		TypeLine:   "Artifact Creature — Construct",
		OracleText: "When this creature enters, if it was cast, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	trigger := face.TriggeredAbilities[0].Trigger
	if trigger.InterveningIf != "if it was cast" || !trigger.InterveningIfEventPermanentWasCast {
		t.Fatalf("trigger = %+v, want was-cast intervening-if", trigger)
	}
}

func TestLowerSelfEnterTriggerSupportsCasterRelativeCondition(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Construct",
		Layout:     "normal",
		TypeLine:   "Artifact Creature — Construct",
		OracleText: "When this creature enters, if you cast it, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	trigger := face.TriggeredAbilities[0]
	if !trigger.Trigger.InterveningIfEventPermanentWasCastByController {
		t.Fatalf(
			"cast conditions = (cast: %v, cast by controller: %v), want caster-relative condition",
			trigger.Trigger.InterveningIfEventPermanentWasCast,
			trigger.Trigger.InterveningIfEventPermanentWasCastByController,
		)
	}
	if len(trigger.Content.Modes) != 1 || len(trigger.Content.Modes[0].Sequence) != 1 {
		t.Fatalf("trigger content = %#v, want one draw instruction", trigger.Content)
	}
	if _, ok := trigger.Content.Modes[0].Sequence[0].Primitive.(game.Draw); !ok {
		t.Fatalf("primitive = %T, want game.Draw", trigger.Content.Modes[0].Sequence[0].Primitive)
	}
}

func TestLowerAttackedThisTurnEnterTriggerFailsClosed(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Warrior",
		Layout:     "normal",
		TypeLine:   "Creature — Warrior",
		OracleText: "When this creature enters, if this creature attacked this turn, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("attacked-this-turn self-enter condition unexpectedly lowered")
	}
}

func TestLowerControlsPermanentEnterTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Artificer",
		Layout:     "normal",
		TypeLine:   "Creature — Artificer",
		OracleText: "When this creature enters, if you control an artifact, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	trigger := face.TriggeredAbilities[0].Trigger
	if trigger.InterveningIf != "if you control an artifact" ||
		!trigger.InterveningCondition.Exists {
		t.Fatalf("trigger = %+v, want controls-artifact intervening-if", trigger)
	}
	selection := trigger.InterveningCondition.Val.ControlsMatching
	if !selection.Exists ||
		!slices.Equal(selection.Val.Selection.RequiredTypes, []types.Card{types.Artifact}) {
		t.Fatalf("condition = %+v, want controls an artifact", trigger.InterveningCondition.Val)
	}
}

func TestLowerEnterTriggerSupportsSubtypeInterveningCondition(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Handler",
		Layout:     "normal",
		TypeLine:   "Creature — Elf",
		OracleText: "When this creature enters, if you control an Elf, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	condition := face.TriggeredAbilities[0].Trigger.InterveningCondition
	if !condition.Exists ||
		!condition.Val.ControlsMatching.Exists ||
		!slices.Equal(condition.Val.ControlsMatching.Val.Selection.SubtypesAny, []types.Sub{types.Elf}) {
		t.Fatalf("condition = %+v, want controlled Elf selection", condition)
	}
}

func TestLowerSagaChapterAbilities(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Saga",
		Layout:     "saga",
		TypeLine:   "Enchantment — Saga",
		OracleText: "I — Draw a card.\nII, III — Draw two cards.",
	})
	if len(face.ChapterAbilities) != 2 {
		t.Fatalf("got %d chapter abilities, want 2", len(face.ChapterAbilities))
	}
	if !slices.Equal(face.ChapterAbilities[0].Chapters, []int{1}) ||
		!slices.Equal(face.ChapterAbilities[1].Chapters, []int{2, 3}) {
		t.Fatalf("chapter numbers = %v, %v", face.ChapterAbilities[0].Chapters, face.ChapterAbilities[1].Chapters)
	}
	draw, ok := face.ChapterAbilities[1].Content.Modes[0].Sequence[0].Primitive.(game.Draw)
	if !ok {
		t.Fatalf("primitive = %T, want game.Draw", face.ChapterAbilities[1].Content.Modes[0].Sequence[0].Primitive)
	}
	if got := draw.Amount; got != game.Fixed(2) {
		t.Fatalf("draw amount = %#v, want 2", got)
	}
}

func TestLowerReadAheadSaga(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Read Ahead Saga",
		Layout:     "saga",
		TypeLine:   "Enchantment — Saga",
		OracleText: "Read ahead (Choose a chapter and start with that many lore counters. Add one after your draw step. Skipped chapters don't trigger.)\nI — Draw a card.\nII — Draw a card.",
	})
	if len(face.StaticAbilities) != 1 || !game.BodyHasKeyword(&face.StaticAbilities[0].Body, game.ReadAhead) {
		t.Fatalf("static abilities = %#v, want ReadAheadStaticBody", face.StaticAbilities)
	}
	if len(face.ChapterAbilities) != 2 {
		t.Fatalf("chapter abilities = %#v, want two", face.ChapterAbilities)
	}
}

func TestLowerReadAheadRejectsNoncanonicalReminder(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Malformed Read Ahead Saga",
		Layout:     "saga",
		TypeLine:   "Enchantment — Saga",
		OracleText: "Read ahead (Choose whichever chapter you want.)\nI — Draw a card.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("noncanonical Read ahead reminder unexpectedly lowered")
	}
}

func TestLowerReadAheadRejectsMismatchedSacrificeChapter(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Mismatched Read Ahead Saga",
		Layout:     "saga",
		TypeLine:   "Enchantment — Saga",
		OracleText: "Read ahead (Choose a chapter and start with that many lore counters. Add one after your draw step. Skipped chapters don't trigger. Sacrifice after IV.)\nI — Draw a card.\nII — Draw a card.\nIII — Draw a card.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("mismatched Read ahead sacrifice chapter unexpectedly lowered")
	}
}

func TestLowerChapterShapedTextRequiresSagaSubtype(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Not a Saga",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "I — Draw a card.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected non-Saga chapter-shaped text to be rejected")
	}
}

func TestOrdinarySagaReminder(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"(As this Saga enters and after your draw step, add a lore counter.)",
		"(As this Saga enters and after your draw step, add a lore counter. Sacrifice after I.)",
		"(As this Saga enters and after your draw step add a lore counter. Sacrifice after III.)",
	} {
		document, _ := parser.Parse(text, parser.Context{Saga: true})
		if len(document.Abilities) != 1 || !document.Abilities[0].SagaReminder {
			t.Errorf("parser SagaReminder for %q = false, want true", text)
		}
	}
	for _, text := range []string{
		"Read ahead (Choose a chapter and start with that many lore counters.)",
		"(As this Saga enters and after your draw step, add a lore counter. Sacrifice after VII.)",
		"(As this Saga enters, add a lore counter.)",
	} {
		document, _ := parser.Parse(text, parser.Context{Saga: true})
		if len(document.Abilities) == 1 && document.Abilities[0].SagaReminder {
			t.Errorf("parser SagaReminder for %q = true, want false", text)
		}
	}
}

func TestLowerSagaChapterConsumesInlineReminderText(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Saga",
		Layout:     "saga",
		TypeLine:   "Enchantment — Saga",
		OracleText: "I — Proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)",
	})
	if len(face.ChapterAbilities) != 1 {
		t.Fatalf("got %d chapter abilities, want 1", len(face.ChapterAbilities))
	}
}
