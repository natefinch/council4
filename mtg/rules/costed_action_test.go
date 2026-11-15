package rules

import (
	"math/rand/v2"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

// CR source: official 2026-11-06, 601.2a–i, 602.2a–b, 605.1a/3a,
// 701.40b/58b, 702.37e/168d and 734.1–2. Normal payments retain a complete
// witness; only a selected opaque mana ability can require reversal.

type costedObservationAgent struct {
	t      *testing.T
	source id.ID
	seen   int
	choose func(PlayerObservation, game.ChoiceRequest) []int
}

func (*costedObservationAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *costedObservationAgent) ChooseChoice(obs PlayerObservation, req game.ChoiceRequest) []int {
	if req.Kind == game.ChoicePayment {
		a.seen++
		if a.source != 0 {
			pending, present := obs.g.Stack.Peek()
			for _, player := range []game.PlayerID{game.Player1, game.Player2} {
				objects := NewObservation(obs.g, player).Stack()
				if !present || pending.SourceID != a.source || len(objects) == 0 || objects[len(objects)-1].ID != pending.ID {
					a.t.Fatal("payment observed without the announced ability on the stack")
				}
			}
		}
	}
	if a.choose != nil {
		return a.choose(obs, req)
	}
	return req.DefaultSelection
}

func TestCostedActionAnnouncesBeforeEveryPaymentObservation(t *testing.T) {
	for _, route := range activationWindowRoutes() {
		t.Run(route.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			e := NewEngine(nil)
			setSorceryTiming(g, game.Player1)
			source := route.build(t, g)
			for _, name := range route.lands {
				addCombatPermanent(g, game.Player1, registryCardDef(t, name))
			}
			act, ok := findRouteActivation(e, g, route, source)
			if !ok {
				t.Fatal("route not offered")
			}
			a := &costedObservationAgent{t: t, source: source}
			if route.name == "hand cycling Raugrin Crystal" {
				a.choose = func(obs PlayerObservation, req game.ChoiceRequest) []int {
					if !slices.ContainsFunc(NewObservation(obs.g, game.Player2).AnnouncedCards(), func(card CardView) bool {
						return card.CardInstanceID == source && card.Name == "Raugrin Crystal"
					}) {
						t.Fatal("opponent could not observe the revealed hidden-zone source")
					}
					return req.DefaultSelection
				}
			}
			if outcome, _ := e.attemptActionWithChoices(g, game.Player1, act, reboundKickerAgents(a), &TurnLog{}, nil); outcome != actionApplied || a.seen == 0 {
				t.Fatalf("outcome/observations = %v/%d", outcome, a.seen)
			}
			route.done(t, g, source)
			if route.name == "hand cycling Raugrin Crystal" && !slices.ContainsFunc(g.Events, func(event game.Event) bool {
				return event.Kind == game.EventCardRevealed && event.CardID == source
			}) {
				t.Fatal("hidden-zone activation source was not revealed")
			}
		})
	}
	t.Run("registered Spellbomb draw", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		e := NewEngine(nil)
		setSorceryTiming(g, game.Player1)
		bomb := addCombatPermanent(g, game.Player1, registryCardDef(t, "Aether Spellbomb"))
		addCombatPermanent(g, game.Player1, registryCardDef(t, "Waterlogged Grove"))
		act, ok := findRouteActivation(e, g, activationWindowRoute{activation: activationOf(0)}, bomb.ObjectID)
		if !ok {
			t.Fatal("draw activation not offered")
		}
		a := &costedObservationAgent{t: t, source: bomb.ObjectID}
		if outcome, _ := e.attemptActionWithChoices(g, game.Player1, act, reboundKickerAgents(a), &TurnLog{}, nil); outcome != actionApplied || a.seen == 0 {
			t.Fatalf("outcome/observations = %v/%d", outcome, a.seen)
		}
	})
}

