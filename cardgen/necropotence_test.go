package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/zone"
)

// necropotenceOracleText is Necropotence's exact Oracle text, the three-line
// engine this card support adds on top of the shared "Skip your draw step."
// static (#2001).
const necropotenceOracleText = "Skip your draw step.\n" +
	"Whenever you discard a card, exile that card from your graveyard.\n" +
	"Pay 1 life: Exile the top card of your library face down. Put that card into your hand at the beginning of your next end step."

// TestGenerateNecropotence proves the whole card renders end to end (parse →
// compile → lower → render) to the exact typed nodes the runtime consumes: the
// shared skip-draw static, the discard-linked exile trigger, and the pay-life
// activated ability whose face-down top-card exile publishes a link the delayed
// controller-keyed end-step return captures.
func TestGenerateNecropotence(t *testing.T) {
	t.Parallel()
	generatedSourceContains(t, &ScryfallCard{
		Name:       "Necropotence",
		Layout:     "normal",
		ManaCost:   "{B}{B}{B}",
		TypeLine:   "Enchantment",
		OracleText: necropotenceOracleText,
	}, []string{
		"game.SkipDrawStepStaticBody",
		"game.EventCardDiscarded",
		"game.ExileTopOfLibrary{",
		`PublishLinked: game.LinkedKey("delayed-top-card-1")`,
		"FaceDown:      true,",
		"game.CreateDelayedTrigger{",
		"game.DelayedAtBeginningOfYourNextEndStep",
		`CapturedCard: opt.Val(game.LinkedObjectReference("delayed-top-card-1"))`,
		"game.CardReference{Kind: game.CardReferenceCaptured}",
	})
}

// TestLowerNecropotenceDiscardExileTrigger proves the discard clause lowers to a
// mandatory "you discard a card" trigger whose single instruction moves the exact
// discarded card (the event back-reference) from the graveyard to exile.
func TestLowerNecropotenceDiscardExileTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Necropotence Discard Tester",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Whenever you discard a card, exile that card from your graveyard.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	trigger := face.TriggeredAbilities[0]
	if trigger.Trigger.Pattern.Event != game.EventCardDiscarded {
		t.Fatalf("trigger event = %v, want EventCardDiscarded", trigger.Trigger.Pattern.Event)
	}
	if trigger.Trigger.Pattern.Player != game.TriggerPlayerYou {
		t.Fatalf("trigger player = %v, want TriggerPlayerYou", trigger.Trigger.Pattern.Player)
	}
	if trigger.Optional {
		t.Fatal("discard trigger must be mandatory")
	}
	seq := trigger.Content.Modes[0].Sequence
	if len(seq) != 1 {
		t.Fatalf("trigger sequence = %d instructions, want 1", len(seq))
	}
	move, ok := seq[0].Primitive.(game.MoveCard)
	if !ok {
		t.Fatalf("instruction = %T, want game.MoveCard", seq[0].Primitive)
	}
	if move.Card.Kind != game.CardReferenceEvent {
		t.Fatalf("moved card reference = %v, want CardReferenceEvent", move.Card.Kind)
	}
	if move.FromZone != zone.Graveyard || move.Destination != zone.Exile {
		t.Fatalf("move %v->%v, want Graveyard->Exile", move.FromZone, move.Destination)
	}
}

// TestLowerNecropotencePayLifeExileDelayedReturn proves the pay-life ability lowers
// to the two-instruction body: a face-down single-top-card exile of the
// controller's own library that publishes a body-scoped link, followed by a
// controller-keyed "your next end step" delayed trigger that captures that linked
// card and moves it from exile to hand.
func TestLowerNecropotencePayLifeExileDelayedReturn(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Necropotence Pay Life Tester",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Pay 1 life: Exile the top card of your library face down. Put that card into your hand at the beginning of your next end step.",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	ability := face.ActivatedAbilities[0]
	if len(ability.AdditionalCosts) != 1 ||
		ability.AdditionalCosts[0].Kind != cost.AdditionalPayLife ||
		ability.AdditionalCosts[0].Amount != 1 {
		t.Fatalf("ability cost = %+v, want a single Pay 1 life", ability.AdditionalCosts)
	}
	seq := ability.Content.Modes[0].Sequence
	if len(seq) != 2 {
		t.Fatalf("ability body = %d instructions, want 2 (exile, delayed return)", len(seq))
	}

	exile, ok := seq[0].Primitive.(game.ExileTopOfLibrary)
	if !ok {
		t.Fatalf("instruction 0 = %T, want game.ExileTopOfLibrary", seq[0].Primitive)
	}
	if !exile.FaceDown {
		t.Fatal("top-card exile must be face down")
	}
	if exile.Amount != game.Fixed(1) {
		t.Fatalf("exile amount = %v, want exactly one top card", exile.Amount)
	}
	if exile.PublishLinked != game.LinkedKey("delayed-top-card-1") {
		t.Fatalf("exile publish key = %q, want delayed-top-card-1", exile.PublishLinked)
	}

	delayed, ok := seq[1].Primitive.(game.CreateDelayedTrigger)
	if !ok {
		t.Fatalf("instruction 1 = %T, want game.CreateDelayedTrigger", seq[1].Primitive)
	}
	if delayed.Trigger.Timing != game.DelayedAtBeginningOfYourNextEndStep {
		t.Fatalf("delayed timing = %v, want DelayedAtBeginningOfYourNextEndStep", delayed.Trigger.Timing)
	}
	if !delayed.Trigger.CapturedCard.Exists {
		t.Fatal("delayed trigger must capture the exiled card")
	}
	if delayed.Trigger.CapturedCard.Val.Kind() != game.ObjectReferenceLinkedObject ||
		delayed.Trigger.CapturedCard.Val.LinkID() != "delayed-top-card-1" {
		t.Fatalf("captured card reference = %#v, want linked object delayed-top-card-1", delayed.Trigger.CapturedCard.Val)
	}
	returnSeq := delayed.Trigger.Content.Modes[0].Sequence
	if len(returnSeq) != 1 {
		t.Fatalf("delayed return body = %d instructions, want 1", len(returnSeq))
	}
	move, ok := returnSeq[0].Primitive.(game.MoveCard)
	if !ok {
		t.Fatalf("delayed instruction = %T, want game.MoveCard", returnSeq[0].Primitive)
	}
	if move.Card.Kind != game.CardReferenceCaptured {
		t.Fatalf("returned card reference = %v, want CardReferenceCaptured", move.Card.Kind)
	}
	if move.FromZone != zone.Exile || move.Destination != zone.Hand {
		t.Fatalf("return move %v->%v, want Exile->Hand", move.FromZone, move.Destination)
	}
}

