package cardgen

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestLowerContextualNumericConditions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, typeLine, text, prefix, object, attribute, op string
		index, slot, value                                  int
		past                                                bool
	}{
		{"permanent possessive", "Instant", "Destroy target creature. If that permanent's mana value was 3 or less, draw a card.", "SpellAbility", "TargetPermanent", "ManaValue", "LessOrEqual", 1, 0, 3, false},
		{"bare target power", "Instant", "Destroy target creature. If its power is 4 or greater, draw a card.", "SpellAbility", "TargetPermanent", "Power", "GreaterOrEqual", 1, 0, 4, false},
		{"bare target toughness", "Instant", "Tap target creature. If its toughness is 1 or greater, you gain 1 life.", "SpellAbility", "TargetPermanent", "Toughness", "GreaterOrEqual", 1, 0, 1, false},
		{"nonzero target after rider", "Instant", "Tap target creature. Destroy target creature. You gain 1 life. If its mana value was 3 or less, draw a card.", "SpellAbility", "TargetPermanent", "ManaValue", "LessOrEqual", 3, 1, 3, false},
		{"countered spell", "Instant", "Counter target spell. If that spell's mana value was 3 or less, proliferate.", "SpellAbility", "TargetStackObject", "ManaValue", "LessOrEqual", 1, 0, 3, true},
		{"second countered spell", "Instant", "Counter target spell. Counter target spell. You gain 1 life. If that spell's mana value was 3 or less, you gain 2 life.", "SpellAbility", "TargetStackObject", "ManaValue", "LessOrEqual", 3, 1, 3, true},
		{"targeted card after exile", "Instant", "Exile target card from a graveyard. If its mana value was 3 or less, you gain 1 life.", "SpellAbility", "TargetCard", "ManaValue", "LessOrEqual", 1, 0, 3, false},
		{"live spell", "Instant", "Counter target spell if its mana value is 3 or less.", "SpellAbility", "TargetStackObject", "ManaValue", "LessOrEqual", 0, 0, 3, false},
		{"event permanent", "Creature", "Whenever another creature enters, draw a card. If its toughness is 3 or greater, you gain 1 life.", "TriggeredAbilities[0].Content", "EventPermanent", "Toughness", "GreaterOrEqual", 1, 0, 3, false},
		{"event spell", "Creature", "Whenever you cast an instant or sorcery spell, you gain 1 life. If that spell has mana value 5 or greater, put a +1/+1 counter on this creature.", "TriggeredAbilities[0].Content", "EventStackObject", "ManaValue", "GreaterOrEqual", 1, 0, 5, false},
		{"event spell after returned permanent", "Creature", "Whenever you cast an instant or sorcery spell, return target creature card from your graveyard to the battlefield. You gain 1 life. If that spell has mana value 5 or greater, you gain 2 life.", "TriggeredAbilities[0].Content", "EventStackObject", "ManaValue", "GreaterOrEqual", 2, 0, 5, false},
		{"activated", "Creature", "{1}: Destroy target creature. If its mana value was 3 or less, you gain 1 life.", "ActivatedAbilities[0].Content", "TargetPermanent", "ManaValue", "LessOrEqual", 1, 0, 3, false},
		{"source after target", "Creature", "When this creature enters, tap target creature. This creature gets +1/+1 until end of turn. If its toughness is 3 or greater, you gain 1 life.", "TriggeredAbilities[0].Content", "SourcePermanent", "Toughness", "GreaterOrEqual", 2, 0, 3, false},
		{"has after mutation", "Creature", "At the beginning of combat on your turn, put a +1/+1 counter on target creature you control. Then if that creature has toughness 6 or greater, transform this creature.", "TriggeredAbilities[0].Content", "TargetPermanent", "Toughness", "GreaterOrEqual", 1, 0, 6, false},
		{"modal", "Instant", "Choose one —\n• Destroy target creature. If its mana value was 3 or less, you gain 1 life.\n• Draw a card.", "SpellAbility", "TargetPermanent", "ManaValue", "LessOrEqual", 1, 0, 3, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{Name: "Numeric Comparison", Layout: "normal", TypeLine: tt.typeLine,
				OracleText: tt.text, Power: new("2"), Toughness: new("2")}
			prefix := tt.prefix
			if prefix == "SpellAbility" {
				prefix += ".Val"
			}
			path := fmt.Sprintf("%s.Modes[0].Sequence[%d].Condition.Val.Condition.Val", prefix, tt.index)
			assertCardPaths(t, card,
				path+".Object.Val.kind = game.ObjectReference"+tt.object,
				fmt.Sprintf("%s.ObjectMatches.Val.%s.Val.Op = compare.%s", path, tt.attribute, tt.op),
				fmt.Sprintf("%s.ObjectMatches.Val.%s.Val.Value = %d", path, tt.attribute, tt.value),
			)
			if tt.slot != 0 {
				assertCardPaths(t, card, fmt.Sprintf("%s.Object.Val.targetIndex = %d", path, tt.slot))
			}
			if tt.past {
				assertCardPaths(t, card, path+".UseCounteredSpellManaValue = true")
			} else {
				assertCardPathsAbsent(t, card, path+".UseCounteredSpellManaValue")
			}
			assertCardPathsAbsent(t, card, "ActivationCondition", "Trigger.InterveningIf")
		})
	}
}