func TestCostedActionGroveOnlyXOne(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	setSorceryTiming(g, game.Player1)
	pinnacle := addCombatPermanent(g, game.Player1, registryCardDef(t, "Mirror Entity"))
	bear := addCombatPermanent(g, game.Player1, registryCardDef(t, "Grizzly Bears"))
	grove := addCombatPermanent(g, game.Player1, registryCardDef(t, "Waterlogged Grove"))
	var chosen action.Action
	for _, act := range e.legalActivateAbilityActions(g, game.Player1) {
		if activation, ok := act.ActivateAbilityPayload(); ok && activation.SourceID == pinnacle.ObjectID && activation.XValue == 1 {
			chosen = act
		}
	}
	if chosen.Kind != action.ActionActivateAbility {
		t.Fatal("X=1 payable through Grove was not offered")
	}
	if !e.announcementLegal(g, game.Player1, chosen) || !e.applyAction(g, game.Player1, chosen) {
		t.Fatal("X=1 could not be announced and paid")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	toughness, _ := effectiveToughness(g, bear)
	if !grove.Tapped || g.Players[game.Player1].Life != 39 || effectivePower(g, bear) != 1 || toughness != 1 {
		t.Fatal("Mirror Entity did not play its X text with Grove's mana")
	}
}

func TestCostedActionGroveOnlyXOneWithTapCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	setSorceryTiming(g, game.Player1)
	magus := addCombatPermanent(g, game.Player1, registryCardDef(t, "Magus of the Candelabra"))
	grove := addCombatPermanent(g, game.Player1, registryCardDef(t, "Waterlogged Grove"))
	var chosen action.Action
	for _, act := range e.legalActivateAbilityActions(g, game.Player1) {
		if activation, ok := act.ActivateAbilityPayload(); ok && activation.SourceID == magus.ObjectID &&
			activation.XValue == 1 && len(activation.Targets) == 1 && activation.Targets[0].PermanentID == grove.ObjectID {
			chosen = act
		}
	}
	if chosen.Kind != action.ActionActivateAbility || !e.applyAction(g, game.Player1, chosen) || !magus.Tapped || !grove.Tapped {
		t.Fatal("X=1 with a tap cost was not offered and paid using only Grove")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	if grove.Tapped || g.Players[game.Player1].Life != 39 {
		t.Fatal("Magus did not untap the Grove that paid its activation")
	}
}

func TestCostedActionManaActivationPaysThroughGrove(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	setSorceryTiming(g, game.Player1)
	filter := addCombatPermanent(g, game.Player1, colossusFilterDef())
	grove := addCombatPermanent(g, game.Player1, registryCardDef(t, "Waterlogged Grove"))
	bear := addCombatPermanent(g, game.Player1, registryCardDef(t, "Grizzly Bears"))
	activate := action.ActivateAbility(filter.ObjectID, 0, nil, 0)
	if !actionsContain(e.legalActivateAbilityActions(g, game.Player1), activate) || !e.applyAction(g, game.Player1, activate) {
		t.Fatal("mana activation payable only through Grove was not offered and paid")
	}
	if !grove.Tapped || g.Players[game.Player1].Life != 39 || g.Players[game.Player1].ManaPool.Amount(mana.C) != 2 ||
		!g.Players[game.Player1].Graveyard.Contains(bear.CardInstanceID) || g.Stack.Size() != 0 {
		t.Fatal("costed mana activation did not pay its costs and resolve immediately")
	}
}

func TestCostedActionCyclingWithTwoAltarActivationOccurrences(t *testing.T) {
	for _, reverseOrder := range []bool{false, true} {
		t.Run(map[bool]string{false: "first creature", true: "second creature"}[reverseOrder], func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			e := NewEngine(nil)
			setSorceryTiming(g, game.Player1)
			crystalDef := registryCardDef(t, "Raugrin Crystal")
			crystal := addCardToHand(g, game.Player1, crystalDef)
			altar := addCombatPermanent(g, game.Player1, registryCardDef(t, "Phyrexian Altar"))
			first := addCombatPermanent(g, game.Player1, registryCardDef(t, "Grizzly Bears"))
			second := addCombatPermanent(g, game.Player1, registryCardDef(t, "Grizzly Bears"))
			draw := addCardToLibrary(g, game.Player1, registryCardDef(t, "Forest"))
			wanted := first.ObjectID
			if reverseOrder {
				wanted = second.ObjectID
			}
			colors := []string{"G", "U"}
			chosenColors, chosenCreature := 0, false
			agent := authoredChoicePolicy{choose: func(obs PlayerObservation, req game.ChoiceRequest) []int {
				if pending, ok := obs.g.Stack.Peek(); !ok || pending.SourceCardID != crystal {
					t.Fatal("cycling ability was not on the stack during payment")
				}
				if req.Prompt == "Sacrifice a creature" {
					for _, option := range req.Options {
						if len(option.Targets) == 1 && option.Targets[0].PermanentID == wanted {
							chosenCreature = true
							return []int{option.Index}
						}
					}
				}
				if req.Prompt == "Choose a color" {
					if chosenColors >= len(colors) {
						t.Fatal("more than two Altar output choices")
					}
					for _, option := range req.Options {
						if option.Label == colors[chosenColors] {
							chosenColors++
							return []int{option.Index}
						}
					}
					t.Fatal("chosen occurrence's mana color was not offered")
				}
				return req.DefaultSelection
			}}
			activate := action.ActivateAbility(crystal, crystalDef.ActivatedAbilityIndex(0), nil, 0)
			if !actionsContain(e.legalCyclingActions(g, game.Player1), activate) {
				t.Fatal("cycling lacked its initial two-activation witness")
			}
			log := &TurnLog{}
			if !e.applyActionWithChoices(g, game.Player1, activate, reboundKickerAgents(agent), log) {
				t.Fatal("cycling could not complete with two Altar activations")
			}
			if !chosenCreature || chosenColors != 2 || g.Players[game.Player1].ManaPool.Total() != 0 ||
				!g.Players[game.Player1].Graveyard.Contains(first.CardInstanceID) ||
				!g.Players[game.Player1].Graveyard.Contains(second.CardInstanceID) ||
				!g.Players[game.Player1].Graveyard.Contains(crystal) {
				t.Fatal("distinct activation costs and outputs were not paid exactly")
			}
			produced := []mana.Color{}
			for _, event := range g.Events {
				if event.Kind == game.EventManaProduced && event.SourceObjectID == altar.ObjectID {
					produced = append(produced, event.ProducedManaColors...)
				}
			}
			if !slices.Equal(produced, []mana.Color{mana.G, mana.U}) {
				t.Fatalf("Altar outputs = %v, want independent G then U", produced)
			}
			logged := 0
			for _, entry := range log.Entries {
				if activation, ok := entry.Action.Action.ActivateAbilityPayload(); entry.Kind == TurnLogEntryAction &&
					ok && activation.SourceID == altar.ObjectID && entry.Action.ManaAbility {
					logged++
				}
			}
			if logged != 2 {
				t.Fatal("both occurrences were not logged")
			}
			requireActivatedFrom(t, g, crystal)
			e.resolveTopOfStack(g, log)
			if !g.Players[game.Player1].Hand.Contains(draw) || g.Stack.Size() != 0 {
				t.Fatal("Raugrin Crystal did not resolve its cycling draw")
			}
		})
	}
}

