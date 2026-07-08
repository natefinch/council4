package rules

import (
	"slices"
	"testing"

	cardt "github.com/natefinch/council4/mtg/cards/t"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

// newSpearCombat puts the real The Spear of Bashenga onto the battlefield under
// game.Player1 attached to a Player1 creature, and stages the declare-attackers
// step with Player1 active so the equipped creature can be declared as an
// attacker. The monarch is left unset; callers set it to toggle the "attacks the
// monarch" trigger.
func newSpearCombat(t *testing.T) (g *game.Game, spear, attacker *game.Permanent) {
	t.Helper()
	g = game.NewGame([game.NumPlayers]game.PlayerConfig{})
	spear = addCombatPermanent(g, game.Player1, cardt.TheSpearOfBashenga)
	attacker = addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	if !attachPermanent(g, spear, attacker) {
		t.Fatal("attachPermanent(spear, attacker) = false")
	}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Turn.ActivePlayer = game.Player1
	g.Combat = &game.CombatState{}
	return g, spear, attacker
}

// spearDestroyTargetSpec returns the real destroy-target spec from The Spear of
// Bashenga's "attacks the monarch" triggered ability (index 1).
func spearDestroyTargetSpec(t *testing.T) game.TargetSpec {
	t.Helper()
	content := cardt.TheSpearOfBashenga.TriggeredAbilities[1].Content
	if len(content.Modes) != 1 || len(content.Modes[0].Targets) != 1 {
		t.Fatalf("unexpected Spear ability 3 shape: %+v", content)
	}
	return content.Modes[0].Targets[0]
}

// TestTheSpearOfBashengaAttacksMonarchDestroysTappedNonland drives the full
// engine path: the equipped creature is declared attacking the living monarch,
// the trigger goes on the stack, and it destroys the tapped nonland permanent
// that monarch controls. The single legal target is chosen through the real
// target enumeration.
func TestTheSpearOfBashengaAttacksMonarchDestroysTappedNonland(t *testing.T) {
	g, _, attacker := newSpearCombat(t)
	engine := NewEngine(nil)
	if !setMonarch(g, game.Player2) {
		t.Fatal("setMonarch(Player2) = false")
	}

	// A tapped nonland permanent the attacked monarch controls: the only legal
	// target on the battlefield.
	victim := addCombatPermanent(g, game.Player2, permanentDef("Monarch Artifact", nil, types.Artifact))
	victim.Tapped = true

	attack := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
	}))
	if !engine.applyDeclareAttackers(g, game.Player1, attack) {
		t.Fatal("applyDeclareAttackers() rejected the equipped creature attacking the monarch")
	}

	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}},
	}
	log := TurnLog{}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log) {
		t.Fatal("attacks-the-monarch trigger was not put on the stack")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want 1 (the destroy trigger)", g.Stack.Size())
	}

	engine.resolveTopOfStackWithChoices(g, agents, &log)

	if _, ok := permanentByObjectID(g, victim.ObjectID); ok {
		t.Fatal("tapped nonland permanent the monarch controls was not destroyed")
	}
	if !g.Players[game.Player2].Graveyard.Contains(victim.CardInstanceID) {
		t.Fatal("destroyed target did not enter its owner's graveyard")
	}
}

// TestTheSpearOfBashengaDoesNotFireAttackingNonMonarch proves the trigger is
// gated on the attacked player being the monarch: when the equipped creature
// attacks a player who is not the monarch, no ability goes on the stack.
func TestTheSpearOfBashengaDoesNotFireAttackingNonMonarch(t *testing.T) {
	g, _, attacker := newSpearCombat(t)
	engine := NewEngine(nil)
	// Player3 holds the crown, but the equipped creature attacks Player2.
	if !setMonarch(g, game.Player3) {
		t.Fatal("setMonarch(Player3) = false")
	}

	addCombatPermanent(g, game.Player2, permanentDef("Player2 Artifact", nil, types.Artifact)).Tapped = true

	attack := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
	}))
	if !engine.applyDeclareAttackers(g, game.Player1, attack) {
		t.Fatal("applyDeclareAttackers() rejected the equipped creature attacking a non-monarch")
	}

	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}},
	}
	if engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("attacks-the-monarch trigger fired when attacking a non-monarch player")
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want 0 (no trigger)", g.Stack.Size())
	}
}

