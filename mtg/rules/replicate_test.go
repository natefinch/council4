package rules

import (
	"math"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: Lose Focus, pinned Scryfall revision
// 985bdb0c-ce6c-4506-8163-76f3b2fdf5fb (MH2 #49), {1}{U}, Instant:
// “Replicate {U} (When you cast this spell, copy it for each time you paid
// its replicate cost. You may choose new targets for the copies.) Counter
// target spell unless its controller pays {2}.” Live Scryfall entry and both
// live rulings agree: the trigger copies after the original leaves the stack,
// and the copies are created on the stack, not cast.
// Official CR, effective 2026-11-06:
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// Read 601.2b/c/f–i (announce choices, determine/pay total, then cast),
// 603.2–4 and 702.56a–b (one trigger per paid instance, copies per payment),
// 113.7a and 608.2h (source independence/LKI), 608.2b/g (target legality
// and mana payment on resolution), and 707.10/.10c/.12 (copied choices,
// new targets, and difference between copying a spell and casting a card copy).
// This is an unregistered gameplay fixture for Engine gap #4028.
func TestLoseFocusReplicateEngine(t *testing.T) {
	g, e := game.NewGame([game.NumPlayers]game.PlayerConfig{}), NewEngine(nil)
	first := addStackSpell(g, game.Player2, "Original target", []types.Card{types.Instant})
	pays := addStackSpell(g, game.Player3, "Controller can pay", []types.Card{types.Instant})
	cannotPay := addStackSpell(g, game.Player4, "Controller cannot pay", []types.Card{types.Instant})
	g.Players[game.Player3].ManaPool.Add(mana.C, 2)

	def := replicateFixture(cost.Mana{cost.U})
	def.Name = "Lose Focus"
	def.ManaCost = opt.Val(cost.Mana{cost.O(1), cost.U})
	def.SpellAbility = opt.Val(game.Mode{
		Targets: []game.TargetSpec{{
			MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowStackObject,
			Predicate: game.TargetPredicate{StackObjectKinds: []game.StackObjectKind{game.StackSpell}},
		}},
		Sequence: unlessPaysCounterInstructions(cost.Mana{cost.O(2)}),
	}.Ability())
	if issues := game.ValidateCardDef(def); len(issues) != 0 {
		t.Fatalf("Lose Focus fixture rejected: %+v", issues)
	}
	card := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	g.Players[game.Player1].ManaPool.Add(mana.U, 3)
	payload := action.CastSpellAction{CardID: card, SourceZone: zone.Hand,
		Targets: []game.Target{game.StackObjectTarget(first.ID)}, ReplicateCounts: []int{2}}
	cast := action.AnnounceSpell(payload)
	if !slices.ContainsFunc(e.legalCastActions(g, game.Player1), func(candidate action.Action) bool {
		return actionsEqual(candidate, cast)
	}) || !e.announcementLegal(g, game.Player1, cast) || !e.applyAction(g, game.Player1, cast) {
		t.Fatal("two Replicate payments were not offered and accepted")
	}
	original, _ := g.Stack.Peek()
	if original.ManaSpentToCast != 4 || original.KickerPaid || !e.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Replicate did not pay {1}{U}{U}{U} and trigger without Kicker")
	}
	trigger, _ := g.Stack.Peek()
	if g.Stack.Size() != 5 || !slices.Equal(trigger.TriggerEvent.ReplicateCounts, []int{2}) {
		t.Fatal("expected only one Replicate trigger with the paid count")
	}

	// Counter on a game clone so the trigger must copy its own spell's LKI.
	clone := g.Clone()
	if !counterStackObject(clone, original.ID) || !clone.Players[game.Player1].Graveyard.Contains(card) || g.Stack.Size() != 5 {
		t.Fatal("countering the source did not affect just the clone")
	}
	targets := []game.Target{game.StackObjectTarget(pays.ID), game.StackObjectTarget(cannotPay.ID)}
	next := 0
	agents := conspireTestAgent(func(req game.ChoiceRequest) []int {
		if req.Prompt == "Choose new targets" {
			if next >= len(targets) {
				t.Fatal("asked for more than two retarget choices")
			}
			for _, option := range req.Options {
				if slices.Equal(option.Targets, targets[next:next+1]) {
					next++
					return []int{option.Index}
				}
			}
			t.Fatal("expected a legal new stack target")
		}
		return req.DefaultSelection
	})
	log := &TurnLog{}
	e.resolveTopOfStackWithChoices(clone, agents, log)
	if next != 2 || clone.Stack.Size() != 5 || spellCastEventCount(clone) != 1 || e.putTriggeredAbilitiesOnStack(clone) {
		t.Fatal("expected two independently retargeted copies and no additional cast")
	}
	for i, copy := range clone.Stack.Objects()[3:] {
		if !copy.Copy || copy.CastCopy || !slices.Equal(copy.ReplicateCounts, []int{2}) ||
			!slices.Equal(copy.Targets, targets[i:i+1]) {
			t.Fatalf("copy %d lost cast choices: %+v", i, copy)
		}
	}
	e.resolveTopOfStack(clone, log)
	e.resolveTopOfStack(clone, log)
	if _, ok := stackObjectByID(clone, pays.ID); !ok {
		t.Fatal("copy countered the spell whose controller paid {2}")
	}
	if _, ok := stackObjectByID(clone, cannotPay.ID); ok || !clone.Players[game.Player4].Graveyard.Contains(cannotPay.SourceID) {
		t.Fatal("copy did not counter the spell whose controller could not pay {2}")
	}
	if _, ok := stackObjectByID(clone, first.ID); !ok {
		t.Fatal("the countered original spell affected its target")
	}
}

