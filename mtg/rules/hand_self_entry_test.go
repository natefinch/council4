package rules

import (
	"testing"

	cardt "github.com/natefinch/council4/mtg/cards/t"
	cardu "github.com/natefinch/council4/mtg/cards/u"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
)

// The hand self-entry family ("{4}: Put this card from your hand onto the
// battlefield.", Talon Gates of Madara) keeps its source card in the hand while
// the activated ability is on the stack and moves it onto the battlefield only
// when the ability resolves. These tests cover the whole activation, using the
// real generated cards, so they exercise the compiler zone recognition, cardgen
// lowering, and engine activation together.

func addManaLands(g *game.Game, controller game.PlayerID, count int) {
	for range count {
		addBasicLandPermanent(g, controller, types.Forest)
	}
}

func TestHandSelfEntryKeepsSourceInHandUntilResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, cardt.TalonGatesOfMadara())
	addManaLands(g, game.Player1, 4)
	setMainPhasePriority(g, game.Player1)

	activation := action.ActivateAbility(cardID, 0, nil, 0)
	if !actionsContain(engine.legalActions(g, game.Player1), activation) {
		t.Fatal("legal actions do not include the hand self-entry activation")
	}
	if !engine.applyAction(g, game.Player1, activation) {
		t.Fatal("applyAction() = false, want the self-entry activation to succeed")
	}

	// The source card must remain in the hand while the ability is on the stack;
	// it is neither discarded as a cost nor already on the battlefield.
	if !g.Players[game.Player1].Hand.Contains(cardID) {
		t.Fatal("self-entry source left the hand at activation, want it to stay until resolution")
	}
	if _, ok := reanimatedPermanent(g, cardID); ok {
		t.Fatal("self-entry source entered the battlefield at activation, want entry at resolution")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.Kind != game.StackActivatedAbility ||
		obj.SourceCardID != cardID {
		t.Fatalf("stack object = %+v, want hand-sourced activated ability", obj)
	}

	engine.resolveTopOfStack(g, nil)

	if g.Players[game.Player1].Hand.Contains(cardID) {
		t.Fatal("self-entry source still in hand after resolution")
	}
	permanent, ok := reanimatedPermanent(g, cardID)
	if !ok {
		t.Fatal("self-entry source did not enter the battlefield at resolution")
	}
	if permanent.Owner != game.Player1 || permanent.Controller != game.Player1 {
		t.Fatalf("entered permanent owner/controller = %v/%v, want Player1/Player1", permanent.Owner, permanent.Controller)
	}
}

func TestHandSelfEntryETBPhasesOutTargetCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, cardt.TalonGatesOfMadara())
	addManaLands(g, game.Player1, 4)
	victim := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Phase Victim",
		Types: []types.Card{types.Creature},
	}})
	setMainPhasePriority(g, game.Player1)

	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{}}
	log := TurnLog{}
	if !engine.applyActionWithChoices(g, game.Player1, action.ActivateAbility(cardID, 0, nil, 0), agents, &log) {
		t.Fatal("self-entry activation failed")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &log)

	if _, ok := reanimatedPermanent(g, cardID); !ok {
		t.Fatal("Talon Gates did not enter the battlefield")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log) {
		t.Fatal("the enters-the-battlefield phase-out trigger was not put on the stack")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want 1 (ETB phase-out trigger)", g.Stack.Size())
	}
	engine.resolveTopOfStackWithChoices(g, agents, &log)

	if !victim.PhasedOut {
		t.Fatal("targeted creature did not phase out from the enters trigger")
	}
}