func TestCostedActionRepeatedManaCostsSeePreviousOccurrenceResources(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	setSorceryTiming(g, game.Player1)
	crystalDef := registryCardDef(t, "Raugrin Crystal")
	crystal := addCardToHand(g, game.Player1, crystalDef)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Counter mana", Types: []types.Card{types.Artifact},
		ManaAbilities: []game.ManaAbility{{
			AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalRemoveCounter, AnyCounterKind: true, Amount: 1}},
			Content: game.Mode{Sequence: []game.Instruction{
				{Primitive: game.AddMana{ManaColor: mana.G, Amount: game.Fixed(1)}},
				{Primitive: game.AddCounter{Object: game.SourcePermanentReference(), CounterKind: counter.Charge, Amount: game.Fixed(1)}},
			}}.Ability(),
		}},
	}})
	source.Counters.Add(counter.PlusOnePlusOne, 1)
	if !e.applyAction(g, game.Player1, action.ActivateAbility(crystal, crystalDef.ActivatedAbilityIndex(0), nil, 0)) {
		t.Fatal("the second activation could not use the counter produced by the first")
	}
	var removed []counter.Kind
	for _, event := range g.Events {
		if event.Kind == game.EventCountersRemoved && event.PermanentID == source.ObjectID {
			removed = append(removed, event.CounterKind)
		}
	}
	if !slices.Equal(removed, []counter.Kind{counter.PlusOnePlusOne, counter.Charge}) ||
		source.Counters.Get(counter.Charge) != 1 || g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("occurrence-specific counter costs were not planned and paid in order")
	}
	requireActivatedFrom(t, g, crystal)
}

