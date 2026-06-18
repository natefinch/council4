package agent

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/rules"
)

func attackAction(target game.PlayerID, attackerID id.ID) action.Action {
	return action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: attackerID, Target: game.AttackTarget{Player: target}},
	})
}

func blockAction(blockerID, attackerID id.ID) action.Action {
	return action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: blockerID, Blocking: attackerID},
	})
}

func TestScoreBlockPrefersKillAndSurvive(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addObservedPermanent(g, game.Player2, creatureCardDef("Attacker", 2, 2))
	survivor := addObservedPermanent(g, game.Player1, creatureCardDef("Survivor", 3, 3)) // kills 2/2, survives
	trader := addObservedPermanent(g, game.Player1, creatureCardDef("Trader", 2, 2))     // kills, but dies
	chump := addObservedPermanent(g, game.Player1, creatureCardDef("Chump", 1, 1))       // dies, does not kill

	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	survive := strategy.ScoreAction(obs, blockAction(survivor.ObjectID, attacker.ObjectID))
	trade := strategy.ScoreAction(obs, blockAction(trader.ObjectID, attacker.ObjectID))
	chumpScore := strategy.ScoreAction(obs, blockAction(chump.ObjectID, attacker.ObjectID))
	noBlock := strategy.ScoreAction(obs, action.DeclareBlockers(nil))

	if !(survive > trade && trade > chumpScore) {
		t.Errorf("block ordering wrong: survive=%v trade=%v chump=%v", survive, trade, chumpScore)
	}
	if chumpScore >= noBlock {
		t.Errorf("chump-blocking a 2/2 (%v) should not beat not blocking (%v)", chumpScore, noBlock)
	}
}

func TestScoreBlockKillAndSurviveDominatesAbsorb(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// A big 11/11 attacker: a blocker that kills it and survives should always
	// be preferred over one that merely walls it without killing.
	attacker := addObservedPermanent(g, game.Player2, creatureCardDef("Giant", 11, 11))
	killer := addObservedPermanent(g, game.Player1, creatureCardDef("Killer", 11, 12)) // kills and survives
	wall := addObservedPermanent(g, game.Player1, creatureCardDef("Wall", 1, 12))      // absorbs, no kill

	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	kill := strategy.ScoreAction(obs, blockAction(killer.ObjectID, attacker.ObjectID))
	absorb := strategy.ScoreAction(obs, blockAction(wall.ObjectID, attacker.ObjectID))
	if kill <= absorb {
		t.Errorf("kill-and-survive (%v) must dominate absorbing the same attacker (%v)", kill, absorb)
	}
}

func TestScoreBlockDeathtouchKillsAnything(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	bigAttacker := addObservedPermanent(g, game.Player2, creatureCardDef("Giant", 8, 8))
	deathtoucher := addObservedPermanent(g, game.Player1, creatureWithKeywords("Deathtoucher", 1, 1, game.Deathtouch))

	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	// A 1/1 deathtouch blocking an 8/8 kills it (but dies): a favorable trade,
	// clearly better than not blocking the 8/8.
	block := strategy.ScoreAction(obs, blockAction(deathtoucher.ObjectID, bigAttacker.ObjectID))
	noBlock := strategy.ScoreAction(obs, action.DeclareBlockers(nil))
	if block <= noBlock {
		t.Errorf("deathtouch block (%v) should beat not blocking the 8/8 (%v)", block, noBlock)
	}
}

func TestScoreBlockFirstStrikeSurvives(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addObservedPermanent(g, game.Player2, creatureCardDef("Attacker", 3, 3))
	firstStriker := addObservedPermanent(g, game.Player1, creatureWithKeywords("First Striker", 3, 3, game.FirstStrike))
	vanilla := addObservedPermanent(g, game.Player1, creatureCardDef("Vanilla", 3, 3))

	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	// A 3/3 first striker kills the 3/3 attacker before taking damage (survives);
	// a vanilla 3/3 trades. First strike should score higher.
	fs := strategy.ScoreAction(obs, blockAction(firstStriker.ObjectID, attacker.ObjectID))
	plain := strategy.ScoreAction(obs, blockAction(vanilla.ObjectID, attacker.ObjectID))
	if fs <= plain {
		t.Errorf("first-strike block (%v) should beat the vanilla trade (%v)", fs, plain)
	}
}

func TestScoreAttackAvoidsProfitableBlocker(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addObservedPermanent(g, game.Player1, creatureCardDef("Attacker", 2, 2))
	// Player2 has a 3/3 untapped blocker that kills the 2/2 and survives.
	addObservedPermanent(g, game.Player2, creatureCardDef("Wall", 3, 3))

	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	attack := strategy.ScoreAction(obs, attackAction(game.Player2, attacker.ObjectID))
	noAttack := strategy.ScoreAction(obs, action.DeclareAttackers(nil))
	if attack >= noAttack {
		t.Errorf("attacking a 2/2 into a profitable 3/3 blocker (%v) should not beat not attacking (%v)", attack, noAttack)
	}
}

func TestScoreAttackRewardsEvasion(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	flyer := addObservedPermanent(g, game.Player1, creatureWithKeywords("Flyer", 3, 3, game.Flying))
	ground := addObservedPermanent(g, game.Player1, creatureCardDef("Ground", 3, 3))
	// Player2 has only a ground blocker that would trade with the ground attacker.
	addObservedPermanent(g, game.Player2, creatureCardDef("Blocker", 3, 3))

	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	flyAttack := strategy.ScoreAction(obs, attackAction(game.Player2, flyer.ObjectID))
	groundAttack := strategy.ScoreAction(obs, attackAction(game.Player2, ground.ObjectID))
	if flyAttack <= groundAttack {
		t.Errorf("attacking with the evasive creature (%v) should beat the ground creature (%v)", flyAttack, groundAttack)
	}
}

func TestScoreAttackPrefersBiggerThreatPlayer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addObservedPermanent(g, game.Player1, creatureWithKeywords("Evasive", 3, 3, game.Flying))
	// Player3 is a bigger threat than Player2.
	addObservedPermanent(g, game.Player3, creatureCardDef("Threat", 7, 7))

	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	atP2 := strategy.ScoreAction(obs, attackAction(game.Player2, attacker.ObjectID))
	atP3 := strategy.ScoreAction(obs, attackAction(game.Player3, attacker.ObjectID))
	if atP3 <= atP2 {
		t.Errorf("attacking the bigger threat Player3 (%v) should beat attacking Player2 (%v)", atP3, atP2)
	}
}