func TestLowerNumericActivationRestrictionRemainsAnnouncementOnly(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name: "Numeric Restriction", Layout: "normal", TypeLine: "Creature",
		OracleText: "Ferocious — {1}: Draw a card. Activate only if you control a creature with power 3 or less.",
		Power:      new("2"), Toughness: new("2"),
	}
	prefix := "CardDef.CardFace.ActivatedAbilities[0]."
	assertCardPaths(t, card,
		prefix+"ActivationCondition.Val.ControlsMatching.Val.Selection.Power.Val.Op = compare.LessOrEqual",
		prefix+"ActivationCondition.Val.ControlsMatching.Val.Selection.Power.Val.Value = 3",
		prefix+"Content.Modes[0].Sequence[0].Primitive",
	)
	assertCardPathsAbsent(t, card, prefix+"Content.Modes[0].Sequence[0].Condition", "TriggeredAbilities")
}

func TestContextualNumericConditionsFailClosed(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"Tap target creature. If that spell's mana value was 3 or less, draw a card.",
		"Counter target spell. If that spell's power is 4 or greater, draw a card.",
		"Whenever you cast a creature spell, if that spell's power is 4 or greater, draw a card.",
		"Destroy two target creatures. If its toughness is 3 or greater, draw a card.",
		"Destroy target creature. If its power is X or greater, draw a card.",
		"Destroy target creature. If its power is greater than this creature's power, draw a card.",
		"Reveal the top two cards of your library. If it had mana value 3 or less, draw a card.",
		"Exile target creature. If it had mana value 3 or less, draw a card.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			assertCardUnsupported(t, &ScryfallCard{Name: "Unavailable Numeric Subject", Layout: "normal",
				TypeLine: "Creature", OracleText: "When this creature enters, " + text,
				Power: new("2"), Toughness: new("2")})
		})
	}

	assertCardUnsupported(t, &ScryfallCard{Name: "Unbound Numeric Subject", Layout: "normal",
		TypeLine: "Instant", OracleText: "If its power is 4 or greater, draw a card."})
}

