package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestLowerDiesTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature dies, draw two cards.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	if face.TriggeredAbilities[0].Trigger.Pattern.Event != game.EventPermanentDied {
		t.Fatalf("event = %v, want EventPermanentDied", face.TriggeredAbilities[0].Trigger.Pattern.Event)
	}
}

func TestLowerDiesTriggerHadNoPlusPlusCounters(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Undying Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature dies, if it had no +1/+1 counters, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	trigger := face.TriggeredAbilities[0].Trigger
	if trigger.InterveningIf != "if it had no +1/+1 counters" ||
		!trigger.InterveningIfEventPermanentHadNoCounterKind.Exists ||
		trigger.InterveningIfEventPermanentHadNoCounterKind.Val != counter.PlusOnePlusOne {
		t.Fatalf("trigger = %+v, want no +1/+1 counters intervening-if", trigger)
	}
}

func TestLowerDiesTriggerHadNoMinusMinusCounters(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Persist Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature dies, if it had no -1/-1 counters on it, it deals 3 damage to any target.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	ability := face.TriggeredAbilities[0]
	trigger := ability.Trigger
	if trigger.InterveningIf != "if it had no -1/-1 counters on it" ||
		!trigger.InterveningIfEventPermanentHadNoCounterKind.Exists ||
		trigger.InterveningIfEventPermanentHadNoCounterKind.Val != counter.MinusOneMinusOne {
		t.Fatalf("trigger = %+v, want no -1/-1 counters intervening-if", trigger)
	}
	damage, ok := ability.Content.Modes[0].Sequence[0].Primitive.(game.Damage)
	if !ok || !damage.DamageSource.Exists ||
		damage.DamageSource.Val != game.EventPermanentReference() {
		t.Fatalf("primitive = %+v, want damage from event permanent", ability.Content.Modes[0].Sequence[0].Primitive)
	}
}

func TestLowerDiesTriggerOptional(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature dies, you may draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	ability := face.TriggeredAbilities[0]
	if !ability.Optional {
		t.Fatal("dies trigger is not optional")
	}
	if _, ok := ability.Content.Modes[0].Sequence[0].Primitive.(game.Draw); !ok {
		t.Fatalf("primitive = %T, want game.Draw", ability.Content.Modes[0].Sequence[0].Primitive)
	}
}

func TestLowerDiesTriggerRejectsAmbiguousCounterAbsence(t *testing.T) {
	t.Parallel()
	for _, condition := range []string{
		"if it had no counters on it",
		"if it had no charge counters on it",
		"if it didn't have a +1/+1 counter on it",
	} {
		t.Run(condition, func(t *testing.T) {
			t.Parallel()
			_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: "When this creature dies, " + condition + ", draw a card.",
				Power:      new("2"),
				Toughness:  new("2"),
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) == 0 {
				t.Fatalf("ambiguous or unsupported condition %q unexpectedly lowered", condition)
			}
		})
	}
}

func TestLowerDiesTriggerReturnsEventCardToOwnersHand(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature dies, return it to its owner's hand.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	primitive := face.TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive
	move, ok := primitive.(game.MoveCard)
	if !ok {
		t.Fatalf("primitive = %T, want game.MoveCard", primitive)
	}
	if move.Card.Kind != game.CardReferenceEvent ||
		move.FromZone != zone.Graveyard ||
		move.Destination != zone.Hand {
		t.Fatalf("move = %+v, want event card from graveyard to hand", move)
	}
}

