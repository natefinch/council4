package cardgen

import (
	goparser "go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

const mimicVatOracleText = "Imprint — Whenever a nontoken creature dies, you may exile that card. If you do, return each other card exiled with this artifact to its owner's graveyard.\n{3}, {T}: Create a token that's a copy of a card exiled with this artifact. It gains haste. Exile it at the beginning of the next end step."

func mimicVatScryfallCard() *ScryfallCard {
	return &ScryfallCard{
		Name:       "Mimic Vat",
		Layout:     "normal",
		ManaCost:   "{3}",
		TypeLine:   "Artifact",
		OracleText: mimicVatOracleText,
	}
}

func TestLowerMimicVatUsesReusableLinkedImprintMechanics(t *testing.T) {
	face := lowerSingleFace(t, mimicVatScryfallCard())
	if len(face.TriggeredAbilities) != 1 || len(face.ActivatedAbilities) != 1 {
		t.Fatalf("abilities = %d triggered/%d activated, want 1/1",
			len(face.TriggeredAbilities), len(face.ActivatedAbilities))
	}
	triggered := face.TriggeredAbilities[0]
	if triggered.Trigger.Pattern.Event != game.EventPermanentDied ||
		!triggered.Trigger.Pattern.SubjectSelection.NonToken {
		t.Fatalf("trigger pattern = %#v, want nontoken permanent died", triggered.Trigger.Pattern)
	}
	imprintInstruction := triggered.Content.Modes[0].Sequence[0]
	imprint, ok := imprintInstruction.Primitive.(game.ReplaceLinkedExiledCard)
	if !ok || !imprintInstruction.Optional || imprint.FromZone != zone.Graveyard ||
		imprint.Card.Kind != game.CardReferenceEvent || imprint.LinkID == "" {
		t.Fatalf("imprint instruction = %#v", imprintInstruction)
	}

	activated := face.ActivatedAbilities[0]
	sequence := activated.Content.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("activation sequence length = %d, want 2", len(sequence))
	}
	create, ok := sequence[0].Primitive.(game.CreateToken)
	if !ok || create.PublishLinked == "" || sequence[0].PublishResult == "" {
		t.Fatalf("create-token instruction = %#v", sequence[0])
	}
	copySpec, ok := create.Source.TokenCopy()
	if !ok || copySpec.Source != game.TokenCopySourceLinkedExiledCard ||
		copySpec.LinkID != imprint.LinkID || len(copySpec.AddKeywords) != 1 ||
		copySpec.AddKeywords[0] != game.Haste {
		t.Fatalf("copy spec = %#v", copySpec)
	}
	delayed, ok := sequence[1].Primitive.(game.CreateDelayedTrigger)
	if !ok || !delayed.Trigger.CapturedObjectGroup.Exists ||
		delayed.Trigger.Timing != game.DelayedAtBeginningOfNextEndStep ||
		!sequence[1].ResultGate.Exists {
		t.Fatalf("delayed cleanup instruction = %#v", sequence[1])
	}

	if issues := game.ValidateCardDef(&game.CardDef{CardFace: game.CardFace{
		Name:                 "Mimic Vat",
		Types:                []types.Card{types.Artifact},
		TriggeredAbilities:   face.TriggeredAbilities,
		ActivatedAbilities:   face.ActivatedAbilities,
		ReplacementAbilities: face.ReplacementAbilities,
		StaticAbilities:      nil,
	}}); len(issues) != 0 {
		t.Fatalf("ValidateCardDef() = %#v", issues)
	}
}

func TestGenerateExecutableMimicVatSource(t *testing.T) {
	source, diagnostics, err := GenerateExecutableCardSource(mimicVatScryfallCard(), "m")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.ReplaceLinkedExiledCard",
		"game.TokenCopySourceLinkedExiledCard",
		"AddKeywords: []game.Keyword{game.Haste}",
		"game.DelayedAtBeginningOfNextEndStep",
		"CapturedObjectGroup:",
		"AdditionalCosts: cost.Tap",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
	if _, err := goparser.ParseFile(token.NewFileSet(), "mimic_vat.go", source, goparser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
}