func TestCostedActionWithoutManaDoesNotOpenManaAbilities(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	setSorceryTiming(g, game.Player1)
	addCombatPermanent(g, game.Player1, registryCardDef(t, "The Great Henge"))
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Life-only cost", Types: []types.Card{types.Artifact},
		ActivatedAbilities: []game.ActivatedAbility{{
			AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalPayLife, Amount: 42}},
			Content:         game.Mode{Sequence: []game.Instruction{{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)}}}}.Ability(),
		}},
	}})
	ids := g.IDGen.Current()
	if actionsContain(e.legalActivateAbilityActions(g, game.Player1), action.ActivateAbility(source.ObjectID, 0, nil, 0)) ||
		e.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, nil, 0)) ||
		g.Players[game.Player1].Life != 40 || g.IDGen.Current() != ids {
		t.Fatal("a nonmana cost illegally activated Henge to make its life payment possible")
	}
}

func TestCostedActionChannelLocksCostBeforeSacrificingReducer(t *testing.T) {
	for _, floatColorless := range []bool{false, true} {
		t.Run(map[bool]string{false: "no extra mana", true: "retain floated C"}[floatColorless], func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			e := NewEngine(nil)
			setSorceryTiming(g, game.Player1)
			boseiju := addCardToHand(g, game.Player1, registryCardDef(t, "Boseiju, Who Endures"))
			thalia := addCombatPermanent(g, game.Player1, registryCardDef(t, "Thalia, Guardian of Thraben"))
			addCombatPermanent(g, game.Player1, registryCardDef(t, "Phyrexian Altar"))
			greaves := addCombatPermanent(g, game.Player2, registryCardDef(t, "Lightning Greaves"))
			if floatColorless {
				g.Players[game.Player1].ManaPool.Add(mana.C, 1)
			}
			var chosen action.Action
			for _, act := range e.legalHandActivateAbilityActions(g, game.Player1) {
				if activation, ok := act.ActivateAbilityPayload(); ok && activation.SourceID == boseiju &&
					len(activation.Targets) == 1 && activation.Targets[0].PermanentID == greaves.ObjectID {
					chosen = act
				}
			}
			if chosen.Kind != action.ActionActivateAbility || !e.applyAction(g, game.Player1, chosen) {
				t.Fatal("channel was not offered and paid")
			}

			if _, ok := permanentByObjectID(g, thalia.ObjectID); ok {
				t.Fatal("Altar did not sacrifice the legendary cost reducer")
			}
			want := 0
			if floatColorless {
				want = 1
			}
			if g.Players[game.Player1].ManaPool.Total() != want {
				t.Fatal("channel recomputed its locked G cost after sacrificing Thalia")
			}
			requireActivatedFrom(t, g, boseiju)
		})
	}
}

func TestCostedActionFaceUpDisclosesBeforePayment(t *testing.T) {
	for _, tc := range []struct {
		name     string
		card     string
		kind     game.FaceDownKind
		showCard bool
	}{
		{"manifest", "Grizzly Bears", game.FaceDownManifest, true},
		{"cloak", "Grizzly Bears", game.FaceDownCloak, true},
		{"morph", "Willbender", game.FaceDownMorph, false},
		{"disguise", "Museum Nightwatch", game.FaceDownDisguise, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			e := NewEngine(nil)
			setSorceryTiming(g, game.Player1)
			permanent := addFaceDownPermanent(g, game.Player1, registryCardDef(t, tc.card), tc.kind)
			addCombatPermanent(g, game.Player1, registryCardDef(t, "Mana Confluence"))
			for _, color := range []mana.Color{mana.C, mana.C, mana.C, mana.W} {
				g.Players[game.Player1].ManaPool.Add(color, 1)
			}
			a := &costedObservationAgent{t: t, choose: func(obs PlayerObservation, req game.ChoiceRequest) []int {
				if req.Kind == game.ChoicePayment && req.Prompt != "Choose turn face-up cost" {
					for _, player := range []game.PlayerID{game.Player1, game.Player2} {
						view := NewObservation(obs.g, player)
						disclosures := view.CostDisclosures()
						if len(disclosures) == 0 || disclosures[len(disclosures)-1].Card.Exists != tc.showCard {
							t.Fatal("face-up cost disclosure absent or leaked the card")
						}
						for _, p := range view.Battlefield() {
							if p.ObjectID == permanent.ObjectID && (!p.FaceDown || p.Power != 2 || p.Toughness != 2 || p.Name != "") {
								t.Fatal("cost disclosure changed face-down characteristics")
							}
						}
					}
				}
				return req.DefaultSelection
			}}
			if !e.applyActionWithChoices(g, game.Player1, action.TurnFaceUp(permanent.ObjectID), reboundKickerAgents(a), &TurnLog{}) || a.seen == 0 || permanent.FaceDown {
				t.Fatal("face-up special action did not complete through disclosed payment")
			}
		})
	}
}

