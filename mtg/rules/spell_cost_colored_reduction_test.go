package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review (#3856): pinned corpus 6da9ced0 Brush Off
// (873694c0-c100-4d24-a74f-1ab26442dd18, printing 151eab82 sos) {2}{U}{U}
// Instant: "This spell costs {1}{U} less to cast if it targets an instant or
// sorcery spell. Counter target spell." Scryfall rulings retrieved 2026-11-14
// (one, 2026-03-20): the total cost starts with the mana or alternative cost,
// adds increases, then applies reductions such as Brush Off's; mana value
// reads only the mana cost. Official CR effective 2026-11-06 (MagicCompRules
// 20261106.txt) read: 601.2c (targets precede the total cost), 601.2f (total
// cost = mana or alternative cost plus additional costs and increases, minus
// reductions; never below {0}; then locked in), 118.7 and 118.7a (generic
// reductions affect only generic), 118.7b (a colored reduction the cost doesn't
// require reduces generic) and 118.7c (colored reduction beyond that color's
// component reduces generic by the difference). These scenarios use a
// test-local Brush Off; the card itself is authored by a later card unit.

var instantOrSorcerySpell = game.TargetPredicate{
	StackObjectKinds:  []game.StackObjectKind{game.StackSpell},
	SpellCardTypesAny: []types.Card{types.Instant, types.Sorcery},
}

// coloredReductionCounterspell is a counterspell costing manaCost that costs
// generic generic mana and colored mana symbols less to cast if it targets an
// instant or sorcery spell.
func coloredReductionCounterspell(name string, manaCost cost.Mana, generic int, colored ...mana.Color) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		Types:    []types.Card{types.Instant},
		Colors:   []color.Color{color.Blue},
		ManaCost: opt.Val(manaCost),
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Constraint: "target spell",
				Allow:      game.TargetAllowStackObject,
				Predicate:  game.TargetPredicate{StackObjectKinds: []game.StackObjectKind{game.StackSpell}},
			}},
			Sequence: []game.Instruction{{Primitive: game.CounterObject{Object: game.TargetStackObjectReference(0)}}},
		}.Ability()),
		StaticAbilities: []game.StaticAbility{{RuleEffects: []game.RuleEffect{{
			Kind:           game.RuleEffectCostModifier,
			AffectedSource: true,
			CostModifier: game.CostModifier{
				Kind:             game.CostModifierSpell,
				GenericReduction: generic,
				ColoredReduction: colored,
				ReductionTarget:  &game.TargetSpec{Allow: game.TargetAllowStackObject, Predicate: instantOrSorcerySpell},
			},
		}}}},
	}}
}

func brushOffShape() *game.CardDef {
	return coloredReductionCounterspell("Test Brush Off", cost.Mana{cost.O(2), cost.U, cost.U}, 1, mana.U)
}

// pushOpponentCardSpell puts a spell Player2 controls onto the stack, with its
// card instance, and returns its stack-object target and card ID.
func pushOpponentCardSpell(g *game.Game, name string, cardType types.Card) (game.Target, game.StackObject) {
	pushSpell(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		Types:    []types.Card{cardType},
		Colors:   []color.Color{color.Red},
		ManaCost: opt.Val(cost.Mana{cost.O(1), cost.R}),
	}}, false)
	obj, _ := g.Stack.Peek()
	return game.StackObjectTarget(obj.ID), *obj
}

func TestBrushOffCostsOneBlueLessWhenTargetingInstantSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	def := brushOffShape()
	if issues := game.ValidateCardDef(def); len(issues) != 0 {
		t.Fatalf("ValidateCardDef = %v", issues)
	}
	spellID := addCardToHand(g, game.Player1, def)
	creatureSpell, _ := pushOpponentCardSpell(g, "Opponent Creature Spell", types.Creature)
	instantSpell, instantObj := pushOpponentCardSpell(g, "Opponent Instant Spell", types.Instant)
	pool := g.Players[game.Player1].ManaPool

	// {1}{U} still requires one blue mana: the reduction removed only one {U}.
	pool.Add(mana.C, 2)
	requireTargetReductionCastRejected(t, g, engine, action.CastSpell(spellID, []game.Target{instantSpell}, 0, nil), spellID)
	pool.Add(mana.U, 1)
	// A creature spell does not qualify, so the full {2}{U}{U} is due.
	requireTargetReductionCastRejected(t, g, engine, action.CastSpell(spellID, []game.Target{creatureSpell}, 0, nil), spellID)

	cast := action.CastSpell(spellID, []game.Target{instantSpell}, 0, nil)
	if !containsAction(engine.legalActions(g, game.Player1), cast) {
		t.Fatal("legal actions did not offer Brush Off at {1}{U} targeting an instant spell")
	}
	requireTargetReductionCastSpent(t, g, engine, cast, 1)
	if pool.Amount(mana.U) != 0 {
		t.Fatalf("blue mana left = %d, want the {U} spent", pool.Amount(mana.U))
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if top, ok := g.Stack.Peek(); !ok || top.ID == instantObj.ID || g.Stack.Size() != 1 || !g.Players[game.Player2].Graveyard.Contains(instantObj.SourceID) {
		t.Fatal("Brush Off did not counter the targeted instant spell into its owner's graveyard")
	}
}