// TestTheSpearOfBashengaMonarchGateRequiresLivingMonarch proves the monarch
// trigger relation is gated on a living monarch: a monarch who has been
// eliminated keeps the IsMonarch flag (it is only cleared by setMonarch), but the
// "attacks the monarch" match must ignore that departed monarch (CR 800.4a).
func TestTheSpearOfBashengaMonarchGateRequiresLivingMonarch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	if !setMonarch(g, game.Player2) {
		t.Fatal("setMonarch(Player2) = false")
	}
	if !triggerPlayerMatches(g, game.Player1, game.TriggerPlayerMonarch, game.Player2) {
		t.Fatal("living monarch did not match the monarch trigger relation")
	}
	g.Players[game.Player2].Eliminated = true
	if !g.Players[game.Player2].IsMonarch {
		t.Fatal("eliminated monarch unexpectedly lost the IsMonarch flag")
	}
	if triggerPlayerMatches(g, game.Player1, game.TriggerPlayerMonarch, game.Player2) {
		t.Fatal("eliminated monarch matched the monarch trigger relation")
	}
}

// TestTheSpearOfBashengaTargetLegality proves the destroy target legality through
// the real target enumeration: only a tapped, nonland permanent controlled by the
// attacked monarch is a candidate. An untapped permanent, a land, and a permanent
// another player controls are all excluded, because "that player" resolves to the
// attack's defending player (the monarch), not the attacker.
func TestTheSpearOfBashengaTargetLegality(t *testing.T) {
	g, spear, attacker := newSpearCombat(t)
	engine := NewEngine(nil)
	if !setMonarch(g, game.Player2) {
		t.Fatal("setMonarch(Player2) = false")
	}

	tappedNonland := addCombatPermanent(g, game.Player2, permanentDef("Monarch Artifact", nil, types.Artifact))
	tappedNonland.Tapped = true
	untappedNonland := addCombatPermanent(g, game.Player2, permanentDef("Monarch Untapped Artifact", nil, types.Artifact))
	tappedLand := addCombatPermanent(g, game.Player2, permanentDef("Monarch Land", nil, types.Land))
	tappedLand.Tapped = true
	// A tapped nonland permanent another player (the attacker) controls: illegal
	// because "that player" is the attacked monarch, not the attacker.
	attackerTappedNonland := addCombatPermanent(g, game.Player1, permanentDef("Attacker Artifact", nil, types.Artifact))
	attackerTappedNonland.Tapped = true

	attack := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
	}))
	if !engine.applyDeclareAttackers(g, game.Player1, attack) {
		t.Fatal("applyDeclareAttackers() rejected the equipped creature attacking the monarch")
	}

	var attackEvent game.Event
	for i := len(g.Events) - 1; i >= 0; i-- {
		if g.Events[i].Kind == game.EventAttackerDeclared {
			attackEvent = g.Events[i]
			break
		}
	}
	if attackEvent.Kind != game.EventAttackerDeclared {
		t.Fatal("no EventAttackerDeclared recorded by the real declare-attackers path")
	}

	spec := spearDestroyTargetSpec(t)
	candidates := targetCandidatesForSpecChosenBy(
		g, game.Player1, game.Player1, cardt.TheSpearOfBashenga, spear.ObjectID, attackEvent, &spec)

	if !slices.Contains(candidates, game.PermanentTarget(tappedNonland.ObjectID)) {
		t.Fatalf("candidates = %+v, want the monarch's tapped nonland permanent", candidates)
	}
	for _, illegal := range []struct {
		name string
		id   id.ID
	}{
		{"untapped nonland permanent", untappedNonland.ObjectID},
		{"tapped land", tappedLand.ObjectID},
		{"attacker's tapped nonland permanent", attackerTappedNonland.ObjectID},
	} {
		if slices.Contains(candidates, game.PermanentTarget(illegal.id)) {
			t.Fatalf("candidates included an illegal %s: %+v", illegal.name, candidates)
		}
	}
}
