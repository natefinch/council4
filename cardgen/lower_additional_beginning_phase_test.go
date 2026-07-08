package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerAdditionalBeginningPhaseSphinx proves the trailing "there is an
// additional beginning phase after this phase." wording (Sphinx of the Second
// Sun) lowers to an AddExtraPhases that queues only an extra beginning phase.
func TestLowerAdditionalBeginningPhaseSphinx(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Sphinx of the Second Sun",
		Layout:     "normal",
		TypeLine:   "Creature — Sphinx",
		ManaCost:   "{6}{U}{U}",
		Power:      new("6"),
		Toughness:  new("6"),
		OracleText: "Flying\nAt the beginning of each of your postcombat main phases, there is an additional beginning phase after this phase.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	seq := face.TriggeredAbilities[0].Content.Modes[0].Sequence
	if len(seq) != 1 {
		t.Fatalf("sequence = %#v, want one instruction", seq)
	}
	extra, ok := seq[0].Primitive.(game.AddExtraPhases)
	if !ok {
		t.Fatalf("primitive = %T, want game.AddExtraPhases", seq[0].Primitive)
	}
	if !extra.Beginning || extra.Combat || extra.Main {
		t.Fatalf("extra phases = %#v, want Beginning only", extra)
	}
}

// TestLowerCyclonusBothFaces proves that Cyclonus, the Saboteur // Cyclonus,
// Cybertronian Fighter lowers both faces:
//
//   - Front: "Whenever Cyclonus deals combat damage to a player, it connives.
//     Then if Cyclonus's power is 5 or greater, convert it." lowers to a connive
//     followed by a Transform of the source gated on the source's power being 5
//     or greater (the power-gated convert).
//   - Back: "Whenever Cyclonus deals combat damage to a player, convert it. If
//     you do, there is an additional beginning phase after this phase." lowers to
//     a Transform that publishes a result, followed by an AddExtraPhases that
//     queues an extra beginning phase gated on that result ("if you do").
func TestLowerCyclonusBothFaces(t *testing.T) {
	t.Parallel()
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:      "Cyclonus, the Saboteur // Cyclonus, Cybertronian Fighter",
		Layout:    "transform",
		TypeLine:  "Legendary Artifact Creature — Robot // Legendary Artifact — Vehicle",
		ManaCost:  "{2}{U}{B}",
		Power:     new("2"),
		Toughness: new("5"),
		CardFaces: []ScryfallCardFace{
			{
				Name:       "Cyclonus, the Saboteur",
				TypeLine:   "Legendary Artifact Creature — Robot",
				ManaCost:   "{2}{U}{B}",
				OracleText: "More Than Meets the Eye {5}{U}{B} (You may cast this card converted for {5}{U}{B}.)\nFlying\nWhenever Cyclonus deals combat damage to a player, it connives. Then if Cyclonus's power is 5 or greater, convert it. (To have a creature connive, draw a card, then discard a card. If you discarded a nonland card, put a +1/+1 counter on that creature.)",
				Power:      new("2"),
				Toughness:  new("5"),
			},
			{
				Name:       "Cyclonus, Cybertronian Fighter",
				TypeLine:   "Legendary Artifact — Vehicle",
				OracleText: "Living metal (During your turn, this Vehicle is also a creature.)\nFlying\nWhenever Cyclonus deals combat damage to a player, convert it. If you do, there is an additional beginning phase after this phase. (The beginning phase includes the untap, upkeep, and draw steps.)",
				Power:      new("5"),
				Toughness:  new("5"),
			},
		},
	})
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
	if len(faces) != 2 {
		t.Fatalf("faces = %d, want 2", len(faces))
	}

	// Front face: connive, then power-gated convert.
	front := faces[0]
	if len(front.TriggeredAbilities) != 1 {
		t.Fatalf("front triggered abilities = %d, want 1", len(front.TriggeredAbilities))
	}
	frontSeq := front.TriggeredAbilities[0].Content.Modes[0].Sequence
	if len(frontSeq) != 2 {
		t.Fatalf("front sequence = %#v, want two instructions", frontSeq)
	}
	if _, ok := frontSeq[0].Primitive.(game.Connive); !ok {
		t.Fatalf("front instruction 0 = %T, want game.Connive", frontSeq[0].Primitive)
	}
	frontTransform, ok := frontSeq[1].Primitive.(game.Transform)
	if !ok {
		t.Fatalf("front instruction 1 = %T, want game.Transform", frontSeq[1].Primitive)
	}
	if frontTransform.Object != game.SourcePermanentReference() {
		t.Fatalf("front transform object = %#v, want source permanent", frontTransform.Object)
	}
	if !frontSeq[1].Condition.Exists {
		t.Fatal("front transform must be gated by the power condition")
	}
	powerGate := frontSeq[1].Condition.Val.Condition
	if !powerGate.Exists || !powerGate.Val.ObjectMatches.Exists || !powerGate.Val.ObjectMatches.Val.Power.Exists {
		t.Fatalf("front transform gate = %#v, want a power selection", frontSeq[1].Condition.Val)
	}
	if got := powerGate.Val.ObjectMatches.Val.Power.Val.Value; got != 5 {
		t.Fatalf("front transform gate power threshold = %d, want 5", got)
	}

	// Back face: convert (publish), then an additional beginning phase ("if you do").
	back := faces[1]
	if len(back.TriggeredAbilities) != 1 {
		t.Fatalf("back triggered abilities = %d, want 1", len(back.TriggeredAbilities))
	}
	backSeq := back.TriggeredAbilities[0].Content.Modes[0].Sequence
	if len(backSeq) != 2 {
		t.Fatalf("back sequence = %#v, want two instructions", backSeq)
	}
	if _, ok := backSeq[0].Primitive.(game.Transform); !ok {
		t.Fatalf("back instruction 0 = %T, want game.Transform", backSeq[0].Primitive)
	}
	if backSeq[0].PublishResult == "" {
		t.Fatal("back transform must publish a result for the if-you-do gate")
	}
	extra, ok := backSeq[1].Primitive.(game.AddExtraPhases)
	if !ok {
		t.Fatalf("back instruction 1 = %T, want game.AddExtraPhases", backSeq[1].Primitive)
	}
	if !extra.Beginning || extra.Combat || extra.Main {
		t.Fatalf("back extra phases = %#v, want Beginning only", extra)
	}
	if !backSeq[1].ResultGate.Exists {
		t.Fatal("back additional beginning phase must be result-gated on the convert")
	}
	if backSeq[1].ResultGate.Val.Key != backSeq[0].PublishResult {
		t.Fatalf("back gate key = %q, want %q", backSeq[1].ResultGate.Val.Key, backSeq[0].PublishResult)
	}
	if backSeq[1].ResultGate.Val.Succeeded != game.TriTrue {
		t.Fatalf("back gate = %#v, want Succeeded TriTrue", backSeq[1].ResultGate.Val)
	}
}
