package rules

// Source review: Engine gap #3439 (sub-issue of #3333) makes cascade a real
// "When you cast this spell" triggered ability. Pinned corpus
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4 and Scryfall
// rulings retrieved for this unit:
//   - Bloodbraid Elf 3f0c9466-5ab9-4205-a84f-b4b27b5a678e {2}{R}{G}: "Haste" /
//     "Cascade (When you cast this spell, exile cards from the top of your library
//     until you exile a nonland card that costs less. You may cast it without
//     paying its mana cost. Put the exiled cards on the bottom in a random
//     order.)". Rulings 2021-06-18: cascade triggers when you cast the spell and
//     resolves before it, so the cast card goes on the stack above the spell; if
//     the spell with cascade is countered, the cascade ability still resolves;
//     the resulting spell must also have lesser mana value.
//   - Maelstrom Wanderer ad9b7fbc-61c8-43ee-a65c-99206fd1e4df {5}{G}{U}{R}:
//     "Cascade, cascade". Rulings 2020-11-10: each instance triggers and resolves
//     separately; the spell cast by the first trigger goes on the stack above the
//     second trigger and resolves before you exile cards for the second; the
//     second still looks for mana value less than Maelstrom Wanderer's.
//   - Shardless Agent 2afbaa9a-c171-4a8b-90f3-5250d8498356 {1}{G}{U}: Cascade.
//   - Etali, Primal Storm 078def07-ae5d-4591-8db6-d156834aab97: "Whenever Etali
//     attacks, exile the top card of each player's library, then you may cast
//     any number of spells from among those cards without paying their mana
//     costs." Ruling 2021-03-19: the spells are cast during resolution; uncast
//     cards remain in exile.
//   - Beast Whisperer 5da7eea8-bb9e-47ce-a554-8a1ee058bd7a: "Whenever you cast a
//     creature spell, draw a card."
//   - Counterspell cc187110-1148-4090-bbb8-e205694a39f5: "Counter target spell."
//   - Stifle b3b00911-ece7-4484-bc36-f211ce72b6cc: "Counter target activated or
//     triggered ability."
// Comprehensive Rules (verified copy effective 2026-09-25): 702.85a (cascade is a
// triggered ability; "costs less" and the resulting spell use the cascade
// spell's mana value), 702.85c (each instance triggers separately), 603.3 and
// 603.3b (triggers wait until a player would receive priority and are ordered by
// their controller in APNAP order), 608.2h (information about an object that left
// its zone uses last known information), 707.10 (non-cast copies are not cast) and
// 707.12 (a cast copy of a card is cast).

