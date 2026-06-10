package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestLegalActionsIncludesNinjutsuWithUnblockedAttacker(t *testing.T) {
	g, engine, ninjaID, _ := ninjutsuGame(t)

	legal := engine.legalActions(g, game.Player1)

	if !actionsContain(legal, action.ActivateAbility(ninjaID, 0, nil, 0)) {
		t.Fatalf("legal actions = %+v, want Ninjutsu activation", legal)
	}
}

func TestNinjutsuReturnsAttackerAndEntersTappedAttacking(t *testing.T) {
	g, engine, ninjaID, attacker := ninjutsuGame(t)
	target := g.Combat.Attackers[0].Target

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(ninjaID, 0, nil, 0)) {
		t.Fatal("applyAction() = false, want true for Ninjutsu")
	}

	if !g.Players[game.Player1].Hand.Contains(attacker.CardInstanceID) {
		t.Fatal("Ninjutsu did not return the selected attacker to its owner's hand")
	}
	if !g.Players[game.Player1].Hand.Contains(ninjaID) {
		t.Fatal("Ninjutsu source left hand before resolution")
	}
	if len(g.Combat.Attackers) != 0 {
		t.Fatalf("attackers after paying Ninjutsu cost = %+v, want none", g.Combat.Attackers)
	}
	obj, ok := g.Stack.Peek()
	if !ok || !obj.Ninjutsu || obj.SourceZone != zone.Hand || obj.NinjutsuAttackTarget != target {
		t.Fatalf("Ninjutsu stack object = %+v, want stored attack target %+v", obj, target)
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player1].Hand.Contains(ninjaID) {
		t.Fatal("Ninjutsu source remained in hand after resolution")
	}
	ninja := permanentWithCardID(g, ninjaID)
	if ninja == nil {
		t.Fatal("Ninjutsu source did not enter the battlefield")
	}
	if !ninja.Tapped {
		t.Fatal("Ninjutsu source entered untapped")
	}
	if len(g.Combat.Attackers) != 1 ||
		g.Combat.Attackers[0].Attacker != ninja.ObjectID ||
		g.Combat.Attackers[0].Target != target {
		t.Fatalf("attackers after resolution = %+v, want Ninja attacking %+v", g.Combat.Attackers, target)
	}
}

func TestNinjutsuSourceLeavingHandBeforeResolutionDoesNotEnter(t *testing.T) {
	g, engine, ninjaID, _ := ninjutsuGame(t)
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(ninjaID, 0, nil, 0)) {
		t.Fatal("applyAction() = false, want true for Ninjutsu")
	}
	g.Players[game.Player1].Hand.Remove(ninjaID)
	g.Players[game.Player1].Graveyard.Add(ninjaID)

	engine.resolveTopOfStack(g, &TurnLog{})

	if permanentWithCardID(g, ninjaID) != nil {
		t.Fatal("Ninjutsu source entered after leaving hand")
	}
	if !g.Players[game.Player1].Graveyard.Contains(ninjaID) {
		t.Fatal("Ninjutsu resolution moved source from its new zone")
	}
}

func TestNinjutsuRequiresLegalTimingResourcesAndAttacker(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*game.Game, *game.Permanent)
	}{
		{
			name: "before blockers",
			mutate: func(g *game.Game, _ *game.Permanent) {
				g.Turn.Step = game.StepDeclareAttackers
			},
		},
		{
			name: "outside combat",
			mutate: func(g *game.Game, _ *game.Permanent) {
				g.Turn.Phase = game.PhasePrecombatMain
			},
		},
		{
			name: "without priority",
			mutate: func(g *game.Game, _ *game.Permanent) {
				g.Turn.PriorityPlayer = game.Player2
			},
		},
		{
			name: "without mana",
			mutate: func(g *game.Game, _ *game.Permanent) {
				for _, permanent := range g.Battlefield {
					permanent.Tapped = true
				}
			},
		},
		{
			name: "attacker was blocked",
			mutate: func(g *game.Game, attacker *game.Permanent) {
				g.Combat.BlockedAttackers = map[game.ObjectID]bool{attacker.ObjectID: true}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, engine, ninjaID, attacker := ninjutsuGame(t)
			tt.mutate(g, attacker)

			if actionsContain(engine.legalActions(g, game.Player1), action.ActivateAbility(ninjaID, 0, nil, 0)) {
				t.Fatal("Ninjutsu activation was legal")
			}
		})
	}
}

func TestBlockedAttackerRemainsBlockedAfterBlockerLeavesCombat(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanent(g, game.Player1)
	g.Combat = &game.CombatState{
		Attackers:        []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}},
		BlockedAttackers: map[game.ObjectID]bool{attacker.ObjectID: true},
	}

	if !attackerWasBlocked(g, attacker.ObjectID) {
		t.Fatal("attacker lost blocked status after blocker left combat")
	}
}

func ninjutsuGame(t *testing.T) (*game.Game, *Engine, game.ObjectID, *game.Permanent) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	ninjaID := addCardToHand(g, game.Player1, ninjutsuCard())
	attacker := addCombatCreaturePermanent(g, game.Player1)
	addBasicLandPermanent(g, game.Player1, types.Island)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{
			Attacker: attacker.ObjectID,
			Target:   game.AttackTarget{Player: game.Player2},
		}},
	}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Turn.PriorityPlayer = game.Player1
	return g, engine, ninjaID, attacker
}

func ninjutsuCard() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Ninjutsu Test Card",
		Types: []types.Card{types.Creature},
		ActivatedAbilities: []game.ActivatedAbility{
			game.NinjutsuActivatedAbility(cost.Mana{cost.O(1)}),
		},
	}}
}

func permanentWithCardID(g *game.Game, cardID game.ObjectID) *game.Permanent {
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == cardID {
			return permanent
		}
	}
	return nil
}
