package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestStaticPTEffectAffectsCombatDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addAnthemPermanent(g, game.Player1)
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	opponentCreature := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
			{Attacker: opponentCreature.ObjectID, Target: game.AttackTarget{Player: game.Player1}},
		},
	}
	log := TurnLog{}

	NewEngine(nil).resolveCombatDamage(g, &log)

	if g.Players[game.Player2].Life != 37 {
		t.Fatalf("defending Player2 life = %d, want 37", g.Players[game.Player2].Life)
	}
	if g.Players[game.Player1].Life != 38 {
		t.Fatalf("defending Player1 life = %d, want 38", g.Players[game.Player1].Life)
	}
}

func TestStaticPTEffectRaisesLethalDamageThreshold(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addAnthemPermanent(g, game.Player2)
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
		Blockers: []game.BlockDeclaration{
			{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID},
		},
	}
	engine := NewEngine(nil)

	engine.resolveCombatDamage(g, &TurnLog{})
	_, deaths := engine.applyStateBasedActionsWithDeaths(g)

	if permanentByObjectID(g, blocker.ObjectID) == nil {
		t.Fatal("anthem-pumped blocker died to nonlethal marked damage")
	}
	for _, death := range deaths {
		if death.Permanent == blocker.ObjectID {
			t.Fatalf("blocker death = %+v, want blocker to survive anthem-raised toughness", death)
		}
	}
}

func TestStaticPTEffectDisappearingChangesLethalDamageThreshold(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	anthem := addAnthemPermanent(g, game.Player1)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	creature.MarkedDamage = 2
	engine := NewEngine(nil)

	_, deaths := engine.applyStateBasedActionsWithDeaths(g)

	if permanentByObjectID(g, creature.ObjectID) == nil {
		t.Fatal("anthem-pumped creature died before anthem left")
	}
	if len(deaths) != 0 {
		t.Fatalf("deaths before anthem leaves = %+v, want none", deaths)
	}

	movePermanentToZone(g, anthem, game.ZoneGraveyard)
	_, deaths = engine.applyStateBasedActionsWithDeaths(g)

	if permanentByObjectID(g, creature.ObjectID) != nil {
		t.Fatal("creature survived after anthem left and marked damage became lethal")
	}
	if len(deaths) != 1 || deaths[0].Permanent != creature.ObjectID || deaths[0].Reason != PermanentDeathReasonLethalDamage {
		t.Fatalf("deaths after anthem leaves = %+v, want creature lethal damage death", deaths)
	}
}

func addAnthemPermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	pt := game.PT{Value: 2}
	return addCombatPermanent(g, controller, &game.CardDef{
		Name:      "Anthem Captain",
		Types:     []game.CardType{game.TypeCreature},
		Power:     &pt,
		Toughness: &pt,
		Abilities: []game.AbilityDef{
			{
				Kind: game.StaticAbility,
				Effects: []game.Effect{
					{
						Type:           game.EffectModifyPT,
						Selector:       game.EffectSelectorOtherCreaturesYouControl,
						PowerDelta:     1,
						ToughnessDelta: 1,
					},
				},
			},
		},
	})
}
