package rules

import (
	"reflect"
	"testing"

	cardsf "github.com/natefinch/council4/mtg/cards/f"
	cardso "github.com/natefinch/council4/mtg/cards/o"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Source review: Fleshbag Marauder (CLU fce2baa4-2976-4bbd-b6c5-a5a3c6a901be)
// and Olórin's Searing Light (LTC 7496ad0f-fc6d-4154-ac42-c2471e72ed88) Oracle
// text from the pinned corpus. https://api.scryfall.com/cards/<id>/rulings
// returned Fleshbag's 2020-11-10 rulings (it may sacrifice itself; the active
// player chooses first, then each other player in turn order "knowing the
// choices made before them", then all are sacrificed simultaneously) and
// Olórin's four 2023-11-03 rulings (tied greatest power is the player's choice;
// APNAP choices then one simultaneous exile; spell mastery checked on
// resolution; damage uses last-known power). CR 101.4 and 101.4b (later
// players know earlier APNAP choices), 107.1b (a comparison uses a negative
// power as itself) and 701.21a (sacrifice) apply.

func TestFleshbagMarauderLaterPlayersSeeEarlierSacrificeChoices(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	keptP1 := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	chosenP2 := addCombatCreaturePermanentWithPower(g, game.Player2, 1)
	forcedP3 := addCombatCreaturePermanentWithPower(g, game.Player3, 4)
	chosenP4 := addCombatCreaturePermanentWithPower(g, game.Player4, 2)
	keptP4 := addCombatCreaturePermanentWithPower(g, game.Player4, 5)

	cardID := addCardToHand(g, game.Player1, cardsf.FleshbagMarauder())
	g.Players[game.Player1].ManaPool.Add(mana.B, 3)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatal("Fleshbag Marauder cast failed")
	}
	engine.resolveTopOfStackWithChoices(g, simPassPolicies(), &TurnLog{})
	fleshbag := permanentByCardID(g, cardID)
	if fleshbag == nil {
		t.Fatal("Fleshbag Marauder did not enter the battlefield")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Fleshbag Marauder's enters trigger was not put on the stack")
	}

	// The active player sacrifices Fleshbag Marauder itself (option 1).
	p1 := &edictChoiceRecorder{g: g, t: t, choice: []int{1}}
	p2 := &edictChoiceRecorder{g: g, t: t, choice: []int{1}, mustRemain: fleshbag.ObjectID}
	p4 := &edictChoiceRecorder{g: g, t: t, choice: []int{0}, mustRemain: chosenP2.ObjectID}
	agents := simPassPolicies()
	agents[game.Player1], agents[game.Player2], agents[game.Player4] = p1, p2, p4
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if len(p1.requests) != 1 || len(p1.requests[0].SacrificeDecisions) != 0 {
		t.Fatalf("Player1 requests = %#v, want one request with no earlier decisions", p1.requests)
	}
	wantP2 := []game.SacrificeDecision{{Player: game.Player1, Permanents: []id.ID{fleshbag.ObjectID}}}
	if len(p2.requests) != 1 || !reflect.DeepEqual(p2.requests[0].SacrificeDecisions, wantP2) {
		t.Fatalf("Player2 requests = %#v, want earlier decisions %#v", p2.requests, wantP2)
	}
	wantP4 := []game.SacrificeDecision{
		{Player: game.Player1, Permanents: []id.ID{fleshbag.ObjectID}},
		{Player: game.Player2, Permanents: []id.ID{chosenP2.ObjectID}},
		{Player: game.Player3, Permanents: []id.ID{forcedP3.ObjectID}},
	}
	if len(p4.requests) != 1 || !reflect.DeepEqual(p4.requests[0].SacrificeDecisions, wantP4) {
		t.Fatalf("Player4 requests = %#v, want earlier decisions %#v", p4.requests, wantP4)
	}
	if !permanentStillPresent(g, keptP1.ObjectID) || !permanentStillPresent(g, keptP4.ObjectID) {
		t.Fatal("an unchosen creature was sacrificed")
	}
	batches := sacrificeBatchIDs(g, fleshbag, chosenP2, forcedP3, chosenP4)
	if len(batches) != 4 || batches[0] == 0 ||
		batches[1] != batches[0] || batches[2] != batches[0] || batches[3] != batches[0] {
		t.Fatalf("sacrifice batches = %v, want one simultaneous batch of four", batches)
	}
}