func TestContextualTypeConditionActualLookSubjectPayment(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{Name: "Wand of Denial", Layout: "normal", TypeLine: "Artifact",
		OracleText: "{T}: Look at the top card of target player's library. If it's a nonland card, you may pay 2 life. If you do, put it into that player's graveyard."})
	mode := face.ActivatedAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 1 || len(mode.Sequence) != 3 {
		t.Fatalf("observed payment = %+v, want one player target and three instructions", mode)
	}
	look, ok := mode.Sequence[0].Primitive.(game.LookAtLibraryTop)
	if !ok || look.Player != game.TargetPlayerReference(0) || look.PublishLinked == "" {
		t.Fatalf("observation = %+v, want exact selected player's published card", mode.Sequence[0])
	}
	pay, ok := mode.Sequence[1].Primitive.(game.Pay)
	gate := effectConditionMatch(t, mode.Sequence[1])
	if !ok || !pay.Payment.Payer.Exists || pay.Payment.Payer.Val != game.ControllerReference() ||
		len(pay.Payment.AdditionalCosts) != 1 || pay.Payment.AdditionalCosts[0].Kind != cost.AdditionalPayLife ||
		pay.Payment.AdditionalCosts[0].Amount != 2 || mode.Sequence[1].Optional ||
		gate.Object.Val != game.LinkedObjectReference(string(look.PublishLinked)) ||
		len(gate.ObjectMatches.Val.ExcludedTypes) != 1 || gate.ObjectMatches.Val.ExcludedTypes[0] != types.Land {
		t.Fatalf("payment = %+v, want controller's exact payment gated on actual observed nonland", mode.Sequence[1])
	}
	move, ok := mode.Sequence[2].Primitive.(game.MoveCard)
	result := mode.Sequence[2].ResultGate
	if !ok || move.Card.Kind != game.CardReferenceLinked || move.Card.LinkID != string(look.PublishLinked) ||
		move.FromZone != zone.Library || move.Destination != zone.Graveyard || !result.Exists ||
		mode.Sequence[1].PublishResult == "" || result.Val.Key != mode.Sequence[1].PublishResult || result.Val.Succeeded != game.TriTrue {
		t.Fatalf("movement = %+v, want same observed incarnation after successful payment", mode.Sequence[2])
	}
	assertCardUnsupported(t, &ScryfallCard{Name: "Ambiguous Observed Payment", Layout: "normal", TypeLine: "Artifact",
		OracleText: "{T}: Reveal the top two cards of target player's library. If it's a nonland card, you may pay 2 life. If you do, put it into that player's graveyard."})
}

func TestContextualNumericConditionActualRevealSubject(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{Name: "Observed Mana Value", Layout: "normal", TypeLine: "Creature",
		OracleText: "When this creature enters, reveal the top card of your library. If it had mana value 3 or less, draw a card."})
	sequence := face.TriggeredAbilities[0].Content.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("sequence = %+v, want observation and numeric consumer", sequence)
	}
	reveal, ok := sequence[0].Primitive.(game.Reveal)
	gate := effectConditionMatch(t, sequence[1])
	draw, draws := sequence[1].Primitive.(game.Draw)
	if !ok || reveal.PublishLinked == "" || reveal.Player != game.ControllerReference() || reveal.Amount != game.Fixed(1) ||
		gate.Object.Val != game.LinkedObjectReference(string(reveal.PublishLinked)) ||
		!gate.ObjectMatches.Val.ManaValue.Exists || gate.ObjectMatches.Val.ManaValue.Val.Op != compare.LessOrEqual ||
		gate.ObjectMatches.Val.ManaValue.Val.Value != 3 || !draws || draw.Player != game.ControllerReference() || draw.Amount != game.Fixed(1) {
		t.Fatalf("observation/consumer = %+v / %+v, want exact observed card and <=3 controller draw", sequence[0], sequence[1])
	}
}

func TestLowerSplitskinDollUsesSharedUpperBound(t *testing.T) {
	t.Parallel()
	assertCardPaths(t, &ScryfallCard{Name: "Splitskin Doll", Layout: "normal", TypeLine: "Artifact Creature",
		OracleText: "When this creature enters, draw a card. Then discard a card unless you control another creature with power 2 or less.",
		Power:      new("2"), Toughness: new("1")},
		"Sequence[1].Condition.Val.Condition.Val.Negate = true",
		"Sequence[1].Condition.Val.Condition.Val.ControlsMatching.Val.Selection.ExcludeSource = true",
		"Sequence[1].Condition.Val.Condition.Val.ControlsMatching.Val.Selection.Power.Val.Op = compare.LessOrEqual",
		"Sequence[1].Condition.Val.Condition.Val.ControlsMatching.Val.Selection.Power.Val.Value = 2")
}

func effectConditionMatch(t *testing.T, instr game.Instruction) game.Condition {
	t.Helper()
	if !instr.Condition.Exists || !instr.Condition.Val.Condition.Exists || !instr.Condition.Val.Condition.Val.ObjectMatches.Exists {
		t.Fatalf("instruction = %#v, want a per-effect ObjectMatches gate", instr)
	}
	return instr.Condition.Val.Condition.Val
}