import (
	"fmt"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

type cascadeTriggerFixture struct {
	registry *cards.Registry
	engine   *Engine
	game     *game.Game
}

func newCascadeTriggerFixture(t *testing.T) cascadeTriggerFixture {
	t.Helper()
	registry := cards.NewDefaultRegistry()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1
	return cascadeTriggerFixture{registry: registry, engine: NewEngineWithPrograms(nil, registry.Programs()), game: g}
}

// def returns the single registered production definition for an exact name.
// These legacy cards are registered by name; their Oracle identities are
// recorded in the source review above.
func (f cascadeTriggerFixture) def(t *testing.T, name string) *game.CardDef {
	t.Helper()
	defs, err := f.registry.LookupAllChecked(name)
	if err != nil || len(defs) != 1 || defs[0].Name != name {
		t.Fatalf("exact production definition %s: %d definitions, %v", name, len(defs), err)
	}
	return defs[0]
}

// castBloodbraid casts Bloodbraid Elf from Player1's hand with exact mana.
func (f cascadeTriggerFixture) castBloodbraid(t *testing.T) (id.ID, *game.StackObject) {
	t.Helper()
	g := f.game
	cardID := addCardToHand(g, game.Player1, f.def(t, "Bloodbraid Elf"))
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	if !f.engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatal("Bloodbraid Elf cast failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != cardID || obj.Kind != game.StackSpell {
		t.Fatalf("stack top = %+v, want Bloodbraid Elf spell", obj)
	}
	return cardID, obj
}

func cascadeTriggersOnStack(g *game.Game) []*game.StackObject {
	var triggers []*game.StackObject
	for _, obj := range g.Stack.Objects() {
		if isCascadeTrigger(obj) {
			triggers = append(triggers, obj)
		}
	}
	return triggers
}

// TestBloodbraidElfCascadeTriggerResolvesAboveSpell proves cascade waits on the
// stack above Bloodbraid Elf with the library untouched, then exiles, casts a
// lesser card for free and bottoms the rest when it resolves; the trigger also
// survives game cloning independently.
func TestBloodbraidElfCascadeTriggerResolvesAboveSpell(t *testing.T) {
	f := newCascadeTriggerFixture(t)
	g := f.game
	hitID := addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Cascade Hit", 3))
	tooBigID := addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Same Mana Value", 4))
	landID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Forest", Types: []types.Card{types.Land}}})
	bloodbraidID, _ := f.castBloodbraid(t)

	if g.Stack.Size() != 1 || g.Players[game.Player1].Library.Size() != 3 || g.Players[game.Player1].Exile.Size() != 0 {
		t.Fatal("cascade exiled cards or used the stack during casting instead of triggering")
	}
	if !f.engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("casting Bloodbraid Elf did not trigger cascade")
	}
	triggers := cascadeTriggersOnStack(g)
	top, _ := g.Stack.Peek()
	if g.Stack.Size() != 2 || len(triggers) != 1 || top != triggers[0] ||
		top.Controller != game.Player1 || top.SourceCardID != bloodbraidID ||
		top.TriggerEvent.ManaValue != opt.Val(4) {
		t.Fatalf("stack = %+v, want one cascade trigger (mana value 4) above Bloodbraid Elf", g.Stack.Objects())
	}
	if g.Players[game.Player1].Library.Size() != 3 {
		t.Fatal("library changed before the cascade trigger resolved")
	}

	clone := g.Clone()
	f.engine.resolveTopOfStack(clone, &TurnLog{})
	if top, _ := clone.Stack.Peek(); top == nil || top.SourceID != hitID {
		t.Fatalf("cloned cascade trigger did not cast the hit: top=%+v", top)
	}
	if g.Stack.Size() != 2 || g.Players[game.Player1].Library.Size() != 3 || !isCascadeTrigger(g.Stack.Objects()[1]) {
		t.Fatal("resolving the cloned cascade trigger changed the original game")
	}

	f.engine.resolveTopOfStack(g, &TurnLog{})
	objects := g.Stack.Objects()
	if len(objects) != 2 || objects[0].SourceID != bloodbraidID || objects[1].SourceID != hitID {
		t.Fatalf("stack = %+v, want cascaded hit above Bloodbraid Elf", objects)
	}
	if got := g.Players[game.Player1].Library.All(); !sameCardIDs(got, []id.ID{tooBigID, landID}) {
		t.Fatalf("library = %v, want same-mana-value card and land bottomed", got)
	}
	if g.Players[game.Player1].Exile.Size() != 0 {
		t.Fatal("cascade left cards in exile")
	}
	f.engine.resolveTopOfStack(g, &TurnLog{})
	f.engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := findPermanentByCardID(g, bloodbraidID); !ok {
		t.Fatal("Bloodbraid Elf did not resolve after its cascade trigger")
	}
}

// TestCascadeTriggerResolvesAfterCounterspellCountersSpell proves the trigger
// is independent of the spell: Counterspell counters Bloodbraid Elf while the
// trigger waits, and the trigger still cascades below mana value 4.
func TestCascadeTriggerResolvesAfterCounterspellCountersSpell(t *testing.T) {
	f := newCascadeTriggerFixture(t)
	g := f.game
	hitID := addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Cascade Hit", 3))
	tooBigID := addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Same Mana Value", 4))
	bloodbraidID, bloodbraid := f.castBloodbraid(t)
	f.engine.putTriggeredAbilitiesOnStack(g)

	counterID := addCardToHand(g, game.Player2, f.def(t, "Counterspell"))
	g.Players[game.Player2].ManaPool.Add(mana.U, 2)
	g.Turn.PriorityPlayer = game.Player2
	if !f.engine.applyAction(g, game.Player2, action.CastSpell(counterID, []game.Target{game.StackObjectTarget(bloodbraid.ID)}, 0, nil)) {
		t.Fatal("Counterspell targeting Bloodbraid Elf failed while the cascade trigger waited")
	}
	f.engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].Graveyard.Contains(bloodbraidID) || g.Stack.Size() != 1 {
		t.Fatalf("Counterspell did not counter Bloodbraid Elf: stack=%+v", g.Stack.Objects())
	}

	resolveCascadeTrigger(t, f.engine, g, nil)
	top, ok := g.Stack.Peek()
	if !ok || top.SourceID != hitID || g.Stack.Size() != 1 {
		t.Fatalf("stack = %+v, want hit cast from the countered spell's cascade", g.Stack.Objects())
	}
	if got := g.Players[game.Player1].Library.All(); !sameCardIDs(got, []id.ID{tooBigID}) {
		t.Fatalf("library = %v, want mana value 4 card skipped and bottomed", got)
	}
}

