package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestMaxSpeedTriggeredSafetyRefusal(t *testing.T) {
	for _, body := range []string{
		"At the beginning of combat on your turn, draw a card.",
		"Whenever another creature you control enters, you gain 2 life.",
		"At the beginning of combat on your turn, target creature gains double strike until end of turn.",
		"At the beginning of combat on your turn, target creature or Vehicle you control gains double strike until end of turn.",
		"At the beginning of your upkeep, if you control an artifact, draw a card.",
	} {
		t.Run(body, func(t *testing.T) {
			card := &ScryfallCard{Name: "Max Speed Safety", Layout: "normal", TypeLine: "Creature", Power: new("2"), Toughness: new("2"), OracleText: "Max speed — " + body}
			defs, diagnostics, err := CompileCardDefs(card)
			if err != nil || len(defs) != 0 {
				t.Fatalf("unmodeled triggered availability emitted a whole CardDef: defs=%d err=%v", len(defs), err)
			}
			doc, parseDiagnostics := parser.Parse(card.OracleText, parser.Context{CardName: card.Name})
			if len(parseDiagnostics) != 0 || doc.Abilities[0].AbilityWord == nil {
				t.Fatal("exact Max speed qualifier missing")
			}
			found := false
			for _, diagnostic := range diagnostics {
				if diagnostic.Summary == "unsupported Max speed triggered ability" &&
					diagnostic.Detail == "conditional triggered availability requires copied static ability sources (issue #3263)" &&
					diagnostic.Span == doc.Abilities[0].AbilityWord.Span {
					found = true
				}
			}
			if !found {
				t.Fatalf("missing exact qualifier/runtime-contract diagnostic: %+v", diagnostics)
			}
			assertCardUnsupported(t, card)
		})
	}
}

func TestMaxSpeedSafetyPreservesOrdinaryAndDecorativeTriggers(t *testing.T) {
	for _, body := range []string{
		"At the beginning of your upkeep, draw a card.",
		"Whenever another creature you control enters, you gain 2 life.",
	} {
		base := lowerSingleFace(t, &ScryfallCard{
			Name: "Decorative Trigger", Layout: "normal", TypeLine: "Creature", Power: new("2"), Toughness: new("2"), OracleText: body,
		})
		for _, label := range []string{"Maximum speed", "Max speeds", "Unrecognized qualifier"} {
			face := lowerSingleFace(t, &ScryfallCard{
				Name: "Decorative Trigger", Layout: "normal", TypeLine: "Creature", Power: new("2"), Toughness: new("2"), OracleText: label + " — " + body,
			})
			if len(face.TriggeredAbilities) != 1 || len(face.StaticAbilities) != 0 {
				t.Fatal("decorative label acquired guessed speed semantics")
			}
			actual, expected := face.TriggeredAbilities[0], base.TriggeredAbilities[0]
			actual.Text, expected.Text = "", ""
			if !reflect.DeepEqual(actual, expected) {
				t.Fatal("decorative label changed the ordinary modeled body")
			}
		}
	}
}

func TestMaxSpeedSafetyUnsupportedPayloadsRemainClosed(t *testing.T) {
	for _, text := range []string{
		"Unrecognized qualifier — At the beginning of your upkeep, draw a card and invent an unmodeled action.",
		"Maximum speed — At the beginning of combat on your turn, target spell gains flying until end of turn.",
		"Max speed — When you cycle this card, draw a card.",
	} {
		assertCardUnsupported(t, &ScryfallCard{Name: "Safety Near Miss", Layout: "normal", TypeLine: "Creature", Power: new("2"), Toughness: new("2"), OracleText: text})
	}
}

func TestMaxSpeedStaticAndActivationRemainModeled(t *testing.T) {
	assertCardPaths(t, &ScryfallCard{
		Name: "Static Speed", Layout: "normal", TypeLine: "Creature", Power: new("2"), Toughness: new("2"),
		OracleText: "Max speed — This creature gets +0/+1 and has vigilance.",
	}, "StaticAbilities[0].Condition.Val.ControllerHasMaxSpeed = true")
	assertCardPaths(t, &ScryfallCard{
		Name: "Activated Speed", Layout: "normal", TypeLine: "Creature", Power: new("2"), Toughness: new("2"),
		OracleText: "Max speed — {2}: Draw a card.",
	}, "ActivatedAbilities[0].ActivationCondition.Val.ControllerHasMaxSpeed = true")
}

func TestMaxSpeedRefusalPrecedesSpecialTriggerPaths(t *testing.T) {
	syntax := &parser.Ability{}
	for _, ability := range []compiler.CompiledAbility{
		{Kind: compiler.AbilityTriggered, AbilityWord: "Max speed", OpponentSecondActionTriplet: true},
		{Kind: compiler.AbilityTriggered, AbilityWord: "Max speed", ExactSequence: compiler.ExactSequenceConditionalLookAtTopReveal},
		{Kind: compiler.AbilityTriggered, AbilityWord: "Max speed"},
	} {
		lowered, diagnostic := lowerExecutableAbility("Max Speed Safety", false, nil, -1, ability, syntax)
		if diagnostic == nil || diagnostic.Summary != "unsupported Max speed triggered ability" ||
			!reflect.DeepEqual(lowered, abilityLowering{}) {
			t.Fatal("special trigger path bypassed known unsupported qualifier")
		}
	}
}

func TestMaxSpeedNewCompositionWholeCardRefuses(t *testing.T) {
	assertCardUnsupported(t, &ScryfallCard{
		Name: "Pride of the Road", Layout: "normal", TypeLine: "Creature", Power: new("3"), Toughness: new("2"),
		OracleText: "Vigilance\nStart your engines!\n" +
			"Max speed — At the beginning of combat on your turn, target creature or Vehicle you control gains double strike until end of turn.",
	})
}
