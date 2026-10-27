// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestRTRemovedCounterQuantitySemantic(t *testing.T) {
	seq := RTCounterQuantity().ActivatedAbilities[0].Content.Modes[0].Sequence
	remove := seq[0].Primitive.(game.RemoveCounter)
	draw := seq[1].Primitive.(game.Draw)
	if remove.AllKinds || remove.CounterKind != counter.Charge ||
		remove.Amount.DynamicAmount().Val.Kind != game.DynamicAmountObjectCounters ||
		seq[0].PublishResult == "" || !seq[1].ResultGate.Val.AmountAvailable ||
		draw.Amount.DynamicAmount().Val.ResultKey != seq[0].PublishResult {
		t.Fatal("named-kind actual quantity publication did not round-trip")
	}
	seq = RTCounterQuantitySpell().SpellAbility.Val.Modes[0].Sequence
	remove = seq[0].Primitive.(game.RemoveCounter)
	draw = seq[1].Primitive.(game.Draw)
	if remove.AllKinds || remove.CounterKind != counter.PlusOnePlusOne ||
		!seq[1].ResultGate.Val.AmountAvailable ||
		draw.Amount.DynamicAmount().Val.ResultKey != seq[0].PublishResult {
		t.Fatal("spell scalar availability did not round-trip")
	}
	ability := Ashling().ActivatedAbilities[0]
	seq = ability.Content.Modes[0].Sequence
	remove = seq[1].Primitive.(game.RemoveCounter)
	if !ability.CountsResolutionsThisTurn || len(seq) != 4 ||
		remove.AllKinds || remove.CounterKind != counter.PlusOnePlusOne ||
		seq[1].Condition.Val.Condition.Val.SourceAbilityResolutionOrdinalThisTurn != 3 ||
		seq[1].PublishCondition == "" {
		t.Fatal("ordinal named-kind removal did not round-trip")
	}
	for _, instruction := range seq[2:] {
		damage := instruction.Primitive.(game.Damage)
		if instruction.ConditionGate != seq[1].PublishCondition ||
			!instruction.ResultGate.Val.AmountAvailable ||
			instruction.ResultGate.Val.Key != seq[1].PublishResult ||
			damage.Amount.DynamicAmount().Val.ResultKey != seq[1].PublishResult {
			t.Fatal("expanded damage scalar and condition identities did not round-trip")
		}
	}
}

func TestRTNumericConditionSemantic(t *testing.T) {
	past := RTCounteredSpellCondition().SpellAbility.Val.Modes[0].Sequence[1].Condition.Val.Condition.Val
	if !past.UseCounteredSpellManaValue ||
		past.Object.Val.Kind() != game.ObjectReferenceTargetStackObject ||
		past.ObjectMatches.Val.ManaValue.Val.Op != compare.LessOrEqual ||
		past.ObjectMatches.Val.ManaValue.Val.Value != 3 {
		t.Fatal("past target-spell information did not round-trip")
	}
	event := RTEventSpellCondition().TriggeredAbilities[0].Content.Modes[0].Sequence[1].Condition.Val.Condition.Val
	if event.UseCounteredSpellManaValue || event.Object.Val.Kind() != game.ObjectReferenceEventStackObject ||
		event.ObjectMatches.Val.ManaValue.Val.Op != compare.GreaterOrEqual ||
		event.ObjectMatches.Val.ManaValue.Val.Value != 5 {
		t.Fatal("exact event-spell numeric reference did not round-trip")
	}
	returnedCompetitor := RTSpellAfterReturn().TriggeredAbilities[0].Content.Modes[0].Sequence[2].Condition.Val.Condition.Val
	if returnedCompetitor.Object.Val.Kind() != game.ObjectReferenceEventStackObject ||
		returnedCompetitor.ObjectMatches.Val.ManaValue.Val.Value != 5 {
		t.Fatal("returned permanent captured the event-spell condition")
	}
	unless := RTNumericUnless().TriggeredAbilities[0].Content.Modes[0].Sequence[1].Condition.Val.Condition.Val
	if !unless.Negate || !unless.ControlsMatching.Val.Selection.ExcludeSource ||
		unless.ControlsMatching.Val.Selection.Power.Val.Op != compare.LessOrEqual ||
		unless.ControlsMatching.Val.Selection.Power.Val.Value != 2 {
		t.Fatal("source-excluding numeric Unless did not round-trip")
	}
	seq := RTReachedTargetCard().ActivatedAbilities[0].Content.Modes[0].Sequence
	card := seq[1].Condition.Val.Condition.Val
	if card.Object.Val.Kind() != game.ObjectReferenceTargetCard ||
		card.TargetCardResultKey == "" || card.TargetCardResultKey != seq[0].PublishResult {
		t.Fatal("actual reached-card publication did not round-trip")
	}
}

