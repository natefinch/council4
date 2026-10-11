package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

func TestLowerEventCardRecipientOwnedClause(t *testing.T) {
	for _, test := range []struct {
		name, text string
		explicit   bool
	}{
		{"opponent put", "Whenever an opponent sacrifices a nontoken permanent, put that card onto the battlefield under your control.", true},
		{"opponent default", "Whenever an opponent sacrifices a nontoken permanent, put that card onto the battlefield.", false},
		{"self return", "When this creature dies, return it to the battlefield under your control.", true},
		{"self owner", "When this creature dies, return it to the battlefield under its owner's control.", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			face := lowerSingleFace(t, &ScryfallCard{
				Name: "Event Recipient", Layout: "normal", TypeLine: "Creature",
				Power: new("1"), Toughness: new("1"), OracleText: test.text,
			})
			put, ok := face.TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.PutOnBattlefield)
			if !ok || put.Source != game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceEvent}) {
				t.Fatal("event body did not preserve exact proved event card")
			}
			if put.Recipient.Exists != test.explicit ||
				test.explicit && put.Recipient.Val != game.ControllerReference() {
				t.Fatalf("recipient=%+v want explicit controller=%v", put.Recipient, test.explicit)
			}
		})
	}
}
func TestLowerEventCardRecipientRefusesUnownedSubjects(t *testing.T) {
	const text = "Whenever an opponent sacrifices a nontoken permanent, put that card onto the battlefield under your control."
	compile := func() (parser.Document, compiler.AbilityContent) {
		document, diagnostics := parser.Parse(text, parser.Context{CardName: "Event Recipient"})
		if len(diagnostics) != 0 {
			t.Fatal(diagnostics)
		}
		compilation, diagnostics := compiler.Compile(document, compiler.Context{})
		if len(diagnostics) != 0 {
			t.Fatal(diagnostics)
		}
		return document, compilation.Abilities[0].Content
	}
	for _, corruption := range []string{"missing proof", "foreign proof", "wrong domain"} {
		t.Run(corruption, func(t *testing.T) {
			document, content := compile()
			switch corruption {
			case "missing proof":
				content.References[0].Subject = compiler.ReferenceSubjectProof{}
			case "foreign proof":
				_, foreign := compile()
				content.References[0] = foreign.References[0]
			case "wrong domain":
				content.References[0].Binding = compiler.ReferenceBindingSource
			default:
				t.Fatal("unknown corruption")
			}
			ctx := contentCtx{content: content, text: text, span: document.Abilities[0].Span,
				enclosingKind: compiler.AbilityTriggered}
			if _, diagnostic := lowerContent("Event Recipient", ctx, &document.Abilities[0]); diagnostic == nil {
				t.Fatal("unowned event subject acquired explicit recipient")
			}
		})
	}
}

func TestLowerEventCardRecipientRejectsUnmodeledControl(t *testing.T) {
	for _, control := range []string{
		" under target opponent's control",
		" under that player's control",
		" under your control or its owner's control",
		" under your control and an opponent's control",
	} {
		t.Run(control, func(t *testing.T) {
			assertCardUnsupported(t, &ScryfallCard{
				Name: "Event Recipient", Layout: "normal", TypeLine: "Creature",
				Power: new("1"), Toughness: new("1"),
				OracleText: "Whenever an opponent sacrifices a nontoken permanent, put that card onto the battlefield" + control + ".",
			})
		})
	}
}

func TestLowerEventCardEntryParametersAcrossVerbs(t *testing.T) {
	for _, verb := range []string{"put that card onto", "return that card to"} {
		for _, rider := range []struct {
			text        string
			transformed bool
			counter     bool
		}{
			{" tapped under your control", false, false},
			{" transformed under your control", true, false},
			{" under your control with a +1/+1 counter on it", false, true},
		} {
			t.Run(verb+rider.text, func(t *testing.T) {
				text := "Whenever an opponent sacrifices a nontoken permanent, " + verb + " the battlefield" + rider.text + "."
				face := lowerSingleFace(t, &ScryfallCard{
					Name: "Event Recipient", Layout: "normal", TypeLine: "Creature",
					Power: new("1"), Toughness: new("1"),
					OracleText: text,
				})
				put, ok := face.TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.PutOnBattlefield)
				if !ok || put.Source != game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceEvent}) ||
					!put.Recipient.Exists || put.Recipient.Val != game.ControllerReference() ||
					put.EntryTapped != (rider.text == " tapped under your control") ||
					put.EntryTransformed != rider.transformed ||
					(len(put.EntryCounters) == 1) != rider.counter {
					t.Fatalf("entry parameters not consumed: %+v", put)
				}
				if rider.counter && (put.EntryCounters[0].Kind != counter.PlusOnePlusOne ||
					put.EntryCounters[0].Amount != 1) {
					t.Fatalf("entry counter kind/amount not preserved: %+v", put.EntryCounters)
				}
			})
		}
	}
}

func TestLowerEventCardPutRefusesUnrepresentedEntryRiders(t *testing.T) {
	for _, rider := range []string{
		" face down under your control",
		" under your control with a charge counter on it",
		" transformed under your control (front face up)",
	} {
		t.Run(rider, func(t *testing.T) {
			assertCardUnsupported(t, &ScryfallCard{
				Name: "Event Recipient", Layout: "normal", TypeLine: "Creature",
				Power: new("1"), Toughness: new("1"),
				OracleText: "Whenever an opponent sacrifices a nontoken permanent, put that card onto the battlefield" + rider + ".",
			})
		})
	}
}