// This is deliberately not a PCG and exposes no serializable source state.
func TestCostedActionContextualOpaqueAutomaticSource(t *testing.T) {
	g, e, act, _ := costedOpaqueFixture(t, rand.New(&costedOpaqueSource{word: 37}), "battlefield")
	var source *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanentChoiceLabel(g, permanent) == "Opaque source" {
			source = permanent
		}
	}
	card, _ := g.GetCardInstance(source.CardInstanceID)
	card.Def.ManaAbilities = []game.ManaAbility{game.TapManaAbility(mana.C)}
	card.Def.TriggeredAbilities = []game.TriggeredAbility{{
		Optional: true,
		Trigger:  game.TriggerCondition{Type: game.TriggerWhen, Pattern: game.TriggerPattern{Event: game.EventAbilityActivated, Source: game.TriggerSourceSelf}},
		Content: game.Mode{Sequence: []game.Instruction{
			{Primitive: game.Tap{Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Land}, Controller: game.ControllerYou})}},
			{Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.C}},
		}}.Ability(),
	}}
	index := card.Def.ManaAbilityIndex(0)
	if _, class := payment.ClassifyManaAbility(&card.Def.ManaAbilities[0]); class != payment.PlannableManaAbility {
		t.Fatal("fixture source is not structurally plannable")
	}
	candidates := (&rulesPaymentState{g: g, engine: e}).ManaActivationCandidates(game.Player1)
	if !e.opaqueManaActivation(g, game.Player1, action.ActivateAbilityAction{SourceID: source.ObjectID, AbilityIndex: index}, candidates) {
		t.Fatal("opaque optional-trigger context was ignored")
	}
	before := g.Clone()
	opaqueChoices, optional := 0, 0
	agent := authoredChoicePolicy{choose: func(_ PlayerObservation, req game.ChoiceRequest) []int {
		if req.Kind == game.ChoiceMay {
			optional++
			return []int{1}
		}
		for _, option := range req.Options {
			if strings.Contains(option.Label, "(opaque)") {
				opaqueChoices++
				return []int{option.Index}
			}
		}
		return req.DefaultSelection
	}}
	if outcome, _ := e.attemptActionWithChoices(g, game.Player1, act, reboundKickerAgents(agent), &TurnLog{}, nil); outcome != actionReversed ||
		opaqueChoices != 1 || optional != 1 || !reflect.DeepEqual(g.Battlefield, before.Battlefield) || g.IDGen.Current() != before.IDGen.Current() {
		t.Fatal("contextual opaque mana was not offered, resolved immediately and reversed")
	}
}

func TestCostedActionNestedOpaquePaymentSharesCheckpoint(t *testing.T) {
	g, e, act, _ := costedOpaqueFixture(t, rand.New(&costedOpaqueSource{word: 37}), "battlefield")
	var outer *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanentChoiceLabel(g, permanent) == "Opaque source" {
			outer = permanent
		}
	}
	card, _ := g.GetCardInstance(outer.CardInstanceID)
	card.Def.ManaAbilities[0].ManaCost = opt.Val(cost.Mana{cost.C})
	helperBody := game.TapManaAbility(mana.C)
	helperBody.Content.Modes[0].Sequence = append(helperBody.Content.Modes[0].Sequence,
		game.Instruction{Primitive: game.Choose{Choice: game.ResolutionChoice{Kind: game.ResolutionChoiceNumber, MinNumber: 1, MaxNumber: 3, AtRandom: true}, PublishChoice: "nested-random"}})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Opaque helper", Types: []types.Card{types.Artifact}, ManaAbilities: []game.ManaAbility{helperBody}}})
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	before := g.Clone()
	selected := 0
	agent := authoredChoicePolicy{choose: func(_ PlayerObservation, req game.ChoiceRequest) []int {
		want := "Opaque source"
		if selected == 1 {
			want = "Opaque helper"
		}
		if selected < 2 {
			for _, option := range req.Options {
				if strings.Contains(option.Label, want) && strings.Contains(option.Label, "(opaque)") {
					selected++
					return []int{option.Index}
				}
			}
		}
		return req.DefaultSelection
	}}
	if outcome, _ := e.attemptActionWithChoices(g, game.Player1, act, reboundKickerAgents(agent), &TurnLog{}, nil); outcome != actionReversed ||
		selected != 2 || g.IDGen.Current() != before.IDGen.Current() || !reflect.DeepEqual(g.Players, before.Players) ||
		!reflect.DeepEqual(g.Battlefield, before.Battlefield) || g.Stack.Size() != 0 {
		t.Fatal("nested opaque mana payment did not restore through its caller's single checkpoint")
	}
}

