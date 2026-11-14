package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review (#3858): pinned corpus 6da9ced0 Lullmage's Domination
// (fdac480b-03b5-4ece-b42a-e6928b609419, printing c9944393 znr) {X}{U}{U}{U}
// Sorcery: "This spell costs {3} less to cast if it targets a creature whose
// controller has eight or more cards in their graveyard. Gain control of
// target creature with mana value X." Scryfall rulings retrieved 2026-11-14
// (2020-09-25): choose X, then a creature with that mana value, then determine
// the cost, subtracting {3} if appropriate; the reduction can't reduce
// {U}{U}{U} (X=2 against a qualifying creature costs exactly {U}{U}{U}); a
// creature's {X} is 0 on the battlefield; the control effect doesn't expire;
// only whether the count is at least eight matters. Official CR effective
// 2026-11-06 (MagicCompRules 20261106.txt) read: 601.2c (targets precede the
// total cost), 601.2f (reductions are applied as the total cost is determined,
// then it is locked in), 118.7a (a generic reduction affects only the generic
// component). These scenarios use a test-local Lullmage's Domination; the card
// itself is authored by a later card unit.

func lullmagesDominationShape() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Test Lullmage's Domination",
		Types:    []types.Card{types.Sorcery},
		Colors:   []color.Color{color.Blue},
		ManaCost: opt.Val(cost.Mana{cost.X, cost.U, cost.U, cost.U}),
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets:       1,
				MaxTargets:       1,
				Constraint:       "target creature with mana value X",
				Allow:            game.TargetAllowPermanent,
				Selection:        opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
				ManaValueEqualsX: true,
			}},
			Sequence: []game.Instruction{{Primitive: game.ApplyContinuous{
				Object: opt.Val(game.TargetPermanentReference(0)),
				ContinuousEffects: []game.ContinuousEffect{{
					Layer:         game.LayerControl,
					NewController: opt.Val(game.Player1),
				}},
				Duration: game.DurationPermanent,
			}}},
		}.Ability()),
		StaticAbilities: []game.StaticAbility{{RuleEffects: []game.RuleEffect{{
			Kind:           game.RuleEffectCostModifier,
			AffectedSource: true,
			CostModifier: game.CostModifier{
				Kind:             game.CostModifierSpell,
				GenericReduction: 3,
				ReductionTarget: &game.TargetSpec{
					Allow: game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{
						RequiredTypes:            []types.Card{types.Creature},
						ControllerGraveyardCount: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 8}),
					}),
				},
			},
		}}}},
	}}
}

func newLullmagesDominationGame(t *testing.T) (*game.Game, *Engine) {
	t.Helper()
	if issues := game.ValidateCardDef(lullmagesDominationShape()); len(issues) != 0 {
		t.Fatalf("ValidateCardDef = %v", issues)
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)
	return g, NewEngine(nil)
}

func fillGraveyard(g *game.Game, playerID game.PlayerID, n int) {
	for range n {
		addCardToGraveyard(g, playerID, &game.CardDef{CardFace: game.CardFace{Name: "Filler", Types: []types.Card{types.Instant}}})
	}
}

func TestLullmagesDominationCostsUUUAgainstCreatureWhoseControllerHasEightGraveyardCards(t *testing.T) {
	g, engine := newLullmagesDominationGame(t)
	cardID := addCardToHand(g, game.Player1, lullmagesDominationShape())
	stolen := addCreatureWithManaValue(g, game.Player2, 2)
	fillGraveyard(g, game.Player2, 8)
	g.Players[game.Player1].ManaPool.Add(mana.U, 3)

	cast := action.CastSpell(cardID, []game.Target{game.PermanentTarget(stolen.ObjectID)}, 2, nil)
	if !containsAction(engine.legalActions(g, game.Player1), cast) {
		t.Fatal("legal actions did not offer X=2 for {U}{U}{U} against a creature whose controller has eight graveyard cards")
	}
	requireTargetReductionCastSpent(t, g, engine, cast, 0)
	// The reduction is locked in once the total cost is determined (CR 601.2f).
	top, _ := g.Players[game.Player2].Graveyard.Top()
	g.Players[game.Player2].Graveyard.Remove(top)
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := effectiveController(g, stolen); got != game.Player1 {
		t.Fatalf("controller after resolution = %v, want Player1", got)
	}
}

func TestLullmagesDominationReductionReadsTargetControllersGraveyard(t *testing.T) {
	g, engine := newLullmagesDominationGame(t)
	cardID := addCardToHand(g, game.Player1, lullmagesDominationShape())
	fillGraveyard(g, game.Player1, 8)
	fillGraveyard(g, game.Player2, 7)
	fillGraveyard(g, game.Player3, 8)
	// Owned by the player with eight cards but controlled by one with seven.
	borrowed := addCreatureWithManaValue(g, game.Player1, 4)
	borrowed.Controller = game.Player2
	// Owned by the player with seven cards but controlled by one with eight.
	lent := addCreatureWithManaValue(g, game.Player2, 4)
	lent.Controller = game.Player3
	g.Players[game.Player1].ManaPool.Add(mana.U, 3)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)

	requireTargetReductionCastRejected(t, g, engine, action.CastSpell(cardID, []game.Target{game.PermanentTarget(borrowed.ObjectID)}, 4, nil), cardID)
	requireTargetReductionCastSpent(t, g, engine, action.CastSpell(cardID, []game.Target{game.PermanentTarget(lent.ObjectID)}, 4, nil), 0)
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := effectiveController(g, lent); got != game.Player1 {
		t.Fatalf("controller after resolution = %v, want Player1", got)
	}
}

func TestLullmagesDominationCostsFullPriceBelowEightGraveyardCards(t *testing.T) {
	g, engine := newLullmagesDominationGame(t)
	cardID := addCardToHand(g, game.Player1, lullmagesDominationShape())
	target := addCreatureWithManaValue(g, game.Player2, 2)
	fillGraveyard(g, game.Player2, 7)
	g.Players[game.Player1].ManaPool.Add(mana.U, 3)

	cast := action.CastSpell(cardID, []game.Target{game.PermanentTarget(target.ObjectID)}, 2, nil)
	requireTargetReductionCastRejected(t, g, engine, cast, cardID)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	requireTargetReductionCastSpent(t, g, engine, cast, 0)
}

// TestControllerGraveyardCountMatchesOnlyBattlefieldPermanents proves the
// predicate fails closed for a subject without a controller.
func TestControllerGraveyardCountMatchesOnlyBattlefieldPermanents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	fillGraveyard(g, game.Player1, 8)
	creature := addCreatureWithManaValue(g, game.Player1, 1)
	sel := game.Selection{ControllerGraveyardCount: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 8})}
	if !matchSelection(&selectionSubject{kind: subjectPermanent, g: g, permanent: creature}, &sel) {
		t.Fatal("permanent whose controller has eight graveyard cards did not match")
	}
	top, _ := g.Players[game.Player1].Graveyard.Top()
	card := g.CardInstances[top]
	if matchSelection(&selectionSubject{kind: subjectCard, g: g, card: card, controller: game.Player1}, &sel) {
		t.Fatal("graveyard card matched a controller-graveyard predicate")
	}
}
