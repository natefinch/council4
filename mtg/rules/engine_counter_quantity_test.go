package rules

import (
	"reflect"
	"slices"
	"strconv"
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

// Source review: Scryfall printing cd2cb84e-c079-486e-87ad-d188fe38bc76,
// Oracle ID 79e91e8a-cee7-411b-b0a3-d3383c0d4a43, current text and revision
// verified from the pinned corpus and live Scryfall card response. Rulings
// retrieved from https://api.scryfall.com/cards/cd2cb84e-c079-486e-87ad-d188fe38bc76/rulings:
// one ruling (2004-10-04), "Can be tapped even if it has no counters."
// The complete response is retained at
// .cardwork/blue-mana-battery/rulings.json.
// Oracle: "{2}, {T}: Put a charge counter on this artifact. {T}, Remove any
// number of charge counters from this artifact: Add {U}, then add an additional
// {U} for each charge counter removed this way."
//
// Read the official Comprehensive Rules effective 2026-09-25, pinned by
// testdata/card-preservation/corpus.json, at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.docx
// (retrieved source retained at
// .cardwork/blue-mana-battery/comprehensive-rules-2026-09-25.docx):
// CR 106.3-4 requires mana output to enter the player's mana pool;
// CR 107.1c says "any number" includes zero; CR 118.3 and 602.2b govern full
// activation-cost payment; CR 605.1a and 605.3a-b make the second ability a
// no-target mana ability that resolves immediately, including in the mana
// window before costs are paid (CR 601.2g-h). Thus the controller must be able
// to choose 0..N charge counters, tap the artifact, and receive one blue plus
// one more for each counter actually removed. The ruling confirms that zero
// counters still permits tapping for the first blue. CR 707.10 distinguishes
// stack copies, which are not cast and do not activate mana abilities again.
// Cross-checked these payment rules and CR 608.2g (effect asks for payment) in
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt.
//
// This unregistered Engine fixture plays only the second ability.
const blueManaBatteryRemovedCountersKey = "removed-charge-counters"

func chooseChargeCounterCost() cost.Additional {
	return cost.Additional{
		Kind: cost.AdditionalRemoveCounter, CounterKind: counter.Charge,
		ChooseCounterCount: true, SubjectKey: blueManaBatteryRemovedCountersKey,
	}
}

func blueManaBatteryCounterRemovalProbe() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Blue Mana Battery counter-removal probe",
		Types: []types.Card{types.Artifact},
		ManaAbilities: []game.ManaAbility{{
			Text: "{T}, Remove any number of charge counters from this artifact: Add {U}, then add an additional {U} for each charge counter removed this way.",
			AdditionalCosts: []cost.Additional{
				{Kind: cost.AdditionalTap},
				chooseChargeCounterCost(),
			},
			Content: game.AbilityContent{
				Modes: []game.Mode{{
					Sequence: []game.Instruction{
						{Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.U}},
						{Primitive: game.AddMana{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:    game.DynamicAmountPaidCostCounters,
								CostKey: blueManaBatteryRemovedCountersKey,
							}),
							ManaColor: mana.U,
						}},
					},
				}},
				MinModes: 1,
				MaxModes: 1,
			},
		}},
	}}
}

func chooseCounterCount(t *testing.T, req game.ChoiceRequest, count int, wantLabels ...string) []int {
	t.Helper()
	if req.Kind != game.ChoicePayment || req.Private {
		t.Fatalf("counter payment choice is not public: %+v", req)
	}
	labels := make([]string, 0, len(req.Options))
	for _, opt := range req.Options {
		labels = append(labels, opt.Label)
	}
	if !slices.Equal(labels, wantLabels) {
		t.Fatalf("offered counts %v, want %v", labels, wantLabels)
	}
	index := slices.Index(labels, strconv.Itoa(count))
	if index < 0 {
		t.Fatalf("count %d not offered", count)
	}
	return []int{index}
}

