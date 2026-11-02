package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: Engine gap #3519 (preserve the actual caster when control of a
// permanent spell changes). Oracle text from the pinned corpus: Iridescent
// Tiger (e3abbc8b-2bf8-478e-a541-f8019d150054) "When this creature enters, if
// you cast it, add {W}{U}{B}{R}{G}."; Aethersnatch
// (0ce562a9-ca35-4f31-8618-bc26d3d40169) "Gain control of target spell. You may
// choose new targets for it. (If that spell becomes a permanent, it enters under
// your control.)"; Myojin of Night's Reach
// (73bc5505-7fa0-4ebf-a691-b4ca5f52b019) "Myojin of Night's Reach enters with a
// divinity counter on it if you cast it from your hand."; Abzan Guide (Morph
// {2}{W}{B}{G}). Scryfall rulings: Iridescent Tiger 2025-04-04 (triggers if you
// cast it from any zone, not if it is put onto the battlefield without being
// cast); Myojin of Night's Reach 2024-11-08 (casting it from the command zone
// isn't casting it from your hand). CR 2026-09-25 from the official TXT
// (https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt):
// 109.5 ("you" on an object is its controller), 110.2 (a permanent's controller
// is the player under whose control it entered), 601.2 (casting a spell),
// 603.4 (intervening "if" conditions), 707.12 (casting a copy) and 708.4 (a
// face-down cast spell becomes a face-down permanent). Gaining control of a
// spell does not make the new controller its caster: the resulting permanent
// enters under the new controller, so "if you cast it" is false for it; if the
// caster regains control first, the condition is true again.

func stolenCasterEngine() *Engine {
	return NewEngineWithPrograms(nil, cards.NewDefaultRegistry().Programs())
}

func castFromHandForCaster(t *testing.T, engine *Engine, g *game.Game, def *game.CardDef, lands int, land types.Sub) (id.ID, *game.StackObject) {
	t.Helper()
	cardID := addCardToHand(g, game.Player1, def)
	for range lands {
		addBasicLandPermanent(g, game.Player1, land)
	}
	giveMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatalf("Player1 could not cast %s", def.Name)
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != cardID || obj.Controller != game.Player1 {
		t.Fatalf("stack object = %+v, want Player1's %s", obj, def.Name)
	}
	return cardID, obj
}

// snatchSpell has thief cast the registered Aethersnatch on spell and
// resolves it.
func snatchSpell(t *testing.T, engine *Engine, g *game.Game, thief game.PlayerID, spell *game.StackObject) {
	t.Helper()
	snatchID := addCardToHand(g, thief, registryCard(t, "Aethersnatch"))
	for range 6 {
		addBasicLandPermanent(g, thief, types.Island)
	}
	g.Turn.PriorityPlayer = thief
	if !engine.applyAction(g, thief, action.CastSpell(snatchID, []game.Target{game.StackObjectTarget(spell.ID)}, 0, nil)) {
		t.Fatal("Aethersnatch cast failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if top, ok := g.Stack.Peek(); !ok || top.ID != spell.ID || top.Controller != thief {
		t.Fatalf("stack object = %+v, want the spell controlled by Player%d", top, thief+1)
	}
}

func assertEntryCaster(t *testing.T, g *game.Game, permanent *game.Permanent, controller game.PlayerID, from zone.Type) {
	t.Helper()
	if permanent == nil || permanent.Controller != controller {
		t.Fatalf("permanent = %+v, want it controlled by Player%d", permanent, controller+1)
	}
	event, ok := enteredEventFor(g, permanent)
	if !ok || !event.EnterWasCast || !event.EnterHasCastController || event.EnterCastController != game.Player1 ||
		event.EnterCastFromZone != from || event.Controller != controller {
		t.Fatalf("enter event = %+v, want cast by Player1 from %v under Player%d", event, from, controller+1)
	}
}

func assertNoTigerMana(t *testing.T, engine *Engine, g *game.Game) {
	t.Helper()
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Iridescent Tiger's \"if you cast it\" triggered for a controller who did not cast it")
	}
	for _, player := range []game.PlayerID{game.Player1, game.Player2} {
		if total := g.Players[player].ManaPool.Total(); total != 0 {
			t.Fatalf("Player%d mana = %d, want none", player+1, total)
		}
	}
}

func TestStolenTigerSpellKeepsItsCaster(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := stolenCasterEngine()
	tigerID, tiger := castFromHandForCaster(t, engine, g, registryCard(t, "Iridescent Tiger"), 5, types.Mountain)
	snatchSpell(t, engine, g, game.Player2, tiger)
	if !tiger.Caster.Exists || tiger.Caster.Val != game.Player1 {
		t.Fatalf("stolen spell caster = %+v, want Player1", tiger.Caster)
	}

	replay := g.Clone()
	for _, state := range []*game.Game{g, replay} {
		if obj, ok := state.Stack.Peek(); !ok || obj.Caster.Val != game.Player1 || obj.Controller != game.Player2 {
			t.Fatalf("stack object = %+v, want Player2's spell cast by Player1", obj)
		}
		engine.resolveTopOfStack(state, &TurnLog{})
		permanent, _ := findPermanentByCardID(state, tigerID)
		assertEntryCaster(t, state, permanent, game.Player2, zone.Hand)
		assertNoTigerMana(t, engine, state)
	}
	if len(g.Events) != len(replay.Events) {
		t.Fatalf("clone replay diverged: %d events, want %d", len(replay.Events), len(g.Events))
	}
}

