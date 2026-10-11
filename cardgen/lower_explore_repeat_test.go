package cardgen

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
)

func exploreRepeatCard(text string) *ScryfallCard {
	return &ScryfallCard{
		Name: "Test Scout", Layout: "normal", ManaCost: "{X}{G}",
		TypeLine: "Creature — Merfolk Scout", Power: new("2"), Toughness: new("2"),
		OracleText: text,
	}
}

func TestLowerExploreRepeatSubjects(t *testing.T) {
	t.Parallel()
	for _, subject := range []struct {
		text string
		ref  game.ObjectReference
	}{
		{"it", game.EventPermanentReference()},
		{"this creature", game.SourcePermanentReference()},
		{"Test Scout", game.SourcePermanentReference()},
		{"target creature you control", game.TargetPermanentReference(0)},
	} {
		for _, count := range []string{"one times", "two times", "twice", "zero times", "X times"} {
			t.Run(subject.text+"/"+count, func(t *testing.T) {
				t.Parallel()
				card := exploreRepeatCard("When this creature enters, " + subject.text + " explores " + count + ".")
				face := lowerSingleFace(t, card)
				mode := face.TriggeredAbilities[0].Content.Modes[0]
				repeat, ok := mode.Sequence[0].Primitive.(game.RepeatProcess)
				if !ok {
					t.Fatalf("primitive = %T, want RepeatProcess", mode.Sequence[0].Primitive)
				}
				body := repeat.Body.Modes[0]
				explore, ok := body.Sequence[0].Primitive.(game.Explore)
				if !ok || explore.Creature != subject.ref || len(body.Sequence) != 1 || len(body.Targets) != 0 {
					t.Fatalf("body = %+v, want one original-subject Explore without nested targets", body)
				}
				targeted := subject.ref.Kind() == game.ObjectReferenceTargetPermanent
				if (len(mode.Targets) == 1) != targeted {
					t.Fatalf("targets = %+v, want targeted=%v", mode.Targets, targeted)
				}
				prefix := "TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.RepeatProcess)"
				assertCardPaths(t, card, prefix+".Body.Modes[0].Sequence[0].Primitive.(game.Explore).Creature.kind")
				assertCardPathsAbsent(t, card, prefix+".Body.Modes[0].Targets")
				switch count {
				case "X times":
					assertCardPaths(t, card, prefix+".Times.dynamic.Kind = game.DynamicAmountX")
				case "zero times":
					assertCardPathsAbsent(t, card, prefix+".Times.fixed", prefix+".Times.dynamic")
				default:
					value := 2
					if count == "one times" {
						value = 1
					}
					assertCardPaths(t, card, fmt.Sprintf("%s.Times.fixed = %d", prefix, value))
				}
			})
		}
	}
}

func TestLowerExploreRepeatCompositions(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"Target creature you control explores two times. Draw a card.",
		"Target creature you control explores X times.",
		"Target creature you control explores X times, where X is the number of artifacts you control.",
		"Target creature you control explores two times. Scry 1. You gain 2 life.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{Name: "Test Survey", Layout: "normal", ManaCost: "{X}{G}", TypeLine: "Sorcery", OracleText: text}
			assertCardPaths(t, card,
				"SpellAbility.Val.Modes[0].Targets[0]",
				"Primitive.(game.RepeatProcess).Body.Modes[0].Sequence[0].Primitive.(game.Explore).Creature.kind = game.ObjectReferenceTargetPermanent",
			)
		})
	}
	for _, text := range []string{
		"Repeat the following process two times. You draw a card.",
		"Repeat the following process X times. You gain 2 life.",
		"Repeat the following process two times. This creature explores.",
		"Repeat the following process two times. Target player draws a card.",
		"Repeat the following process X times. Tap target creature.",
		"Repeat the following process two times. Target creature explores.",
		"Repeat the following process two times. Target player draws a card and target creature explores.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{Name: "Test Process", Layout: "normal", ManaCost: "{X}{G}", TypeLine: "Sorcery", OracleText: text}
			if text == "Repeat the following process two times. This creature explores." {
				card = exploreRepeatCard("{1}: " + text)
			}
			assertCardPaths(t, card, "Primitive.(game.RepeatProcess).Body.Modes[0].Sequence[0].Primitive")
		})
	}
}

func TestLowerExploreRepeatRefusesUnsupportedParameters(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"Target creature explores -2 times.",
		"Target creature explores Y times.",
		"Target creature explores up to two times.",
		"Target creature explores X plus one times.",
		"Target creature explores two times and dances.",
		"Two target creatures explore two times.",
		"Target creatures explores two times.",
		"When this creature enters, this creature and another creature explore two times.",
		"Target creature card in your graveyard explores two times.",
		"Target player explores two times.",
		"Each creature you control explores two times.",
		"Target creature that crewed it this turn explores two times.",
		"When this creature enters, they explore two times.",
		"When this creature enters, it explores X times, where X is the number of sandwiches you control.",
		"When this creature enters, it explores X times, where X is the number of cards exiled this way.",
		"Target creature explores X times, where X is the number of creature cards in target player's hand with mana value less than or equal to the number of creatures that player controls.",
		"When this creature enters, it explores X times, where X is the number of creature cards in that player's hand with mana value less than or equal to the number of creatures that player controls.",
		"Repeat the following process two times. They explore.",
		"Repeat the following process two times. Target creature card in your graveyard explores.",
		"Repeat the following process two times. Each creature explores.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			assertCardUnsupported(t, exploreRepeatCard(text))
		})
	}
	for _, amount := range []compiler.CompiledAmount{
		{Known: true, Value: -1}, {RangeKnown: true},
		{VariableX: true, Multiplier: 2}, {VariableX: true, Addend: 1},
		{Known: true, Value: 2, VariableX: true},
		{DynamicKind: compiler.DynamicAmountSourcePower},
		{DynamicKind: compiler.DynamicAmountRemovedCounterCount},
	} {
		if _, ok := repeatProcessQuantity(amount); ok {
			t.Fatalf("unsupported repeat amount accepted: %+v", amount)
		}
	}
}

func TestJadelightSpelunkerWholeDefinition(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name: "Jadelight Spelunker", Layout: "normal", ManaCost: "{X}{G}",
		TypeLine: "Creature — Merfolk Scout", Power: new("1"), Toughness: new("1"),
		OracleText: "When Jadelight Spelunker enters, it explores X times.",
	}
	assertCardPaths(t, card,
		"ManaCost.Val[0] = {X}", "ManaCost.Val[1] = {G}",
		"Power.Val.Value = 1", "Toughness.Val.Value = 1",
		"Trigger.Pattern.Event = game.EventPermanentEnteredBattlefield", "Trigger.Pattern.Source = game.TriggerSourceSelf",
		"TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.RepeatProcess).Times.dynamic.Kind = game.DynamicAmountX",
		"Primitive.(game.Explore).Creature.kind = game.ObjectReferenceEventPermanent",
	)
}