// TestStifleCountersCascadeTrigger proves the cascade trigger is an ordinary
// triggered ability: countering it exiles nothing and Bloodbraid Elf resolves.
func TestStifleCountersCascadeTrigger(t *testing.T) {
	f := newCascadeTriggerFixture(t)
	g := f.game
	hitID := addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Cascade Hit", 3))
	bloodbraidID, _ := f.castBloodbraid(t)
	f.engine.putTriggeredAbilitiesOnStack(g)
	trigger, _ := g.Stack.Peek()
	if !isCascadeTrigger(trigger) {
		t.Fatalf("stack top = %+v, want cascade trigger", trigger)
	}

	stifleID := addCardToHand(g, game.Player2, f.def(t, "Stifle"))
	g.Players[game.Player2].ManaPool.Add(mana.U, 1)
	g.Turn.PriorityPlayer = game.Player2
	if !f.engine.applyAction(g, game.Player2, action.CastSpell(stifleID, []game.Target{game.StackObjectTarget(trigger.ID)}, 0, nil)) {
		t.Fatal("Stifle could not target the cascade trigger")
	}
	f.engine.resolveTopOfStack(g, &TurnLog{})
	if len(cascadeTriggersOnStack(g)) != 0 || g.Stack.Size() != 1 {
		t.Fatalf("Stifle did not counter the cascade trigger: stack=%+v", g.Stack.Objects())
	}
	if !g.Players[game.Player1].Library.Contains(hitID) || g.Players[game.Player1].Exile.Size() != 0 {
		t.Fatal("countered cascade trigger still exiled cards")
	}
	f.engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := findPermanentByCardID(g, bloodbraidID); !ok {
		t.Fatal("Bloodbraid Elf did not resolve after its cascade trigger was countered")
	}
}

// TestMaelstromWandererCascadesTwiceWithChainBetween proves "Cascade, cascade"
// creates two triggers; the spell cast by the first (Shardless Agent) and its
// own chained cascade trigger resolve before the second trigger exiles, and the
// second trigger still uses Maelstrom Wanderer's mana value 8.
func TestMaelstromWandererCascadesTwiceWithChainBetween(t *testing.T) {
	f := newCascadeTriggerFixture(t)
	g := f.game
	bigHitID := addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Big Hit", 7))
	smallHitID := addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Small Hit", 1))
	agentID := addCardToLibrary(g, game.Player1, f.def(t, "Shardless Agent"))
	wandererID := addCardToHand(g, game.Player1, f.def(t, "Maelstrom Wanderer"))
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 5)
	if !f.engine.applyAction(g, game.Player1, action.CastSpell(wandererID, nil, 0, nil)) {
		t.Fatal("Maelstrom Wanderer cast failed")
	}
	f.engine.putTriggeredAbilitiesOnStack(g)
	triggers := cascadeTriggersOnStack(g)
	if g.Stack.Size() != 3 || len(triggers) != 2 ||
		triggers[0].TriggerEvent.ManaValue != opt.Val(8) || triggers[1].TriggerEvent.ManaValue != opt.Val(8) {
		t.Fatalf("stack = %+v, want two separate cascade triggers above Maelstrom Wanderer", g.Stack.Objects())
	}

	f.engine.resolveTopOfStack(g, &TurnLog{})
	top, _ := g.Stack.Peek()
	if top == nil || top.SourceID != agentID || top.Kind != game.StackSpell || g.Stack.Size() != 3 {
		t.Fatalf("stack = %+v, want Shardless Agent above the second cascade trigger", g.Stack.Objects())
	}
	if !g.Players[game.Player1].Library.Contains(smallHitID) || !g.Players[game.Player1].Library.Contains(bigHitID) {
		t.Fatal("Shardless Agent's cascade resolved during the first cascade's resolution")
	}

	resolveCascadeTrigger(t, f.engine, g, nil)
	if top, _ := g.Stack.Peek(); top == nil || top.SourceID != smallHitID {
		t.Fatalf("Shardless Agent's cascade did not cast the small hit: stack=%+v", g.Stack.Objects())
	}
	if !g.Players[game.Player1].Library.Contains(bigHitID) {
		t.Fatal("Shardless Agent's cascade ignored its own mana value 3")
	}
	f.engine.resolveTopOfStack(g, &TurnLog{})
	f.engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := findPermanentByCardID(g, agentID); !ok {
		t.Fatal("Shardless Agent did not resolve before the second cascade trigger")
	}
	if top, _ := g.Stack.Peek(); !isCascadeTrigger(top) || !g.Players[game.Player1].Library.Contains(bigHitID) {
		t.Fatal("second cascade trigger exiled before the first trigger's spell resolved")
	}

	f.engine.resolveTopOfStack(g, &TurnLog{})
	if top, _ := g.Stack.Peek(); top == nil || top.SourceID != bigHitID {
		t.Fatalf("second cascade trigger did not use Maelstrom Wanderer's mana value: stack=%+v", g.Stack.Objects())
	}
}

