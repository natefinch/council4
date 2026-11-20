package rules

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: Pinned Oracle revisions and live /cards/<revision>/rulings
// at https://api.scryfall.com (number of rulings in parentheses):
// Lose Focus 985bdb0c-ce6c-4506-8163-76f3b2fdf5fb (2);
// Super Combo 52bed0d4-3bcf-4770-b9c9-2a1d0666d220 (2);
// Train of Thought a9c1c674-0742-45be-afd5-2dbcd6b499ae (0);
// Pyromatics 795a7929-2c32-4006-bdf1-def0a4400f59 (0);
// Gigadrowse bf9f85b3-eb8c-43de-aaac-8a887e9c58da (0);
// Leap of Flame c011b432-ec39-44c8-bf2a-1c3fec32a576 (0);
// Vacuumelt 32dfe9aa-8303-460b-a4c3-e5533af2e1fd (0);
// Shattering Spree d6dcff21-5900-43c4-a38b-cdc19c704ce4 (2).
// Official CR effective 2026-11-06:
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// 601.2b/f–i, 702.33a/c–e: announce the repeatable cost, pay it before
// becoming cast. 603.2–4, 702.56a–b: Replicate triggers only when paid and
// copies per payment of that instance; it is not Kicker. This fixture uses
// real Multikicker to exercise the #3821 Engine seam, not to register Replicate.
// 113.7a, 608.2h, 707.10/.10c/.12 and the Lose Focus/Shattering Spree rulings:
// copies resolve independently even after the spell is countered, can choose
// new targets, keep cast choices, and are not themselves cast.
func repeatableCastCopyFixture() *game.CardDef {
	def := simpleGainLifeInstant("Repeatable cast-cost test spell")
	def.ManaCost = opt.Val(cost.Mana{cost.O(1)})
	def.StaticAbilities = []game.StaticAbility{{
		KeywordAbilities: []game.KeywordAbility{game.KickerKeyword{Cost: cost.Mana{cost.O(1)}, Multi: true}},
	}}
	def.TriggeredAbilities = []game.TriggeredAbility{{
		Trigger: game.TriggerCondition{Pattern: game.TriggerPattern{
			Event: game.EventSpellCast, Source: game.TriggerSourceSelf, SelfWasCast: true, RequireKickerPaid: true,
		}},
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.CopyStackObject{
			Object:              game.EventStackObjectReference(),
			DynamicCount:        game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountEventSpellTimesKicked}),
			MayChooseNewTargets: true,
			UseTriggerSpellLKI:  true,
		}}}}.Ability(),
	}}
	return def
}

