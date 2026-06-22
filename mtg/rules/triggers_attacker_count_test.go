package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestAttackAlonePatternMatchesOnlySoleAttacker checks that an "attacks alone"
// trigger matches when its source is the only declared attacker and stops
// matching once a second creature joins the attack.
func TestAttackAlonePatternMatchesOnlySoleAttacker(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanent(g, game.Player1)
	pattern := &game.TriggerPattern{
		Event:       game.EventAttackerDeclared,
		Source:      game.TriggerSourceSelf,
		AttackAlone: true,
	}
	event := game.Event{
		Kind:           game.EventAttackerDeclared,
		Controller:     game.Player1,
		PermanentID:    source.ObjectID,
		SourceObjectID: source.ObjectID,
	}
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{{Attacker: source.ObjectID}}}
	if !triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("attacks-alone pattern did not match the sole attacker")
	}

	other := addCombatCreaturePermanent(g, game.Player1)
	g.Combat.Attackers = append(g.Combat.Attackers, game.AttackDeclaration{Attacker: other.ObjectID})
	if triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("attacks-alone pattern matched when two creatures attacked")
	}
}

// TestAttackAlonePatternFailsClosedWithoutCombatState ensures the relation does
// not match when there is no combat state to read the attacker count from.
func TestAttackAlonePatternFailsClosedWithoutCombatState(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanent(g, game.Player1)
	pattern := &game.TriggerPattern{
		Event:       game.EventAttackerDeclared,
		Source:      game.TriggerSourceSelf,
		AttackAlone: true,
	}
	event := game.Event{
		Kind:           game.EventAttackerDeclared,
		Controller:     game.Player1,
		PermanentID:    source.ObjectID,
		SourceObjectID: source.ObjectID,
	}
	if triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("attacks-alone pattern matched with no combat state")
	}
}

// TestAttackerCountAtLeastPatternRequiresThreshold checks that a
// "you attack with two or more creatures" trigger only matches once the
// declared attacker count reaches the threshold.
func TestAttackerCountAtLeastPatternRequiresThreshold(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanent(g, game.Player1)
	first := addCombatCreaturePermanent(g, game.Player1)
	pattern := &game.TriggerPattern{
		Event:                game.EventAttackerDeclared,
		Controller:           game.TriggerControllerYou,
		SubjectSelection:     game.Selection{RequiredTypes: []types.Card{types.Creature}},
		OneOrMore:            true,
		AttackerCountAtLeast: 2,
	}
	event := game.Event{
		Kind:        game.EventAttackerDeclared,
		Controller:  game.Player1,
		PermanentID: first.ObjectID,
	}
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{{Attacker: first.ObjectID}}}
	if triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("attack-with-two-or-more pattern matched with only one attacker")
	}

	second := addCombatCreaturePermanent(g, game.Player1)
	g.Combat.Attackers = append(g.Combat.Attackers, game.AttackDeclaration{Attacker: second.ObjectID})
	if !triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("attack-with-two-or-more pattern did not match with two attackers")
	}
}

// TestAttackerCountAtLeastTriggerFiresOncePerCombat verifies that the relation
// coalesces into a single trigger when several creatures attack together.
func TestAttackerCountAtLeastTriggerFiresOncePerCombat(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:                game.EventAttackerDeclared,
		Controller:           game.TriggerControllerYou,
		SubjectSelection:     game.Selection{RequiredTypes: []types.Card{types.Creature}},
		OneOrMore:            true,
		AttackerCountAtLeast: 2,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	first := addCombatCreaturePermanent(g, game.Player1)
	second := addCombatCreaturePermanent(g, game.Player1)
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{
		{Attacker: first.ObjectID},
		{Attacker: second.ObjectID},
	}}
	batchID := g.IDGen.Next()
	emitEvent(g, game.Event{Kind: game.EventAttackerDeclared, Controller: game.Player1, PermanentID: first.ObjectID, SimultaneousID: batchID})
	emitEvent(g, game.Event{Kind: game.EventAttackerDeclared, Controller: game.Player1, PermanentID: second.ObjectID, SimultaneousID: batchID})

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("attack-with-two-or-more trigger was not put on stack")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want a single coalesced trigger", got)
	}
}

// TestAttackerCountAtLeastTriggerDoesNotFireBelowThreshold verifies that a
// single attacker does not satisfy a "two or more creatures" relation.
func TestAttackerCountAtLeastTriggerDoesNotFireBelowThreshold(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:                game.EventAttackerDeclared,
		Controller:           game.TriggerControllerYou,
		SubjectSelection:     game.Selection{RequiredTypes: []types.Card{types.Creature}},
		OneOrMore:            true,
		AttackerCountAtLeast: 2,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	only := addCombatCreaturePermanent(g, game.Player1)
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{{Attacker: only.ObjectID}}}
	emitEvent(g, game.Event{Kind: game.EventAttackerDeclared, Controller: game.Player1, PermanentID: only.ObjectID})

	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("attack-with-two-or-more trigger fired with only one attacker")
	}
	if got := g.Stack.Size(); got != 0 {
		t.Fatalf("stack size = %d, want no trigger", got)
	}
}

// TestBattalionSelfSourceAttackerCountMatchesSourceAttackingWithOthers covers
// the Battalion relation "Whenever this creature and at least two other
// creatures attack": a self-source attacker-declared pattern with
// AttackerCountAtLeast = 3 matches when the source is declared as an attacker
// alongside at least two other creatures, and fails to match while fewer than
// three creatures are attacking.
func TestBattalionSelfSourceAttackerCountMatchesSourceAttackingWithOthers(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanent(g, game.Player1)
	pattern := &game.TriggerPattern{
		Event:                game.EventAttackerDeclared,
		Source:               game.TriggerSourceSelf,
		AttackerCountAtLeast: 3,
	}
	event := game.Event{
		Kind:           game.EventAttackerDeclared,
		Controller:     game.Player1,
		PermanentID:    source.ObjectID,
		SourceObjectID: source.ObjectID,
	}

	// Source attacking alone: below threshold, no match.
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{{Attacker: source.ObjectID}}}
	if triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("battalion pattern matched with only the source attacking")
	}

	// Source plus one other: still below the three-attacker threshold.
	second := addCombatCreaturePermanent(g, game.Player1)
	g.Combat.Attackers = append(g.Combat.Attackers, game.AttackDeclaration{Attacker: second.ObjectID})
	if triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("battalion pattern matched with only two attackers")
	}

	// Source plus two others: threshold met, match.
	third := addCombatCreaturePermanent(g, game.Player1)
	g.Combat.Attackers = append(g.Combat.Attackers, game.AttackDeclaration{Attacker: third.ObjectID})
	if !triggerMatchesEvent(g, source, pattern, event) {
		t.Fatal("battalion pattern did not match the source attacking with two others")
	}
}