func TestCostedActionOpaquePreservesIrreversibleLibraryBoundary(t *testing.T) {
	g, e, act, source := costedOpaqueFixture(t, rand.New(&costedOpaqueSource{word: 37}), "battlefield")
	var opaque *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanentChoiceLabel(g, permanent) == "Opaque source" {
			opaque = permanent
		}
	}
	card, _ := g.GetCardInstance(opaque.CardInstanceID)
	card.Def.ManaAbilities[0].AdditionalCosts = []cost.Additional{{Kind: cost.AdditionalSacrificeSource, Text: "Sacrifice this artifact", Amount: 1}}
	card.Def.ReplacementAbilities = homingGolemDef().ReplacementAbilities
	ids := g.IDGen.Current()
	agent := authoredChoicePolicy{choose: func(_ PlayerObservation, req game.ChoiceRequest) []int {
		for _, option := range req.Options {
			if strings.Contains(option.Label, "(opaque)") {
				return []int{option.Index}
			}
		}
		return req.DefaultSelection
	}}
	if outcome, _ := e.attemptActionWithChoices(g, game.Player1, act, reboundKickerAgents(agent), &TurnLog{}, nil); outcome != actionReversed {
		t.Fatalf("outcome = %v, want the unpaid ability reversed", outcome)
	}
	if !g.Players[game.Player1].Library.Contains(card.ID) || g.Stack.Size() != 0 || g.IDGen.Current() <= ids {
		t.Fatal("an irreversible library move or its surviving IDs were undone")
	}
	if _, ok := permanentByObjectID(g, source); !ok || !slices.ContainsFunc(g.Events, func(event game.Event) bool {
		return event.Kind == game.EventManaProduced && event.SourceObjectID == opaque.ObjectID
	}) || slices.ContainsFunc(g.Events, func(event game.Event) bool { return event.Kind == game.EventObjectBecameTarget }) {
		t.Fatal("the illegal announcement survived or legal irreversible mana was erased")
	}
}

func TestCostedActionIrreversibleFallbackDropsUnpaidChoices(t *testing.T) {
	g, e, act, source := costedOpaqueFixture(t, rand.New(&costedOpaqueSource{word: 37}), "graveyard")
	card, _ := g.GetCardInstance(source)
	card.Def.ActivatedAbilities[0].AdditionalCosts = []cost.Additional{{Kind: cost.AdditionalDiscard, Amount: 1, Text: "Discard a card"}}
	addCardToHand(g, game.Player1, registryCardDef(t, "Grizzly Bears"))
	addCardToHand(g, game.Player1, registryCardDef(t, "Grizzly Bears"))
	for _, permanent := range g.Battlefield {
		if permanentChoiceLabel(g, permanent) == "Opaque source" {
			opaque, _ := g.GetCardInstance(permanent.CardInstanceID)
			opaque.Def.ManaAbilities[0].AdditionalCosts = []cost.Additional{{Kind: cost.AdditionalSacrificeSource, Amount: 1}}
			opaque.Def.ReplacementAbilities = homingGolemDef().ReplacementAbilities
		}
	}
	agent := authoredChoicePolicy{choose: func(_ PlayerObservation, req game.ChoiceRequest) []int {
		for _, option := range req.Options {
			if strings.Contains(option.Label, "(opaque)") {
				return []int{option.Index}
			}
		}
		return req.DefaultSelection
	}}
	log := &TurnLog{}
	if outcome, _ := e.attemptActionWithChoices(g, game.Player1, act, reboundKickerAgents(agent), log, nil); outcome != actionReversed {
		t.Fatal("irreversible fixture was not reversed")
	}
	if g.Players[game.Player1].Hand.Size() != 2 || g.Stack.Size() != 0 || slices.ContainsFunc(log.Choices, func(choice game.ChoiceDecision) bool {
		return choice.Request.Prompt == "Discard a card"
	}) {
		t.Fatal("an unpaid hidden-card choice survived the canceled announcement")
	}
	if !slices.ContainsFunc(log.Entries, func(entry TurnLogEntry) bool { return entry.Kind == TurnLogEntryAction }) {
		t.Fatal("the legally retained opaque mana activation lost its log")
	}
}