func TestSpellCastTriggerCopiesOncePerMultikickerPayment(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	def := repeatableCastCopyFixture()
	def.SpellAbility = opt.Val(game.Mode{
		Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent}},
		Sequence: []game.Instruction{{Primitive: game.Tap{
			Object: game.TargetPermanentReference(0),
		}}},
	}.Ability())
	if issues := game.ValidateCardDef(def); len(issues) != 0 {
		t.Fatalf("fixture rejected: %+v", issues)
	}
	card := addCardToHand(g, game.Player1, def)
	first := addCreaturePermanent(g, game.Player2)
	second := addCreaturePermanent(g, game.Player2)
	third := addCreaturePermanent(g, game.Player2)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	cast := action.CastMultikickedSpellFaceFromZone(card, zone.Hand, game.FaceFront, []game.Target{game.PermanentTarget(first.ObjectID)}, 0, nil, 2)
	play := func(g *game.Game, agents [game.NumPlayers]PlayerAgent) *TurnLog {
		t.Helper()
		log := &TurnLog{}
		if !e.applyActionWithChoices(g, game.Player1, cast, agents, log) {
			t.Fatal("cast with two payments failed")
		}
		original, _ := g.Stack.Peek()
		if !e.putTriggeredAbilitiesOnStackWithChoices(g, agents, log) || g.Stack.Size() != 2 {
			t.Fatal("cast did not put one trigger above the spell")
		}
		// Copy the game while the cast trigger is pending on the stack, then
		// exercise the departure branch and choice replay independently.
		g = g.Clone()
		if !counterStackObject(g, original.ID) || !g.Players[game.Player1].Graveyard.Contains(card) {
			t.Fatal("the original spell was not countered")
		}
		e.resolveTopOfStackWithChoices(g, agents, log)
		copies := g.Stack.Objects()
		if len(copies) != 2 {
			t.Fatalf("cast with two payments produced %d copies, want 2", len(copies))
		}
		for i, target := range []id.ID{second.ObjectID, third.ObjectID} {
			copy := copies[i]
			if !copy.Copy || copy.CastCopy || copy.KickerCount != 2 || copy.ID == original.ID ||
				len(copy.Targets) != 1 || copy.Targets[0].PermanentID != target {
				t.Fatalf("copy %d lost cast choices or its independent target: %+v", i, copy)
			}
		}
		if copies[0].ID == copies[1].ID || spellCastEventCount(g) != 1 || e.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("copies reused an ID or were treated as cast")
		}
		e.resolveTopOfStack(g, log)
		e.resolveTopOfStack(g, log)
		for i, target := range []id.ID{first.ObjectID, second.ObjectID, third.ObjectID} {
			perm, ok := permanentByObjectID(g, target)
			if !ok || perm.Tapped != (i != 0) {
				t.Fatalf("target %d did not reflect the independent resolutions", i)
			}
		}
		return log
	}
	initial := g.Clone()
	nextTarget := 0
	log := play(g, conspireTestAgent(func(req game.ChoiceRequest) []int {
		if req.Prompt == "Choose new targets" {
			if nextTarget >= 2 {
				t.Fatal("unexpected extra new-target choice")
			}
			target := []id.ID{second.ObjectID, third.ObjectID}[nextTarget]
			nextTarget++
			for _, option := range req.Options {
				if len(option.Targets) == 1 && option.Targets[0].PermanentID == target {
					return []int{option.Index}
				}
			}
			t.Fatalf("new target %d was not offered", target)
		}
		return req.DefaultSelection
	}))
	cursor := 0
	replay := play(initial, conspireTestAgent(func(req game.ChoiceRequest) []int {
		if cursor == len(log.Choices) {
			t.Fatal("replay requested an extra decision")
		}
		record := log.Choices[cursor]
		cursor++
		if !reflect.DeepEqual(req, record.Request) {
			t.Fatalf("replay request differs: got %+v, recorded %+v", req, record.Request)
		}
		return record.Selected
	}))
	if cursor != len(log.Choices) || !reflect.DeepEqual(log.Choices, replay.Choices) {
		t.Fatal("replaying the cast and copy choices diverged")
	}
}

func TestSpellCastTriggerMultikickerCombinedCosts(t *testing.T) {
	for _, tc := range []struct {
		name        string
		kicks, mana int
		wantCast    bool
	}{
		{"paid", 2, 4, true},
		{"declined", 0, 4, true},
		{"unaffordable", 2, 3, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			e := NewEngine(nil)
			def := repeatableCastCopyFixture()
			def.AdditionalCosts = []cost.Additional{{Kind: cost.AdditionalDiscard, Amount: 1}}
			def.AdditionalCostChoices = []cost.AdditionalChoice{{Options: []cost.AdditionalChoiceOption{
				{Label: "One mana", Mana: cost.Mana{cost.O(1)}},
				{Label: "Two mana", Mana: cost.Mana{cost.O(2)}},
			}}}
			card := addCardToHand(g, game.Player1, def)
			fodder := addCardToHand(g, game.Player1, vanillaCreature("Discard this", 1, 1))
			player := g.Players[game.Player1]
			player.ManaPool.Add(mana.C, tc.mana)
			cast := action.CastMultikickedSpellFaceFromZone(card, zone.Hand, game.FaceFront, nil, 0, nil, tc.kicks)
			agents := conspireTestAgent(func(req game.ChoiceRequest) []int {
				if req.Prompt == "Choose spell cost" {
					for _, option := range req.Options {
						if option.Label == "One mana" {
							return []int{option.Index}
						}
					}
					t.Fatal("missing mana option for additional cost")
				}
				return req.DefaultSelection
			})
			if got := e.applyActionWithChoices(g, game.Player1, cast, agents, &TurnLog{}); got != tc.wantCast {
				t.Fatalf("cast = %v, want %v", got, tc.wantCast)
			}
			if !tc.wantCast {
				if !player.Hand.Contains(card) || !player.Hand.Contains(fodder) || spellCastEventCount(g) != 0 || !g.Stack.IsEmpty() {
					t.Fatal("an unaffordable cast paid a partial cost or became cast")
				}
				return
			}
			original, _ := g.Stack.Peek()
			if original.ManaSpentToCast != 2+tc.kicks || !player.Graveyard.Contains(fodder) {
				t.Fatalf("paid %d mana, discarded=%v; want %d mana and a discard", original.ManaSpentToCast,
					player.Graveyard.Contains(fodder), 2+tc.kicks)
			}
			if got := e.putTriggeredAbilitiesOnStack(g); got != (tc.kicks > 0) {
				t.Fatalf("presence of paid-only self trigger = %v with %d payments", got, tc.kicks)
			}
			if tc.kicks > 0 {
				e.resolveTopOfStack(g, &TurnLog{})
			}
			if got := g.Stack.Size() - 1; got != tc.kicks {
				t.Fatalf("copies = %d, want %d", got, tc.kicks)
			}
		})
	}
}

