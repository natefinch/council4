package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestDamagePreventionToPlusOneCountersSelfUngated proves the ungated self form
// "If damage would be dealt to this permanent, prevent that damage and put that
// many +1/+1 counters on it." (Anti-Venom): the damage is fully prevented, none
// is marked, and the permanent gains that many +1/+1 counters.
func TestDamagePreventionToPlusOneCountersSelfUngated(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := &game.CardDef{CardFace: game.CardFace{
		Name: "Anti-Venom",
		ReplacementAbilities: []game.ReplacementAbility{
			game.DamagePreventionToPlusOneCountersReplacement("anti-venom", false, opt.V[game.Condition]{}),
		},
	}}
	target := addCombatPermanent(g, game.Player1, def)
	registerPermanentReplacementEffects(g, target)
	sourceID := addColoredSourceCard(g, game.Player2, color.Red)

	if dealt := dealPermanentDamage(g, sourceID, 0, game.Player2, target, 3, false); dealt != 0 {
		t.Fatalf("dealt = %d, want 0 (prevented)", dealt)
	}
	if target.MarkedDamage != 0 {
		t.Fatalf("marked damage = %d, want 0", target.MarkedDamage)
	}
	if got := target.Counters.Get(counter.PlusOnePlusOne); got != 3 {
		t.Fatalf("+1/+1 counters = %d, want 3", got)
	}
}

// TestDamagePreventionRemovesCounterPhantom proves the Phantom self form "If
// damage would be dealt to this creature, prevent that damage. Remove a +1/+1
// counter from this creature." (Phantom Tiger): the damage is fully prevented,
// none is marked, and exactly one +1/+1 counter is removed per prevented event
// regardless of the damage amount, doing nothing once no counter remains.
func TestDamagePreventionRemovesCounterPhantom(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := &game.CardDef{CardFace: game.CardFace{
		Name: "Phantom Tiger",
		ReplacementAbilities: []game.ReplacementAbility{
			game.DamagePreventionRemovesCounterReplacement("phantom", false, opt.V[game.Condition]{}),
		},
	}}
	target := addCombatPermanent(g, game.Player1, def)
	target.Counters.Add(counter.PlusOnePlusOne, 2)
	registerPermanentReplacementEffects(g, target)
	sourceID := addColoredSourceCard(g, game.Player2, color.Red)

	// A 3-damage event is fully prevented and removes exactly one counter (not
	// three), leaving one.
	if dealt := dealPermanentDamage(g, sourceID, 0, game.Player2, target, 3, false); dealt != 0 {
		t.Fatalf("dealt = %d, want 0 (prevented)", dealt)
	}
	if target.MarkedDamage != 0 {
		t.Fatalf("marked damage = %d, want 0", target.MarkedDamage)
	}
	if got := target.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("+1/+1 counters after first hit = %d, want 1", got)
	}

	// A second event removes the last counter.
	target.MarkedDamage = 0
	if dealt := dealPermanentDamage(g, sourceID, 0, game.Player2, target, 2, false); dealt != 0 {
		t.Fatalf("dealt = %d, want 0 (prevented)", dealt)
	}
	if got := target.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("+1/+1 counters after second hit = %d, want 0", got)
	}

	// With no counter left the damage is still prevented and the removal is a
	// no-op.
	target.MarkedDamage = 0
	if dealt := dealPermanentDamage(g, sourceID, 0, game.Player2, target, 5, false); dealt != 0 {
		t.Fatalf("dealt with no counters = %d, want 0 (still prevented)", dealt)
	}
	if got := target.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("+1/+1 counters after third hit = %d, want 0", got)
	}
}