func TestAnyNumberEdictLaterPlayersSeeEarlierSacrificeChoices(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	chosenP2 := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	keptP2 := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	keptP3 := addCombatCreaturePermanentWithPower(g, game.Player3, 1)
	chosenP4 := addCombatCreaturePermanentWithPower(g, game.Player4, 1)

	obj := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell, Controller: game.Player1, SourceCardID: g.IDGen.Next()}
	content := game.Mode{Sequence: []game.Instruction{{Primitive: game.SacrificePermanents{
		PlayerGroup: game.OpponentsReference(),
		Selection:   edictCreatureSelection,
		AnyNumber:   true,
	}}}}.Ability()
	p2 := &edictChoiceRecorder{g: g, t: t, choice: []int{0}}
	p3 := &edictChoiceRecorder{g: g, t: t, mustRemain: chosenP2.ObjectID}
	p4 := &edictChoiceRecorder{g: g, t: t, choice: []int{0}, mustRemain: chosenP2.ObjectID}
	agents := simPassPolicies()
	agents[game.Player2], agents[game.Player3], agents[game.Player4] = p2, p3, p4
	engine.resolveAbilityContentWithChoices(g, obj, content, agents, &TurnLog{})

	if len(p2.requests) != 1 || len(p2.requests[0].SacrificeDecisions) != 0 {
		t.Fatalf("Player2 requests = %#v, want one request with no earlier decisions", p2.requests)
	}
	wantP3 := []game.SacrificeDecision{{Player: game.Player2, Permanents: []id.ID{chosenP2.ObjectID}}}
	if len(p3.requests) != 1 || !reflect.DeepEqual(p3.requests[0].SacrificeDecisions, wantP3) {
		t.Fatalf("Player3 requests = %#v, want earlier decisions %#v", p3.requests, wantP3)
	}
	wantP4 := []game.SacrificeDecision{
		{Player: game.Player2, Permanents: []id.ID{chosenP2.ObjectID}},
		{Player: game.Player3},
	}
	if len(p4.requests) != 1 || !reflect.DeepEqual(p4.requests[0].SacrificeDecisions, wantP4) {
		t.Fatalf("Player4 requests = %#v, want earlier decisions %#v", p4.requests, wantP4)
	}
	if permanentStillPresent(g, chosenP2.ObjectID) || permanentStillPresent(g, chosenP4.ObjectID) ||
		!permanentStillPresent(g, keptP2.ObjectID) || !permanentStillPresent(g, keptP3.ObjectID) {
		t.Fatal("the any-number edict did not sacrifice exactly the chosen creatures")
	}
}

func TestOlorinsSearingLightComparesSignedNegativePower(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// Player2: a -1/3 and a 0/2; only the 0/2 has the greatest power.
	negativeP2 := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	zeroP2 := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	modifyPowerUntilEndOfTurn(g, negativeP2, -4)
	modifyPowerUntilEndOfTurn(g, zeroP2, -2)
	// Player3: a -2/1 and a -1/1; every power is negative, -1 is still greatest.
	lowerP3 := addCombatCreaturePermanentWithPower(g, game.Player3, 1)
	higherP3 := addCombatCreaturePermanentWithPower(g, game.Player3, 1)
	modifyPowerUntilEndOfTurn(g, lowerP3, -3)
	modifyPowerUntilEndOfTurn(g, higherP3, -2)

	cardID := addCardToHand(g, game.Player1, cardso.OlórinSSearingLight())
	g.Players[game.Player1].ManaPool.Add(mana.R, 2)
	g.Players[game.Player1].ManaPool.Add(mana.W, 2)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatal("Olórin's Searing Light cast failed")
	}
	p2 := &edictChoiceRecorder{g: g, t: t}
	p3 := &edictChoiceRecorder{g: g, t: t}
	agents := simPassPolicies()
	agents[game.Player2], agents[game.Player3] = p2, p3
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	for _, recorder := range []*edictChoiceRecorder{p2, p3} {
		if len(recorder.requests) != 1 || len(recorder.requests[0].Options) != 1 {
			t.Fatalf("requests = %#v, want only the signed greatest-power creature offered", recorder.requests)
		}
	}
	if !permanentStillPresent(g, negativeP2.ObjectID) || permanentStillPresent(g, zeroP2.ObjectID) {
		t.Fatal("Player2 did not exile the 0-power creature over the -1-power one")
	}
	if !permanentStillPresent(g, lowerP3.ObjectID) || permanentStillPresent(g, higherP3.ObjectID) {
		t.Fatal("Player3 did not exile the -1-power creature over the -2-power one")
	}
}
