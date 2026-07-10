package rules

import (
	"testing"

	cards "github.com/natefinch/council4/mtg/cards/s"
	"github.com/natefinch/council4/mtg/game"
)

// newSlicerFront puts the real Slicer, Hired Muscle card onto the controller's
// battlefield as its front face so its "At the beginning of each opponent's
// upkeep, you may have that player gain control of Slicer until end of turn. If
// you do, untap Slicer, goad it, and it can't be sacrificed this turn. If you
// don't, convert it." trigger runs through the real resolution path.
func newSlicerFront(g *game.Game, controller game.PlayerID) *game.Permanent {
	permanent := addCombatPermanent(g, controller, cards.SlicerHiredMuscle())
	permanent.Face = game.FaceFront
	return permanent
}

// newSlicerBack puts the real card onto the battlefield already converted to its
// back face, Slicer, High-Speed Antagonist, so its "Whenever Slicer deals combat
// damage to a player, convert it at end of combat." trigger runs through the
// real path.
func newSlicerBack(g *game.Game, controller game.PlayerID) *game.Permanent {
	permanent := addCombatPermanent(g, controller, cards.SlicerHiredMuscle())
	permanent.Face = game.FaceBack
	permanent.Transformed = true
	return permanent
}

// TestSlicerOpponentUpkeepControlGrantAccepted proves the taken arm of the front
// trigger: on an opponent's upkeep the controller accepts the "you may have that
// player gain control of Slicer until end of turn" offer, so control of Slicer
// passes to that opponent and the "If you do" consequences fire — Slicer is
// untapped, goaded, and can't be sacrificed this turn — while Slicer stays on its
// front face (the else-branch convert does not run).
func TestSlicerOpponentUpkeepControlGrantAccepted(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	slicer := newSlicerFront(g, game.Player1)
	slicer.Tapped = true

	g.Turn.ActivePlayer = game.Player2
	emitBeginningOfStepEvent(g, game.StepUpkeep)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("each-opponent upkeep control-grant trigger was not put on the stack")
	}
	// The ability's controller (Player1) accepts the "you may" offer with "Yes".
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &sequencedChoiceAgent{choices: [][]int{{1}}},
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := effectiveController(g, slicer); got != game.Player2 {
		t.Fatalf("effective controller = %v, want Player2 (the opponent whose upkeep it is gained control)", got)
	}
	if slicer.Tapped {
		t.Fatal("Slicer was not untapped by the \"If you do\" consequence")
	}
	if !isGoadedNow(g, slicer) {
		t.Fatal("Slicer was not goaded by the \"If you do\" consequence")
	}
	if !permanentCantBeSacrificed(g, slicer) {
		t.Fatal("Slicer is not shielded from sacrifice by the \"If you do\" consequence")
	}
	if slicer.Face != game.FaceFront || slicer.Transformed {
		t.Fatalf("Slicer face/transformed = %v/%v, want front/false (accepted arm does not convert)", slicer.Face, slicer.Transformed)
	}
}

// TestSlicerOpponentUpkeepControlGrantDeclined proves the else arm of the front
// trigger: declining the "you may" offer leaves control with the source's
// controller and runs "If you don't, convert it", flipping Slicer to its back
// face without untapping, goading, or shielding it.
func TestSlicerOpponentUpkeepControlGrantDeclined(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	slicer := newSlicerFront(g, game.Player1)
	slicer.Tapped = true

	g.Turn.ActivePlayer = game.Player2
	emitBeginningOfStepEvent(g, game.StepUpkeep)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("each-opponent upkeep control-grant trigger was not put on the stack")
	}
	// The ability's controller (Player1) declines the "you may" offer with "No".
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &sequencedChoiceAgent{choices: [][]int{{0}}},
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := effectiveController(g, slicer); got != game.Player1 {
		t.Fatalf("effective controller = %v, want Player1 (declined offer keeps control)", got)
	}
	if slicer.Face != game.FaceBack || !slicer.Transformed {
		t.Fatalf("Slicer face/transformed = %v/%v, want back/true (declined arm converts)", slicer.Face, slicer.Transformed)
	}
	if !slicer.Tapped {
		t.Fatal("Slicer was untapped despite the declined offer")
	}
	if isGoadedNow(g, slicer) {
		t.Fatal("Slicer was goaded despite the declined offer")
	}
	if permanentCantBeSacrificed(g, slicer) {
		t.Fatal("Slicer was shielded from sacrifice despite the declined offer")
	}
}

// TestSlicerBackCombatDamageSchedulesConvert proves the scheduling arm of the
// back trigger: when Slicer deals combat damage to a player, its "convert it at
// end of combat" trigger schedules a delayed end-of-combat self-convert without
// converting immediately.
func TestSlicerBackCombatDamageSchedulesConvert(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	slicer := newSlicerBack(g, game.Player1)

	dealPlayerDamage(g, slicer.CardInstanceID, slicer.ObjectID, game.Player1, game.Player2, 3, true)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("combat-damage convert trigger was not put on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("delayed triggers = %d, want 1 (convert scheduled at end of combat)", len(g.DelayedTriggers))
	}
	if slicer.Face != game.FaceBack || !slicer.Transformed {
		t.Fatal("Slicer converted immediately instead of at end of combat")
	}
}

// TestSlicerBackScheduledConvertFiresAtEndOfCombat proves the timing arm of the
// back trigger: once scheduled, the delayed self-convert stays pending through
// the rest of combat and flips Slicer to its front face at the end-of-combat
// step, reusing the shared delayed-at-end-of-combat infrastructure.
func TestSlicerBackScheduledConvertFiresAtEndOfCombat(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	engine := NewEngine(nil)
	slicer := newSlicerBack(g, game.Player1)

	// Living metal makes the Vehicle a creature on its controller's turn and its
	// haste lets it attack immediately. Slicer attacks the opponent directly and
	// deals combat damage, so its own "Whenever Slicer deals combat damage to a
	// player, convert it at end of combat" trigger schedules the delayed
	// self-convert; that trigger then fires at the end-of-combat step and flips
	// Slicer to its front face, reusing the shared delayed-at-end-of-combat
	// infrastructure.
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	markAttacking(g, game.Player2, slicer)

	combatEngine{engine}.resolveCombatAfterAttackers(g, allFirstLegalAgents(), &TurnLog{})

	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("delayed triggers after end of combat = %d, want 0", len(g.DelayedTriggers))
	}
	converted, ok := permanentByObjectID(g, slicer.ObjectID)
	if !ok {
		t.Fatal("Slicer left the battlefield")
	}
	if converted.Face != game.FaceFront || converted.Transformed {
		t.Fatalf("Slicer face/transformed = %v/%v, want front/false (converts at end of combat)", converted.Face, converted.Transformed)
	}
}