func TestTigerReturnedToItsCasterTriggers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := stolenCasterEngine()
	tigerID, tiger := castFromHandForCaster(t, engine, g, registryCard(t, "Iridescent Tiger"), 5, types.Mountain)
	snatchSpell(t, engine, g, game.Player2, tiger)
	snatchSpell(t, engine, g, game.Player1, tiger)
	if tiger.Caster.Val != game.Player1 {
		t.Fatalf("caster after two control changes = %+v, want Player1", tiger.Caster)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	permanent, _ := findPermanentByCardID(g, tigerID)
	assertEntryCaster(t, g, permanent, game.Player1, zone.Hand)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Iridescent Tiger did not trigger under its caster's control")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !wubrgPool(g) {
		t.Fatalf("Player1 mana = %+v, want {W}{U}{B}{R}{G}", g.Players[game.Player1].ManaPool)
	}
}

func TestStolenCastTigerCopyKeepsItsCaster(t *testing.T) {
	g := castCopyProvenanceGame()
	engine := stolenCasterEngine()
	tigerCardID := addCardToExile(g, game.Player1, registryCard(t, "Iridescent Tiger"))
	castCopy := castCopyOfExiledCard(t, engine, g, tigerCardID)
	snatchSpell(t, engine, g, game.Player2, castCopy)
	if castCopy.Caster.Val != game.Player1 || castCopy.CopyOwner.Val != game.Player1 {
		t.Fatalf("stolen cast copy caster=%+v owner=%+v, want Player1 for both", castCopy.Caster, castCopy.CopyOwner)
	}

	replay := g.Clone()
	for _, state := range []*game.Game{g, replay} {
		engine.resolveTopOfStack(state, &TurnLog{})
		token := newestTokenNamed(state, "Iridescent Tiger")
		assertEntryCaster(t, state, token, game.Player2, zone.Exile)
		if token.Owner != game.Player1 {
			t.Fatalf("token owner = Player%d, want the copy's caster Player1", token.Owner+1)
		}
		assertNoTigerMana(t, engine, state)
	}
}

func TestUncastCopyOfStolenSpellHasNoCaster(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := stolenCasterEngine()
	_, tiger := castFromHandForCaster(t, engine, g, registryCard(t, "Iridescent Tiger"), 5, types.Mountain)
	snatchSpell(t, engine, g, game.Player2, tiger)
	addEffectSpellToStack(g, game.Player2, game.CopyStackObject{Object: game.TargetStackObjectReference(0)},
		[]game.Target{game.StackObjectTarget(tiger.ID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	copyObj, _ := g.Stack.Peek()
	if !copyObj.Copy || copyObj.CastCopy || copyObj.Caster.Exists {
		t.Fatalf("copy = %+v, want an uncast copy without a caster", copyObj)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	token := newestTokenNamed(g, "Iridescent Tiger")
	if event, ok := enteredEventFor(g, token); !ok || event.EnterWasCast || event.EnterHasCastController {
		t.Fatalf("uncast copy enter event = %+v, want no cast facts", event)
	}
	assertNoTigerMana(t, engine, g)
}

func TestFaceDownSpellEntryKeepsItsCaster(t *testing.T) {
	for _, tc := range []struct {
		name       string
		thieves    []game.PlayerID
		controller game.PlayerID
	}{
		{name: "caster keeps control", controller: game.Player1},
		{name: "stolen spell", thieves: []game.PlayerID{game.Player2}, controller: game.Player2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := stolenCasterEngine()
			cardID := addCardToHand(g, game.Player1, registryCard(t, "Abzan Guide"))
			for range 3 {
				addBasicLandPermanent(g, game.Player1, types.Plains)
			}
			giveMainPhasePriority(g, game.Player1)
			if !engine.applyAction(g, game.Player1, action.CastFaceDown(cardID, game.FaceFront, game.FaceDownMorph)) {
				t.Fatal("casting Abzan Guide face down failed")
			}
			spell, _ := g.Stack.Peek()
			for _, thief := range tc.thieves {
				snatchSpell(t, engine, g, thief, spell)
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			permanent, _ := findPermanentByCardID(g, cardID)
			if permanent == nil || !permanent.FaceDown {
				t.Fatalf("permanent = %+v, want a face-down Abzan Guide", permanent)
			}
			assertEntryCaster(t, g, permanent, tc.controller, zone.Hand)
		})
	}
}

func TestMyojinCastFromHandReplacementReadsCaster(t *testing.T) {
	for _, tc := range []struct {
		name         string
		thieves      []game.PlayerID
		controller   game.PlayerID
		wantDivinity int
	}{
		{name: "caster keeps control", controller: game.Player1, wantDivinity: 1},
		{name: "stolen spell", thieves: []game.PlayerID{game.Player2}, controller: game.Player2},
		{name: "returned to caster", thieves: []game.PlayerID{game.Player2, game.Player1}, controller: game.Player1, wantDivinity: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := stolenCasterEngine()
			cardID, spell := castFromHandForCaster(t, engine, g, registryCard(t, "Myojin of Night's Reach"), 8, types.Swamp)
			for _, thief := range tc.thieves {
				snatchSpell(t, engine, g, thief, spell)
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			permanent, _ := findPermanentByCardID(g, cardID)
			assertEntryCaster(t, g, permanent, tc.controller, zone.Hand)
			if got := permanent.Counters.Get(counter.Divinity); got != tc.wantDivinity {
				t.Fatalf("divinity counters = %d, want %d", got, tc.wantDivinity)
			}
		})
	}
}
