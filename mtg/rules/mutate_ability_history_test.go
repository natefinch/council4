package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards/g"
	"github.com/natefinch/council4/mtg/cards/r"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review (#3390):
//   - Rumor Gatherer (a7782044-616e-4d4f-b38f-93320ba19797): "Alliance —
//     Whenever another creature you control enters, scry 1. If this is the
//     second time this ability has resolved this turn, draw a card instead."
//     No Scryfall rulings.
//   - Gemrazer (0095245c-a30e-4e2a-88c9-632c678e9f03): "Mutate {1}{G}{G} ...
//     Reach, trample. Whenever this creature mutates, destroy target artifact
//     or enchantment an opponent controls." No Scryfall rulings.
//   - CR 702.140c: a legal mutating creature spell merges with its target and
//     doesn't enter. CR 702.140e: the mutated permanent has all abilities of
//     each component. CR 730.2c: a merged permanent is the same object it was
//     before. CR 113.7a: a triggered ability on the stack exists independently
//     of its source.
//
// Gemrazer over Rumor Gatherer puts Gemrazer's four abilities first, so the
// Alliance ability moves from index 0 to 4 and must keep its resolution count.

func TestMutateOverKeepsTriggeredAbilityResolutionOrdinal(t *testing.T) {
	fx := newMutateHistoryFixture(t)

	fx.enterCreature(t)
	fx.resolveTriggers(t)
	if got := fx.drawn(); got != 0 {
		t.Fatalf("cards drawn after first Alliance resolution = %d, want scry only", got)
	}

	fx.mutateOver(t)
	fx.resolveTriggers(t)

	fx.enterCreature(t)
	fx.resolveTriggers(t)
	if got := fx.drawn(); got != 1 {
		t.Fatalf("cards drawn after second Alliance resolution = %d, want draw instead of scry", got)
	}
	moved := game.TriggeredAbilityUse{SourceID: fx.gatherer.ObjectID, AbilityIndex: fx.offset}
	if got := fx.game.ResolvedTriggeredAbilitiesThisTurn[moved]; got != 2 {
		t.Fatalf("resolution count at moved index = %d, want 2 (all counts %+v)", got, fx.game.ResolvedTriggeredAbilitiesThisTurn)
	}
	stale := game.TriggeredAbilityUse{SourceID: fx.gatherer.ObjectID, AbilityIndex: 0}
	if got := fx.game.ResolvedTriggeredAbilitiesThisTurn[stale]; got != 0 {
		t.Fatalf("Gemrazer's ability at index 0 was credited %d Alliance resolutions", got)
	}
}

func TestMutateOverKeepsResolutionOrdinalForTriggerAlreadyOnStack(t *testing.T) {
	fx := newMutateHistoryFixture(t)

	fx.enterCreature(t)
	fx.resolveTriggers(t)
	if !fx.engine.applyAction(fx.game, game.Player1, action.CastMutateSpell(fx.gemrazerID, fx.gatherer.ObjectID)) {
		t.Fatal("Gemrazer mutate cast failed")
	}
	fx.enterCreature(t)
	if !fx.engine.putTriggeredAbilitiesOnStack(fx.game) {
		t.Fatal("second Alliance trigger was not put on the stack")
	}
	// Order the stack as if Gemrazer were cast in response to the trigger.
	trigger, _ := fx.game.Stack.Pop()
	mutate, _ := fx.game.Stack.Pop()
	if trigger.Kind != game.StackTriggeredAbility || trigger.SourceID != fx.gatherer.ObjectID || mutate.Kind != game.StackSpell {
		t.Fatalf("stack = trigger %+v over spell %+v, want Alliance trigger over Gemrazer", trigger, mutate)
	}
	fx.game.Stack.Push(trigger)
	fx.game.Stack.Push(mutate)

	fx.engine.resolveTopOfStackWithChoices(fx.game, fx.overAgents(), &TurnLog{})
	if trigger.AbilityIndex != fx.offset {
		t.Fatalf("pending Alliance trigger index = %d, want moved to %d", trigger.AbilityIndex, fx.offset)
	}
	fx.engine.putTriggeredAbilitiesOnStack(fx.game)
	fx.engine.resolveTopOfStack(fx.game, &TurnLog{})
	if got := fx.drawn(); got != 1 {
		t.Fatalf("cards drawn after second Alliance resolution = %d, want draw instead of scry", got)
	}
}

func TestMutateUnderKeepsTriggeredAbilityResolutionOrdinal(t *testing.T) {
	fx := newMutateHistoryFixture(t)

	fx.enterCreature(t)
	fx.resolveTriggers(t)
	if !fx.engine.applyAction(fx.game, game.Player1, action.CastMutateSpell(fx.gemrazerID, fx.gatherer.ObjectID)) {
		t.Fatal("Gemrazer mutate cast failed")
	}
	under := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}}}
	fx.engine.resolveTopOfStackWithChoices(fx.game, under, &TurnLog{})
	fx.resolveTriggers(t)

	fx.enterCreature(t)
	fx.resolveTriggers(t)
	if got := fx.drawn(); got != 1 {
		t.Fatalf("cards drawn after second Alliance resolution = %d, want draw instead of scry", got)
	}
}