// TestEtaliFinishesBeforeFreeCastBloodbraidCascades proves a spell cast during
// another ability's resolution waits to cascade until that ability finishes;
// Etali's uncast land stays in exile and is never touched by cascade.
func TestEtaliFinishesBeforeFreeCastBloodbraidCascades(t *testing.T) {
	f := newCascadeTriggerFixture(t)
	g := f.game
	etaliDef := f.def(t, "Etali, Primal Storm")
	etali := addCombatPermanent(g, game.Player1, etaliDef)
	hitID := addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Cascade Hit", 2))
	bloodbraidID := addCardToLibrary(g, game.Player1, f.def(t, "Bloodbraid Elf"))
	landID := addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Island", Types: []types.Card{types.Land}}})
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: castFirstResolutionAgent{},
		game.Player2: defaultChoiceAgent{},
		game.Player3: defaultChoiceAgent{},
		game.Player4: defaultChoiceAgent{},
	}

	f.engine.resolveAbilityContentWithChoices(g, triggeredObjFor(etali), etaliDef.TriggeredAbilities[0].Content, agents, &TurnLog{})
	top, _ := g.Stack.Peek()
	if top == nil || top.SourceID != bloodbraidID || g.Stack.Size() != 1 {
		t.Fatalf("stack = %+v, want Bloodbraid Elf cast by Etali", g.Stack.Objects())
	}
	if !g.Players[game.Player1].Library.Contains(hitID) {
		t.Fatal("Bloodbraid Elf cascaded during Etali's resolution")
	}
	if !g.Players[game.Player2].Exile.Contains(landID) {
		t.Fatal("Etali's uncast card did not stay in exile")
	}

	resolveCascadeTrigger(t, f.engine, g, nil)
	if top, _ := g.Stack.Peek(); top == nil || top.SourceID != hitID {
		t.Fatalf("Bloodbraid Elf's cascade did not cast the hit after Etali: stack=%+v", g.Stack.Objects())
	}
	if !g.Players[game.Player2].Exile.Contains(landID) {
		t.Fatal("cascade moved Etali's exiled card")
	}
}

