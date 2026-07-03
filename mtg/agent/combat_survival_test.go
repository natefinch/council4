package agent

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/rules"
)

func declareAttack(g *game.Game, attacker *game.Permanent, defender game.PlayerID) {
	if g.Combat == nil {
		g.Combat = &game.CombatState{BlockedAttackers: make(map[id.ID]bool)}
	}
	g.Combat.AttackersDeclared = true
	g.Combat.Attackers = append(g.Combat.Attackers, game.AttackDeclaration{
		Attacker: attacker.ObjectID,
		Target:   game.AttackTarget{Player: defender},
	})
}

func TestChumpBlocksToSurviveLethal(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Players[game.Player1].Life = 3
	attacker := addObservedPermanent(g, game.Player2, creatureCardDef("Ogre", 5, 5))
	blocker := addObservedPermanent(g, game.Player1, creatureCardDef("Goblin", 1, 1))
	declareAttack(g, attacker, game.Player1)

	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	noBlock := strategy.ScoreAction(obs, action.DeclareBlockers(nil))
	chump := strategy.ScoreAction(obs, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID},
	}))
	if chump <= noBlock {
		t.Fatalf("chump block scored %v, want above taking lethal unblocked %v", chump, noBlock)
	}
}

func TestNoChumpWhenNotLethal(t *testing.T) {
	// At a healthy life total a lone attacker is not lethal, so the agent should
	// not throw a creature away — the survival term stays neutral and trade
	// quality decides. Blocking a 2/2 with a 1/1 (a chump that changes nothing)
	// should not beat simply taking two damage.
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addObservedPermanent(g, game.Player2, creatureCardDef("Bear", 2, 2))
	blocker := addObservedPermanent(g, game.Player1, creatureCardDef("Rat", 1, 1))
	declareAttack(g, attacker, game.Player1)

	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	noBlock := strategy.ScoreAction(obs, action.DeclareBlockers(nil))
	chump := strategy.ScoreAction(obs, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID},
	}))
	if chump >= noBlock {
		t.Fatalf("pointless chump scored %v, want no better than taking 2 damage %v", chump, noBlock)
	}
}

func TestChumpBlocksToSurviveCommanderDamage(t *testing.T) {
	// Life is high, but the attacker is a commander that has already dealt 15
	// commander damage; a 7-power hit reaches 22 >= 21 and kills. The agent must
	// block to survive even though its life total is fine.
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Players[game.Player1].Life = 40
	commander := addObservedPermanent(g, game.Player2, creatureCardDef("General", 7, 7))
	g.Players[game.Player2].CommanderInstanceID = commander.CardInstanceID
	g.Players[game.Player1].CommanderDamage = map[id.ID]int{commander.CardInstanceID: 15}
	blocker := addObservedPermanent(g, game.Player1, creatureCardDef("Wall", 0, 4))
	declareAttack(g, commander, game.Player1)

	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	noBlock := strategy.ScoreAction(obs, action.DeclareBlockers(nil))
	block := strategy.ScoreAction(obs, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: blocker.ObjectID, Blocking: commander.ObjectID},
	}))
	if block <= noBlock {
		t.Fatalf("blocking lethal commander damage scored %v, want above taking it %v", block, noBlock)
	}
}
