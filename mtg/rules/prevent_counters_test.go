package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
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