func TestRTLandSemantic(t *testing.T) {
	if RTLand().CardFace.Name != "RT Land" {
		t.Fatalf("name = %q", RTLand().CardFace.Name)
	}
	if len(RTLand().CardFace.ManaAbilities) != 1 {
		t.Fatalf("mana abilities = %d", len(RTLand().CardFace.ManaAbilities))
	}
	prim := RTLand().CardFace.ManaAbilities[0].Content.Modes[0].Sequence[0].Primitive
	add, ok := prim.(game.AddMana)
	if !ok {
		t.Fatalf("primitive type = %T", prim)
	}
	if add.ManaColor != mana.G {
		t.Fatalf("mana color = %q", add.ManaColor)
	}
}

func TestRTCounterOutcomeSemantic(t *testing.T) {
	sequences := [][]game.Instruction{
		RTCounterOutcome().SpellAbility.Val.Modes[0].Sequence,
		RTCounterActivated().ActivatedAbilities[0].Content.Modes[0].Sequence,
		RTCounterTriggered().TriggeredAbilities[0].Content.Modes[0].Sequence,
		RTCounterModal().SpellAbility.Val.Modes[0].Sequence,
	}
	for i, sequence := range sequences {
		index := 0
		if i == 0 {
			index = 1
			if _, ok := sequence[0].Primitive.(game.Pay); !ok ||
				!sequence[1].ResultGate.Exists ||
				sequence[1].ResultGate.Val.Key != sequence[0].PublishResult {
				t.Fatalf("counter tax wiring = %#v", sequence)
			}
		}
		counter, ok := sequence[index].Primitive.(game.CounterObject)
		if !ok || !counter.ExileInstead || sequence[index].PublishResult == "" ||
			!sequence[index+1].ResultGate.Exists ||
			sequence[index+1].ResultGate.Val.Key != sequence[index].PublishResult {
			t.Fatalf("counter outcome round trip = %#v", sequence)
		}
	}
	if sequences[0][3].ResultGate.Exists {
		t.Fatal("independent counter rider became success-gated")
	}
}

func TestRTFixedPhaseSubjectSemantic(t *testing.T) {
	product := RTCapturedProduct().SpellAbility.Val.Modes[0].Sequence
	create := product[0].Primitive.(game.CreateToken)
	trigger := product[2].Primitive.(game.CreateDelayedTrigger).Trigger
	if create.PublishLinked == "" || !trigger.CapturedObjectGroup.Exists ||
		trigger.CapturedObjectGroup.Val != game.LinkedObjectReference(string(create.PublishLinked)) {
		t.Fatal("actual product group did not round-trip")
	}
	card := RTCapturedCard().SpellAbility.Val.Modes[0].Sequence
	move := card[0].Primitive.(game.MoveCard)
	trigger = card[2].Primitive.(game.CreateDelayedTrigger).Trigger
	if !card[0].ClearLinkedBeforeGate || !move.ReplacePublishedLinked || move.PublishLinked == "" ||
		!trigger.CapturedCard.Exists || trigger.CapturedCard.Val != game.LinkedObjectReference(string(move.PublishLinked)) {
		t.Fatal("transient card publication/capture did not round-trip")
	}
	target := RTCapturedTarget().SpellAbility.Val.Modes[0].Sequence
	trigger = target[3].Primitive.(game.CreateDelayedTrigger).Trigger
	if !trigger.CapturedObject.Exists || trigger.CapturedObject.Val != game.TargetPermanentReference(1) {
		t.Fatal("nonzero captured target occurrence did not round-trip")
	}
	future := RTCapturedFuture().SpellAbility.Val.Modes[0].Sequence
	trigger = future[1].Primitive.(game.CreateDelayedTrigger).Trigger
	condition := trigger.Content.Modes[0].Sequence[0].Condition
	if future[1].Condition.Exists || !condition.Exists ||
		condition.Val.Condition.Val.Object.Val != game.CapturedObjectReference() {
		t.Fatal("future captured-subject condition evaluated at the wrong time")
	}
	choice := RTCapturedFutureChoice().SpellAbility.Val.Modes[0].Sequence
	trigger = choice[1].Primitive.(game.CreateDelayedTrigger).Trigger
	if choice[1].Optional || !trigger.Optional || !trigger.CapturedObject.Exists {
		t.Fatal("future captured-subject choice did not round-trip")
	}
}

