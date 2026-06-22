package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TestClassLevelGatedAbilityActiveOnlyAtLevel verifies the gating semantics the
// Class level-up slice produces: an ability gated with SourceClassLevelAtLeast
// is inactive until the source Class reaches that level, and a level-up ability
// gated with SourceClassLevelLessThan can only raise the level by one.
func TestClassLevelGatedAbilityActiveOnlyAtLevel(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Class",
		Types:    []types.Card{types.Enchantment},
		Subtypes: []types.Sub{types.Class}},
	})
	card := g.CardInstances[cardID]
	source, ok := createCardPermanent(g, card, game.Player1, zone.Hand)
	if !ok {
		t.Fatal("createCardPermanent failed")
	}

	levelTwoGate := opt.Val(game.Condition{SourceClassLevelAtLeast: 2})
	levelUpGate := opt.Val(game.Condition{SourceClassLevelLessThan: 2})

	if !activationConditionSatisfied(g, game.Player1, source, levelUpGate) {
		t.Fatal("level-up ability should be available at level 1")
	}
	if activationConditionSatisfied(g, game.Player1, source, levelTwoGate) {
		t.Fatal("level-2 band ability should be inactive at level 1")
	}

	obj := &game.StackObject{
		Kind:         game.StackActivatedAbility,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   game.Player1,
	}
	resolveInstruction(engine, g, obj, game.SetClassLevel{Amount: game.Fixed(2), Object: game.SourcePermanentReference()}, &TurnLog{})

	if got := source.ClassLevel; got != 2 {
		t.Fatalf("class level = %d, want 2", got)
	}
	if !activationConditionSatisfied(g, game.Player1, source, levelTwoGate) {
		t.Fatal("level-2 band ability should be active at level 2")
	}
	if activationConditionSatisfied(g, game.Player1, source, levelUpGate) {
		t.Fatal("level-up-to-2 ability should be unavailable once already level 2")
	}
}

// TestClassBecameLevelTriggerFiresAtTargetLevel verifies the "When this Class
// becomes level N" trigger fires exactly when the Class reaches level N: a
// becomes-level-2 trigger resolves (drawing a card) when the level rises to 2,
// and a becomes-level-3 trigger does not fire on the same level rise.
func TestClassBecameLevelTriggerFiresAtTargetLevel(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:            game.EventClassLevelGained,
		Source:           game.TriggerSourceSelf,
		ClassBecameLevel: 2,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	source.ClassLevel = 1

	obj := &game.StackObject{
		Kind:         game.StackActivatedAbility,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   game.Player1,
	}
	resolveInstruction(engine, g, obj, game.SetClassLevel{Amount: game.Fixed(2), Object: game.SourcePermanentReference()}, &TurnLog{})

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("becomes-level-2 trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want becomes-level-2 trigger to draw one card", got)
	}
}

func TestClassBecameLevelTriggerDoesNotFireAtWrongLevel(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:            game.EventClassLevelGained,
		Source:           game.TriggerSourceSelf,
		ClassBecameLevel: 3,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	source.ClassLevel = 1

	obj := &game.StackObject{
		Kind:         game.StackActivatedAbility,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   game.Player1,
	}
	resolveInstruction(engine, g, obj, game.SetClassLevel{Amount: game.Fixed(2), Object: game.SourcePermanentReference()}, &TurnLog{})

	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("becomes-level-3 trigger fired when the Class only reached level 2")
	}
}
