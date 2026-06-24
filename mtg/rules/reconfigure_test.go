package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// reconfigureEquipmentCreature builds a test Artifact Creature — Equipment whose
// only activated ability is Reconfigure {G}, mirroring a printed Reconfigure
// card such as Rabbit Battery.
func reconfigureEquipmentCreature() *game.CardDef {
	pt := game.PT{Value: 1}
	return &game.CardDef{CardFace: game.CardFace{Name: "Test Reconfigure Gear",
		Types:              []types.Card{types.Artifact, types.Creature},
		Subtypes:           []types.Sub{types.Equipment},
		Power:              opt.Val(pt),
		Toughness:          opt.Val(pt),
		ActivatedAbilities: []game.ActivatedAbility{game.ReconfigureActivatedAbility(cost.Mana{cost.G})},
	}}
}

// TestReconfigureAbilityUsesStackAndAttachesOnResolution verifies the
// Reconfigure activated ability is dispatched like Equip: it uses the stack and
// attaches the source to the targeted creature you control when it resolves.
func TestReconfigureAbilityUsesStackAndAttachesOnResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	gear := addCombatPermanent(g, game.Player1, reconfigureEquipmentCreature())
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Runeclaw Bear",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2})},
	})
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	act := action.ActivateAbility(gear.ObjectID, 0, []game.Target{game.PermanentTarget(creature.ObjectID)}, 0)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("reconfigure activation was not legal")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(reconfigure) = false, want true")
	}
	if !forest.Tapped {
		t.Fatal("forest was not tapped to pay reconfigure cost")
	}
	if gear.AttachedTo.Exists {
		t.Fatal("gear attached before reconfigure ability resolved")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want 1", g.Stack.Size())
	}
	engine.resolveTopOfStack(g, nil)
	if !gear.AttachedTo.Exists || gear.AttachedTo.Val != creature.ObjectID {
		t.Fatalf("gear attached to = %v, want %v", gear.AttachedTo, creature.ObjectID)
	}
	if !permanentIDsContain(creature.Attachments, gear.ObjectID) {
		t.Fatal("reconfigured creature does not reference the gear")
	}
}

// TestReconfigureAbilityOnlyAsSorceryToCreatureYouControl verifies Reconfigure
// shares Equip's timing and target restrictions: sorcery speed only, and only a
// creature you control.
func TestReconfigureAbilityOnlyAsSorceryToCreatureYouControl(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	gear := addCombatPermanent(g, game.Player1, reconfigureEquipmentCreature())
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Runeclaw Bear",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2})},
	})
	opponentCreature := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Silvercoat Lion",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2})},
	})
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhaseBeginning
	g.Turn.Step = game.StepUpkeep

	if containsAction(engine.legalActions(g, game.Player1), action.ActivateAbility(gear.ObjectID, 0, []game.Target{game.PermanentTarget(creature.ObjectID)}, 0)) {
		t.Fatal("reconfigure activation was legal outside sorcery speed")
	}
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	if containsAction(engine.legalActions(g, game.Player1), action.ActivateAbility(gear.ObjectID, 0, []game.Target{game.PermanentTarget(opponentCreature.ObjectID)}, 0)) {
		t.Fatal("reconfigure activation was legal targeting opponent's creature")
	}
}