func TestSpellCastTriggerMultikickerExileRoutes(t *testing.T) {
	for _, tc := range []struct {
		name       string
		free, copy bool
	}{
		{"paid", false, false},
		{"free", true, false},
		{"cast-card-copy", true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			e := NewEngine(nil)
			cardID := addCardToHand(g, game.Player1, repeatableCastCopyFixture())
			if !moveCardBetweenZones(g, game.Player1, cardID, zone.Hand, zone.Exile) {
				t.Fatal("could not exile the spell card")
			}
			player := g.Players[game.Player1]
			wantMana := 2
			if !tc.free {
				wantMana++
			}
			player.ManaPool.Add(mana.C, wantMana)
			cast := action.CastSpellAction{CardID: cardID, SourceZone: zone.Exile}
			agents := conspireTestAgent(func(req game.ChoiceRequest) []int {
				if req.Prompt == "Choose how many times to pay multikicker." {
					return []int{2}
				}
				return req.DefaultSelection
			})
			log := &TurnLog{}
			if tc.free {
				if !e.castFreeOptionWithCopy(g, player, game.Player1, g.CardInstances[cardID], cast, false, tc.copy, agents, log) {
					t.Fatal("free cast from exile with two optional payments failed")
				}
			} else if !e.castPaidSpellFromSource(g, player, game.Player1, cardID, zone.Exile, agents, log) {
				t.Fatal("paid cast from exile failed")
			}
			original, _ := g.Stack.Peek()
			if original.CastCopy != tc.copy || original.ManaSpentToCast != wantMana {
				t.Fatalf("wrong cast route or payment: %+v", original)
			}
			if !e.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("paid cast from exile did not trigger")
			}
			e.resolveTopOfStack(g, log)
			if g.Stack.Size() != 3 || spellCastEventCount(g) != 1 || e.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("expected a genuine cast followed by two non-cast copies")
			}
			for range 3 {
				e.resolveTopOfStack(g, log)
			}
			if player.Exile.Contains(cardID) != tc.copy || player.Graveyard.Contains(cardID) == tc.copy {
				t.Fatal("cast/copy resolution used the wrong physical card destination")
			}
		})
	}
}

func TestSpellCastTriggerMultikickerStormCopy(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	def := repeatableCastCopyFixture()
	def.StaticAbilities = append(def.StaticAbilities, game.StormStaticBody)
	first := addCardToHand(g, game.Player1, simpleGainLifeInstant("Earlier spell"))
	card := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	if !e.applyAction(g, game.Player1, action.CastSpell(first, nil, 0, nil)) {
		t.Fatal("earlier cast failed")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	if !e.applyAction(g, game.Player1, action.CastMultikickedSpellFaceFromZone(card, zone.Hand, game.FaceFront, nil, 0, nil, 2)) ||
		!e.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 3 {
		t.Fatal("expected separate Storm and paid-cost cast triggers")
	}
	for range 2 {
		trigger, _ := g.Stack.Peek()
		want := 2
		if isStormTrigger(trigger) {
			want = 1
		}
		depth := g.Stack.Size()
		e.resolveTopOfStack(g, &TurnLog{})
		if got := g.Stack.Size() - depth + 1; got != want {
			t.Fatalf("trigger produced %d copies, want %d", got, want)
		}
		for range want {
			copy, _ := g.Stack.Peek()
			if !copy.Copy || copy.CastCopy || copy.KickerCount != 2 {
				t.Fatal("copy lost cast decisions or was treated as cast")
			}
			e.resolveTopOfStack(g, &TurnLog{})
		}
	}
	if spellCastEventCount(g) != 2 || e.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("copies counted as casts or retriggered the spell's abilities")
	}
}
