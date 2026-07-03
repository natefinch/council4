package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

func TestObservationCombatViewReportsAttackers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatPermanent(g, game.Player2, vanillaCreature("Ogre", 5, 5))
	blocker := addCombatPermanent(g, game.Player1, vanillaCreature("Goblin", 1, 1))
	g.Combat = &game.CombatState{
		AttackersDeclared: true,
		BlockedAttackers:  map[id.ID]bool{attacker.ObjectID: true},
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player1}},
		},
		Blockers: []game.BlockDeclaration{
			{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID},
		},
	}

	view := NewObservation(g, game.Player1).Combat()
	if len(view.Attackers) != 1 {
		t.Fatalf("Combat() reported %d attackers, want 1", len(view.Attackers))
	}
	got := view.Attackers[0]
	if !got.AttacksPlayerDirectly || got.DefendingPlayer != game.Player1 {
		t.Fatalf("attacker view = %+v, want attacking Player1 directly", got)
	}
	if got.Attacker.Power != 5 {
		t.Fatalf("attacker power = %d, want 5", got.Attacker.Power)
	}
	if !got.Blocked || len(got.Blockers) != 1 {
		t.Fatalf("attacker blocked=%v blockers=%d, want blocked with one blocker", got.Blocked, len(got.Blockers))
	}

	against := NewObservation(g, game.Player1).AttackersAgainst(game.Player1)
	if len(against) != 1 {
		t.Fatalf("AttackersAgainst(Player1) = %d, want 1", len(against))
	}
	if none := NewObservation(g, game.Player1).AttackersAgainst(game.Player3); len(none) != 0 {
		t.Fatalf("AttackersAgainst(Player3) = %d, want 0", len(none))
	}
}