func replicateFixture(replicateCosts ...cost.Mana) *game.CardDef {
	def := simpleGainLifeInstant("Replicate fixture")
	def.ManaCost = opt.Val(cost.Mana{cost.O(1)})
	for _, replicateCost := range replicateCosts {
		def.StaticAbilities = append(def.StaticAbilities, game.StaticAbility{
			KeywordAbilities: []game.KeywordAbility{game.ReplicateKeyword{Cost: replicateCost}},
		})
	}
	return def
}

func TestReplicateCombinedCosts(t *testing.T) {
	for _, tc := range []struct {
		name         string
		counts       []int
		generic      int
		wantTriggers int
		wantCast     bool
	}{
		{"paid-two-instances", []int{2, 1}, 5, 2, true},
		{"declined", nil, 3, 0, true},
		{"unaffordable", []int{2, 1}, 4, 0, false},
		{"missing-instance", []int{2}, 5, 0, false},
		{"negative-count", []int{-1, 0}, 5, 0, false},
		{"oversized-cost", []int{math.MaxInt, 0}, 5, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, e := game.NewGame([game.NumPlayers]game.PlayerConfig{}), NewEngine(nil)
			def := replicateFixture(cost.Mana{cost.U}, cost.Mana{cost.O(2)})
			def.StaticAbilities = append(def.StaticAbilities, game.StaticAbility{
				KeywordAbilities: []game.KeywordAbility{game.KickerKeyword{Cost: cost.Mana{cost.O(1)}}},
			})
			def.AdditionalCosts = []cost.Additional{{Kind: cost.AdditionalDiscard, Amount: 1}}
			def.AdditionalCostChoices = []cost.AdditionalChoice{{Options: []cost.AdditionalChoiceOption{
				{Label: "One mana", Mana: cost.Mana{cost.O(1)}},
				{Label: "Two mana", Mana: cost.Mana{cost.O(2)}},
			}}}
			card := addCardToHand(g, game.Player1, def)
			fodder := addCardToHand(g, game.Player1, vanillaCreature("Discard this", 1, 1))
			p := g.Players[game.Player1]
			p.ManaPool.Add(mana.C, tc.generic)
			p.ManaPool.Add(mana.U, 2)
			cast := action.AnnounceSpell(action.CastSpellAction{CardID: card, SourceZone: zone.Hand,
				KickerPaid: true, ReplicateCounts: tc.counts})
			if got := e.applyAction(g, game.Player1, cast); got != tc.wantCast {
				t.Fatalf("cast = %v, want %v", got, tc.wantCast)
			}
			if !tc.wantCast {
				if !p.Hand.Contains(card) || !p.Hand.Contains(fodder) || p.ManaPool.Total() != tc.generic+2 || spellCastEventCount(g) != 0 {
					t.Fatal("rejected announcement paid costs or became cast")
				}
				return
			}
			original, _ := g.Stack.Peek()
			wantMana := 3
			if tc.wantTriggers != 0 {
				wantMana = 7
			}
			if !original.KickerPaid || original.ManaSpentToCast != wantMana || !p.Graveyard.Contains(fodder) {
				t.Fatal("the combined mana, Kicker, choice-group or discard payment was lost")
			}
			e.putTriggeredAbilitiesOnStack(g)
			if g.Stack.Size() != 1+tc.wantTriggers {
				t.Fatal("expected one trigger per paid instance and none when declined")
			}
			var batches []int
			for range tc.wantTriggers {
				depth := g.Stack.Size()
				e.resolveTopOfStack(g, &TurnLog{})
				count := g.Stack.Size() - depth + 1
				batches = append(batches, count)
				for range count {
					e.resolveTopOfStack(g, &TurnLog{})
				}
			}
			slices.Sort(batches)
			if tc.wantTriggers == 2 && !slices.Equal(batches, []int{1, 2}) {
				t.Fatalf("Replicate instances copied using one another's payments: %v", batches)
			}
		})
	}
}