func TestRTPublishedSubjectSemantic(t *testing.T) {
	returned := RTReturnedSubject().SpellAbility.Val.Modes[0].Sequence
	put := returned[0].Primitive.(game.PutOnBattlefield)
	want := game.LinkedObjectReference(string(put.PublishLinked))
	if put.PublishLinked == "" || returned[1].Condition.Val.Condition.Val.Object.Val != want ||
		returned[1].Primitive.(game.AddCounter).Object != want {
		t.Fatal("returned subject publication did not round-trip")
	}
	blinked := RTBlinkedSubject().SpellAbility.Val.Modes[0].Sequence
	put = blinked[1].Primitive.(game.PutOnBattlefield)
	input, ok := put.Source.LinkedKey()
	condition := blinked[2].Condition.Val.Condition.Val
	if !ok || put.PublishLinked == "" || input == put.PublishLinked ||
		condition.Object.Val != game.LinkedObjectReference(string(put.PublishLinked)) ||
		len(condition.ObjectMatches.Val.SubtypesAny) != 4 {
		t.Fatal("blink input, output or full subject selection did not round-trip")
	}
	modal := RTModalSubject().SpellAbility.Val.Modes[0].Sequence
	put = modal[1].Primitive.(game.PutOnBattlefield)
	want = game.LinkedObjectReference(string(put.PublishLinked))
	if put.PublishLinked == "" || modal[2].Primitive.(game.Untap).Object != want ||
		modal[2].Condition.Val.Condition.Val.Object.Val != want {
		t.Fatal("modal subject and consequence identity did not round-trip")
	}
}

func TestRTBearSemantic(t *testing.T) {
	if len(RTBear().CardFace.StaticAbilities) != 2 {
		t.Fatalf("static abilities = %d", len(RTBear().CardFace.StaticAbilities))
	}
	keywords := RTBear().CardFace.StaticAbilities[0].KeywordAbilities
	if len(keywords) != 1 {
		t.Fatalf("keyword abilities = %d", len(keywords))
	}
	keyword, ok := keywords[0].(game.SimpleKeyword)
	if !ok || keyword.Kind != game.Flying {
		t.Fatalf("keyword[0] = %#v", keywords[0])
	}
}

func TestRTBoltSemantic(t *testing.T) {
	if !RTBolt().CardFace.SpellAbility.Exists {
		t.Fatal("spell ability missing")
	}
	mode := RTBolt().CardFace.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %d", len(mode.Targets))
	}
	damage, ok := mode.Sequence[0].Primitive.(game.Damage)
	if !ok {
		t.Fatalf("primitive type = %T", mode.Sequence[0].Primitive)
	}
	if damage.Amount.Value() != 3 {
		t.Fatalf("damage amount = %d", damage.Amount.Value())
	}
}

func TestRTConditionGroupSemantic(t *testing.T) {
	seq := RTConditionGroup().SpellAbility.Val.Modes[0].Sequence
	if len(seq) != 2 || !seq[0].Condition.Exists ||
		!seq[0].Condition.Val.Condition.Val.ControllerHandEmpty ||
		seq[0].PublishCondition == "" ||
		seq[1].Condition.Exists || seq[1].ConditionGate != seq[0].PublishCondition {
		t.Fatal("scoped condition evaluation did not round-trip")
	}
}

func TestRTConditionAbilitySemantic(t *testing.T) {
	seq := RTConditionAbility().ActivatedAbilities[0].Content.Modes[0].Sequence
	if len(seq) != 2 || !seq[0].Condition.Exists ||
		seq[0].PublishCondition == "" ||
		seq[1].Condition.Exists || seq[1].ConditionGate != seq[0].PublishCondition {
		t.Fatal("nested scoped condition evaluation did not round-trip")
	}
}

func TestRTOrdinalModesSemantic(t *testing.T) {
	ability := RTOrdinalModes().ActivatedAbilities[0]
	if !ability.CountsResolutionsThisTurn || ability.ActivationCondition.Exists ||
		len(ability.Content.Modes) != 2 {
		t.Fatal("resolution-count shell metadata did not round-trip")
	}
	seq := ability.Content.Modes[0].Sequence
	if len(seq) != 2 || !seq[0].Condition.Exists ||
		seq[0].Condition.Val.Condition.Val.SourceAbilityResolutionOrdinalThisTurn != 2 ||
		seq[0].PublishCondition == "" || seq[1].ConditionGate != seq[0].PublishCondition {
		t.Fatal("modal resolution ordinal and grouped consumer did not round-trip")
	}
}

