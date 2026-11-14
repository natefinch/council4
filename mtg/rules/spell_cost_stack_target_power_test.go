package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review (#3857): pinned corpus 6da9ced0 Not of This World
// (5da5d8a3-009e-4d8a-977e-6ebc73ba2682, printing d18ce39e cmm) {7} Kindred
// Instant — Eldrazi: "Counter target spell or ability that targets a permanent
// you control. This spell costs {7} less to cast if it targets a spell or
// ability that targets a creature you control with power 7 or greater."
// Scryfall rulings retrieved 2026-11-14 (2010-06-15): only who controls the
// targeted object and its power are checked, not whether it is still a legal
// target of that spell or ability; the targeted spell or ability may have any
// number of targets as long as one is a permanent you control; one that targets
// only you can't be targeted; both abilities check the current state of those
// targets, so a control change in response makes the counter's target illegal
// on resolution; the power check happens as it is cast, not when the targeted
// spell or ability was cast. Official CR effective 2026-11-06 (MagicCompRules
// 20261106.txt) read: 115.9b ("that targets" checks current state, ignoring a
// target that left its zone), 208.3 (a noncreature permanent has no power),
// 601.2c (targets precede the total cost), 601.2f (reductions, never below
// {0}), 608.2b (illegal targets on resolution). These scenarios use a
// test-local Not of This World; the card itself is authored by a later card
// unit.

func stackObjectThatTargetsPredicate(requirement game.SpellTargetRequirement) game.TargetPredicate {
	return game.TargetPredicate{
		StackObjectKinds: []game.StackObjectKind{game.StackSpell, game.StackActivatedAbility, game.StackTriggeredAbility},
		SpellTargets:     []game.SpellTargetRequirement{requirement},
	}
}

func notOfThisWorldTarget() game.TargetSpec {
	return game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Constraint: "target spell or ability that targets a permanent you control",
		Allow:      game.TargetAllowStackObject,
		Predicate: stackObjectThatTargetsPredicate(game.SpellTargetRequirement{
			Kind:       game.SpellTargetRequirementPermanent,
			Controller: game.ControllerYou,
		}),
	}
}

func notOfThisWorldShape() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Test Not of This World",
		Types:    []types.Card{types.Kindred, types.Instant},
		Subtypes: []types.Sub{types.Eldrazi},
		ManaCost: opt.Val(cost.Mana{cost.O(7)}),
		SpellAbility: opt.Val(game.Mode{
			Targets:  []game.TargetSpec{notOfThisWorldTarget()},
			Sequence: []game.Instruction{{Primitive: game.CounterObject{Object: game.TargetStackObjectReference(0)}}},
		}.Ability()),
		StaticAbilities: []game.StaticAbility{{RuleEffects: []game.RuleEffect{{
			Kind:           game.RuleEffectCostModifier,
			AffectedSource: true,
			CostModifier: game.CostModifier{
				Kind:             game.CostModifierSpell,
				GenericReduction: 7,
				ReductionTarget: &game.TargetSpec{
					Allow: game.TargetAllowStackObject,
					Predicate: stackObjectThatTargetsPredicate(game.SpellTargetRequirement{
						Kind:          game.SpellTargetRequirementPermanent,
						RequiredTypes: []types.Card{types.Creature},
						Controller:    game.ControllerYou,
						Power:         opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 7}),
					}),
				},
			},
		}}}},
	}}
}

func newNotOfThisWorldGame(t *testing.T) (*game.Game, *Engine) {
	t.Helper()
	if issues := game.ValidateCardDef(notOfThisWorldShape()); len(issues) != 0 {
		t.Fatalf("ValidateCardDef = %v", issues)
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)
	return g, NewEngine(nil)
}

// opponentSpellTargeting puts a Player2 instant spell on the stack whose chosen
// targets are targets.
func opponentSpellTargeting(g *game.Game, name string, targets ...game.Target) *game.StackObject {
	obj := addStackSpell(g, game.Player2, name, []types.Card{types.Instant})
	obj.Targets = targets
	return obj
}

func TestNotOfThisWorldCostsZeroCounteringSpellThatTargetsYourPowerSevenCreature(t *testing.T) {
	g, engine := newNotOfThisWorldGame(t)
	cardID := addCardToHand(g, game.Player1, notOfThisWorldShape())
	big := addCombatCreaturePermanentWithPower(g, game.Player1, 7)
	small := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	theirs := addCombatCreaturePermanentWithPower(g, game.Player2, 9)
	atYou := opponentSpellTargeting(g, "Targets You", game.PlayerTarget(game.Player1))
	atTheirs := opponentSpellTargeting(g, "Targets Their Creature", game.PermanentTarget(theirs.ObjectID))
	atSmall := opponentSpellTargeting(g, "Targets Small Creature", game.PermanentTarget(small.ObjectID))
	// Forked Bolt-style: one target is you, the other your power-7 creature.
	atBig := opponentSpellTargeting(g, "Targets You And Big Creature", game.PlayerTarget(game.Player1), game.PermanentTarget(big.ObjectID))

	spec := notOfThisWorldTarget()
	candidates := targetCandidatesForSpecChosenBy(g, game.Player1, game.Player1, notOfThisWorldShape(), 0, game.Event{}, &spec)
	for _, illegal := range []*game.StackObject{atYou, atTheirs} {
		if slices.Contains(candidates, game.StackObjectTarget(illegal.ID)) {
			t.Fatalf("candidates included %d, which targets no permanent its caster controls", illegal.ID)
		}
	}

	// With no mana, only the spell targeting the power-7 creature is castable.
	requireTargetReductionCastRejected(t, g, engine, action.CastSpell(cardID, []game.Target{game.StackObjectTarget(atSmall.ID)}, 0, nil), cardID)
	cast := action.CastSpell(cardID, []game.Target{game.StackObjectTarget(atBig.ID)}, 0, nil)
	if !containsAction(engine.legalActions(g, game.Player1), cast) {
		t.Fatal("legal actions did not offer Not of This World for {0} against a spell targeting a power-7 creature")
	}
	requireTargetReductionCastSpent(t, g, engine, cast, 0)
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := stackObjectByID(g, atBig.ID); ok || !g.Players[game.Player2].Graveyard.Contains(atBig.SourceID) {
		t.Fatal("Not of This World did not counter the spell targeting the power-7 creature")
	}
}