func TestBrushOffCostsFullPriceWhenTargetingCreatureSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	spellID := addCardToHand(g, game.Player1, brushOffShape())
	creatureSpell, creatureObj := pushOpponentCardSpell(g, "Opponent Creature Spell", types.Creature)
	g.Players[game.Player1].ManaPool.Add(mana.U, 2)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)

	requireTargetReductionCastSpent(t, g, engine, action.CastSpell(spellID, []game.Target{creatureSpell}, 0, nil), 0)
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player2].Graveyard.Contains(creatureObj.SourceID) {
		t.Fatal("Brush Off did not counter the targeted creature spell")
	}
}

// TestColoredReductionAppliesAfterCostIncreases proves the reduction comes off
// the total cost after every increase (CR 601.2f): {2}{U}{U} plus a {1}{U} tax
// minus {1}{U} still requires {2}{U}{U}.
func TestColoredReductionAppliesAfterCostIncreases(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	g.CostModifiers = append(g.CostModifiers, game.CostModifier{
		Kind:            game.CostModifierSpell,
		GenericIncrease: 1,
		ColoredIncrease: []mana.Color{mana.U},
	})
	spellID := addCardToHand(g, game.Player1, brushOffShape())
	instantSpell, _ := pushOpponentCardSpell(g, "Opponent Instant Spell", types.Instant)
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)

	requireTargetReductionCastRejected(t, g, engine, action.CastSpell(spellID, []game.Target{instantSpell}, 0, nil), spellID)
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	requireTargetReductionCastSpent(t, g, engine, action.CastSpell(spellID, []game.Target{instantSpell}, 0, nil), 1)
}

// TestColoredReductionOfAbsentColorReducesGenericWithKicker proves a colored
// reduction lowers the color of a kicker cost it is added to, and lowers
// generic mana when the total cost has no symbol of that color (CR 118.7b):
// {1}{R} costs {R} unkicked and {1}{R} with kicker {U}.
func TestColoredReductionOfAbsentColorReducesGenericWithKicker(t *testing.T) {
	newGame := func() (*game.Game, *Engine, game.Target, *game.CardDef) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		setMainPhasePriority(g, game.Player1)
		def := coloredReductionCounterspell("Test Red Counter", cost.Mana{cost.O(1), cost.R}, 0, mana.U)
		def.StaticAbilities = append(def.StaticAbilities, game.StaticAbility{
			KeywordAbilities: []game.KeywordAbility{game.KickerKeyword{Cost: cost.Mana{cost.U}}},
		})
		if issues := game.ValidateCardDef(def); len(issues) != 0 {
			t.Fatalf("ValidateCardDef = %v", issues)
		}
		target, _ := pushOpponentCardSpell(g, "Opponent Sorcery Spell", types.Sorcery)
		return g, NewEngine(nil), target, def
	}

	g, engine, target, def := newGame()
	spellID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	requireTargetReductionCastSpent(t, g, engine, action.CastSpell(spellID, []game.Target{target}, 0, nil), 0)

	g, engine, target, def = newGame()
	spellID = addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	requireTargetReductionCastRejected(t, g, engine, action.CastKickedSpell(spellID, []game.Target{target}, 0, nil), spellID)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	requireTargetReductionCastSpent(t, g, engine, action.CastKickedSpell(spellID, []game.Target{target}, 0, nil), 0)
	if obj, ok := g.Stack.Peek(); !ok || !obj.KickerPaid {
		t.Fatalf("stack object = %+v, want KickerPaid", obj)
	}
}

