package rules

// Source review: Engine gap #3441 (sub-issue of #3333) adds the shared "as you
// cascade" land option; no card is admitted (Averna is a test-local definition).
// Pinned corpus 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4
// and Scryfall rulings retrieved for this unit:
//   - Averna, the Chaos Bloom 23695bc2-86d3-4d78-b374-9e83c4cc9cc4 {G}{U}{R}
//     Legendary Creature — Elemental Shaman 4/2: "As you cascade, you may put a
//     land card from among the exiled cards onto the battlefield tapped. (Do
//     this after the last card is exiled but before deciding whether to cast a
//     nonland card.)" followed by a separate "Whenever you cast a spell with
//     cascade" ability (out of scope). Rulings 2020-11-10: if you don't cast the
//     nonland card or don't reveal one, you can still put a land onto the
//     battlefield before you finish cascading; putting the land isn't playing a
//     land, so it works after your land drop or on another player's turn.
//   - Bloodbraid Elf 3f0c9466-5ab9-4205-a84f-b4b27b5a678e {2}{R}{G}: Haste,
//     Cascade.
//   - Amulet of Vigor (registered): "Whenever a permanent you control enters
//     tapped, untap it."
// Comprehensive Rules (verified copy effective 2026-09-25): 702.85a (cascade
// exiles until a lesser-mana-value nonland card, may cast it, bottoms the rest
// in a random order), 702.85b ("as you cascade" actions happen after you finish
// exiling cards and before you choose whether to cast the last exiled card, or,
// if none was found, before the exiled cards are put on the bottom), 702.85c
// (each instance triggers separately) and 305.4 (putting a land onto the
// battlefield isn't playing a land).

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

const cascadeLandPrompt = "As you cascade, you may put a land card from among the exiled cards onto the battlefield tapped"

func cascadeLandOptionAverna() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       "Averna, the Chaos Bloom",
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{{
			Text:        "As you cascade, you may put a land card from among the exiled cards onto the battlefield tapped.",
			RuleEffects: []game.RuleEffect{{Kind: game.RuleEffectCascadeLandOption}},
		}},
	}}
}

func cascadeTestLand(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: name, Types: []types.Card{types.Land}}}
}

// cascadeChoiceAgent answers the cascade land option and free-cast decision
// from separate scripts, takes the default for any other choice and records
// every cascade-related request in order.
type cascadeChoiceAgent struct {
	lands    [][]int
	casts    [][]int
	requests []game.ChoiceRequest
}

func (*cascadeChoiceAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *cascadeChoiceAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	switch {
	case request.Prompt == cascadeLandPrompt:
		a.requests = append(a.requests, request)
		return nextScripted(&a.lands)
	case strings.HasPrefix(request.Prompt, "Choose how to cast "):
		a.requests = append(a.requests, request)
		return nextScripted(&a.casts)
	}
	return request.DefaultSelection
}

func nextScripted(script *[][]int) []int {
	if len(*script) == 0 {
		return nil
	}
	next := (*script)[0]
	*script = (*script)[1:]
	return next
}

func landOptionRequests(requests []game.ChoiceRequest) []game.ChoiceRequest {
	var found []game.ChoiceRequest
	for _, request := range requests {
		if request.Prompt == cascadeLandPrompt {
			found = append(found, request)
		}
	}
	return found
}

func castCascadeSpellFor(t *testing.T, engine *Engine, g *game.Game, playerID game.PlayerID, manaValue int) id.ID {
	t.Helper()
	cardID := addCardToHand(g, playerID, cascadeSpell(manaValue))
	g.Players[playerID].ManaPool.Add(mana.C, manaValue)
	g.Turn.PriorityPlayer = playerID
	if !engine.applyAction(g, playerID, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatal("cascade spell cast failed")
	}
	return cardID
}

func assertTappedLandUnder(t *testing.T, g *game.Game, cardID id.ID, controller game.PlayerID) *game.Permanent {
	t.Helper()
	permanent, ok := findPermanentByCardID(g, cardID)
	if !ok {
		t.Fatalf("land %v is not on the battlefield", cardID)
	}
	if !permanent.Tapped || permanent.Controller != controller {
		t.Fatalf("land permanent tapped=%v controller=%v, want tapped under %v", permanent.Tapped, permanent.Controller, controller)
	}
	return permanent
}

