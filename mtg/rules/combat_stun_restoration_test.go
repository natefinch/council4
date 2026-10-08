package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
)

func TestCompiledCombatStunUsesOtherCombatantAndItsNextUntap(t *testing.T) {
	def := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Combat Stun", Layout: "normal", TypeLine: "Creature", Power: new("0"), Toughness: new("4"),
		OracleText: "Whenever this creature blocks a creature, tap that creature. That creature doesn't untap during its controller's next untap step.",
	})
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, def)
	other := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	decoy := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	emitEvent(g, game.Event{
		Kind: game.EventBlockerDeclared, Controller: source.Controller,
		PermanentID: source.ObjectID, RelatedPermanentID: other.ObjectID, BlockedAttackerID: other.ObjectID,
	})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("compiled blocker trigger did not fire")
	}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if !other.Tapped || !other.Exerted || source.Tapped || source.Exerted || decoy.Tapped || decoy.Exerted {
		t.Fatal("stun did not isolate the related combatant")
	}
	g.Turn.ActivePlayer = game.Player1
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Draw"}})
	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if !other.Tapped || !other.Exerted {
		t.Fatal("source controller's untap consumed another player's stun")
	}
	g.Turn.ActivePlayer = game.Player2
	for step := range 2 {
		addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Draw"}})
		engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		if other.Tapped != (step == 0) || other.Exerted {
			t.Fatal("related creature did not skip exactly its next untap")
		}
	}
}

func TestCompiledEquipmentConditionTracksActualAttachment(t *testing.T) {
	def := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Equipment Condition", Layout: "normal", TypeLine: "Artifact Creature", Power: new("4"), Toughness: new("4"),
		OracleText: "This creature can't attack or block unless it's equipped.",
	})
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, def)
	source.SummoningSick = false
	decoy := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	equipment := addEquipmentPermanent(g, game.Player1)
	for _, attachment := range []*game.Permanent{nil, decoy, source, nil} {
		detachPermanent(g, equipment)
		if attachment != nil && !attachPermanent(g, equipment, attachment) {
			t.Fatal("equipment attachment failed")
		}
		want := attachment == source
		if canAttackWith(g, source, game.Player1) != want || canBlockWith(g, source, game.Player1) != want {
			t.Fatalf("combat legality with attachment %v did not match source equipment", attachment)
		}
	}
}