// TestCascadeTriggerOrderedWithBeastWhispererTrigger proves the cascade trigger
// is ordered with the controller's other cast triggers (CR 603.3b): the chosen
// order decides whether cascade exiles before or after Beast Whisperer draws.
func TestCascadeTriggerOrderedWithBeastWhispererTrigger(t *testing.T) {
	for _, cascadeFirst := range []bool{true, false} {
		t.Run(fmt.Sprintf("cascade-resolves-first=%v", cascadeFirst), func(t *testing.T) {
			f := newCascadeTriggerFixture(t)
			g := f.game
			addCombatPermanent(g, game.Player1, f.def(t, "Beast Whisperer"))
			secondID := addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Second Card", 1))
			topID := addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Top Card", 2))
			_, bloodbraid := f.castBloodbraid(t)

			var orderRequests int
			script := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
				if request.Kind != game.ChoiceOrder {
					return fallbackChoice(request)
				}
				orderRequests++
				cascadeLabel := fmt.Sprintf("source=%v ", bloodbraid.SourceID)
				var cascade, other []int
				for _, option := range request.Options {
					if strings.HasPrefix(option.Label, cascadeLabel) {
						cascade = append(cascade, option.Index)
					} else {
						other = append(other, option.Index)
					}
				}
				if len(cascade) != 1 || len(other) != 1 {
					t.Fatalf("order options = %+v, want cascade and Beast Whisperer", request.Options)
				}
				// The ability placed last resolves first.
				if cascadeFirst {
					return append(other, cascade...)
				}
				return append(cascade, other...)
			}}
			agents := [game.NumPlayers]PlayerAgent{script, script, script, script}
			f.engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
			if orderRequests != 1 || g.Stack.Size() != 3 || len(cascadeTriggersOnStack(g)) != 1 {
				t.Fatalf("order requests=%d stack=%+v, want both triggers ordered by Player1", orderRequests, g.Stack.Objects())
			}
			if top, _ := g.Stack.Peek(); isCascadeTrigger(top) != cascadeFirst {
				t.Fatalf("stack top = %+v, want chosen order", top)
			}

			drawn, cast := secondID, topID
			if !cascadeFirst {
				drawn, cast = topID, secondID
				f.engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
				if !g.Players[game.Player1].Hand.Contains(drawn) {
					t.Fatalf("hand = %v, want Beast Whisperer to draw %v first", g.Players[game.Player1].Hand.All(), drawn)
				}
			}
			f.engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
			if top, _ := g.Stack.Peek(); top == nil || top.SourceID != cast {
				t.Fatalf("stack = %+v, want cascade to cast %v", g.Stack.Objects(), cast)
			}
			if cascadeFirst {
				f.engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
				f.engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
				if !g.Players[game.Player1].Hand.Contains(drawn) {
					t.Fatalf("hand = %v, want Beast Whisperer to draw %v after cascade", g.Players[game.Player1].Hand.All(), drawn)
				}
			}
		})
	}
}

// TestCascadeUsesManaValueIncludingX proves the trigger captures the spell's
// mana value on the stack, including the chosen X (CR 202.3e, 702.85a).
func TestCascadeUsesManaValueIncludingX(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spell := cascadeSpell(1)
	spell.ManaCost = opt.Val(cost.Mana{cost.X, cost.O(1)})
	spellID := addCardToHand(g, game.Player1, spell)
	hitID := addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Cascade Hit", 3))
	skippedID := addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Same Mana Value", 4))
	g.Players[game.Player1].ManaPool.Add(mana.C, 4)
	g.Turn.PriorityPlayer = game.Player1
	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 3, nil)) {
		t.Fatal("X cascade spell cast failed")
	}
	engine.putTriggeredAbilitiesOnStack(g)
	if top, _ := g.Stack.Peek(); !isCascadeTrigger(top) || top.TriggerEvent.ManaValue != opt.Val(4) {
		t.Fatalf("stack top = %+v, want cascade trigger capturing mana value 4", top)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if top, _ := g.Stack.Peek(); top == nil || top.SourceID != hitID {
		t.Fatalf("stack = %+v, want hit below X-inclusive mana value", g.Stack.Objects())
	}
	if got := g.Players[game.Player1].Library.All(); !sameCardIDs(got, []id.ID{skippedID}) {
		t.Fatalf("library = %v, want mana value 4 card skipped", got)
	}
}

// TestStormCopiesOfCascadeSpellDoNotCascade proves only the cast spell
// triggers: storm copies are not cast (CR 707.10).
func TestStormCopiesOfCascadeSpellDoNotCascade(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	warmupID := addCardToHand(g, game.Player1, simpleGainLifeInstantWithManaValue("Warmup", 1))
	spell := cascadeSpell(3)
	spell.StaticAbilities = append(spell.StaticAbilities, game.StormStaticBody)
	spellID := addCardToHand(g, game.Player1, spell)
	addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Second Hit", 1))
	addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("First Hit", 2))
	g.Players[game.Player1].ManaPool.Add(mana.C, 4)
	g.Turn.PriorityPlayer = game.Player1
	if !engine.applyAction(g, game.Player1, action.CastSpell(warmupID, nil, 0, nil)) ||
		!engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("casting warmup and storm cascade spell failed")
	}
	copies := 0
	for _, obj := range g.Stack.Objects() {
		if obj.Copy && obj.SourceID == spellID {
			copies++
		}
	}
	if copies != 1 {
		t.Fatalf("storm copies = %d, want 1", copies)
	}
	engine.putTriggeredAbilitiesOnStack(g)
	if got := len(cascadeTriggersOnStack(g)); got != 1 {
		t.Fatalf("cascade triggers = %d, want only the cast spell's trigger", got)
	}
}

