package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// wolfSkullShamanOracleText is the complete printed Wolf-Skull Shaman rules
// text. The Kinship ability word is a rules-free prefix; the body is the upkeep
// look/reveal/payoff sequence shared by every Kinship card.
const wolfSkullShamanOracleText = "Kinship — At the beginning of your upkeep, you may look at the top card of your library. " +
	"If it shares a creature type with this creature, you may reveal it. " +
	"If you do, create a 2/2 green Wolf creature token."

func TestLowerKinshipLookRevealCreate(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Wolf-Skull Shaman",
		Layout:     "normal",
		TypeLine:   "Creature — Elf Shaman",
		OracleText: wolfSkullShamanOracleText,
	})

	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want one upkeep Kinship sequence", len(face.TriggeredAbilities))
	}
	trigger := face.TriggeredAbilities[0]
	if trigger.Trigger.Type != game.TriggerAt ||
		trigger.Trigger.Pattern.Event != game.EventBeginningOfStep ||
		trigger.Trigger.Pattern.Step != game.StepUpkeep ||
		trigger.Trigger.Pattern.Controller != game.TriggerControllerYou {
		t.Fatalf("trigger = %#v, want controller upkeep", trigger.Trigger)
	}
	if trigger.Optional {
		t.Fatal("trigger should not be optional; the optional choices are on the look and reveal")
	}
	if len(trigger.Content.Modes) != 1 {
		t.Fatalf("trigger modes = %d, want one", len(trigger.Content.Modes))
	}
	sequence := trigger.Content.Modes[0].Sequence
	if len(sequence) != 3 {
		t.Fatalf("sequence length = %d, want look/reveal/create", len(sequence))
	}

	look, ok := sequence[0].Primitive.(game.LookAtLibraryTop)
	if !ok || look.PublishLinked == "" {
		t.Fatalf("sequence[0] = %#v, want LookAtLibraryTop with a published link", sequence[0].Primitive)
	}
	if !sequence[0].Optional {
		t.Fatalf("look instruction = %#v, want optional \"you may look\"", sequence[0])
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
		!condition.Val.Selection.SharesCreatureTypeWithSource ||
		condition.Val.Card.LinkID != string(look.PublishLinked) {
		t.Fatalf("reveal card condition = %#v, want shares-creature-type gate on the looked-at card", sequence[1].CardCondition)
	}

	create, ok := sequence[2].Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("sequence[2] = %#v, want CreateToken", sequence[2].Primitive)
	}
	def, defOK := create.Source.TokenDefRef()
	if !defOK || def.Name != "Wolf" {
		t.Fatalf("create token source = %#v, want a Wolf token definition", create.Source)
	}
	gate := sequence[2].ResultGate
	if !gate.Exists ||
		gate.Val.Key != sequence[1].PublishResult ||
		gate.Val.Accepted != game.TriTrue ||
		gate.Val.Succeeded != game.TriTrue {
		t.Fatalf("create result gate = %#v, want gated on the reveal succeeding", sequence[2].ResultGate)
	}
}

// TestLowerKinshipDamagePayoff confirms the look/reveal prefix is text-blind to
// the trailing payoff: Pyroclast Consul's "deal 2 damage to each creature"
// lowers through the same path as Wolf-Skull Shaman's token creation.
func TestLowerKinshipDamagePayoff(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Pyroclast Consul",
		Layout:   "normal",
		TypeLine: "Creature — Giant Shaman",
		OracleText: "Kinship — At the beginning of your upkeep, you may look at the top card of your library. " +
			"If it shares a creature type with this creature, you may reveal it. " +
			"If you do, Pyroclast Consul deals 2 damage to each creature.",
	})

	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want one upkeep Kinship sequence", len(face.TriggeredAbilities))
	}
	sequence := face.TriggeredAbilities[0].Content.Modes[0].Sequence
	if len(sequence) != 3 {
		t.Fatalf("sequence length = %d, want look/reveal/damage", len(sequence))
	}
	if _, ok := sequence[2].Primitive.(game.Damage); !ok {
		t.Fatalf("sequence[2] = %#v, want Damage payoff", sequence[2].Primitive)
	}
	if !sequence[2].ResultGate.Exists {
		t.Fatalf("damage instruction = %#v, want gated on the reveal succeeding", sequence[2])
	}
}

// TestLowerKinshipFailsClosedOnUnrepresentablePayoff confirms the look/reveal
// prefix fails closed when the trailing payoff does not lower. Leaf-Crowned
// Elder's "you may play that card" references the looked-at card, which the
// per-effect payoff path does not lower, so the whole ability is rejected
// rather than silently dropping the payoff.
func TestLowerKinshipFailsClosedOnUnrepresentablePayoff(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:     "Leaf-Crowned Elder",
		Layout:   "normal",
		TypeLine: "Creature — Treefolk Shaman",
		OracleText: "Kinship — At the beginning of your upkeep, you may look at the top card of your library. " +
			"If it shares a creature type with this creature, you may reveal it. " +
			"If you do, you may play that card without paying its mana cost.",
	})
}