type costedRetryAgent struct {
	source   id.ID
	attempts int
	endless  bool
}

func (a *costedRetryAgent) ChooseAction(_ PlayerObservation, legal []action.Action) action.Action {
	for _, act := range legal {
		if activate, ok := act.ActivateAbilityPayload(); ok && activate.SourceID == a.source && len(activate.Targets) == 1 {
			a.attempts++
			return act
		}
	}
	return action.Pass()
}

func (a *costedRetryAgent) ChooseChoice(_ PlayerObservation, req game.ChoiceRequest) []int {
	if a.endless || a.attempts <= 2 {
		for _, option := range req.Options {
			if strings.Contains(option.Label, "(opaque)") {
				return []int{option.Index}
			}
		}
	}
	return req.DefaultSelection
}

func TestCostedActionOpaquePriorityRetryAfterTwoReversals(t *testing.T) {
	g, e, _, source := costedOpaqueFixture(t, rand.New(&costedOpaqueSource{word: 37}), "battlefield")
	agent := &costedRetryAgent{source: source}
	e.runPriorityLoop(g, reboundKickerAgents(agent), &TurnLog{})
	if agent.attempts != 3 || g.Players[game.Player2].Hand.Size() != 1 || g.Players[game.Player1].Life != 39 {
		t.Fatalf("attempts=%d opponent hand=%d life=%d, want two reversals then a legally distinct success",
			agent.attempts, g.Players[game.Player2].Hand.Size(), g.Players[game.Player1].Life)
	}
}

func TestCostedActionOpaqueNonprogressUsesExecutionBudget(t *testing.T) {
	g, e, _, source := costedOpaqueFixture(t, rand.New(&costedOpaqueSource{word: 37}), "battlefield")
	agent := &costedRetryAgent{source: source, endless: true}
	e.DecisionLimit = 24
	defer func() {
		failure, ok := recover().(WorkBudgetError)
		if !ok || failure.Execution.Decisions != e.DecisionLimit || agent.attempts < 3 {
			t.Fatalf("nonprogress termination = %v, attempts=%d; want work-budget termination, never withholding", failure, agent.attempts)
		}
	}()
	e.runPriorityLoop(g, reboundKickerAgents(agent), &TurnLog{})
	t.Fatal("nonprogressing opaque retries were silently suppressed")
}

type costedOpaqueSource struct{ word uint64 }

func (s *costedOpaqueSource) Uint64() uint64 {
	s.word = s.word*6364136223846793005 + 1442695040888963407
	return s.word
}

func costedOpaqueFixture(t *testing.T, rng *rand.Rand, route string) (*game.Game, *Engine, action.Action, id.ID) {
	t.Helper()
	g := game.NewGameWithRand([game.NumPlayers]game.PlayerConfig{}, rng)
	e := NewEngine(nil)
	setSorceryTiming(g, game.Player1)
	addCombatPermanent(g, game.Player1, registryCardDef(t, "Waterlogged Grove"))
	target := addCombatPermanent(g, game.Player2, registryCardDef(t, "Grizzly Bears"))
	body := game.ActivatedAbility{ManaCost: opt.Val(cost.Mana{cost.U}), Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)}}}}.Ability()}
	var source id.ID
	var chosen action.Action
	if route == "battlefield" {
		bomb := addCombatPermanent(g, game.Player1, registryCardDef(t, "Aether Spellbomb"))
		source = bomb.ObjectID
		for _, act := range e.legalActivateAbilityActions(g, game.Player1) {
			if activation, ok := act.ActivateAbilityPayload(); ok && activation.SourceID == source && len(activation.Targets) == 1 && activation.Targets[0].PermanentID == target.ObjectID {
				chosen = act
			}
		}
	} else {
		def := &game.CardDef{CardFace: game.CardFace{Name: "Opaque rollback route", Types: []types.Card{types.Creature}}}
		if route == "hand" {
			body.ZoneOfFunction = zone.Hand
			body.AdditionalCosts = []cost.Additional{{Kind: cost.AdditionalDiscard, Text: "Discard this card", Amount: 1}}
			def.ActivatedAbilities = []game.ActivatedAbility{body}
			source = addCardToHand(g, game.Player1, def)
		} else {
			body.ZoneOfFunction = zone.Graveyard
			def.ActivatedAbilities = []game.ActivatedAbility{body}
			source = addCardToGraveyard(g, game.Player1, def)
		}
		chosen = action.ActivateAbility(source, 0, nil, 0)
	}
	opaque := game.ManaAbility{AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalTap}}, Content: game.Mode{Sequence: []game.Instruction{
		{Primitive: game.Choose{Choice: game.ResolutionChoice{Kind: game.ResolutionChoiceNumber, MinNumber: 1, MaxNumber: 3, AtRandom: true}, PublishChoice: "random-tokens"}},
		{Primitive: game.CreateToken{Amount: game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountChosenNumber, ResultKey: "random-tokens"}), Source: game.TokenDef(vanillaCreature("Opaque token", 1, 1))}},
		{Primitive: game.Tap{Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Land}, Controller: game.ControllerYou})}},
		{Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.C}},
	}}.Ability()}
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Opaque source", Types: []types.Card{types.Artifact}, ManaAbilities: []game.ManaAbility{opaque}}})
	if _, class := payment.ClassifyManaAbility(&opaque); class != payment.OpaqueManaAbility || !e.announcementLegal(g, game.Player1, chosen) {
		t.Fatalf("opaque rollback fixture: class=%v legal=%v chosen=%v", class, e.announcementLegal(g, game.Player1, chosen), chosen.Kind)
	}
	return g, e, chosen, source
}