func TestBlueManaBatteryCanChooseAnyNumberOfChargeCounters(t *testing.T) {
	for _, tc := range []struct {
		name             string
		available, count int
	}{
		{"none available", 0, 0},
		{"choose zero", 2, 0},
		{"choose one of two", 2, 1},
		{"choose all", 2, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := mainPhaseGame(game.Player1)
			setMainPhasePriority(g, game.Player1)
			e := NewEngine(nil)
			source := addCombatPermanent(g, game.Player1, blueManaBatteryCounterRemovalProbe())
			source.Counters.Add(counter.Charge, tc.available)
			source.Counters.Add(counter.Stun, 1)
			act := action.ActivateAbility(source.ObjectID, 0, nil, 0)
			if !actionsContain(e.legalActions(g, game.Player1), act) {
				t.Fatal("counter-removal mana ability not offered")
			}
			prompted := false
			agent := &authoredCopyChoiceAgent{choose: func(req game.ChoiceRequest) []int {
				if req.Prompt != payment.AdditionalCostText(chooseChargeCounterCost()) {
					return req.DefaultSelection
				}
				prompted = true
				if source.Tapped || source.Counters.Get(counter.Charge) != tc.available {
					t.Fatal("resources changed before the payment choice")
				}
				return chooseCounterCount(t, req, tc.count, "0", "1", "2")
			}}
			if !e.applyActionWithChoices(g, game.Player1, act, reboundKickerAgents(agent), &TurnLog{}) {
				t.Fatal("counter-removal mana ability activation rejected")
			}
			if prompted != (tc.available > 0) {
				t.Fatalf("count prompted = %v for %d available", prompted, tc.available)
			}
			if !source.Tapped || g.Stack.Size() != 0 || source.Counters.Get(counter.Stun) != 1 ||
				source.Counters.Get(counter.Charge) != tc.available-tc.count ||
				g.Players[game.Player1].ManaPool.Amount(mana.U) != 1+tc.count {
				t.Fatalf("activation left tapped=%v, stack=%d, counters=%v, blue=%d",
					source.Tapped, g.Stack.Size(), source.Counters.All(), g.Players[game.Player1].ManaPool.Amount(mana.U))
			}
			removals := countEventsOfKind(g.Events, game.EventCountersRemoved)
			if tc.count == 0 {
				if removals != 0 {
					t.Fatal("paying zero emitted a counter removal event")
				}
			} else if event := lastCountersRemovedEvent(t, g); removals != 1 || event.PermanentID != source.ObjectID ||
				event.CounterKind != counter.Charge || event.Amount != tc.count || event.PreviousCounterAmount != tc.available {
				t.Fatalf("removal event = %+v, want %d charge counters removed", event, tc.count)
			}
		})
	}
}

func TestChosenCounterManaFundsCastAndReplays(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	setMainPhasePriority(g, game.Player1)
	e := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, blueManaBatteryCounterRemovalProbe())
	source.Counters.Add(counter.Charge, 2)
	spell := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Two-Mana Probe", Types: []types.Card{types.Artifact}, ManaCost: opt.Val(cost.Mana{cost.U, cost.U}),
	}})
	act := action.CastSpell(spell, nil, 0, nil)
	replay := g.Clone()
	run := func(g *game.Game, agent PlayerAgent) *TurnLog {
		t.Helper()
		log := &TurnLog{}
		if !actionsContain(e.legalActions(g, game.Player1), act) ||
			!e.applyActionWithChoices(g, game.Player1, act, reboundKickerAgents(agent), log) {
			t.Fatal("spell requiring two blue mana could not be cast")
		}
		battery, _ := permanentByObjectID(g, source.ObjectID)
		top, _ := g.Stack.Peek()
		if !battery.Tapped || battery.Counters.Get(counter.Charge) != 1 ||
			g.Players[game.Player1].ManaPool.Total() != 0 || g.Stack.Size() != 1 ||
			top.SourceID != spell || top.ManaSpentToCast != 2 || len(top.PaidCostSubjects) != 0 ||
			spellCastEvents(g, top.ID) != 1 {
			t.Fatalf("cast: source=%v tapped=%v counters=%d pool=%d stack=%d spent=%d paid subjects=%v",
				top.SourceID, battery.Tapped, battery.Counters.Get(counter.Charge), g.Players[game.Player1].ManaPool.Total(),
				g.Stack.Size(), top.ManaSpentToCast, top.PaidCostSubjects)
		}
		return log
	}
	asked := false
	log := run(g, &authoredCopyChoiceAgent{choose: func(req game.ChoiceRequest) []int {
		if req.Prompt == payment.AdditionalCostText(chooseChargeCounterCost()) {
			asked = true
			return chooseCounterCount(t, req, 1, "1", "2")
		}
		return req.DefaultSelection
	}})
	cursor := 0
	replayed := run(replay, &authoredCopyChoiceAgent{choose: func(req game.ChoiceRequest) []int {
		// The payment-window source menu is unlogged; here there is only one
		// possible source. Replay the recorded counter amount itself.
		if req.Prompt != payment.AdditionalCostText(chooseChargeCounterCost()) {
			return req.DefaultSelection
		}
		if cursor >= len(log.Choices) {
			t.Fatal("counter count was not recorded for replay")
		}
		cursor++
		return log.Choices[cursor-1].Selected
	}})
	if !asked || cursor != 1 || len(log.Choices) != cursor || !reflect.DeepEqual(log.Choices, replayed.Choices) {
		t.Fatal("counter count choice did not replay identically on the clone")
	}
}

