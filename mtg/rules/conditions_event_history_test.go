package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestEventHistoryConditionCurrentTurn(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source1 := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Bear",
		Types: []types.Card{types.Creature},
	}})
	source2 := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Opponent Bear",
		Types: []types.Card{types.Creature},
	}})

	attackedCond := opt.Val(game.Condition{
		EventHistory: opt.Val(game.EventHistoryCondition{
			Pattern: game.TriggerPattern{
				Event:      game.EventAttackerDeclared,
				Controller: game.TriggerControllerYou,
			},
			Window: game.EventHistoryCurrentTurn,
		}),
	})

	ctx1 := conditionContext{controller: game.Player1, source: source1}
	ctx2 := conditionContext{controller: game.Player2, source: source2}
	if conditionSatisfied(g, ctx1, attackedCond) {
		t.Fatal("condition satisfied before any attacks")
	}

	// Player1 attacks — satisfied for Player1's source but not Player2's.
	emitEvent(g, game.Event{Kind: game.EventAttackerDeclared, Controller: game.Player1})

	if !conditionSatisfied(g, ctx1, attackedCond) {
		t.Fatal("condition not satisfied after Player1 attacked")
	}
	if conditionSatisfied(g, ctx2, attackedCond) {
		t.Fatal("condition satisfied for Player2 source when only Player1 attacked")
	}
}

func TestEventHistoryConditionAttackerCount(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Bear",
		Types: []types.Card{types.Creature},
	}})

	twoAttackersCond := opt.Val(game.Condition{
		EventHistory: opt.Val(game.EventHistoryCondition{
			Pattern: game.TriggerPattern{
				Event:      game.EventAttackerDeclared,
				Controller: game.TriggerControllerYou,
			},
			Window:   game.EventHistoryCurrentTurn,
			MinCount: 2,
		}),
	})

	ctx := conditionContext{controller: game.Player1, source: source}

	// One attacker you control is not enough for "two or more creatures".
	emitEvent(g, game.Event{Kind: game.EventAttackerDeclared, Controller: game.Player1})
	if conditionSatisfied(g, ctx, twoAttackersCond) {
		t.Fatal("condition satisfied after only one attacker")
	}

	// An opponent's attacker does not count toward your controller-scoped total.
	emitEvent(g, game.Event{Kind: game.EventAttackerDeclared, Controller: game.Player2})
	if conditionSatisfied(g, ctx, twoAttackersCond) {
		t.Fatal("condition satisfied counting an opponent's attacker")
	}

	// A second attacker you control reaches the required count.
	emitEvent(g, game.Event{Kind: game.EventAttackerDeclared, Controller: game.Player1})
	if !conditionSatisfied(g, ctx, twoAttackersCond) {
		t.Fatal("condition not satisfied after attacking with two creatures")
	}
}

func TestEventHistoryConditionPreviousTurn(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Bear",
		Types: []types.Card{types.Creature},
	}})

	lifeLostCond := opt.Val(game.Condition{
		EventHistory: opt.Val(game.EventHistoryCondition{
			Pattern: game.TriggerPattern{
				Event:  game.EventLifeLost,
				Player: game.TriggerPlayerOpponent,
			},
			Window: game.EventHistoryPreviousTurn,
		}),
	})

	ctx := conditionContext{controller: game.Player1, source: source}
	if conditionSatisfied(g, ctx, lifeLostCond) {
		t.Fatal("condition satisfied before any events")
	}

	// Emit life-loss on this turn then advance to the next turn so it becomes
	// "previous turn" for the upkeep check.
	emitEvent(g, game.Event{Kind: game.EventLifeLost, Player: game.Player2, Amount: 3})
	g.EventTurnStarts = append(g.EventTurnStarts, len(g.Events))
	g.Turn.TurnNumber++

	if !conditionSatisfied(g, ctx, lifeLostCond) {
		t.Fatal("condition not satisfied after opponent lost life last turn")
	}
}