func TestLowerDiesTriggerGrantsAdventureCastFromGraveyard(t *testing.T) {
	t.Parallel()
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:   "Test Dreadknight // Test Whispers",
		Layout: "adventure",
		CardFaces: []ScryfallCardFace{
			{
				Name:       "Test Dreadknight",
				ManaCost:   "{1}{G}",
				TypeLine:   "Creature — Human Knight",
				OracleText: "When Test Dreadknight dies, you may cast it from your graveyard as an Adventure until the end of your next turn.",
				Power:      new("2"),
				Toughness:  new("1"),
			},
			{
				Name:       "Test Whispers",
				ManaCost:   "{1}{B}",
				TypeLine:   "Sorcery — Adventure",
				OracleText: "Draw a card.",
			},
		},
	})
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	ability := faces[0].TriggeredAbilities[0]
	if !ability.Optional {
		t.Fatal("cast-permission dies trigger is not optional")
	}
	primitive := ability.Content.Modes[0].Sequence[0].Primitive
	permission, ok := primitive.(game.GrantCastPermission)
	if !ok {
		t.Fatalf("primitive = %T, want game.GrantCastPermission", primitive)
	}
	if permission.Card.Kind != game.CardReferenceEvent ||
		permission.FromZone != zone.Graveyard ||
		permission.Face != game.FaceAlternate ||
		permission.Duration != game.DurationUntilEndOfYourNextTurn {
		t.Fatalf("permission = %+v, want event Adventure cast through next turn", permission)
	}
}

func TestLowerDiesTriggerRejectsAmbiguousEventCardReference(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"When this creature dies, return it to the battlefield.",
		"When this creature dies, cast it.",
		"When this creature dies, you may cast it from your graveyard.",
		"When this creature dies, return it to its owner's hand or the battlefield.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: text,
				Power:      new("2"),
				Toughness:  new("2"),
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) == 0 {
				t.Fatalf("ambiguous event-card reference unexpectedly lowered: %q", text)
			}
		})
	}
}

func TestLowerDiesTriggerRejectsEnterOnlyInterveningConditions(t *testing.T) {
	t.Parallel()
	for _, condition := range []string{
		"if it was kicked",
		"if it was cast",
		"if you cast it",
		"if this creature attacked this turn",
	} {
		t.Run(condition, func(t *testing.T) {
			t.Parallel()
			_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Bear",
				Layout:     "normal",
				TypeLine:   "Creature — Bear",
				OracleText: "When this creature dies, " + condition + ", draw a card.",
				Power:      new("2"),
				Toughness:  new("2"),
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) == 0 {
				t.Fatalf("self-dies trigger unexpectedly lowered with %q", condition)
			}
		})
	}
}

func TestLowerSelfDiesDamageTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Devil",
		Layout:     "normal",
		TypeLine:   "Creature — Devil",
		OracleText: "When this creature dies, it deals 3 damage to any target.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("got %d targets, want 1", len(mode.Targets))
	}
	damage, ok := mode.Sequence[0].Primitive.(game.Damage)
	if !ok ||
		damage.Amount.Value() != 3 ||
		!damage.DamageSource.Exists ||
		damage.DamageSource.Val != game.EventPermanentReference() {
		t.Fatalf("primitive = %+v, want damage from event permanent", mode.Sequence[0].Primitive)
	}
}

// TestLowerDamageEventPermanentSourceInNonZoneChangeTrigger verifies that the
// shared damage-lowering path handles "It deals X damage to any target." bodies
// in non-zone-change trigger shells, routing the source through
// lowerObjectReference and preserving DamageSource as EventPermanentReference.
func TestLowerDamageEventPermanentSourceInNonZoneChangeTrigger(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		oracle string
	}{
		{
			name:   "attack trigger it deals",
			oracle: "Whenever a creature attacks, it deals 1 damage to any target.",
		},
		{
			name:   "tapped trigger it deals",
			oracle: "Whenever a creature becomes tapped, it deals 2 damage to any target.",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Devil",
				Layout:     "normal",
				TypeLine:   "Creature — Devil",
				OracleText: tc.oracle,
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			mode := face.TriggeredAbilities[0].Content.Modes[0]
			damage, ok := mode.Sequence[0].Primitive.(game.Damage)
			if !ok {
				t.Fatalf("primitive = %T, want game.Damage", mode.Sequence[0].Primitive)
			}
			if !damage.DamageSource.Exists || damage.DamageSource.Val != game.EventPermanentReference() {
				t.Fatalf("DamageSource = %+v, want EventPermanentReference", damage.DamageSource)
			}
		})
	}
}

