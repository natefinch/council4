package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// newClassPermanent puts a bare Class enchantment onto the battlefield for
// Player1 and returns the created permanent, exercising the ordinary
// enters-the-battlefield initialization path (CR 716.2a: a Class enters with its
// level equal to 1).
func newClassPermanent(t *testing.T, g *game.Game) *game.Permanent {
	t.Helper()
	cardID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Class",
		Types:    []types.Card{types.Enchantment},
		Subtypes: []types.Sub{types.Class},
	}})
	permanent, ok := createCardPermanent(g, g.CardInstances[cardID], game.Player1, zone.Hand)
	if !ok {
		t.Fatal("createCardPermanent failed")
	}
	return permanent
}

// raiseClassLevel resolves a source-targeted SetClassLevel to the given level,
// mirroring the activated level-up ability lowered from "{cost}: Level N".
func raiseClassLevel(t *testing.T, engine *Engine, g *game.Game, source *game.Permanent, level int) {
	t.Helper()
	obj := &game.StackObject{
		Kind:         game.StackActivatedAbility,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   source.Controller,
	}
	resolveInstruction(engine, g, obj, game.SetClassLevel{
		Amount: game.Fixed(level),
		Object: game.SourcePermanentReference(),
	}, &TurnLog{})
}

// TestClassEntersAtLevelOne verifies a Class permanent enters the battlefield at
// level 1 while a non-Class permanent carries no class level (CR 716.2a).
func TestClassEntersAtLevelOne(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	class := newClassPermanent(t, g)
	if got := class.ClassLevel; got != 1 {
		t.Fatalf("class level on enter = %d, want 1", got)
	}

	plainID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Plain",
		Types: []types.Card{types.Enchantment},
	}})
	plain, ok := createCardPermanent(g, g.CardInstances[plainID], game.Player1, zone.Hand)
	if !ok {
		t.Fatal("createCardPermanent failed for non-Class")
	}
	if got := plain.ClassLevel; got != 0 {
		t.Fatalf("non-Class class level = %d, want 0", got)
	}
}

// TestClassLevelResetsOnNewObject verifies class level is not carried across
// objects: a fresh permanent created from the same card (as when a Class is
// blinked or otherwise re-enters) starts at level 1 even after a prior object of
// that card reached a higher level (CR 716.2a; level is object state, not a
// copiable value, so it resets on a new object).
func TestClassLevelResetsOnNewObject(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	cardID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Class",
		Types:    []types.Card{types.Enchantment},
		Subtypes: []types.Sub{types.Class},
	}})
	card := g.CardInstances[cardID]

	first, ok := createCardPermanent(g, card, game.Player1, zone.Hand)
	if !ok {
		t.Fatal("createCardPermanent failed")
	}
	raiseClassLevel(t, engine, g, first, 3)
	if got := first.ClassLevel; got != 3 {
		t.Fatalf("first object class level = %d, want 3", got)
	}

	second, ok := createCardPermanent(g, card, game.Player1, zone.Hand)
	if !ok {
		t.Fatal("createCardPermanent failed on re-enter")
	}
	if second.ObjectID == first.ObjectID {
		t.Fatal("re-entered permanent reused the prior object ID")
	}
	if got := second.ClassLevel; got != 1 {
		t.Fatalf("re-entered class level = %d, want 1 (level resets on a new object)", got)
	}
}

// TestClassLevelSurvivesControlChange verifies class level is object state that
// persists through a control change: the raised level and its gated band
// abilities remain active after the permanent changes controller (CR 716.2a; a
// control change does not create a new object).
func TestClassLevelSurvivesControlChange(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	source := newClassPermanent(t, g)
	raiseClassLevel(t, engine, g, source, 2)

	source.Controller = game.Player2

	if got := source.ClassLevel; got != 2 {
		t.Fatalf("class level after control change = %d, want 2", got)
	}
	levelTwoGate := opt.Val(game.Condition{SourceClassLevelAtLeast: 2})
	if !activationConditionSatisfied(g, game.Player2, source, levelTwoGate) {
		t.Fatal("level-2 band ability should remain active for the new controller")
	}
}

// TestClassLevelAbilitiesCumulative verifies that once a Class reaches level 3,
// the abilities of every achieved level are simultaneously active (CR 716.4:
// Class levels add abilities cumulatively) and no further level-up activation is
// available (no skipping or repeating past the maximum printed level).
func TestClassLevelAbilitiesCumulative(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	source := newClassPermanent(t, g)
	raiseClassLevel(t, engine, g, source, 3)

	levelTwoBand := opt.Val(game.Condition{SourceClassLevelAtLeast: 2})
	levelThreeBand := opt.Val(game.Condition{SourceClassLevelAtLeast: 3})
	if !activationConditionSatisfied(g, game.Player1, source, levelTwoBand) {
		t.Fatal("level-2 band ability should be active at level 3")
	}
	if !activationConditionSatisfied(g, game.Player1, source, levelThreeBand) {
		t.Fatal("level-3 band ability should be active at level 3")
	}

	reachTwoGate := opt.Val(game.Condition{SourceClassLevelLessThan: 2})
	reachThreeGate := opt.Val(game.Condition{SourceClassLevelAtLeast: 2, SourceClassLevelLessThan: 3})
	if activationConditionSatisfied(g, game.Player1, source, reachTwoGate) {
		t.Fatal("level-up-to-2 ability must be unavailable at level 3")
	}
	if activationConditionSatisfied(g, game.Player1, source, reachThreeGate) {
		t.Fatal("level-up-to-3 ability must be unavailable at level 3")
	}
}

// TestClassLevelUpDoesNotSkipOrLower verifies SetClassLevel only raises the level
// toward its target and never lowers it: a level-up band can only reach the next
// level, and re-resolving a lower target is a no-op (CR 716.2c/716.2d).
func TestClassLevelUpDoesNotSkipOrLower(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	source := newClassPermanent(t, g)
	raiseClassLevel(t, engine, g, source, 2)
	if got := source.ClassLevel; got != 2 {
		t.Fatalf("class level = %d, want 2", got)
	}

	// Re-resolving a level-2 target while already at level 2 does not change it.
	raiseClassLevel(t, engine, g, source, 2)
	if got := source.ClassLevel; got != 2 {
		t.Fatalf("class level after repeat level-2 = %d, want 2", got)
	}

	// A lower target never reduces the level.
	raiseClassLevel(t, engine, g, source, 1)
	if got := source.ClassLevel; got != 2 {
		t.Fatalf("class level after lower target = %d, want 2", got)
	}
}