func TestNotOfThisWorldCostsZeroCounteringAbilityThatTargetsYourPowerSevenCreature(t *testing.T) {
	g, engine := newNotOfThisWorldGame(t)
	big := addCombatCreaturePermanentWithPower(g, game.Player1, 6)
	// Its current power, not its printed power, is checked.
	big.Counters.Add(counter.PlusOnePlusOne, 1)
	source := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Opponent Tapper",
		Types: []types.Card{types.Creature},
		ActivatedAbilities: []game.ActivatedAbility{{
			Content: game.Mode{
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
				}},
				Sequence: []game.Instruction{{Primitive: game.Tap{Object: game.TargetPermanentReference(0)}}},
			}.Ability(),
		}},
	}})
	setMainPhasePriority(g, game.Player2)
	if !engine.applyAction(g, game.Player2, action.ActivateAbility(source.ObjectID, 0, []game.Target{game.PermanentTarget(big.ObjectID)}, 0)) {
		t.Fatal("opponent's activated ability failed to reach the stack")
	}
	ability, _ := g.Stack.Peek()
	if ability.Kind != game.StackActivatedAbility {
		t.Fatal("fixture did not create an activated ability")
	}
	g.Turn.PriorityPlayer = game.Player1
	cardID := addCardToHand(g, game.Player1, notOfThisWorldShape())

	requireTargetReductionCastSpent(t, g, engine, action.CastSpell(cardID, []game.Target{game.StackObjectTarget(ability.ID)}, 0, nil), 0)
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Stack.Size() != 0 || big.Tapped {
		t.Fatal("Not of This World did not counter the activated ability before it tapped the creature")
	}
}

func TestNotOfThisWorldCostsFullPriceWithoutAQualifyingCreature(t *testing.T) {
	g, engine := newNotOfThisWorldGame(t)
	cardID := addCardToHand(g, game.Player1, notOfThisWorldShape())
	small := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	// An uncrewed Vehicle is not a creature and has no power (CR 208.3).
	vehicle := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Big Vehicle",
		Types:     []types.Card{types.Artifact},
		Subtypes:  []types.Sub{types.Vehicle},
		Power:     opt.Val(game.PT{Value: 8}),
		Toughness: opt.Val(game.PT{Value: 8}),
	}})
	atVehicle := opponentSpellTargeting(g, "Targets Vehicle", game.PermanentTarget(vehicle.ObjectID))
	trigger := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackTriggeredAbility,
		Controller: game.Player2,
		Targets:    []game.Target{game.PermanentTarget(small.ObjectID)},
	}
	g.Stack.Push(trigger)

	g.Players[game.Player1].ManaPool.Add(mana.C, 6)
	requireTargetReductionCastRejected(t, g, engine, action.CastSpell(cardID, []game.Target{game.StackObjectTarget(atVehicle.ID)}, 0, nil), cardID)
	requireTargetReductionCastRejected(t, g, engine, action.CastSpell(cardID, []game.Target{game.StackObjectTarget(trigger.ID)}, 0, nil), cardID)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	requireTargetReductionCastSpent(t, g, engine, action.CastSpell(cardID, []game.Target{game.StackObjectTarget(trigger.ID)}, 0, nil), 0)
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := stackObjectByID(g, trigger.ID); ok {
		t.Fatal("Not of This World did not counter the triggered ability targeting a permanent its caster controls")
	}
}

// TestNotOfThisWorldFizzlesWhenTargetedPermanentChangesControl proves the
// counter's target is rechecked on resolution against the current controller
// of the permanent its target targets (CR 608.2b, 115.9b).
func TestNotOfThisWorldFizzlesWhenTargetedPermanentChangesControl(t *testing.T) {
	g, engine := newNotOfThisWorldGame(t)
	cardID := addCardToHand(g, game.Player1, notOfThisWorldShape())
	big := addCombatCreaturePermanentWithPower(g, game.Player1, 7)
	atBig := opponentSpellTargeting(g, "Targets Big Creature", game.PermanentTarget(big.ObjectID))

	requireTargetReductionCastSpent(t, g, engine, action.CastSpell(cardID, []game.Target{game.StackObjectTarget(atBig.ID)}, 0, nil), 0)
	big.Controller = game.Player2
	engine.resolveTopOfStack(g, &TurnLog{})
	if top, ok := g.Stack.Peek(); !ok || top.ID != atBig.ID || !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("Not of This World countered a spell that no longer targets a permanent its caster controls")
	}
}
