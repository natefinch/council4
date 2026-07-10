package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestGenerateUltraMagnusTacticianFront proves the front face lowers its attack
// trigger into the optional "you may put an artifact creature card from your
// hand onto the battlefield tapped and attacking" put (mechanic #1: an
// artifact-creature-filtered hand selection with the entering-tapped-and-
// attacking riders, wrapped in an optional "may") followed by the gated
// "If you do, convert Ultra Magnus at end of combat" delayed self-convert
// (mechanic #2: a delayed end-of-combat trigger that transforms the source,
// scheduled only when the optional put actually happened).
func TestGenerateUltraMagnusTacticianFront(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:      "Ultra Magnus, Tactician",
		Layout:    "transform",
		TypeLine:  "Legendary Artifact Creature — Robot",
		ManaCost:  "{2}{R}{G}{W}",
		Power:     new("7"),
		Toughness: new("7"),
		OracleText: "More Than Meets the Eye {2}{R}{G}{W} (You may cast this card converted for {2}{R}{G}{W}.)\n" +
			"Ward {2}\n" +
			"Whenever Ultra Magnus attacks, you may put an artifact creature card from your hand onto the battlefield tapped and attacking. If you do, convert Ultra Magnus at end of combat.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	ability := face.TriggeredAbilities[0]
	if ability.Trigger.Type != game.TriggerWhenever ||
		ability.Trigger.Pattern.Event != game.EventAttackerDeclared ||
		ability.Trigger.Pattern.Source != game.TriggerSourceSelf {
		t.Fatalf("trigger = %#v, want whenever this attacks", ability.Trigger)
	}
	mode := ability.Content.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence len = %d, want 2", len(mode.Sequence))
	}

	// Mechanic #1: optional artifact-creature put-from-hand tapped and attacking.
	put := mode.Sequence[0]
	choose, ok := put.Primitive.(game.ChooseFromZone)
	if !ok {
		t.Fatalf("sequence[0] = %T, want game.ChooseFromZone", put.Primitive)
	}
	if choose.SourceZone != zone.Hand {
		t.Fatalf("choose source zone = %v, want hand", choose.SourceZone)
	}
	if choose.Destination.Zone != zone.Battlefield {
		t.Fatalf("choose destination = %v, want battlefield", choose.Destination.Zone)
	}
	wantTypes := []types.Card{types.Artifact, types.Creature}
	if len(choose.Filter.RequiredTypes) != len(wantTypes) {
		t.Fatalf("choose filter types = %v, want %v", choose.Filter.RequiredTypes, wantTypes)
	}
	for i, want := range wantTypes {
		if choose.Filter.RequiredTypes[i] != want {
			t.Fatalf("choose filter types = %v, want %v", choose.Filter.RequiredTypes, wantTypes)
		}
	}
	if !choose.Riders.EntersTapped || !choose.Riders.EntersAttacking {
		t.Fatalf("choose riders = %#v, want enters tapped and attacking", choose.Riders)
	}
	if !put.Optional {
		t.Fatal("put instruction is not optional (the \"you may\")")
	}
	if put.PublishResult == "" {
		t.Fatal("put instruction does not publish its result for the \"If you do\" gate")
	}

	// Mechanic #2: gated delayed convert at end of combat.
	convert := mode.Sequence[1]
	delayed, ok := convert.Primitive.(game.CreateDelayedTrigger)
	if !ok {
		t.Fatalf("sequence[1] = %T, want game.CreateDelayedTrigger", convert.Primitive)
	}
	if delayed.Trigger.Timing != game.DelayedAtEndOfCombat {
		t.Fatalf("delayed timing = %v, want DelayedAtEndOfCombat", delayed.Trigger.Timing)
	}
	inner := delayed.Trigger.Content.Modes[0]
	if len(inner.Sequence) != 1 {
		t.Fatalf("delayed sequence len = %d, want 1", len(inner.Sequence))
	}
	transform, ok := inner.Sequence[0].Primitive.(game.Transform)
	if !ok {
		t.Fatalf("delayed primitive = %T, want game.Transform", inner.Sequence[0].Primitive)
	}
	if transform.Object != game.SourceCardPermanentReference() {
		t.Fatalf("delayed transform object = %#v, want SourceCardPermanentReference()", transform.Object)
	}
	gate := convert.ResultGate
	if !gate.Exists {
		t.Fatal("delayed convert has no result gate")
	}
	if gate.Val.Key != put.PublishResult {
		t.Fatalf("gate key = %q, want %q (the put's published result)", gate.Val.Key, put.PublishResult)
	}
	if gate.Val.Succeeded != game.TriTrue {
		t.Fatalf("gate Succeeded = %v, want TriTrue (convert only if the put happened)", gate.Val.Succeeded)
	}
}