func TestEventHistoryConditionNegatedNoSpells(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Bear",
		Types: []types.Card{types.Creature},
	}})

	noSpellsCond := opt.Val(game.Condition{
		Negate: true,
		EventHistory: opt.Val(game.EventHistoryCondition{
			Pattern: game.TriggerPattern{Event: game.EventSpellCast},
			Window:  game.EventHistoryPreviousTurn,
		}),
	})

	ctx := conditionContext{controller: game.Player1, source: source}

	// No previous turn yet — EventsPreviousTurn returns nil, no spells found →
	// condition (negated) is satisfied.
	if !conditionSatisfied(g, ctx, noSpellsCond) {
		t.Fatal("negated condition not satisfied when no previous-turn events exist")
	}

	// Emit a spell cast on "last turn" then advance.
	emitEvent(g, game.Event{Kind: game.EventSpellCast, Controller: game.Player1})
	g.EventTurnStarts = append(g.EventTurnStarts, len(g.Events))
	g.Turn.TurnNumber++

	if conditionSatisfied(g, ctx, noSpellsCond) {
		t.Fatal("negated condition satisfied when a spell was cast last turn")
	}
}

func TestEventHistoryConditionCreatureDiedCurrentTurn(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Bear",
		Types: []types.Card{types.Creature},
	}})

	creatureDiedCond := opt.Val(game.Condition{
		EventHistory: opt.Val(game.EventHistoryCondition{
			Pattern: game.TriggerPattern{
				Event:            game.EventPermanentDied,
				SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
			},
			Window: game.EventHistoryCurrentTurn,
		}),
	})

	ctx := conditionContext{controller: game.Player1, source: source}
	if conditionSatisfied(g, ctx, creatureDiedCond) {
		t.Fatal("condition satisfied before any creature deaths")
	}

	// A non-creature permanent died — should not satisfy.
	addAndEmitArtifactDied(g)
	if conditionSatisfied(g, ctx, creatureDiedCond) {
		t.Fatal("condition satisfied after non-creature died")
	}

	// A creature dies — now satisfied.
	emitCreatureDiedEvent(g)
	if !conditionSatisfied(g, ctx, creatureDiedCond) {
		t.Fatal("condition not satisfied after creature died")
	}
}

func TestEventHistoryConditionFailsClosedWithNilSource(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	emitEvent(g, game.Event{Kind: game.EventAttackerDeclared, Controller: game.Player1})

	cond := opt.Val(game.Condition{
		EventHistory: opt.Val(game.EventHistoryCondition{
			Pattern: game.TriggerPattern{
				Event:      game.EventAttackerDeclared,
				Controller: game.TriggerControllerYou,
			},
			Window: game.EventHistoryCurrentTurn,
		}),
	})
	if conditionSatisfied(g, conditionContext{controller: game.Player1, source: nil}, cond) {
		t.Fatal("condition satisfied with nil source; should fail closed")
	}
}

// addAndEmitArtifactDied registers and emits an EventPermanentDied for a
// non-creature artifact so tests can verify creature-type filtering.
func addAndEmitArtifactDied(g *game.Game) {
	perm := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Dead Relic",
		Types: []types.Card{types.Artifact},
	}})
	emitEvent(g, game.Event{
		Kind:        game.EventPermanentDied,
		PermanentID: perm.ObjectID,
		CardTypes:   []types.Card{types.Artifact},
	})
}

// emitCreatureDiedEvent emits an EventPermanentDied that looks like a creature
// died by recording creature card types on the event directly.
func emitCreatureDiedEvent(g *game.Game) {
	perm := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Dead Bear",
		Types: []types.Card{types.Creature},
	}})
	emitEvent(g, game.Event{
		Kind:        game.EventPermanentDied,
		PermanentID: perm.ObjectID,
		CardTypes:   []types.Card{types.Creature},
	})
}