// TestLowerNecropotenceDiscardExileRejectsTargetForm proves the discard-exile
// lowering is strict about the non-targeted "that card" back-reference: a targeted
// variant ("exile target card from your graveyard") is not coerced into
// Necropotence's shape but instead lowers through the general targeted-exile path,
// keeping a real target and a CardReferenceTarget move rather than the
// event-back-reference the Necropotence trigger uses.
func TestLowerNecropotenceDiscardExileRejectsTargetForm(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Targeted Discard Exile Tester",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Whenever you discard a card, exile target card from your graveyard.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %d, want 1 (the targeted form keeps its target)", len(mode.Targets))
	}
	move, ok := mode.Sequence[0].Primitive.(game.MoveCard)
	if !ok {
		t.Fatalf("instruction = %T, want game.MoveCard", mode.Sequence[0].Primitive)
	}
	if move.Card.Kind == game.CardReferenceEvent {
		t.Fatal("targeted form was coerced into the non-targeted discard event reference")
	}
	if move.Card.Kind != game.CardReferenceTarget {
		t.Fatalf("moved card reference = %v, want CardReferenceTarget", move.Card.Kind)
	}
}

// TestNecropotenceNearMissesFailClosed proves the strict near-misses around each
// new construct fail closed (report diagnostics and lower no ability) rather than
// being coerced into Necropotence's shape. Each case is Necropotence's own wording
// with a single deviation that must break the match.
func TestNecropotenceNearMissesFailClosed(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		oracleText string
	}{
		{
			// Face-up: the hidden-information return requires the card be exiled
			// face down, so the face-up predecessor cannot be linked.
			name:       "exile top card face up",
			oracleText: "Pay 1 life: Exile the top card of your library. Put that card into your hand at the beginning of your next end step.",
		},
		{
			// Non-controller-keyed "the next end step" is a different timing than
			// the controller-keyed "your next end step".
			name:       "shared next end step timing",
			oracleText: "Pay 1 life: Exile the top card of your library face down. Put that card into your hand at the beginning of the next end step.",
		},
		{
			// Next upkeep is a different delayed timing entirely.
			name:       "next upkeep timing",
			oracleText: "Pay 1 life: Exile the top card of your library face down. Put that card into your hand at the beginning of the next turn's upkeep.",
		},
		{
			// A different destination zone is not the "into your hand" return.
			name:       "return to graveyard",
			oracleText: "Pay 1 life: Exile the top card of your library face down. Put that card into your graveyard at the beginning of your next end step.",
		},
		{
			// Multiple top cards is not the singular "that card" return.
			name:       "multiple top cards",
			oracleText: "Pay 1 life: Exile the top two cards of your library face down. Put those cards into your hand at the beginning of your next end step.",
		},
		{
			// A face-down rider is recognized only for the controller's own
			// library, not a targeted player's.
			name:       "face down on targeted library",
			oracleText: "Pay 1 life: Exile the top card of target player's library face down. Put that card into your hand at the beginning of your next end step.",
		},
		{
			// The exile is from the graveyard where the discarded card rests, not
			// from another zone.
			name:       "discard exile wrong zone",
			oracleText: "Whenever you discard a card, exile that card from your library.",
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
				Name:       "Near Miss " + test.name,
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: test.oracleText,
			})
			if !face.empty() {
				t.Fatalf("near-miss lowered an ability instead of failing closed: %#v", face)
			}
		})
	}
}