func TestReplicateExileRoutes(t *testing.T) {
	for _, route := range []struct {
		name                string
		free, copy, counter bool
		unaffordable        bool
	}{
		{name: "paid"},
		{name: "free"},
		{name: "free-countered", free: true, counter: true},
		{name: "cast-card-copy", free: true, copy: true},
		{name: "unaffordable-free-copy", free: true, copy: true, unaffordable: true},
	} {
		t.Run(route.name, func(t *testing.T) {
			g, e := game.NewGame([game.NumPlayers]game.PlayerConfig{}), NewEngine(nil)
			def := replicateFixture(cost.Mana{cost.U})
			if route.copy {
				def.StaticAbilities = append(def.StaticAbilities, game.CascadeStaticBody)
			}
			card := addCardToHand(g, game.Player1, def)
			if !moveCardBetweenZones(g, game.Player1, card, zone.Hand, zone.Exile) {
				t.Fatal("could not exile the spell")
			}
			p := g.Players[game.Player1]
			if route.unaffordable {
				p.ManaPool.Add(mana.U, 1)
			} else {
				p.ManaPool.Add(mana.U, 2)
			}
			if !route.free {
				p.ManaPool.Add(mana.C, 1)
			}
			agents := conspireTestAgent(func(req game.ChoiceRequest) []int {
				if req.Kind == game.ChoiceNumber && strings.Contains(req.Prompt, "replicate") {
					return []int{2}
				}
				return req.DefaultSelection
			})
			log := &TurnLog{}
			if route.free {
				cast := action.CastSpellAction{CardID: card, SourceZone: zone.Exile}
				castOK := e.castFreeOptionWithCopy(g, p, game.Player1, g.CardInstances[card], cast, !route.counter, route.copy, agents, log)
				if castOK == route.unaffordable {
					t.Fatalf("free cast from exile = %v, want %v", castOK, !route.unaffordable)
				}
				if !castOK {
					if !p.Exile.Contains(card) || p.ManaPool.Total() != 1 || spellCastEventCount(g) != 0 || !g.Stack.IsEmpty() {
						t.Fatal("unaffordable card-copy cast did not roll back")
					}
					return
				}
			} else if !e.castPaidSpellFromSource(g, p, game.Player1, card, zone.Exile, agents, log) {
				t.Fatal("paid cast from exile with Replicate rejected")
			}
			original, _ := g.Stack.Peek()
			if p.ManaPool.Total() != 0 || !slices.Equal(original.ReplicateCounts, []int{2}) || original.CastCopy != route.copy ||
				!e.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("exile route did not pay Replicate and emit its cast trigger")
			}
			if route.copy {
				objects := g.Stack.Objects()
				if len(objects) != 3 {
					t.Fatal("casting the card copy should trigger both Cascade and Replicate")
				}
				for _, obj := range objects[1:] {
					if isCascadeTrigger(obj) {
						counterStackObject(g, obj.ID)
					}
				}
			}
			if g.Stack.Size() != 2 {
				t.Fatal("expected only the Replicate trigger above the original")
			}
			if route.counter && !counterStackObject(g, original.ID) {
				t.Fatal("failed to counter the free cast")
			}
			depth := g.Stack.Size()
			e.resolveTopOfStack(g, log)
			if g.Stack.Size() != depth+1 || spellCastEventCount(g) != 1 || e.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("expected one genuine cast and two non-cast copies")
			}
			for !g.Stack.IsEmpty() {
				e.resolveTopOfStack(g, log)
			}
			wantExile := route.free && !route.counter
			if p.Exile.Contains(card) != wantExile || p.Graveyard.Contains(card) == wantExile {
				t.Fatalf("physical card in exile=%v, graveyard=%v, want exile=%v",
					p.Exile.Contains(card), p.Graveyard.Contains(card), wantExile)
			}
		})
	}
}