type mutateHistoryFixture struct {
	game       *game.Game
	engine     *Engine
	gatherer   *game.Permanent
	gemrazerID game.ObjectID
	offset     int
}

func newMutateHistoryFixture(t *testing.T) mutateHistoryFixture {
	t.Helper()
	gm := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	fx := mutateHistoryFixture{
		game:       gm,
		engine:     NewEngine(nil),
		gatherer:   addCombatPermanent(gm, game.Player1, r.RumorGatherer()),
		gemrazerID: addCardToHand(gm, game.Player1, g.Gemrazer()),
		offset:     g.Gemrazer().AbilityCount(),
	}
	for range 3 {
		addBasicLandPermanent(gm, game.Player1, types.Forest)
		addCardToLibrary(gm, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Library Card"}})
	}
	gm.Turn.Phase = game.PhasePrecombatMain
	gm.Turn.Step = game.StepNone
	gm.Turn.PriorityPlayer = game.Player1
	return fx
}

func (fx mutateHistoryFixture) enterCreature(t *testing.T) {
	t.Helper()
	cardID := addCardToHand(fx.game, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Entering Bear", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Bear},
	}})
	card, _ := fx.game.GetCardInstance(cardID)
	fx.game.Players[game.Player1].Hand.Remove(cardID)
	if _, ok := createCardPermanent(fx.game, card, game.Player1, zone.Hand); !ok {
		t.Fatal("creature failed to enter")
	}
}

func (fx mutateHistoryFixture) mutateOver(t *testing.T) {
	t.Helper()
	if !fx.engine.applyAction(fx.game, game.Player1, action.CastMutateSpell(fx.gemrazerID, fx.gatherer.ObjectID)) {
		t.Fatal("Gemrazer mutate cast failed")
	}
	fx.engine.resolveTopOfStackWithChoices(fx.game, fx.overAgents(), &TurnLog{})
	if fx.gatherer.CardInstanceID != fx.gemrazerID {
		t.Fatal("Gemrazer was not put over Rumor Gatherer")
	}
}

func (fx mutateHistoryFixture) drawn() int {
	return 3 - fx.game.Players[game.Player1].Library.Size()
}

func (fx mutateHistoryFixture) overAgents() [game.NumPlayers]PlayerAgent {
	return [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}}}
}

func (fx mutateHistoryFixture) resolveTriggers(t *testing.T) {
	t.Helper()
	fx.engine.putTriggeredAbilitiesOnStack(fx.game)
	for fx.game.Stack.Size() > 0 {
		fx.engine.resolveTopOfStack(fx.game, &TurnLog{})
		fx.engine.putTriggeredAbilitiesOnStack(fx.game)
	}
}