// TestMutateCastTriggersCascade proves the mutate cast route is a genuine cast
// that triggers cascade using the mutating spell's own mana value.
func TestMutateCastTriggersCascade(t *testing.T) {
	fixture := newMutateFixture(t)
	g := fixture.game
	card, _ := g.GetCardInstance(fixture.mutatorID)
	card.Def.StaticAbilities = append(card.Def.StaticAbilities, game.CascadeStaticBody)
	hitID := addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Cascade Hit", 4))
	if !fixture.engine.applyAction(g, game.Player1, action.CastMutateSpell(fixture.mutatorID, fixture.target.ObjectID)) {
		t.Fatal("mutate cast failed")
	}
	if !g.Players[game.Player1].Library.Contains(hitID) {
		t.Fatal("mutate cast cascaded before the trigger resolved")
	}
	resolveCascadeTrigger(t, fixture.engine, g, nil)
	top, _ := g.Stack.Peek()
	if top == nil || top.SourceID != hitID || g.Stack.Objects()[0].SourceZone != zone.Hand || !g.Stack.Objects()[0].Mutate {
		t.Fatalf("stack = %+v, want hit above the mutating spell", g.Stack.Objects())
	}
}

// TestPaidCastDuringResolutionTriggersCascadeAfterward proves a spell cast and
// paid for during another object's resolution (the Conduit-like route) waits to
// cascade until that resolution ends, and the cascade still obeys the casting
// ability's "can't cast additional spells" lock.
func TestPaidCastDuringResolutionTriggersCascadeAfterward(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToGraveyard(g, game.Player1, cascadeSpell(2))
	hitID := addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Cascade Hit", 1))
	addConduitLikeSource(g, game.Player1, spellID, t)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)

	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: true}}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if top, _ := g.Stack.Peek(); top == nil || top.SourceID != spellID || g.Stack.Size() != 1 {
		t.Fatalf("stack = %+v, want the paid cascade spell", g.Stack.Objects())
	}
	if !g.Players[game.Player1].Library.Contains(hitID) {
		t.Fatal("paid resolution cast cascaded during the casting ability's resolution")
	}
	before := len(g.Events)
	resolveCascadeTrigger(t, engine, g, nil)
	exiled := false
	for _, event := range g.Events[before:] {
		exiled = exiled || event.CardID == hitID && event.FromZone == zone.Library && event.ToZone == zone.Exile
	}
	// The Conduit-like lock forbids additional spells this turn, so cascade
	// exiles the hit, cannot cast it and bottoms it.
	if !exiled || !g.Players[game.Player1].Library.Contains(hitID) || g.Stack.Size() != 1 {
		t.Fatalf("cascade after the paid cast did not exile and bottom the uncastable hit: stack=%+v", g.Stack.Objects())
	}
}

// TestPreparedFaceCastCopyTriggersCascade proves casting a copy of a prepared
// permanent's spell face is a cast (CR 707.12) whose face's cascade triggers.
func TestPreparedFaceCastCopyTriggersCascade(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	def := vanillaCreature("Prepared Cascade Source", 2, 2)
	def.Layout, def.Alternate = game.LayoutPrepare, opt.Val(cascadeSpell(3).CardFace)
	permanent := addCombatPermanent(g, game.Player1, def)
	permanent.Prepared = true
	hitID := addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Cascade Hit", 2))
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	cast := action.CastSpellFaceFromZone(permanent.CardInstanceID, zone.Battlefield, game.FaceAlternate, nil, 0, nil)
	if !engine.applyAction(g, game.Player1, cast) {
		t.Fatal("prepared face cast failed")
	}
	if !g.Players[game.Player1].Library.Contains(hitID) {
		t.Fatal("prepared face cast cascaded before its trigger resolved")
	}
	resolveCascadeTrigger(t, engine, g, nil)
	objects := g.Stack.Objects()
	if len(objects) != 2 || !objects[0].Copy || objects[1].SourceID != hitID {
		t.Fatalf("stack = %+v, want cascaded hit above the prepared copy", objects)
	}
}

// cascadeCastChoiceAgent answers cascade's free-cast choice: it records each
// request and either declines (the "Don't cast" option) or picks option accept.
// Every other choice takes its default.
type cascadeCastChoiceAgent struct {
	requests []game.ChoiceRequest
	decline  bool
	accept   int
}