// TestDamagePreventionRemovesCounterPhantomGangBlock proves that when a Phantom
// is dealt combat damage by multiple sources in the same combat damage step
// (gang-blocking), all of the damage is prevented but only one +1/+1 counter is
// removed total. The per-step latch is combat-only and reset each pass, so a
// later pass (double strike) and any noncombat source each remove another
// counter.
func TestDamagePreventionRemovesCounterPhantomGangBlock(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := &game.CardDef{CardFace: game.CardFace{
		Name: "Phantom Tiger",
		ReplacementAbilities: []game.ReplacementAbility{
			game.DamagePreventionRemovesCounterReplacement("phantom", false, opt.V[game.Condition]{}),
		},
	}}
	target := addCombatPermanent(g, game.Player1, def)
	target.Counters.Add(counter.PlusOnePlusOne, 3)
	registerPermanentReplacementEffects(g, target)
	sourceID := addColoredSourceCard(g, game.Player2, color.Red)

	// Two simultaneous combat sources in one step: both prevented, one counter
	// removed total.
	if dealt := dealPermanentDamage(g, sourceID, 0, game.Player2, target, 3, true); dealt != 0 {
		t.Fatalf("dealt from first blocker = %d, want 0 (prevented)", dealt)
	}
	target.MarkedDamage = 0
	if dealt := dealPermanentDamage(g, sourceID, 0, game.Player2, target, 2, true); dealt != 0 {
		t.Fatalf("dealt from second blocker = %d, want 0 (prevented)", dealt)
	}
	if got := target.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("+1/+1 counters after gang block = %d, want 2 (one removed total)", got)
	}

	// A new combat damage pass resets the latch (e.g. double strike), removing
	// another counter.
	target.DamagePreventionCounterRemovedThisStep = false
	target.MarkedDamage = 0
	if dealt := dealPermanentDamage(g, sourceID, 0, game.Player2, target, 4, true); dealt != 0 {
		t.Fatalf("dealt in second pass = %d, want 0 (prevented)", dealt)
	}
	if got := target.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("+1/+1 counters after second pass = %d, want 1", got)
	}

	// Noncombat damage is never latched and always removes a counter.
	target.MarkedDamage = 0
	if dealt := dealPermanentDamage(g, sourceID, 0, game.Player2, target, 1, false); dealt != 0 {
		t.Fatalf("noncombat dealt = %d, want 0 (prevented)", dealt)
	}
	if got := target.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("+1/+1 counters after noncombat = %d, want 0", got)
	}
}

// TestDamagePreventionRemovesCounterPhantomGangBlockEndToEnd drives a real
// gang-block through the combat damage step: a Phantom attacker blocked by two
// creatures has all combat damage prevented and loses exactly one +1/+1 counter.
// The latch is pre-set to true so the assertion also guards the per-pass reset in
// resolveDamagePass: without the reset the stale latch would suppress the removal
// (leaving 3), and without the dedup both blockers would remove one (leaving 1).
func TestDamagePreventionRemovesCounterPhantomGangBlockEndToEnd(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	pt := game.PT{Value: 1}
	def := &game.CardDef{CardFace: game.CardFace{
		Name:      "Phantom Tiger",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
		ReplacementAbilities: []game.ReplacementAbility{
			game.DamagePreventionRemovesCounterReplacement("phantom", false, opt.V[game.Condition]{}),
		},
	}}
	attacker := addCombatPermanent(g, game.Player1, def)
	attacker.Counters.Add(counter.PlusOnePlusOne, 3)
	// A stale latch from a prior step must be cleared before this pass.
	attacker.DamagePreventionCounterRemovedThisStep = true
	registerPermanentReplacementEffects(g, attacker)
	first := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	second := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
		Blockers: []game.BlockDeclaration{
			{Blocker: first.ObjectID, Blocking: attacker.ObjectID},
			{Blocker: second.ObjectID, Blocking: attacker.ObjectID},
		},
		BlockerOrder: map[id.ID][]id.ID{
			attacker.ObjectID: {first.ObjectID, second.ObjectID},
		},
	}

	log := TurnLog{}
	NewEngine(nil).resolveCombatDamage(g, &log)

	if attacker.MarkedDamage != 0 {
		t.Fatalf("Phantom marked damage = %d, want 0 (all prevented)", attacker.MarkedDamage)
	}
	if got := attacker.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("+1/+1 counters after gang block = %d, want 2 (one removed total, stale latch reset)", got)
	}
}