// TestGenerateUltraMagnusArmoredCarrierBack proves the back face lowers its
// attack trigger into the grant of indestructible to the attacking creatures
// you control until end of turn (mechanic #3: a keyword grant to the attacking
// battlefield group) followed by the "If those creatures have total power 8 or
// greater, convert Ultra Magnus" conditional self-convert (mechanic #4: a
// total-power-of-a-creature-group condition gating the transform).
func TestGenerateUltraMagnusArmoredCarrierBack(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:      "Ultra Magnus, Armored Carrier",
		Layout:    "transform",
		TypeLine:  "Legendary Artifact — Vehicle",
		Power:     new("4"),
		Toughness: new("7"),
		OracleText: "Living metal (During your turn, this Vehicle is also a creature.)\n" +
			"Haste\n" +
			"Formidable — Whenever Ultra Magnus attacks, attacking creatures you control gain indestructible until end of turn. If those creatures have total power 8 or greater, convert Ultra Magnus.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	ability := face.TriggeredAbilities[0]
	if ability.Trigger.Type != game.TriggerWhenever ||
		ability.Trigger.Pattern.Event != game.EventAttackerDeclared ||
		ability.Trigger.Pattern.Source != game.TriggerSourceSelf {
		t.Fatalf("trigger = %#v, want whenever this attacks", ability.Trigger)
	}
	mode := ability.Content.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence len = %d, want 2", len(mode.Sequence))
	}

	// Mechanic #3: grant indestructible to the attacking creatures you control.
	grant, ok := mode.Sequence[0].Primitive.(game.ApplyContinuous)
	if !ok {
		t.Fatalf("sequence[0] = %T, want game.ApplyContinuous", mode.Sequence[0].Primitive)
	}
	if grant.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("grant duration = %v, want DurationUntilEndOfTurn", grant.Duration)
	}
	if len(grant.ContinuousEffects) != 1 {
		t.Fatalf("grant continuous effects = %d, want 1", len(grant.ContinuousEffects))
	}
	effect := grant.ContinuousEffects[0]
	if effect.Layer != game.LayerAbility {
		t.Fatalf("grant layer = %v, want LayerAbility", effect.Layer)
	}
	if len(effect.AddKeywords) != 1 || effect.AddKeywords[0] != game.Indestructible {
		t.Fatalf("grant keywords = %v, want [Indestructible]", effect.AddKeywords)
	}
	if effect.Group.Selection().Controller != game.ControllerYou ||
		effect.Group.Selection().CombatState != game.CombatStateAttacking {
		t.Fatalf("grant group selection = %#v, want attacking creatures you control", effect.Group.Selection())
	}

	// Mechanic #4: convert gated on the attacking group's total power >= 8.
	convert := mode.Sequence[1]
	transform, ok := convert.Primitive.(game.Transform)
	if !ok {
		t.Fatalf("sequence[1] = %T, want game.Transform", convert.Primitive)
	}
	if transform.Object != game.SourcePermanentReference() {
		t.Fatalf("transform object = %#v, want SourcePermanentReference()", transform.Object)
	}
	if !convert.Condition.Exists || !convert.Condition.Val.Condition.Exists {
		t.Fatalf("convert has no gating condition: %#v", convert.Condition)
	}
	control := convert.Condition.Val.Condition.Val.ControlsMatching
	if !control.Exists {
		t.Fatalf("convert condition is not a controls-matching condition: %#v", convert.Condition.Val.Condition.Val)
	}
	if control.Val.Selection.CombatState != game.CombatStateAttacking {
		t.Fatalf("condition selection = %#v, want attacking creatures", control.Val.Selection)
	}
	total := control.Val.TotalPower
	if !total.Exists {
		t.Fatal("convert condition has no total-power comparison")
	}
	if total.Val.Op != compare.GreaterOrEqual || total.Val.Value != 8 {
		t.Fatalf("total-power comparison = %#v, want >= 8", total.Val)
	}
}