func (*cascadeCastChoiceAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *cascadeCastChoiceAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind != game.ChoiceResolution || !strings.Contains(request.Prompt, "without paying its mana cost") {
		return request.DefaultSelection
	}
	a.requests = append(a.requests, request)
	if !a.decline {
		return []int{a.accept}
	}
	for _, option := range request.Options {
		if strings.HasPrefix(option.Label, "Don't cast") {
			return []int{option.Index}
		}
	}
	return nil
}

func declineOptionCount(request game.ChoiceRequest) int {
	count := 0
	for _, option := range request.Options {
		if strings.HasPrefix(option.Label, "Don't cast") {
			count++
		}
	}
	return count
}

// TestBloodbraidElfCascadeMayDeclineLegalHit proves "You may cast it" is
// optional (CR 702.85a): with a castable Llanowar Elves hit the controller is
// offered a decline alongside the cast, and declining casts nothing and bottoms
// the Elves with the other exiled cards.
func TestBloodbraidElfCascadeMayDeclineLegalHit(t *testing.T) {
	f := newCascadeTriggerFixture(t)
	g := f.game
	elvesID := addCardToLibrary(g, game.Player1, f.def(t, "Llanowar Elves"))
	landID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Forest", Types: []types.Card{types.Land}}})
	bloodbraidID, _ := f.castBloodbraid(t)
	casts := spellCastEventCount(g)
	agent := &cascadeCastChoiceAgent{decline: true}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}

	resolveCascadeTrigger(t, f.engine, g, &agents)

	if len(agent.requests) != 1 || len(agent.requests[0].Options) != 2 || declineOptionCount(agent.requests[0]) != 1 {
		t.Fatalf("cascade requests = %+v, want one choice offering casting Llanowar Elves or declining", agent.requests)
	}
	if objects := g.Stack.Objects(); len(objects) != 1 || objects[0].SourceID != bloodbraidID {
		t.Fatalf("stack = %+v, want only Bloodbraid Elf after declining", objects)
	}
	if spellCastEventCount(g) != casts {
		t.Fatal("declining the cascade hit still cast a spell")
	}
	if got := g.Players[game.Player1].Library.All(); !sameCardIDs(got, []id.ID{elvesID, landID}) {
		t.Fatalf("library = %v, want declined Llanowar Elves and the land bottomed", got)
	}
	if g.Players[game.Player1].Exile.Size() != 0 {
		t.Fatal("declined cascade left cards in exile")
	}
}

// TestCascadeDeclineOfferedForTargetedHit proves the decline is a single extra
// option beside every legal way to cast a hit with several target choices
// (Lightning Bolt), and that the same choice can instead cast it at a chosen
// target.
func TestCascadeDeclineOfferedForTargetedHit(t *testing.T) {
	for _, decline := range []bool{true, false} {
		t.Run(fmt.Sprintf("decline=%v", decline), func(t *testing.T) {
			f := newCascadeTriggerFixture(t)
			g := f.game
			boltID := addCardToLibrary(g, game.Player1, f.def(t, "Lightning Bolt"))
			bloodbraidID, _ := f.castBloodbraid(t)
			agent := &cascadeCastChoiceAgent{decline: decline}
			agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}

			resolveCascadeTrigger(t, f.engine, g, &agents)

			if len(agent.requests) != 1 || declineOptionCount(agent.requests[0]) != 1 {
				t.Fatalf("cascade requests = %+v, want one choice with exactly one decline", agent.requests)
			}
			request := agent.requests[0]
			castOptions := len(request.Options) - 1
			if castOptions < 2 || !strings.HasPrefix(request.Options[castOptions].Label, "Don't cast") {
				t.Fatalf("options = %+v, want several Lightning Bolt targets followed by the decline", request.Options)
			}
			objects := g.Stack.Objects()
			if decline {
				if len(objects) != 1 || objects[0].SourceID != bloodbraidID || !g.Players[game.Player1].Library.Contains(boltID) {
					t.Fatalf("stack = %+v, want declined Lightning Bolt bottomed", objects)
				}
				return
			}
			want := request.Options[0].Targets
			if len(objects) != 2 || objects[1].SourceID != boltID || len(objects[1].Targets) != 1 ||
				len(want) != 1 || objects[1].Targets[0] != want[0] {
				t.Fatalf("stack = %+v, want Lightning Bolt cast at %v above Bloodbraid Elf", objects, want)
			}
		})
	}
}