func TestHandSelfEntryETBOptionalWithNoTargetsResolves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, cardt.TalonGatesOfMadara())
	addManaLands(g, game.Player1, 4)
	setMainPhasePriority(g, game.Player1)

	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{}}
	log := TurnLog{}
	if !engine.applyActionWithChoices(g, game.Player1, action.ActivateAbility(cardID, 0, nil, 0), agents, &log) {
		t.Fatal("self-entry activation failed")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &log)
	if _, ok := reanimatedPermanent(g, cardID); !ok {
		t.Fatal("Talon Gates did not enter the battlefield")
	}
	// "up to one target creature" with no creatures in play resolves with no
	// target and phases nothing out; the trigger must not error or strand.
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log)
	for !g.Stack.IsEmpty() {
		engine.resolveTopOfStackWithChoices(g, agents, &log)
	}
}

func TestHandSelfEntryCounteredLeavesSourceInHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, cardt.TalonGatesOfMadara())
	addManaLands(g, game.Player1, 4)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(cardID, 0, nil, 0)) {
		t.Fatal("self-entry activation failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("self-entry ability was not on the stack")
	}
	if !counterStackObject(g, obj.ID) {
		t.Fatal("countering the self-entry ability failed")
	}

	if !g.Players[game.Player1].Hand.Contains(cardID) {
		t.Fatal("countered self-entry ability moved its source out of the hand")
	}
	if _, ok := reanimatedPermanent(g, cardID); ok {
		t.Fatal("countered self-entry ability still put its source onto the battlefield")
	}
	if !g.Stack.IsEmpty() {
		t.Fatal("stack not empty after countering the self-entry ability")
	}
}

func TestHandSelfEntryFailsClosedWhenSourceLeavesHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, cardt.TalonGatesOfMadara())
	addManaLands(g, game.Player1, 4)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(cardID, 0, nil, 0)) {
		t.Fatal("self-entry activation failed")
	}
	// The source card leaves the hand before the ability resolves (e.g. it was
	// discarded or exiled by another effect). Resolution must put nothing onto
	// the battlefield rather than move a card no longer in the hand.
	g.Players[game.Player1].Hand.Remove(cardID)
	g.Players[game.Player1].Graveyard.Add(cardID)

	engine.resolveTopOfStack(g, nil)

	if _, ok := reanimatedPermanent(g, cardID); ok {
		t.Fatal("self-entry resolution put a card onto the battlefield that had left the hand")
	}
}

func TestUrbanRetreatSelfEntryPaysReturnCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, cardu.UrbanRetreat())
	addManaLands(g, game.Player1, 2)
	returned := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Tapped Creature",
		Types: []types.Card{types.Creature},
	}})
	returned.Tapped = true
	returnedCardID := returned.CardInstanceID
	setMainPhasePriority(g, game.Player1)

	agents := [game.NumPlayers]PlayerAgent{game.Player1: defaultChoiceAgent{}}
	log := TurnLog{}
	activation := action.ActivateAbility(cardID, 0, nil, 0)
	if !actionsContain(engine.legalActions(g, game.Player1), activation) {
		t.Fatal("Urban Retreat self-entry with a return-to-hand cost was not legal")
	}
	if !engine.applyActionWithChoices(g, game.Player1, activation, agents, &log) {
		t.Fatal("Urban Retreat self-entry activation failed")
	}

	// The return-to-hand additional cost is paid at activation: the tapped
	// creature leaves the battlefield for its owner's hand, while the source
	// Urban Retreat stays in the hand until the ability resolves.
	if _, ok := permanentByObjectID(g, returned.ObjectID); ok {
		t.Fatal("return-to-hand cost did not remove the tapped creature from the battlefield")
	}
	if !g.Players[game.Player1].Hand.Contains(returnedCardID) {
		t.Fatal("returned creature did not go to its owner's hand")
	}
	if !g.Players[game.Player1].Hand.Contains(cardID) {
		t.Fatal("Urban Retreat left the hand before resolution")
	}

	engine.resolveTopOfStackWithChoices(g, agents, &log)

	permanent, ok := reanimatedPermanent(g, cardID)
	if !ok {
		t.Fatal("Urban Retreat did not enter the battlefield at resolution")
	}
	if !permanent.Tapped {
		t.Fatal("Urban Retreat did not enter tapped")
	}
}
