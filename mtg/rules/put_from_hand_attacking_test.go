package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func putCreatureFromHandAttackingInstruction() *game.Instruction {
	return &game.Instruction{
		Primitive: game.PutFromHandChoice(
			game.ControllerReference(),
			game.Selection{RequiredTypes: []types.Card{types.Creature}},
			game.Fixed(1),
			true, // tapped
			true, // attacking
		),
	}
}

// During the controller's combat, a creature put onto the battlefield from hand
// with the tapped-and-attacking riders enters tapped and is declared attacking a
// defending player (CR 508.4), the Preeminent Captain / Ultra Magnus front
// mechanic.
func TestPutFromHandTappedAndAttackingDuringCombat(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	attacker := addCombatCreaturePermanent(g, game.Player1)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{
			Attacker: attacker.ObjectID,
			Target:   game.AttackTarget{Player: game.Player2},
		}},
		AttackersDeclared: true,
	}
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat

	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Source",
		Types: []types.Card{types.Artifact},
	}})
	obj := triggeredObjFor(source)

	soldier := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Soldier",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Soldier},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})

	agents := [game.NumPlayers]PlayerAgent{game.Player1: defaultChoiceAgent{}}
	engine.resolveInstructionWithChoices(g, obj, putCreatureFromHandAttackingInstruction(), agents, &TurnLog{})

	if g.Players[game.Player1].Hand.Contains(soldier) {
		t.Fatal("chosen creature still in hand")
	}
	permanent, ok := reanimatedPermanent(g, soldier)
	if !ok {
		t.Fatal("creature was not put onto the battlefield")
	}
	if !permanent.Tapped {
		t.Fatal("put creature entered untapped despite the tapped rider")
	}
	if len(g.Combat.Attackers) != 2 {
		t.Fatalf("attackers = %+v, want the original attacker plus the put creature", g.Combat.Attackers)
	}
	last := g.Combat.Attackers[len(g.Combat.Attackers)-1]
	if last.Attacker != permanent.ObjectID || last.Target.Player != game.Player2 {
		t.Fatalf("put creature attack declaration = %+v, want it attacking Player2", last)
	}
}

// Outside combat, the attacking rider is a no-op: the creature still enters from
// hand but joins no combat.
func TestPutFromHandTappedAndAttackingOutsideCombatEntersNormally(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player1

	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Source",
		Types: []types.Card{types.Artifact},
	}})
	obj := triggeredObjFor(source)

	soldier := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Soldier",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})

	agents := [game.NumPlayers]PlayerAgent{game.Player1: defaultChoiceAgent{}}
	engine.resolveInstructionWithChoices(g, obj, putCreatureFromHandAttackingInstruction(), agents, &TurnLog{})

	permanent, ok := reanimatedPermanent(g, soldier)
	if !ok {
		t.Fatal("creature was not put onto the battlefield")
	}
	if !permanent.Tapped {
		t.Fatal("put creature entered untapped despite the tapped rider")
	}
	if g.Combat != nil {
		t.Fatalf("combat state = %+v, want none outside combat", g.Combat)
	}
}
