package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

const minscBooOracle = `When Minsc & Boo enters and at the beginning of your upkeep, you may create Boo, a legendary 1/1 red Hamster creature token with trample and haste.
+1: Put three +1/+1 counters on up to one target creature with trample or haste.
−2: Sacrifice a creature. When you do, Minsc & Boo deals X damage to any target, where X is that creature's power. If the sacrificed creature was a Hamster, draw X cards.
Minsc & Boo, Timeless Heroes can be your commander.`

func TestLowerMinscBooTimelessHeroes(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Minsc & Boo, Timeless Heroes",
		Layout:     "normal",
		ManaCost:   "{2}{R}{G}",
		TypeLine:   "Legendary Planeswalker — Minsc",
		OracleText: minscBooOracle,
		Loyalty:    new("3"),
	})
	if !face.CanBeCommander {
		t.Fatal("CanBeCommander = false")
	}
	if len(face.TriggeredAbilities) != 2 {
		t.Fatalf("triggered abilities = %d, want 2", len(face.TriggeredAbilities))
	}
	for i := range face.TriggeredAbilities {
		mode := face.TriggeredAbilities[i].Content.Modes[0]
		if len(mode.Sequence) != 1 ||
			(!face.TriggeredAbilities[i].Optional && !mode.Sequence[0].Optional) {
			t.Fatalf("trigger %d sequence = %#v, want one optional instruction", i, mode.Sequence)
		}
		create, ok := mode.Sequence[0].Primitive.(game.CreateToken)
		if !ok {
			t.Fatalf("trigger %d primitive = %T, want game.CreateToken", i, mode.Sequence[0].Primitive)
		}
		def, ok := create.Source.TokenDefRef()
		if !ok || def.Name != "Boo" ||
			!def.HasSupertype(types.Legendary) ||
			!def.HasSubtype(types.Hamster) ||
			!def.HasKeyword(game.Trample) ||
			!def.HasKeyword(game.Haste) {
			t.Fatalf("trigger %d Boo token = %#v", i, def)
		}
	}
	if len(face.LoyaltyAbilities) != 2 {
		t.Fatalf("loyalty abilities = %d, want 2", len(face.LoyaltyAbilities))
	}
	plus := face.LoyaltyAbilities[0]
	if plus.LoyaltyCost != 1 {
		t.Fatalf("+1 cost = %d", plus.LoyaltyCost)
	}
	plusMode := plus.Content.Modes[0]
	if len(plusMode.Targets) != 1 {
		t.Fatalf("+1 targets = %#v", plusMode.Targets)
	}
	target := plusMode.Targets[0]
	if target.MinTargets != 0 || target.MaxTargets != 1 || !target.Selection.Exists {
		t.Fatalf("+1 target = %#v", target)
	}
	selection := target.Selection.Val
	hasCreatureType := (len(selection.RequiredTypes) == 1 && selection.RequiredTypes[0] == types.Creature) ||
		(len(selection.RequiredTypesAny) == 1 && selection.RequiredTypesAny[0] == types.Creature)
	if !hasCreatureType ||
		len(selection.AnyOf) != 2 ||
		selection.AnyOf[0].Keyword != game.Trample ||
		selection.AnyOf[1].Keyword != game.Haste {
		t.Fatalf("+1 selection = %#v", selection)
	}
	add, ok := plusMode.Sequence[0].Primitive.(game.AddCounter)
	if !ok || add.CounterKind != counter.PlusOnePlusOne || add.Amount.IsDynamic() || add.Amount.Value() != 3 {
		t.Fatalf("+1 primitive = %#v", plusMode.Sequence[0].Primitive)
	}

	minus := face.LoyaltyAbilities[1]
	if minus.LoyaltyCost != -2 {
		t.Fatalf("-2 cost = %d", minus.LoyaltyCost)
	}
	minusMode := minus.Content.Modes[0]
	if len(minusMode.Targets) != 0 || len(minusMode.Sequence) != 2 {
		t.Fatalf("-2 mode = %#v", minusMode)
	}
	sacrifice, ok := minusMode.Sequence[0].Primitive.(game.SacrificePermanents)
	if !ok || !sacrifice.PublishObjectBinding || sacrifice.PublishLinked == "" {
		t.Fatalf("-2 sacrifice = %#v", minusMode.Sequence[0].Primitive)
	}
	reflexive, ok := minusMode.Sequence[1].Primitive.(game.CreateReflexiveTrigger)
	if !ok || !minusMode.Sequence[1].ResultGate.Exists {
		t.Fatalf("-2 reflexive instruction = %#v", minusMode.Sequence[1])
	}
	body := reflexive.Trigger.Content.Modes[0]
	if len(body.Targets) != 1 ||
		body.Targets[0].Allow != game.TargetAllowPermanent|game.TargetAllowPlayer ||
		len(body.Sequence) != 2 {
		t.Fatalf("reflexive body = %#v", body)
	}
	damage, ok := body.Sequence[0].Primitive.(game.Damage)
	if !ok || !damage.Amount.IsDynamic() || !damage.DamageSource.Exists {
		t.Fatalf("reflexive damage = %#v", body.Sequence[0].Primitive)
	}
	draw, ok := body.Sequence[1].Primitive.(game.Draw)
	if !ok || !draw.Amount.IsDynamic() || !body.Sequence[1].Condition.Exists {
		t.Fatalf("reflexive draw = %#v", body.Sequence[1])
	}
	condition := body.Sequence[1].Condition.Val.Condition
	if !condition.Exists || !condition.Val.ObjectMatches.Exists ||
		len(condition.Val.ObjectMatches.Val.SubtypesAny) != 1 ||
		condition.Val.ObjectMatches.Val.SubtypesAny[0] != types.Hamster {
		t.Fatalf("Hamster condition = %#v", body.Sequence[1].Condition.Val)
	}
}