// The land option comes after the last card is exiled and before the cast
// decision; the chosen land enters tapped, isn't a land play, and the rest
// (including a declined hit) are bottomed as one batch.
func TestCascadeLandOptionPutsChosenLandBeforeCastDecision(t *testing.T) {
	f := newCascadeTriggerFixture(t)
	g := f.game
	averna := addCombatPermanent(g, game.Player1, cascadeLandOptionAverna())
	hitID := addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Cascade Hit", 1))
	islandID := addCardToLibrary(g, game.Player1, f.def(t, "Island"))
	forestID := addCardToLibrary(g, game.Player1, f.def(t, "Forest"))
	g.Turn.LandsPlayedThisTurn = 1
	g.Turn.LandsAllowedThisTurn = 1
	f.castBloodbraid(t)
	agent := &cascadeChoiceAgent{lands: [][]int{{1}}, casts: [][]int{{1}}}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}

	resolveCascadeTrigger(t, f.engine, g, &agents)

	if len(agent.requests) != 2 || agent.requests[0].Prompt != cascadeLandPrompt ||
		!strings.HasPrefix(agent.requests[1].Prompt, "Choose how to cast Cascade Hit") {
		t.Fatalf("choice prompts = %v, want land option then cast decision", choicePrompts(agent.requests))
	}
	land := agent.requests[0]
	if land.Player != game.Player1 || land.MinChoices != 0 || land.MaxChoices != 1 || len(land.Options) != 2 ||
		land.Options[0].Card.Val.CardID != forestID || land.Options[1].Card.Val.CardID != islandID || land.Subject.Val.CardID != averna.CardInstanceID {
		t.Fatalf("land option request = %+v, want optional choice of exiled Forest/Island from Averna", land)
	}
	assertTappedLandUnder(t, g, islandID, game.Player1)
	if g.Turn.LandsPlayedThisTurn != 1 {
		t.Fatalf("lands played = %d, want the put land not counted as a land play", g.Turn.LandsPlayedThisTurn)
	}
	if stackContainsSource(g, hitID) || g.Players[game.Player1].Exile.Size() != 0 {
		t.Fatal("declined hit was cast or cards remained in exile")
	}
	assertLibraryBottomSet(t, g, game.Player1, forestID, hitID)
	assertOneExileBatch(t, g, 2)
}

func TestCascadeLandOptionDeclinedBottomsAllLands(t *testing.T) {
	f := newCascadeTriggerFixture(t)
	g := f.game
	addCombatPermanent(g, game.Player1, cascadeLandOptionAverna())
	hitID := addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Cascade Hit", 1))
	islandID := addCardToLibrary(g, game.Player1, f.def(t, "Island"))
	forestID := addCardToLibrary(g, game.Player1, f.def(t, "Forest"))
	f.castBloodbraid(t)
	agent := &cascadeChoiceAgent{lands: [][]int{{}}, casts: [][]int{{0}}}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}

	resolveCascadeTrigger(t, f.engine, g, &agents)

	if len(landOptionRequests(agent.requests)) != 1 {
		t.Fatalf("choice prompts = %v, want one land option", choicePrompts(agent.requests))
	}
	if _, ok := findPermanentByCardID(g, islandID); ok {
		t.Fatal("declined land option put Island onto the battlefield")
	}
	if !stackContainsSource(g, hitID) {
		t.Fatal("cascade hit was not cast after declining the land option")
	}
	assertLibraryBottomSet(t, g, game.Player1, forestID, islandID)
	assertOneExileBatch(t, g, 2)
}

// With no nonland hit the land option is still offered before bottoming.
func TestCascadeLandOptionOfferedWhenLibraryExhausted(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, cascadeLandOptionAverna())
	expensiveID := addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Too Expensive", 7))
	landID := addCardToLibrary(g, game.Player1, cascadeTestLand("Exhausted Land"))
	castCascadeSpellFor(t, engine, g, game.Player1, 3)
	agent := &cascadeChoiceAgent{lands: [][]int{{0}}, casts: [][]int{{0}}}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}

	resolveCascadeTrigger(t, engine, g, &agents)

	if len(agent.requests) != 1 || agent.requests[0].Prompt != cascadeLandPrompt {
		t.Fatalf("choice prompts = %v, want only the land option", choicePrompts(agent.requests))
	}
	assertTappedLandUnder(t, g, landID, game.Player1)
	if got := g.Players[game.Player1].Library.All(); !sameCardIDs(got, []id.ID{expensiveID}) || g.Stack.Size() != 1 {
		t.Fatalf("library = %v stack = %d, want only the expensive card bottomed and no cast", got, g.Stack.Size())
	}
}