func TestRTResultCountSemantic(t *testing.T) {
	for _, seq := range [][]game.Instruction{
		RTResultCount().SpellAbility.Val.Modes[0].Sequence,
		RTResultAbility().ActivatedAbilities[0].Content.Modes[0].Sequence,
	} {
		gate := seq[1].ResultGate.Val
		if seq[0].PublishResult == "" || gate.Key != seq[0].PublishResult ||
			!gate.ObjectSelection.Exists || !gate.CardOnly || !gate.ObjectCountRange.Exists ||
			gate.ObjectCountRange.Val.Min != 2 || gate.ObjectCountRange.Val.Max != 2 {
			t.Fatal("actual matching-result count did not round-trip")
		}
	}
	complement := RTResultCount().SpellAbility.Val.Modes[0].Sequence[2].ResultGate.Val
	if !complement.Negate || complement.ObjectCountRange.Val.Min != 2 {
		t.Fatal("count predicate complement did not round-trip")
	}
}

func TestRTOptionalGroupSemantic(t *testing.T) {
	group := RTOptionalGroup().SpellAbility.Val.Modes[0].Sequence
	expanded := RTExpandedOptional().ActivatedAbilities[0].Content.Modes[0].Sequence
	for _, seq := range [][]game.Instruction{group, expanded} {
		if !seq[0].Optional || seq[0].PublishOptionalDecision == "" ||
			seq[1].Optional || seq[1].OptionalDecisionGate != seq[0].PublishOptionalDecision ||
			seq[0].PublishResult != "" || seq[1].ResultGate.Exists {
			t.Fatal("single decision independent of actual results did not round-trip")
		}
	}
	if group[0].PublishCondition == "" || group[1].ConditionGate != group[0].PublishCondition ||
		group[2].Optional || group[2].OptionalDecisionGate != "" || group[2].ConditionGate != "" {
		t.Fatal("conditional optional group swallowed the independent rider")
	}
}

func TestRTOrdinaryManaSemantic(t *testing.T) {
	ability := RTOrdinalMana().ActivatedAbilities[0]
	seq := ability.Content.Modes[0].Sequence
	if !ability.CountsResolutionsThisTurn || ability.ActivationCondition.Exists ||
		len(RTOrdinalMana().ManaAbilities) != 0 || len(seq) != 2 || seq[0].Condition.Exists {
		t.Fatal("ordinary fixed-mana shell did not round-trip")
	}
	add, ok := seq[1].Primitive.(game.AddMana)
	if !ok || add.Amount.Value() != 8 || add.ManaColor != mana.R || !seq[1].Optional ||
		seq[1].Condition.Val.Condition.Val.SourceAbilityResolutionOrdinalThisTurn != 3 {
		t.Fatal("one optional fixed-mana action did not round-trip")
	}
	group := RTGatedMana().SpellAbility.Val.Modes[0].Sequence
	if len(group) != 3 || group[0].PublishCondition == "" ||
		group[1].ConditionGate != group[0].PublishCondition || group[2].ConditionGate != group[0].PublishCondition {
		t.Fatal("expanded fixed-mana group decision did not round-trip")
	}
	for i, color := range []mana.Color{mana.R, mana.G} {
		add, ok := group[i+1].Primitive.(game.AddMana)
		if !ok || add.Amount.Value() != 1 || add.ManaColor != color {
			t.Fatal("fixed-mana color or amount did not round-trip")
		}
	}
}

func TestRTPaidCostSubjectSemantic(t *testing.T) {
	ability := RTSacrificeCostSubject().ActivatedAbilities[0]
	key := ability.AdditionalCosts[0].SubjectKey
	ref := ability.Content.Modes[0].Sequence[1].Condition.Val.Condition.Val.Object.Val
	if key == "" || ref != game.PaidCostReference(key, game.PaidCostSacrifice, types.Creature) {
		t.Fatal("sacrifice component identity, domain and noun did not round-trip")
	}
	card := RTDiscardCostSubject()
	key = card.AdditionalCosts[0].SubjectKey
	condition := card.SpellAbility.Val.Modes[0].Sequence[1].Condition.Val.Condition.Val
	if key == "" || condition.Object.Val != game.PaidCostReference(key, game.PaidCostDiscard, "") || !condition.Negate {
		t.Fatalf("discard cost key=%q reference key=%q negated=%v", key, condition.Object.Val.CostKey(), condition.Negate)
	}
	numeric := RTNumericCostSubject().ActivatedAbilities[0]
	key = numeric.AdditionalCosts[1].SubjectKey
	sequence := numeric.Content.Modes[0].Sequence
	condition = sequence[0].Condition.Val.Condition.Val
	if key == "" || condition.Object.Val != game.PaidCostReference(key, game.PaidCostSacrifice, types.Creature) ||
		condition.ObjectMatches.Val.Toughness.Val != (compare.Int{Op: compare.GreaterOrEqual, Value: 4}) ||
		!condition.Negate || sequence[0].PublishCondition == "" ||
		sequence[1].ConditionGate != sequence[0].PublishCondition || !sequence[1].ConditionGateNegate {
		t.Fatal("numeric paid-cost predicate and exclusive replacement envelope did not round-trip")
	}
}
