package cardgen

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

const nightmareShepherdOracleText = "Flying\nWhenever another nontoken creature you control dies, you may exile it. If you do, create a token that's a copy of that creature, except it's 1/1 and it's a Nightmare in addition to its other types."

func nightmareShepherdCard() *ScryfallCard {
	return &ScryfallCard{
		Name:       "Nightmare Shepherd",
		Layout:     "normal",
		ManaCost:   "{2}{B}{B}",
		TypeLine:   "Enchantment Creature — Demon",
		OracleText: nightmareShepherdOracleText,
		Power:      new("4"),
		Toughness:  new("4"),
	}
}

func TestLowerNightmareShepherdComposesEventExileAndLKICopy(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, nightmareShepherdCard())
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	trigger := face.TriggeredAbilities[0]
	pattern := trigger.Trigger.Pattern
	if pattern.Event != game.EventPermanentDied ||
		pattern.Controller != game.TriggerControllerYou ||
		!pattern.ExcludeSelf ||
		!pattern.SubjectSelection.NonToken ||
		len(pattern.SubjectSelection.RequiredTypes) != 1 ||
		pattern.SubjectSelection.RequiredTypes[0] != types.Creature {
		t.Fatalf("trigger pattern = %#v", pattern)
	}
	sequence := trigger.Content.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("sequence length = %d, want exile then copy", len(sequence))
	}
	move, ok := sequence[0].Primitive.(game.MoveCard)
	if !ok || move.Card.Kind != game.CardReferenceEvent ||
		move.FromZone != zone.Graveyard || move.Destination != zone.Exile ||
		!move.ReplacePublishedLinked ||
		!move.IncludeEventPermanentComponents ||
		!sequence[0].Optional || sequence[0].PublishResult == "" {
		t.Fatalf("exile instruction = %#v", sequence[0])
	}
	create, ok := sequence[1].Primitive.(game.CreateToken)
	if !ok || !sequence[1].ResultGate.Exists ||
		sequence[1].ResultGate.Val.Key != sequence[0].PublishResult ||
		sequence[1].ResultGate.Val.Succeeded != game.TriTrue {
		t.Fatalf("create instruction = %#v", sequence[1])
	}
	spec, ok := create.Source.TokenCopy()
	if !ok || spec.Source != game.TokenCopySourceObject ||
		spec.Object.Kind() != game.ObjectReferenceLinkedObject ||
		!spec.SetPower.Exists || spec.SetPower.Val.Value != 1 ||
		!spec.SetToughness.Exists || spec.SetToughness.Val.Value != 1 ||
		len(spec.AddSubtypes) != 1 || spec.AddSubtypes[0] != types.Nightmare {
		t.Fatalf("copy spec = %#v", spec)
	}
}

func TestGenerateExecutableNightmareShepherdSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(nightmareShepherdCard(), "n")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.EventPermanentDied",
		"ExcludeSelf:",
		"NonToken: true",
		"game.CardReferenceEvent",
		"ReplacePublishedLinked:",
		"IncludeEventPermanentComponents:",
		"game.LinkedObjectReference(\"event-card-exile-copy\")",
		"SetPower:",
		"SetToughness:",
		"types.Nightmare",
		"PublishResult:",
		"ResultGate:",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "nightmare_shepherd.go", source, parser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
}