// An opponent's Averna doesn't apply to your cascade.
func TestCascadeLandOptionIgnoresOpponentsEffect(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, cascadeLandOptionAverna())
	hitID := addCardToLibrary(g, game.Player2, simpleGainLifeInstantWithManaValue("Cascade Hit", 1))
	landID := addCardToLibrary(g, game.Player2, cascadeTestLand("Opponent Land"))
	castCascadeSpellFor(t, engine, g, game.Player2, 3)
	recorders := [game.NumPlayers]*cascadeChoiceAgent{}
	var agents [game.NumPlayers]PlayerAgent
	for i := range agents {
		recorders[i] = &cascadeChoiceAgent{lands: [][]int{{0}}, casts: [][]int{{0}}}
		agents[i] = recorders[i]
	}

	resolveCascadeTrigger(t, engine, g, &agents)

	for i, recorder := range recorders {
		if len(landOptionRequests(recorder.requests)) != 0 {
			t.Fatalf("player %d was offered the land option on an opponent's cascade", i)
		}
	}
	if _, ok := findPermanentByCardID(g, landID); ok || !stackContainsSource(g, hitID) {
		t.Fatal("opponent's cascade put a land onto the battlefield or failed to cast its hit")
	}
	assertLibraryBottomSet(t, g, game.Player2, landID)
}

// Without Averna cascade asks nothing about lands (regression).
func TestCascadeWithoutLandOptionAsksNoLandChoice(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	hitID := addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Cascade Hit", 1))
	landID := addCardToLibrary(g, game.Player1, cascadeTestLand("Plain Land"))
	castCascadeSpellFor(t, engine, g, game.Player1, 3)
	agent := &cascadeChoiceAgent{lands: [][]int{{0}}, casts: [][]int{{0}}}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}

	resolveCascadeTrigger(t, engine, g, &agents)

	if len(landOptionRequests(agent.requests)) != 0 {
		t.Fatalf("choice prompts = %v, want no land option without Averna", choicePrompts(agent.requests))
	}
	if !stackContainsSource(g, hitID) {
		t.Fatal("cascade hit was not cast")
	}
	assertLibraryBottomSet(t, g, game.Player1, landID)
}

// Each instance of the effect independently lets one more land enter.
func TestCascadeLandOptionEachInstancePutsOneLand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, cascadeLandOptionAverna())
	addCombatPermanent(g, game.Player1, cascadeLandOptionAverna())
	addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Cascade Hit", 1))
	thirdID := addCardToLibrary(g, game.Player1, cascadeTestLand("Third Land"))
	secondID := addCardToLibrary(g, game.Player1, cascadeTestLand("Second Land"))
	firstID := addCardToLibrary(g, game.Player1, cascadeTestLand("First Land"))
	castCascadeSpellFor(t, engine, g, game.Player1, 3)
	agent := &cascadeChoiceAgent{lands: [][]int{{2}, {0}}, casts: [][]int{{0}}}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}

	resolveCascadeTrigger(t, engine, g, &agents)

	requests := landOptionRequests(agent.requests)
	if len(requests) != 2 || len(requests[0].Options) != 3 || len(requests[1].Options) != 2 {
		t.Fatalf("land option requests = %+v, want one per Averna over the lands still exiled", requests)
	}
	assertTappedLandUnder(t, g, thirdID, game.Player1)
	assertTappedLandUnder(t, g, firstID, game.Player1)
	assertLibraryBottomSet(t, g, game.Player1, secondID)
}

// The land enters through the shared battlefield path, so "enters tapped"
// triggers see it.
func TestCascadeLandOptionLandEntersTappedForTriggers(t *testing.T) {
	f := newCascadeTriggerFixture(t)
	g := f.game
	addCombatPermanent(g, game.Player1, cascadeLandOptionAverna())
	addCombatPermanent(g, game.Player1, f.def(t, "Amulet of Vigor"))
	addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Cascade Hit", 1))
	forestID := addCardToLibrary(g, game.Player1, f.def(t, "Forest"))
	f.castBloodbraid(t)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &cascadeChoiceAgent{lands: [][]int{{0}}, casts: [][]int{{1}}}}

	resolveCascadeTrigger(t, f.engine, g, &agents)
	forest := assertTappedLandUnder(t, g, forestID, game.Player1)
	if g.Players[game.Player1].Exile.Contains(forestID) || forest.Owner != game.Player1 {
		t.Fatal("Forest did not move from exile onto the battlefield")
	}
	log := &TurnLog{}
	if !f.engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, log) {
		t.Fatal("Amulet of Vigor did not trigger on the land entering tapped")
	}
	f.engine.resolveTopOfStackWithChoices(g, agents, log)
	if forest.Tapped {
		t.Fatal("Amulet of Vigor did not untap the cascaded land")
	}
}

func choicePrompts(requests []game.ChoiceRequest) []string {
	prompts := make([]string, len(requests))
	for i, request := range requests {
		prompts[i] = request.Prompt
	}
	return prompts
}
