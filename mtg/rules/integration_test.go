package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
)

func TestPriorityLoopCastsAndResolvesCreatureSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	forestID := addCardToHand(g, game.Player1, basicLand())
	creatureID := addCardToHand(g, game.Player1, greenCreature())
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	log := TurnLog{}

	engine.runPriorityLoop(g, allFirstLegalAgents(), &log)

	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want 0", g.Stack.Size())
	}
	forest := permanentForCard(g, forestID)
	if forest == nil {
		t.Fatal("forest was not played")
	}
	if !forest.Tapped {
		t.Fatal("forest was not tapped to pay for creature")
	}
	creature := permanentForCard(g, creatureID)
	if creature == nil {
		t.Fatal("creature did not resolve to battlefield")
	}
	if g.Players[game.Player1].Hand.Contains(creatureID) {
		t.Fatal("creature card remained in hand after resolving")
	}
	if creature.Controller != game.Player1 {
		t.Fatalf("creature controller = %v, want %v", creature.Controller, game.Player1)
	}
	if !creature.SummoningSick {
		t.Fatal("creature was not summoning sick")
	}
	if !sawAction(log.Actions, action.ActionCastSpell) {
		t.Fatal("cast action was not logged")
	}
	if len(log.Resolves) == 0 || log.Resolves[0].SourceID != creatureID {
		t.Fatalf("resolve logs = %+v, want source %v", log.Resolves, creatureID)
	}
}

func TestPriorityLoopCastsAndResolvesPlayerDamageSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	forestID := addCardToHand(g, game.Player1, basicLand())
	spellID := addCardToHand(g, game.Player1, greenPlayerDamageSpell())
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	log := TurnLog{}
	agents := allFirstLegalAgents()
	agents[game.Player1] = targetPlayerAgent{target: game.Player2}

	engine.runPriorityLoop(g, agents, &log)

	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want 0", g.Stack.Size())
	}
	forest := permanentForCard(g, forestID)
	if forest == nil || !forest.Tapped {
		t.Fatalf("forest permanent = %+v, want tapped", forest)
	}
	if g.Players[game.Player2].Life != 37 {
		t.Fatalf("player 2 life = %d, want 37", g.Players[game.Player2].Life)
	}
	if !g.Players[game.Player1].Graveyard.Contains(spellID) {
		t.Fatal("damage spell did not move to graveyard")
	}
	if !sawAction(log.Actions, action.ActionCastSpell) {
		t.Fatal("cast action was not logged")
	}
	if len(log.Resolves) == 0 || log.Resolves[0].SourceID != spellID {
		t.Fatalf("resolve logs = %+v, want source %v", log.Resolves, spellID)
	}
}

func TestCombatVerticalSliceDeclaresAttackersAndDealsDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	engine := NewEngine(nil)
	log := TurnLog{}

	engine.runCombatPhase(g, allFirstLegalAgents(), &log)

	if !sawAttackWith(log.Actions, attacker.ObjectID) {
		t.Fatalf("declare attackers action for attacker %v was not logged", attacker.ObjectID)
	}
	if g.Players[game.Player2].Life != 38 {
		t.Fatalf("player 2 life = %d, want 38", g.Players[game.Player2].Life)
	}
	if len(log.CombatDamage) != 1 {
		t.Fatalf("combat damage logs = %d, want 1", len(log.CombatDamage))
	}
	if log.CombatDamage[0].DefendingPlayer != game.Player2 || log.CombatDamage[0].Damage != 2 {
		t.Fatalf("combat damage log = %+v, want Player2 damage 2", log.CombatDamage[0])
	}
}

type targetPlayerAgent struct {
	target game.PlayerID
}

func (a targetPlayerAgent) ChooseAction(obs PlayerObservation, legal []action.Action) action.Action {
	for _, act := range legal {
		if act.Kind == action.ActionPlayLand {
			return act
		}
	}
	for _, act := range legal {
		cast, ok := act.CastSpellPayload()
		if !ok || len(cast.Targets) != 1 {
			continue
		}
		target := cast.Targets[0]
		if target.Kind == game.TargetPlayer && target.PlayerID == a.target {
			return act
		}
	}
	return action.Pass()
}

func allFirstLegalAgents() [game.NumPlayers]PlayerAgent {
	return [game.NumPlayers]PlayerAgent{
		game.Player1: firstLegalAgent{},
		game.Player2: firstLegalAgent{},
		game.Player3: firstLegalAgent{},
		game.Player4: firstLegalAgent{},
	}
}

func permanentForCard(g *game.Game, cardID id.ID) *game.Permanent {
	for _, permanent := range g.Battlefield {
		if permanent != nil && permanent.CardInstanceID == cardID {
			return permanent
		}
	}
	return nil
}

func sawAction(actions []ActionLog, kind action.Kind) bool {
	for i := range actions {
		logged := &actions[i]
		if logged.Action.Kind == kind {
			return true
		}
	}
	return false
}

func sawAttackWith(actions []ActionLog, attacker id.ID) bool {
	for i := range actions {
		logged := &actions[i]
		if logged.Action.Kind != action.ActionDeclareAttackers {
			continue
		}
		payload, ok := logged.Action.DeclareAttackersPayload()
		if !ok {
			continue
		}
		for _, declaration := range payload.Attackers {
			if declaration.Attacker == attacker {
				return true
			}
		}
	}
	return false
}

func greenPlayerDamageSpell() *game.CardDef {
	spell := playerDamageSpell()
	spell.ManaCost = greenCost()
	spell.ManaCost = greenCost()
	return spell
}