// TestLowerGroupDamageEventPermanentSourceInTrigger verifies that the shared
// group-damage-lowering path handles "It deals X damage to each {group}."
// bodies in non-zone-change trigger shells, preserving DamageSource/LKI.
func TestLowerGroupDamageEventPermanentSourceInTrigger(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		oracle          string
		wantPlayerGroup bool
	}{
		{
			name:            "each opponent",
			oracle:          "Whenever a creature enters, it deals 1 damage to each opponent.",
			wantPlayerGroup: true,
		},
		{
			name:            "each player",
			oracle:          "Whenever a creature enters, it deals 1 damage to each player.",
			wantPlayerGroup: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Devil",
				Layout:     "normal",
				TypeLine:   "Creature — Devil",
				OracleText: tc.oracle,
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
			}
			mode := face.TriggeredAbilities[0].Content.Modes[0]
			damage, ok := mode.Sequence[0].Primitive.(game.Damage)
			if !ok {
				t.Fatalf("primitive = %T, want game.Damage", mode.Sequence[0].Primitive)
			}
			if !damage.DamageSource.Exists || damage.DamageSource.Val != game.EventPermanentReference() {
				t.Fatalf("DamageSource = %+v, want EventPermanentReference", damage.DamageSource)
			}
			_, isPlayerGroup := damage.Recipient.PlayerGroupReference()
			if tc.wantPlayerGroup && !isPlayerGroup {
				t.Fatal("Recipient has no PlayerGroupReference, want player group recipient")
			}
		})
	}
}

// TestLowerDamageEventPermanentSourceFailsClosed verifies that the damage
// lowerer fails closed when the body uses "It deals" but the source reference
// is not ReferenceBindingEventPermanent, and when the text is wrong.
func TestLowerDamageEventPermanentSourceFailsClosed(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		// "It deals" without a bound event-permanent source cannot lower.
		"{1}: It deals 1 damage to any target.",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Test Devil",
				Layout:     "normal",
				TypeLine:   "Creature — Devil",
				OracleText: oracleText,
				Power:      new("2"),
				Toughness:  new("2"),
			}, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) == 0 {
				t.Fatalf("expected diagnostic for %q but none produced", oracleText)
			}
		})
	}
}

func TestLowerManaParameterizedKeywords(t *testing.T) {
	t.Parallel()

	kicker := lowerKeywordForTest(t, "Kicker {1}{G}", game.Kicker)
	kickerKeyword, ok := kicker.(game.KickerKeyword)
	if !ok || kickerKeyword.Cost.String() != "{1}{G}" {
		t.Fatalf("Kicker keyword = %#v, want {1}{G}", kicker)
	}

	madness := lowerKeywordForTest(t, "Madness {2}{B}", game.Madness)
	madnessKeyword, ok := madness.(game.MadnessKeyword)
	if !ok || madnessKeyword.Cost.String() != "{2}{B}" {
		t.Fatalf("Madness keyword = %#v, want {2}{B}", madness)
	}

	morph := lowerKeywordForTest(t, "Morph {3}{U}", game.Morph)
	morphKeyword, ok := morph.(game.MorphKeyword)
	if !ok || morphKeyword.Cost.String() != "{3}{U}" {
		t.Fatalf("Morph keyword = %#v, want {3}{U}", morph)
	}

	disguise := lowerKeywordForTest(t, "Disguise {4}{W}", game.Disguise)
	disguiseKeyword, ok := disguise.(game.DisguiseKeyword)
	if !ok || disguiseKeyword.Cost.String() != "{4}{W}" {
		t.Fatalf("Disguise keyword = %#v, want {4}{W}", disguise)
	}
}

func TestLowerToxicKeyword(t *testing.T) {
	t.Parallel()
	keyword := lowerKeywordForTest(t, "Toxic 2", game.Toxic)
	toxic, ok := keyword.(game.ToxicKeyword)
	if !ok || toxic.Amount != 2 {
		t.Fatalf("Toxic keyword = %#v, want amount 2", keyword)
	}
}

func TestLowerParameterizedKeywordRejectsVariableCost(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Variable Morph",
		Layout:     "normal",
		TypeLine:   "Creature — Test",
		OracleText: "Morph {X}{U}",
		Power:      new("2"),
		Toughness:  new("2"),
	}, "v")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 || diagnostics[0].Summary != "unsupported parameterized keyword" {
		t.Fatalf("diagnostics = %#v, want unsupported parameterized keyword", diagnostics)
	}
}
