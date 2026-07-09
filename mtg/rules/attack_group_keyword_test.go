package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestTriggeringAttackersAgainstDefenderGroupScopesToOpponents verifies the
// recipient-scoped group produced for "Whenever one or more creatures attack one
// of your opponents or a planeswalker they control, those creatures gain menace
// until end of turn." (Frontier Warmonger). "Those creatures" is the declared
// attackers whose defending player is an opponent of the ability's controller, so
// in multiplayer an attacker declared against the controller itself (part of the
// same simultaneous batch) is excluded, while attackers against two different
// opponents are both included.
func TestTriggeringAttackersAgainstDefenderGroupScopesToOpponents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	vsOpponent := makeCreaturePermanent(g, game.Player2, "Attacks Opponent Two")
	vsController := makeCreaturePermanent(g, game.Player2, "Attacks the Controller")
	vsOtherOpponent := makeCreaturePermanent(g, game.Player3, "Attacks Opponent Three")

	// One simultaneous declaration in which Player2/Player3 attack Player1 (the
	// ability's controller), Player3, and Player2 respectively.
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: vsOpponent.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
			{Attacker: vsController.ObjectID, Target: game.AttackTarget{Player: game.Player1}},
			{Attacker: vsOtherOpponent.ObjectID, Target: game.AttackTarget{Player: game.Player3}},
		},
	}
	batchID := g.IDGen.Next()
	defenders := map[id.ID]game.PlayerID{
		vsOpponent.ObjectID:      game.Player2,
		vsController.ObjectID:    game.Player1,
		vsOtherOpponent.ObjectID: game.Player3,
	}
	for _, permanent := range []*game.Permanent{vsOpponent, vsController, vsOtherOpponent} {
		emitEvent(g, game.Event{
			Kind:           game.EventAttackerDeclared,
			PermanentID:    permanent.ObjectID,
			Controller:     permanent.Controller,
			Player:         defenders[permanent.ObjectID],
			SimultaneousID: batchID,
		})
	}

	obj := &game.StackObject{
		ID:              g.IDGen.Next(),
		Kind:            game.StackTriggeredAbility,
		Controller:      game.Player1,
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:           game.EventAttackerDeclared,
			PermanentID:    vsOpponent.ObjectID,
			Controller:     game.Player2,
			Player:         game.Player2,
			SimultaneousID: batchID,
		},
	}
	r := &effectResolver{engine: NewEngine(nil), game: g, obj: obj, agents: [game.NumPlayers]PlayerAgent{}, log: &TurnLog{}}

	resolved := handleApplyContinuous(r, game.ApplyContinuous{
		ContinuousEffects: []game.ContinuousEffect{{
			Layer: game.LayerAbility,
			Group: game.TriggeringAttackersAgainstDefenderGroup(
				game.Selection{RequiredTypes: []types.Card{types.Creature}},
				game.TriggerControllerOpponent,
			),
			AddKeywords: []game.Keyword{game.Menace},
		}},
		Duration: game.DurationUntilEndOfTurn,
	})
	if !resolved.succeeded {
		t.Fatal("ApplyContinuous over the opponent-attacking attackers did not apply")
	}

	if !hasKeyword(g, vsOpponent, game.Menace) {
		t.Fatal("an attacker against an opponent did not gain menace")
	}
	if !hasKeyword(g, vsOtherOpponent, game.Menace) {
		t.Fatal("an attacker against a second opponent did not gain menace")
	}
	if hasKeyword(g, vsController, game.Menace) {
		t.Fatal("an attacker against the ability's controller wrongly gained menace")
	}
}

// runtime shape produced for "Whenever one or more creatures you control attack,
// they gain <keyword> until end of turn." (Angelic Guardian): an ApplyContinuous
// granting a keyword to the TriggeringAttackers group. It binds the creatures
// declared as attackers in the triggering attack (the trigger event's
// simultaneous batch), filtered to the ability controller — not the current set
// of attacking creatures. It confirms:
//   - each declared attacker the resolving player controls gains the keyword;
//   - an opponent's creature in the same batch is excluded (Controller: You);
//   - a declared attacker removed from combat before resolution still gains it
//     (the affected set is the declared batch, snapshotted at resolution);
//   - a creature you control that is attacking but was not part of the
//     triggering declaration (a later, separate attack) is not included.
func TestTriggeringAttackersGroupKeywordGrantBindsDeclaredAttackers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	declaredFirst := makeCreaturePermanent(g, game.Player1, "Declared First")
	declaredSecond := makeCreaturePermanent(g, game.Player1, "Declared Second")
	leftCombat := makeCreaturePermanent(g, game.Player1, "Left Combat")
	opponentInBatch := makeCreaturePermanent(g, game.Player2, "Opponent Attacker")
	lateAttacker := makeCreaturePermanent(g, game.Player1, "Late Attacker")

	// The triggering attack: three of the controller's creatures and an opponent
	// creature are declared as attackers simultaneously.
	batchID := g.IDGen.Next()
	for _, permanent := range []*game.Permanent{declaredFirst, declaredSecond, leftCombat, opponentInBatch} {
		emitEvent(g, game.Event{
			Kind:           game.EventAttackerDeclared,
			PermanentID:    permanent.ObjectID,
			Controller:     permanent.Controller,
			SimultaneousID: batchID,
		})
	}
	// A separate, later attack declaration that did not trigger this ability.
	emitEvent(g, game.Event{
		Kind:           game.EventAttackerDeclared,
		PermanentID:    lateAttacker.ObjectID,
		Controller:     game.Player1,
		SimultaneousID: g.IDGen.Next(),
	})

	obj := &game.StackObject{
		ID:              g.IDGen.Next(),
		Kind:            game.StackTriggeredAbility,
		Controller:      game.Player1,
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:           game.EventAttackerDeclared,
			PermanentID:    declaredFirst.ObjectID,
			Controller:     game.Player1,
			SimultaneousID: batchID,
		},
	}
	r := &effectResolver{engine: NewEngine(nil), game: g, obj: obj, agents: [game.NumPlayers]PlayerAgent{}, log: &TurnLog{}}

	resolved := handleApplyContinuous(r, game.ApplyContinuous{
		ContinuousEffects: []game.ContinuousEffect{{
			Layer: game.LayerAbility,
			Group: game.TriggeringAttackersGroup(game.Selection{
				RequiredTypes: []types.Card{types.Creature},
				Controller:    game.ControllerYou,
			}),
			AddKeywords: []game.Keyword{game.Indestructible},
		}},
		Duration: game.DurationUntilEndOfTurn,
	})
	if !resolved.succeeded {
		t.Fatal("ApplyContinuous over the triggering attackers did not apply")
	}

	if !hasKeyword(g, declaredFirst, game.Indestructible) ||
		!hasKeyword(g, declaredSecond, game.Indestructible) {
		t.Fatal("a declared attacker you control did not gain indestructible")
	}
	if !hasKeyword(g, leftCombat, game.Indestructible) {
		t.Fatal("a declared attacker that left combat lost the grant; the declared batch was not bound")
	}
	if hasKeyword(g, opponentInBatch, game.Indestructible) {
		t.Fatal("an opponent's attacker in the batch gained indestructible")
	}
	if hasKeyword(g, lateAttacker, game.Indestructible) {
		t.Fatal("a creature attacking from a separate declaration gained indestructible")
	}
}
