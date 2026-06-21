package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

// addCreaturePermanentWithSupertype adds a vanilla creature permanent to the
// battlefield under controller, carrying the given supertypes so group
// selections that filter on supertype (such as "each legendary creature you
// control") can match it.
func addCreaturePermanentWithSupertype(g *game.Game, controller game.PlayerID, supertypes ...types.Super) *game.Permanent {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID: cardID,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:       "Test Creature",
			Types:      []types.Card{types.Creature},
			Supertypes: supertypes,
		}},
		Owner: controller,
	}
	permanent := &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: cardID,
		Owner:          controller,
		Controller:     controller,
	}
	g.Battlefield = append(g.Battlefield, permanent)
	return permanent
}

// TestDynamicMaxOfPicksGreaterOperand covers Willowdusk, Essence Seer's "the
// amount of life you gained this turn or the amount of life you lost this turn,
// whichever is greater": a DynamicAmountMaxOf evaluates every operand against
// the same context and returns the greatest, so the result tracks whichever of
// the controller's life gained or lost this turn is larger.
func TestDynamicMaxOfPicksGreaterOperand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obj := &game.StackObject{Controller: game.Player1}
	maxAmount := game.DynamicAmount{
		Kind: game.DynamicAmountMaxOf,
		Operands: []game.DynamicAmount{
			{Kind: game.DynamicAmountLifeGainedThisTurn},
			{Kind: game.DynamicAmountLifeLostThisTurn},
		},
	}

	if got := dynamicAmountValue(g, obj, game.Player1, maxAmount); got != 0 {
		t.Fatalf("max = %d, want 0 before any life change", got)
	}

	gainLife(g, game.Player1, 4)
	if got := dynamicAmountValue(g, obj, game.Player1, maxAmount); got != 4 {
		t.Fatalf("max = %d, want 4 (gained 4, lost 0)", got)
	}

	loseLife(g, game.Player1, 9)
	if got := dynamicAmountValue(g, obj, game.Player1, maxAmount); got != 9 {
		t.Fatalf("max = %d, want 9 (lost 9 > gained 4)", got)
	}

	gainLife(g, game.Player1, 10)
	if got := dynamicAmountValue(g, obj, game.Player1, maxAmount); got != 14 {
		t.Fatalf("max = %d, want 14 (gained 14 > lost 9)", got)
	}
}

// TestDynamicMaxOfCounterPlacementOnTarget proves the full Willowdusk path: a
// single-target AddCounter whose count is a DynamicAmountMaxOf over life gained
// and lost this turn places exactly the greater amount of +1/+1 counters on the
// chosen creature.
func TestDynamicMaxOfCounterPlacementOnTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	chosen := addCreaturePermanent(g, game.Player1)

	gainLife(g, game.Player1, 3)
	loseLife(g, game.Player1, 7)

	addEffectSpellToStack(g, game.Player1, game.AddCounter{
		Amount: game.Dynamic(game.DynamicAmount{
			Kind: game.DynamicAmountMaxOf,
			Operands: []game.DynamicAmount{
				{Kind: game.DynamicAmountLifeGainedThisTurn},
				{Kind: game.DynamicAmountLifeLostThisTurn},
			},
		}),
		Object:      game.TargetPermanentReference(0),
		CounterKind: counter.PlusOnePlusOne,
	}, []game.Target{game.PermanentTarget(chosen.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := chosen.Counters.Get(counter.PlusOnePlusOne); got != 7 {
		t.Fatalf("chosen +1/+1 counters = %d, want 7 (max of gained 3, lost 7)", got)
	}
}

// TestDynamicCounterPlacementOnGroup proves the group counter path Aerith
// Gainsborough needs: an AddCounter targeting "each legendary creature you
// control" places the resolved dynamic count on every matching permanent the
// controller controls, while leaving non-legendary and opponent-controlled
// creatures untouched. The count is resolved once and applied to each member.
func TestDynamicCounterPlacementOnGroup(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	legendaryA := addCreaturePermanentWithSupertype(g, game.Player1, types.Legendary)
	legendaryB := addCreaturePermanentWithSupertype(g, game.Player1, types.Legendary)
	nonLegendary := addCreaturePermanent(g, game.Player1)
	opponentLegendary := addCreaturePermanentWithSupertype(g, game.Player2, types.Legendary)

	gainLife(g, game.Player1, 5)

	addEffectSpellToStack(g, game.Player1, game.AddCounter{
		Amount: game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountLifeGainedThisTurn}),
		Group: game.BattlefieldGroup(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			Supertypes:    []types.Super{types.Legendary},
			Controller:    game.ControllerYou,
		}),
		CounterKind: counter.PlusOnePlusOne,
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	for _, member := range []*game.Permanent{legendaryA, legendaryB} {
		if got := member.Counters.Get(counter.PlusOnePlusOne); got != 5 {
			t.Fatalf("legendary member +1/+1 counters = %d, want 5", got)
		}
	}
	if got := nonLegendary.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("non-legendary creature +1/+1 counters = %d, want 0", got)
	}
	if got := opponentLegendary.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("opponent legendary +1/+1 counters = %d, want 0", got)
	}
}
