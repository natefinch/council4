package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestAttackingCreaturesYouControlGroupKeywordGrantSnapshots verifies the
// runtime shape produced for "Whenever one or more creatures you control attack,
// they gain <keyword> until end of turn." (Angelic Guardian): an ApplyContinuous
// granting a keyword to the "attacking creatures you control" group. It confirms
// only the attacking creatures the resolving player controls gain the keyword
// (not their non-attacking creatures nor an opponent's attackers), and that the
// affected set is snapshotted at resolution (CR 611.2c) so a creature keeps the
// keyword after it leaves combat.
func TestAttackingCreaturesYouControlGroupKeywordGrantSnapshots(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	myAttacker := makeCreaturePermanent(g, game.Player1, "My Attacker")
	myBencher := makeCreaturePermanent(g, game.Player1, "My Bencher")
	opponentAttacker := makeCreaturePermanent(g, game.Player2, "Opponent Attacker")
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{
		{Attacker: myAttacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		{Attacker: opponentAttacker.ObjectID, Target: game.AttackTarget{Player: game.Player1}},
	}}
	addEffectSpellToStack(g, game.Player1, game.ApplyContinuous{
		ContinuousEffects: []game.ContinuousEffect{{
			Layer: game.LayerAbility,
			Group: game.BattlefieldGroup(game.Selection{
				RequiredTypes: []types.Card{types.Creature},
				Controller:    game.ControllerYou,
				CombatState:   game.CombatStateAttacking,
			}),
			AddKeywords: []game.Keyword{game.Indestructible},
		}},
		Duration: game.DurationUntilEndOfTurn,
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	if !hasKeyword(g, myAttacker, game.Indestructible) {
		t.Fatal("attacking creature you control did not gain indestructible")
	}
	if hasKeyword(g, myBencher, game.Indestructible) {
		t.Fatal("non-attacking creature you control gained indestructible")
	}
	if hasKeyword(g, opponentAttacker, game.Indestructible) {
		t.Fatal("opponent's attacking creature gained indestructible")
	}

	// The affected set is locked at resolution, so leaving combat does not end
	// the grant.
	g.Combat.Attackers = nil
	if !hasKeyword(g, myAttacker, game.Indestructible) {
		t.Fatal("indestructible ended when the creature left combat; group was not snapshotted")
	}
}
