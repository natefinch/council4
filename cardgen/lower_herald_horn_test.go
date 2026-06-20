package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// heraldsHornOracleText is the complete printed Herald's Horn rules text. The
// category lowers the entry-time creature-type choice, the chosen-type spell
// cost reduction, and the upkeep look/reveal/move sequence together.
const heraldsHornOracleText = "As this artifact enters, choose a creature type.\n" +
	"Creature spells you cast of the chosen type cost {1} less to cast.\n" +
	"At the beginning of your upkeep, look at the top card of your library. " +
	"If it's a creature card of the chosen type, you may reveal it and put it into your hand."

func TestLowerHeraldsHornCategory(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Herald's Horn",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: heraldsHornOracleText,
	})

	if len(face.ReplacementAbilities) != 1 {
		t.Fatalf("replacement abilities = %d, want one entry-time creature-type choice", len(face.ReplacementAbilities))
	}

	if len(face.StaticAbilities) != 1 || len(face.StaticAbilities[0].Body.RuleEffects) != 1 {
		t.Fatalf("static abilities = %#v, want one chosen-type cost modifier", face.StaticAbilities)
	}
	modifier := face.StaticAbilities[0].Body.RuleEffects[0].CostModifier
	if modifier.Kind != game.CostModifierSpell ||
		!modifier.MatchCardType ||
		modifier.CardType != types.Creature ||
		!modifier.ChosenSubtypeFromEntryChoice ||
		modifier.GenericReduction != 1 {
		t.Fatalf("cost modifier = %#v, want chosen creature-type {1} reduction", modifier)
	}

	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want one upkeep look sequence", len(face.TriggeredAbilities))
	}
	trigger := face.TriggeredAbilities[0]
	if trigger.Trigger.Type != game.TriggerAt ||
		trigger.Trigger.Pattern.Event != game.EventBeginningOfStep ||
		trigger.Trigger.Pattern.Step != game.StepUpkeep ||
		trigger.Trigger.Pattern.Controller != game.TriggerControllerYou {
		t.Fatalf("trigger = %#v, want controller upkeep", trigger.Trigger)
	}
	if trigger.Optional {
		t.Fatal("trigger should not be optional; the optional choice is on the reveal")
	}
	if len(trigger.Content.Modes) != 1 {
		t.Fatalf("trigger modes = %d, want one", len(trigger.Content.Modes))
	}
	sequence := trigger.Content.Modes[0].Sequence
	if len(sequence) != 3 {
		t.Fatalf("sequence length = %d, want look/reveal/move", len(sequence))
	}

	look, ok := sequence[0].Primitive.(game.LookAtLibraryTop)
	if !ok || look.PublishLinked == "" {
		t.Fatalf("sequence[0] = %#v, want LookAtLibraryTop with a published link", sequence[0].Primitive)
	}

	reveal, ok := sequence[1].Primitive.(game.Reveal)
	if !ok || reveal.Card.Kind != game.CardReferenceLinked || reveal.Card.LinkID != string(look.PublishLinked) {
		t.Fatalf("sequence[1] = %#v, want Reveal of the looked-at card", sequence[1].Primitive)
	}
	if !sequence[1].Optional || sequence[1].PublishResult == "" {
		t.Fatalf("reveal instruction = %#v, want optional with a published result", sequence[1])
	}
	condition := sequence[1].CardCondition
	if !condition.Exists ||
		condition.Val.ChosenSubtypeFrom != game.EntryTypeChoiceKey ||
		len(condition.Val.Types) != 1 || condition.Val.Types[0] != types.Creature {
		t.Fatalf("reveal card condition = %#v, want creature of the chosen subtype", sequence[1].CardCondition)
	}

	move, ok := sequence[2].Primitive.(game.MoveCard)
	if !ok ||
		move.Card.Kind != game.CardReferenceLinked || move.Card.LinkID != string(look.PublishLinked) ||
		move.FromZone != zone.Library || move.Destination != zone.Hand {
		t.Fatalf("sequence[2] = %#v, want MoveCard library->hand of the looked-at card", sequence[2].Primitive)
	}
	gate := sequence[2].ResultGate
	if !gate.Exists ||
		gate.Val.Key != sequence[1].PublishResult ||
		gate.Val.Accepted != game.TriTrue ||
		gate.Val.Succeeded != game.TriTrue {
		t.Fatalf("move result gate = %#v, want gated on the reveal succeeding", sequence[2].ResultGate)
	}
}

// TestLowerChosenTypeLibraryTopRejectsNonUpkeepTrigger confirms the text-blind
// lowering fails closed when the recognized body is attached to a trigger other
// than the controller's own upkeep.
func TestLowerChosenTypeLibraryTopRejectsNonUpkeepTrigger(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:     "Herald's Horn",
		Layout:   "normal",
		TypeLine: "Artifact",
		OracleText: "As this artifact enters, choose a creature type.\n" +
			"At the beginning of each opponent's upkeep, look at the top card of your library. " +
			"If it's a creature card of the chosen type, you may reveal it and put it into your hand.",
	})
}