// TestColoredReductionAppliesToFlashbackCast proves the reduction also lowers
// an alternative cost paid while casting from the graveyard (CR 601.2f):
// flashback {3}{U}{U} costs {2}{U} targeting an instant spell.
func TestColoredReductionAppliesToFlashbackCast(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	def := brushOffShape()
	def.AlternativeCosts = []cost.Alternative{{
		Label:    "Flashback",
		Mechanic: cost.AlternativeMechanicFlashback,
		ManaCost: opt.Val(cost.Mana{cost.O(3), cost.U, cost.U}),
	}}
	def.StaticAbilities = append(def.StaticAbilities, game.StaticAbility{KeywordAbilities: game.SimpleKeywords(game.Flashback)})
	spellID := addCardToGraveyard(g, game.Player1, def)
	creatureSpell, _ := pushOpponentCardSpell(g, "Opponent Creature Spell", types.Creature)
	instantSpell, _ := pushOpponentCardSpell(g, "Opponent Instant Spell", types.Instant)
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)

	requireTargetReductionCastRejected(t, g, engine, action.CastSpellFromZone(spellID, zone.Graveyard, []game.Target{creatureSpell}, 0, nil), spellID)
	requireTargetReductionCastSpent(t, g, engine, action.CastSpellFromZone(spellID, zone.Graveyard, []game.Target{instantSpell}, 0, nil), 0)
	if obj, ok := g.Stack.Peek(); !ok || !obj.Flashback {
		t.Fatalf("stack object = %+v, want flashback cast", obj)
	}
}

// TestColoredReductionLowersXSpellTotalCost proves legal-action enumeration
// offers X values only affordable once a colored reduction removes a colored
// symbol: {X}{U}{U} minus {U} with {U}{C}{C} affords X=2.
func TestColoredReductionLowersXSpellTotalCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	def := coloredReductionCounterspell("Test X Counter", cost.Mana{cost.X, cost.U, cost.U}, 0, mana.U)
	spellID := addCardToHand(g, game.Player1, def)
	creatureSpell, _ := pushOpponentCardSpell(g, "Opponent Creature Spell", types.Creature)
	instantSpell, _ := pushOpponentCardSpell(g, "Opponent Instant Spell", types.Instant)
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)

	maxX := map[game.Target]int{}
	offered := map[game.Target]bool{}
	for _, act := range engine.legalActions(g, game.Player1) {
		payload, ok := act.CastSpellPayload()
		if !ok || payload.CardID != spellID || len(payload.Targets) != 1 {
			continue
		}
		offered[payload.Targets[0]] = true
		maxX[payload.Targets[0]] = max(maxX[payload.Targets[0]], payload.XValue)
	}
	if maxX[instantSpell] != 2 || offered[creatureSpell] {
		t.Fatalf("offered instant max X %d, creature offered %v; want 2 and false", maxX[instantSpell], offered[creatureSpell])
	}
	requireTargetReductionCastSpent(t, g, engine, action.CastSpell(spellID, []game.Target{instantSpell}, 2, nil), 0)
}

// TestConditionalColoredReduction proves a condition-gated source spell
// reduction can remove a colored symbol ("This spell costs {U} less to cast if
// <condition>.", Ertai's Scorn): {1}{U} costs {1} while the caster controls a
// Wizard.
func TestConditionalColoredReduction(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	def := coloredReductionCounterspell("Test Conditional Counter", cost.Mana{cost.O(1), cost.U}, 0, mana.U)
	modifier := &def.StaticAbilities[0].RuleEffects[0].CostModifier
	modifier.ReductionTarget = nil
	modifier.ReductionCondition = opt.Val(controlsWizardCondition())
	if issues := game.ValidateCardDef(def); len(issues) != 0 {
		t.Fatalf("ValidateCardDef = %v", issues)
	}
	spellID := addCardToHand(g, game.Player1, def)
	target, _ := pushOpponentCardSpell(g, "Opponent Creature Spell", types.Creature)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)

	requireTargetReductionCastRejected(t, g, engine, action.CastSpell(spellID, []game.Target{target}, 0, nil), spellID)
	addWizardPermanent(g, game.Player1)
	requireTargetReductionCastSpent(t, g, engine, action.CastSpell(spellID, []game.Target{target}, 0, nil), 0)
}
