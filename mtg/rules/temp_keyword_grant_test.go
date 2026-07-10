package rules

import (
	"testing"

	cardf "github.com/natefinch/council4/mtg/cards/f"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestFeignDeathGrantsDeathTriggerUntilEndOfTurn proves the temporary
// granted-quoted-ability generalization end to end with the real Feign Death
// card ("Until end of turn, target creature gains 'When this creature dies,
// return it to the battlefield ...'"): resolving the spell adds the quoted death
// trigger to the chosen creature only, and the grant expires at the end-of-turn
// cleanup.
func TestFeignDeathGrantsDeathTriggerUntilEndOfTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	bystander := addCombatCreaturePermanentWithPower(g, game.Player1, 2)

	if got := countGrantedTriggeredAbilities(g, target); got != 0 {
		t.Fatalf("granted triggered abilities before resolution = %d, want 0", got)
	}

	addImplementationSpellToStack(g, game.Player1, cardf.FeignDeath(),
		[]game.Target{game.PermanentTarget(target.ObjectID)})
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := countGrantedTriggeredAbilities(g, target); got != 1 {
		t.Fatalf("granted triggered abilities on target after resolution = %d, want 1", got)
	}
	if got := countGrantedTriggeredAbilities(g, bystander); got != 0 {
		t.Fatalf("granted triggered abilities on non-target = %d, want 0 (grant must not spread)", got)
	}

	// The until-end-of-turn grant expires at the cleanup step.
	expireCleanupDurations(g)
	if got := countGrantedTriggeredAbilities(g, target); got != 0 {
		t.Fatalf("granted triggered abilities after cleanup = %d, want 0 (grant must expire)", got)
	}
}

// TestGroupKeywordGrantUntilYourNextTurn proves the group keyword grant with the
// until-your-next-turn duration (Elspeth, Storm Slayer's "Those creatures gain
// flying until your next turn."): the keyword is granted to every creature the
// resolving player controls at resolution — and to no one else — survives the
// end-of-turn cleanup, and only expires at the start of that player's next turn.
func TestGroupKeywordGrantUntilYourNextTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	mine1 := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	mine2 := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	theirs := addCombatCreaturePermanentWithPower(g, game.Player2, 2)

	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{Primitive: game.ApplyContinuous{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer:       game.LayerAbility,
				Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
				AddKeywords: []game.Keyword{game.Flying},
			}},
			Duration: game.DurationUntilYourNextTurn,
		}},
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	if !hasKeyword(g, mine1, game.Flying) || !hasKeyword(g, mine2, game.Flying) {
		t.Fatal("creatures the resolving player controls did not gain flying")
	}
	if hasKeyword(g, theirs, game.Flying) {
		t.Fatal("an opponent's creature gained flying (grant must be limited to the controller's creatures)")
	}

	// End-of-turn cleanup must not touch an until-your-next-turn grant.
	expireCleanupDurations(g)
	if !hasKeyword(g, mine1, game.Flying) || !hasKeyword(g, mine2, game.Flying) {
		t.Fatal("until-your-next-turn grant expired at cleanup; it must last until the controller's next turn")
	}

	// At the start of the controller's next turn the grant expires.
	g.Turn.TurnNumber = 2
	g.Turn.ActivePlayer = game.Player1
	expireTurnStartDurations(g)
	if hasKeyword(g, mine1, game.Flying) || hasKeyword(g, mine2, game.Flying) {
		t.Fatal("until-your-next-turn grant did not expire at the start of the controller's next turn")
	}
}