func TestChosenCounterCostReservesOtherCostsAndRejectsStaleCount(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	e := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, blueManaBatteryCounterRemovalProbe())
	source.Counters.Add(counter.Charge, 2)
	player := g.Players[game.Player1]
	life := player.Life
	player.ManaPool.Add(mana.R, 1)
	req := payment.ForAbility(payment.AbilityRequest{
		PlayerID: game.Player1, Source: source, SourceZone: zone.Battlefield,
		ManaCost: opt.Val(cost.Mana{cost.R}),
		AdditionalCosts: []cost.Additional{
			{Kind: cost.AdditionalTap},
			chooseChargeCounterCost(),
			{Kind: cost.AdditionalRemoveCounter, CounterKind: counter.Charge, Amount: 1},
			{Kind: cost.AdditionalPayLife, Amount: 2, ChoiceGroup: 1},
			{Kind: cost.AdditionalPayLife, Amount: life + 1, ChoiceGroup: 1},
		},
	})
	step := payment.Step{Kind: payment.StepComponent, Index: 1}
	planner := e.PaymentPlanner(g)
	out := planner.Options(req, payment.Partial{}, step)
	if len(out) != 2 || out[0].Value != 0 || out[1].Value != 1 {
		t.Fatalf("count choices %+v, want 0 and 1 with one counter reserved", out)
	}
	for _, bad := range []int{-1, 2} {
		partial := payment.Partial{}.With(step, payment.Choice{Component: payment.ComponentChoice{CounterCount: opt.Val(bad), Finished: true}})
		if _, ok, _ := planner.Complete(req, partial); ok {
			t.Fatalf("accepted an illegal count %d", bad)
		}
	}
	source.Counters.Remove(counter.Charge, 1)
	if _, ok, _ := planner.Apply(out[1].Plan); ok || source.Tapped || player.Life != life ||
		player.ManaPool.Amount(mana.R) != 1 || source.Counters.Get(counter.Charge) != 1 {
		t.Fatal("stale count was changed or partially paid")
	}
	paid, ok, _ := planner.Apply(out[0].Plan)
	subjects := paid.Ability.PaidCostSubjects
	if !ok || !source.Tapped || player.ManaPool.Total() != 0 || player.Life != life-2 ||
		source.Counters.Get(counter.Charge) != 0 || len(subjects) != 1 ||
		subjects[0].Key != blueManaBatteryRemovedCountersKey || subjects[0].Kind != game.PaidCostCounterRemoval || subjects[0].Amount != 0 {
		t.Fatal("zero payment did not preserve the fixed counter, mana, and choice-group costs and its receipt")
	}
}