func TestReplicateCombinesWithFlashback(t *testing.T) {
	g, e := game.NewGame([game.NumPlayers]game.PlayerConfig{}), NewEngine(nil)
	def := replicateFixture(cost.Mana{cost.U})
	def.StaticAbilities = append(def.StaticAbilities, game.StaticAbility{
		KeywordAbilities: []game.KeywordAbility{game.FlashbackKeyword{Cost: cost.Mana{cost.R}}},
	})
	card := addCardToHand(g, game.Player1, def)
	if !moveCardBetweenZones(g, game.Player1, card, zone.Hand, zone.Graveyard) {
		t.Fatal("could not move the fixture to the graveyard")
	}
	p := g.Players[game.Player1]
	p.ManaPool.Add(mana.R, 1)
	p.ManaPool.Add(mana.U, 2)
	cast := action.AnnounceSpell(action.CastSpellAction{CardID: card, SourceZone: zone.Graveyard, ReplicateCounts: []int{2}})
	if !e.applyAction(g, game.Player1, cast) {
		t.Fatal("cast did not accept Replicate plus Flashback")
	}
	original, _ := g.Stack.Peek()
	if !original.Flashback || p.ManaPool.Total() != 0 || !e.putTriggeredAbilitiesOnStack(g) ||
		!counterStackObject(g, original.ID) || !p.Exile.Contains(card) {
		t.Fatal("cast did not pay {R}{U}{U}, trigger, and exile the countered original")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	if g.Stack.Size() != 2 || spellCastEventCount(g) != 1 {
		t.Fatal("Flashback cast did not make two non-cast copies after being countered")
	}
}

func TestReplicateStormAndOrdinaryStackCopies(t *testing.T) {
	g, e := game.NewGame([game.NumPlayers]game.PlayerConfig{}), NewEngine(nil)
	def := replicateFixture(cost.Mana{cost.O(1)})
	def.StaticAbilities = append(def.StaticAbilities, game.StormStaticBody)
	first := addCardToHand(g, game.Player1, simpleGainLifeInstant("Earlier spell"))
	card := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	log := &TurnLog{}
	if !e.applyAction(g, game.Player1, action.CastSpell(first, nil, 0, nil)) {
		t.Fatal("earlier cast failed")
	}
	e.resolveTopOfStack(g, log)
	cast := action.AnnounceSpell(action.CastSpellAction{CardID: card, SourceZone: zone.Hand, ReplicateCounts: []int{2}})
	if !e.applyAction(g, game.Player1, cast) || !e.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 3 {
		t.Fatal("wanted separate Storm and Replicate triggers")
	}
	for range 2 {
		trigger, _ := g.Stack.Peek()
		want := 2
		if isStormTrigger(trigger) {
			want = 1
		}
		depth := g.Stack.Size()
		e.resolveTopOfStack(g, log)
		if g.Stack.Size() != depth-1+want {
			t.Fatalf("trigger made the wrong number of copies, want %d", want)
		}
		for range want {
			copy, _ := g.Stack.Peek()
			if !copy.Copy || copy.CastCopy || !slices.Equal(copy.ReplicateCounts, []int{2}) {
				t.Fatal("a Storm or Replicate copy lost the original cast choices")
			}
			e.resolveTopOfStack(g, log)
		}
	}
	if spellCastEventCount(g) != 2 || e.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("non-cast copies counted as cast or recursively triggered")
	}
}