// TestDamagePreventionToPlusOneCountersSelfMonarchGated proves the monarch-gated
// self form "If damage would be dealt to Jared while you're the monarch, prevent
// that damage and put that many +1/+1 counters on it." (Jared Carthalion): the
// prevention (and the counters) apply only while the controller is the monarch.
func TestDamagePreventionToPlusOneCountersSelfMonarchGated(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := &game.CardDef{CardFace: game.CardFace{
		Name: "Jared Carthalion",
		ReplacementAbilities: []game.ReplacementAbility{
			game.DamagePreventionToPlusOneCountersReplacement("jared", false, opt.Val(game.Condition{ControllerIsMonarch: true})),
		},
	}}
	target := addCombatPermanent(g, game.Player1, def)
	registerPermanentReplacementEffects(g, target)
	sourceID := addColoredSourceCard(g, game.Player2, color.Red)

	if dealt := dealPermanentDamage(g, sourceID, 0, game.Player2, target, 2, false); dealt != 2 {
		t.Fatalf("dealt while not monarch = %d, want 2 (unprevented)", dealt)
	}
	if got := target.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("+1/+1 counters while not monarch = %d, want 0", got)
	}

	setMonarch(g, game.Player1)
	target.MarkedDamage = 0
	if dealt := dealPermanentDamage(g, sourceID, 0, game.Player2, target, 4, false); dealt != 0 {
		t.Fatalf("dealt while monarch = %d, want 0 (prevented)", dealt)
	}
	if got := target.Counters.Get(counter.PlusOnePlusOne); got != 4 {
		t.Fatalf("+1/+1 counters while monarch = %d, want 4", got)
	}
}

// TestDamagePreventionToPlusOneCountersAttached proves the attached form "If
// equipped creature would be dealt damage, prevent that damage and put that many
// +1/+1 counters on it." (Panther Habit): damage to the equipped creature is
// prevented and that creature gains that many +1/+1 counters, while the
// Equipment itself and an unattached bystander are unaffected.
func TestDamagePreventionToPlusOneCountersAttached(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	equipped := addCombatCreaturePermanent(g, game.Player1)
	def := &game.CardDef{CardFace: game.CardFace{
		Name: "Panther Habit",
		ReplacementAbilities: []game.ReplacementAbility{
			game.DamagePreventionToPlusOneCountersReplacement("panther habit", true, opt.V[game.Condition]{}),
		},
	}}
	equipment := addCombatPermanent(g, game.Player1, def)
	// Real ordering: the Equipment enters (and registers its replacement) while
	// unattached, then attaches later. The prevention must still apply.
	registerPermanentReplacementEffects(g, equipment)
	equipment.AttachedTo = opt.Val(equipped.ObjectID)
	equipped.Attachments = append(equipped.Attachments, equipment.ObjectID)
	sourceID := addColoredSourceCard(g, game.Player2, color.Red)

	if dealt := dealPermanentDamage(g, sourceID, 0, game.Player2, equipped, 3, false); dealt != 0 {
		t.Fatalf("dealt to equipped creature = %d, want 0 (prevented)", dealt)
	}
	if got := equipped.Counters.Get(counter.PlusOnePlusOne); got != 3 {
		t.Fatalf("equipped creature +1/+1 counters = %d, want 3", got)
	}

	// A creature the Equipment is not attached to is unaffected.
	bystander := addCombatCreaturePermanent(g, game.Player1)
	if dealt := dealPermanentDamage(g, sourceID, 0, game.Player2, bystander, 2, false); dealt != 2 {
		t.Fatalf("dealt to bystander = %d, want 2 (unprevented)", dealt)
	}
	if got := bystander.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("bystander +1/+1 counters = %d, want 0", got)
	}
}