func TestChosenCounterManaFundsAlternateCastRoutes(t *testing.T) {
	for _, route := range []struct {
		name string
		cast func(*Engine, *game.Game, id.ID) bool
		copy bool
	}{
		{"paid from exile during resolution", func(e *Engine, g *game.Game, card id.ID) bool {
			return e.castPaidSpellFromSource(g, g.Players[game.Player1], game.Player1, card, zone.Exile, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		}, false},
		{"free cast with tax", func(e *Engine, g *game.Game, card id.ID) bool {
			g.CostModifiers = []game.CostModifier{{Kind: game.CostModifierSpell, GenericIncrease: 2}}
			return e.castFreeOption(g, g.Players[game.Player1], game.Player1, g.CardInstances[card],
				action.CastSpellAction{CardID: card, SourceZone: zone.Exile}, false, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		}, false},
		{"free cast of copied card with tax", func(e *Engine, g *game.Game, card id.ID) bool {
			g.CostModifiers = []game.CostModifier{{Kind: game.CostModifierSpell, GenericIncrease: 2}}
			return e.castFreeCopyOfCard(g, game.Player1, card, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		}, true},
	} {
		t.Run(route.name, func(t *testing.T) {
			g := mainPhaseGame(game.Player1)
			setMainPhasePriority(g, game.Player2)
			e := NewEngine(nil)
			source := addCombatPermanent(g, game.Player1, blueManaBatteryCounterRemovalProbe())
			source.Counters.Add(counter.Charge, 1)
			def := simpleGainLifeInstant("Resolution cast payment probe")
			def.ManaCost = opt.Val(cost.Mana{cost.U, cost.U})
			card := addExiledInstant(g, game.Player1, def)
			if !route.cast(e, g, card) {
				t.Fatal("Battery was not available to pay during this cast")
			}
			top, _ := g.Stack.Peek()
			if g.Stack.Size() != 1 || top.Copy != route.copy || top.CastCopy != route.copy ||
				top.ManaSpentToCast != 2 || !source.Tapped || source.Counters.Get(counter.Charge) != 0 ||
				spellCastEvents(g, top.ID) != 1 || g.Players[game.Player1].Exile.Contains(card) != route.copy ||
				len(top.PaidCostSubjects) != 0 {
				t.Fatal("casting did not pay exactly two Battery mana or preserve spell/copy identity")
			}
		})
	}
}

func TestChosenCounterManaDoesNotRepayForSpellCopies(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	setMainPhasePriority(g, game.Player1)
	e := NewEngine(nil)
	castInstantFor(t, e, g, game.Player1, "Earlier spell")
	e.resolveTopOfStack(g, &TurnLog{})
	source := addCombatPermanent(g, game.Player1, blueManaBatteryCounterRemovalProbe())
	source.Counters.Add(counter.Charge, 2)
	spell := simpleGainLifeInstant("Storm mana payment probe")
	spell.ManaCost = opt.Val(cost.Mana{cost.U, cost.U})
	spell.StaticAbilities = []game.StaticAbility{game.StormStaticBody}
	card := addCardToHand(g, game.Player1, spell)
	if !e.applyAction(g, game.Player1, action.CastSpell(card, nil, 0, nil)) {
		t.Fatal("original spell did not cast with Battery mana")
	}
	original, _ := g.Stack.Peek()
	if !e.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("paid spell did not trigger storm")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	addEffectSpellToStack(g, game.Player1, game.CopyStackObject{Object: game.TargetStackObjectReference(0)},
		[]game.Target{game.StackObjectTarget(original.ID)})
	e.resolveTopOfStack(g, &TurnLog{})
	copies := stormCopiesOf(g, card)
	if len(copies) != 2 || !counterStackObjectWithDestination(g, original.ID, false, game.CounteredSpellGraveyard) {
		t.Fatal("expected one storm copy, one direct copy and an independently counterable original")
	}
	for _, copy := range copies {
		if copy.CastCopy || spellCastEvents(g, copy.ID) != 0 || len(copy.PaidCostSubjects) != 0 {
			t.Fatal("copy acquired cast or counter-payment state")
		}
	}
	for range copies {
		e.resolveTopOfStack(g, &TurnLog{})
	}
	if !source.Tapped || source.Counters.Get(counter.Charge) != 1 || countEventsOfKind(g.Events, game.EventCountersRemoved) != 1 ||
		!g.Players[game.Player1].Graveyard.Contains(card) || g.Stack.Size() != 0 {
		t.Fatal("countering or copying repaid the mana cost or changed the original's destination")
	}
}

func TestChosenCounterManaPaysResolutionCostOnlyWhenAcceptedAndAffordable(t *testing.T) {
	for _, tc := range []struct {
		name               string
		available, blue    int
		agree, wantConsent bool
		wantPaid           bool
	}{
		{"accept", 1, 2, true, true, true},
		{"zero counters still funds one blue", 0, 1, true, true, true},
		{"decline", 1, 2, false, true, false},
		{"cannot afford", 0, 2, true, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := mainPhaseGame(game.Player1)
			setMainPhasePriority(g, game.Player2)
			e := NewEngine(nil)
			source := addCombatPermanent(g, game.Player1, blueManaBatteryCounterRemovalProbe())
			source.Counters.Add(counter.Charge, tc.available)
			consented := false
			agent := &authoredCopyChoiceAgent{choose: func(req game.ChoiceRequest) []int {
				if req.Kind != game.ChoiceMay {
					return req.DefaultSelection
				}
				consented = true
				if tc.agree {
					return []int{1}
				}
				return []int{0}
			}}
			obj := &game.StackObject{Controller: game.Player1}
			manaCost := cost.Mana{cost.U}
			if tc.blue == 2 {
				manaCost = append(manaCost, cost.U)
			}
			accepted, paid := e.resolveResolutionPaymentValue(g, obj, &game.ResolutionPayment{
				ManaCost: opt.Val(manaCost),
			}, reboundKickerAgents(agent), &TurnLog{})
			remaining := tc.available
			if tc.wantPaid {
				remaining -= tc.blue - 1
			}
			if consented != tc.wantConsent || accepted != tc.wantPaid || paid != tc.wantPaid ||
				source.Tapped != tc.wantPaid || source.Counters.Get(counter.Charge) != remaining ||
				g.Players[game.Player1].ManaPool.Total() != 0 || g.Stack.Size() != 0 {
				t.Fatalf("payment: consent prompt=%v accepted=%v paid=%v tapped=%v counters=%d",
					consented, accepted, paid, source.Tapped, source.Counters.Get(counter.Charge))
			}
		})
	}
}

func TestChosenCounterManaPreservesCountersReservedForSpellCost(t *testing.T) {
	for _, available := range []int{1, 2} {
		t.Run(strconv.Itoa(available), func(t *testing.T) {
			g := mainPhaseGame(game.Player1)
			setMainPhasePriority(g, game.Player1)
			e := NewEngine(nil)
			source := addCombatPermanent(g, game.Player1, blueManaBatteryCounterRemovalProbe())
			source.Counters.Add(counter.Charge, available)
			spell := simpleGainLifeInstant("Combined counter payment probe")
			spell.ManaCost = opt.Val(cost.Mana{cost.U, cost.U})
			spell.AdditionalCosts = []cost.Additional{
				{Kind: cost.AdditionalRemoveCounterAmong, CounterKind: counter.Charge, Amount: 1},
				{Kind: cost.AdditionalPayLife, Amount: 2, ChoiceGroup: 1},
				{Kind: cost.AdditionalPayLife, Amount: 100, ChoiceGroup: 1},
			}
			player := g.Players[game.Player1]
			life := player.Life
			card := addCardToHand(g, game.Player1, spell)
			act := action.CastSpell(card, nil, 0, nil)
			offered := actionsContain(e.legalActions(g, game.Player1), act)
			applied := e.applyAction(g, game.Player1, act)
			if available == 1 {
				if offered || applied || source.Tapped || source.Counters.Get(counter.Charge) != 1 ||
					player.Life != life || !player.Hand.Contains(card) {
					t.Fatal("spell proposed or attempted to spend the same counter twice")
				}
			} else if !offered || !applied || !source.Tapped || source.Counters.Get(counter.Charge) != 0 ||
				player.Life != life-2 || countEventsOfKind(g.Events, game.EventCountersRemoved) != 2 {
				t.Fatal("spell did not pay separate counters for its mana and additional costs")
			}
		})
	}
}