func TestCostedActionOpaqueRestoresWorldIDsLogsAndRealRandomStream(t *testing.T) {
	for _, route := range []string{"battlefield", "hand", "graveyard"} {
		for _, custom := range []bool{false, true} {
			t.Run(route+map[bool]string{false: "/PCG", true: "/opaque source"}[custom], func(t *testing.T) {
				rng := rand.New(rand.NewPCG(37, 91))
				reference := rand.New(rand.NewPCG(37, 91))
				if custom {
					rng = rand.New(&costedOpaqueSource{word: 19})
					reference = rand.New(&costedOpaqueSource{word: 19})
				}
				g, e, act, source := costedOpaqueFixture(t, rng, route)
				for range 7 {
					if g.RNG.Uint64() != reference.Uint64() {
						t.Fatal("random warmup differs")
					}
				}
				before := g.Clone()
				log := &TurnLog{Entries: []TurnLogEntry{{Kind: TurnLogEntryStep}}}
				beforeLog := *log
				a := &costedObservationAgent{t: t, source: source, choose: func(_ PlayerObservation, req game.ChoiceRequest) []int {
					for _, option := range req.Options {
						if strings.Contains(option.Label, "(opaque)") {
							return []int{option.Index}
						}
					}
					return req.DefaultSelection
				}}
				for attempt := range 2 {
					outcome, _ := e.attemptActionWithChoices(g, game.Player1, act, reboundKickerAgents(a), log, &castAttempts{})
					if outcome != actionReversed || g.IDGen.Current() != before.IDGen.Current() ||
						!reflect.DeepEqual(g.Events, before.Events) || !reflect.DeepEqual(*log, beforeLog) ||
						!reflect.DeepEqual(g.Battlefield, before.Battlefield) || !reflect.DeepEqual(g.Players, before.Players) || g.Stack.Size() != 0 {
						t.Fatalf("attempt %d restore: outcome=%v ids=%v events=%v log=%v battlefield=%v players=%v stack=%d",
							attempt, outcome, g.IDGen.Current() == before.IDGen.Current(), reflect.DeepEqual(g.Events, before.Events),
							reflect.DeepEqual(*log, beforeLog), reflect.DeepEqual(g.Battlefield, before.Battlefield), reflect.DeepEqual(g.Players, before.Players), g.Stack.Size())
					}
					for range 3 {
						if g.RNG.Uint64() != reference.Uint64() {
							t.Fatal("reversal did not resume the real opaque random stream")
						}
					}
				}
				if outcome, _ := e.attemptActionWithChoices(g, game.Player1, act, reboundKickerAgents(&costedObservationAgent{t: t, source: source}), log, &castAttempts{}); outcome != actionApplied {
					t.Fatal("legal third attempt was withheld")
				}
				pending, _ := g.Stack.Peek()
				if pending == nil || pending.SourceID != source {
					t.Fatal("successful third attempt did not retain its ability")
				}
			})
		}
	}
}
