package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: Engine gap #3652 (choose the counter kind when paying an
// any-kind removal cost), verified 2026-09-25 against the pinned corpus,
// Scryfall rulings and the Comprehensive Rules effective 2026-09-25.
//
// CR 601.2h / 602.2b: the player pays the total cost in any order and makes
// the choices that payment requires, including which counters to remove.
// CR 118.3: a cost can be paid only with resources the player has. CR 122.1:
// counters of different kinds are distinct; an unqualified "a counter" means a
// counter of any kind. CR 310.12b / 704.5v: a Siege is defeated, and its
// ability triggers, when its last defense counter is removed.
//
// Power Conduit (oracle 2e358f7e-c282-4b92-8255-1436c99cda49), {2} Artifact:
// "{T}, Remove a counter from a permanent you control: Choose one —
// • Put a charge counter on target artifact. • Put a +1/+1 counter on target
// creature." Rulings: none.
// Moonlit Lamenter (oracle ce009bd4-5ffb-46a4-a9dd-b9c060926467), {2}{W}
// Creature — Treefolk Cleric 2/5: "This creature enters with a -1/-1 counter
// on it. {1}{W}, Remove a counter from this creature: Draw a card. Activate
// only as a sorcery." Rulings: none.
// Gnarlbark Elm (oracle a848ea46-347d-49c7-85c8-023cf836d0b3), {2}{B}
// Creature — Treefolk Warlock 3/4: "This creature enters with two -1/-1
// counters on it. {2}{B}, Remove two counters from this creature: Target
// creature gets -2/-2 until end of turn. Activate only as a sorcery."
// Rulings: none.
// Hopeful Initiate (oracle 317169b8-9014-48e6-862b-ca21b706846e), {W}
// Creature — Human Warlock 1/2: "Training ... {2}{W}, Remove two +1/+1
// counters from among creatures you control: Destroy target artifact or
// enchantment." Rulings concern training only.
// Chisei, Heart of Oceans (oracle f91407fe-3c3b-4ffc-81de-cb8c5b0fe526):
// "At the beginning of your upkeep, sacrifice Chisei unless you remove a
// counter from a permanent you control." Rulings: none.
// Invasion of Segovia (oracle 7b3c0d89-55ed-43d7-a00a-7b060857ec4e), {2}{U}
// Battle — Siege, defense 4; rulings concern Caetus's convoke only.
//
// Before #3652 the payment layer recorded only the permanent for an any-kind
// removal and always removed its lowest-sorted kind, so Power Conduit could
// never remove a Siege's defense counter while a charge counter was present,
// and "remove a counter from this creature" never asked which kind.

// counterLabelAgent answers payment prompts by choosing, in order, the first
// unused option with each wanted label; "may" prompts are accepted and other
// prompts take their default.
type counterLabelAgent struct {
	picks    [][]string
	next     int
	requests []game.ChoiceRequest
}

func (*counterLabelAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *counterLabelAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	a.requests = append(a.requests, request)
	switch {
	case request.Kind == game.ChoiceMay:
		return []int{1}
	case request.Kind != game.ChoicePayment || a.next >= len(a.picks):
		return request.DefaultSelection
	}
	want := a.picks[a.next]
	a.next++
	var selected []int
	for _, label := range want {
		for _, option := range request.Options {
			if option.Label == label && !slices.Contains(selected, option.Index) {
				selected = append(selected, option.Index)
				break
			}
		}
	}
	return selected
}

func (a *counterLabelAgent) paymentLabels() [][]string {
	var labels [][]string
	for _, request := range a.requests {
		if request.Kind != game.ChoicePayment {
			continue
		}
		var row []string
		for _, option := range request.Options {
			row = append(row, option.Label)
		}
		labels = append(labels, row)
	}
	return labels
}

func counterKindCostAgents(agent *counterLabelAgent) [game.NumPlayers]PlayerAgent {
	return [game.NumPlayers]PlayerAgent{game.Player1: agent}
}

func chargeAndDefenseSiege(t *testing.T, g *game.Game) *game.Permanent {
	t.Helper()
	siege := addCombatPermanent(g, game.Player1, registryCardDef(t, "Invasion of Segovia"))
	siege.Counters.Add(counter.Charge, 1)
	siege.Counters.Add(counter.Defense, 1)
	siege.BattleProtector = opt.Val(game.Player2)
	return siege
}

// TestPowerConduitChoosesCounterKindToRemove proves registered Power Conduit
// can remove either kind from Invasion of Segovia carrying one charge and one
// defense counter, that removing the last defense counter triggers the Siege
// in the game and in a clone taken after payment, and that removing the charge
// counter doesn't (CR 601.2h, 122.1, 310.12b).
func TestPowerConduitChoosesCounterKindToRemove(t *testing.T) {
	for _, kind := range []counter.Kind{counter.Charge, counter.Defense} {
		t.Run(kind.String(), func(t *testing.T) {
			g := mainPhaseGame(game.Player1)
			setMainPhasePriority(g, game.Player1)
			e := choicesProgramEngine(t)
			siege := chargeAndDefenseSiege(t, g)
			def := registryCardDef(t, "Power Conduit")
			conduit := addCombatPermanent(g, game.Player1, def)
			agent := &counterLabelAgent{picks: [][]string{{"Invasion of Segovia: " + kind.String() + " counter"}}}
			activate := action.ActivateAbilityWithModes(conduit.ObjectID, def.ActivatedAbilityIndex(0), []game.Target{game.PermanentTarget(conduit.ObjectID)}, 0, []int{0})
			if !e.applyActionWithChoices(g, game.Player1, activate, counterKindCostAgents(agent), &TurnLog{}) {
				t.Fatal("registered Power Conduit activation rejected")
			}
			wantLabels := [][]string{{"Invasion of Segovia: charge counter", "Invasion of Segovia: defense counter"}}
			if got := agent.paymentLabels(); !slices.EqualFunc(got, wantLabels, slices.Equal[[]string]) {
				t.Fatalf("payment prompts = %q, want %q", got, wantLabels)
			}
			if !conduit.Tapped {
				t.Fatal("Power Conduit wasn't tapped as part of the cost")
			}
			other := counter.Defense
			if kind == counter.Defense {
				other = counter.Charge
			}
			if siege.Counters.Get(kind) != 0 || siege.Counters.Get(other) != 1 {
				t.Fatalf("Siege counters = %v, want the chosen %v counter removed and the %v counter kept", siege.Counters, kind, other)
			}
			removed := lastCountersRemovedEvent(t, g)
			if removed.PermanentID != siege.ObjectID || removed.CounterKind != kind || removed.Amount != 1 {
				t.Fatalf("counters removed event = %+v, want one %v counter from the Siege", removed, kind)
			}
			if kind == counter.Charge {
				e.applyStateBasedActions(g)
				e.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
				if siegeDefeatAbilities(g) != 0 {
					t.Fatal("removing the charge counter triggered the Siege")
				}
				return
			}
			clone := g.Clone()
			cloneSiege, ok := permanentByObjectID(clone, siege.ObjectID)
			if !ok {
				t.Fatal("Siege missing from the clone")
			}
			defeatedOrWaitingSiege(t, clone, e, cloneSiege)
			siegeDefeatAbilityOnTop(t, clone, cloneSiege)
			defeatedOrWaitingSiege(t, g, e, siege)
			siegeDefeatAbilityOnTop(t, g, siege)
		})
	}
}

// TestKindSpecificAmongCostOffersOnlyThatKind proves registered Hopeful
// Initiate's "Remove two +1/+1 counters from among creatures you control"
// offers only +1/+1 counters, one option per counter, and removes the chosen
// ones while other kinds stay (no regression for kind-specific costs).
func TestKindSpecificAmongCostOffersOnlyThatKind(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	setMainPhasePriority(g, game.Player1)
	e := choicesProgramEngine(t)
	def := registryCardDef(t, "Hopeful Initiate")
	initiate := addCombatPermanent(g, game.Player1, def)
	initiate.Counters.Add(counter.PlusOnePlusOne, 1)
	initiate.Counters.Add(counter.Charge, 1)
	bear := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Counter Bear",
		Types: []types.Card{types.Creature}, Power: opt.Val(game.PT{Value: 2}), Toughness: opt.Val(game.PT{Value: 2})}})
	bear.Counters.Add(counter.PlusOnePlusOne, 2)
	bear.Counters.Add(counter.Charge, 1)
	relic := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Target Relic", Types: []types.Card{types.Artifact}}})
	g.Players[game.Player1].ManaPool.Add(mana.W, 3)
	agent := &counterLabelAgent{picks: [][]string{{"Hopeful Initiate: +1/+1 counter", "Counter Bear: +1/+1 counter"}}}
	activate := action.ActivateAbility(initiate.ObjectID, def.ActivatedAbilityIndex(0), []game.Target{game.PermanentTarget(relic.ObjectID)}, 0)
	if !e.applyActionWithChoices(g, game.Player1, activate, counterKindCostAgents(agent), &TurnLog{}) {
		t.Fatal("registered Hopeful Initiate activation rejected")
	}
	wantLabels := [][]string{{"Hopeful Initiate: +1/+1 counter", "Counter Bear: +1/+1 counter", "Counter Bear: +1/+1 counter"}}
	if got := agent.paymentLabels(); !slices.EqualFunc(got, wantLabels, slices.Equal[[]string]) {
		t.Fatalf("payment prompts = %q, want %q", got, wantLabels)
	}
	if initiate.Counters.Get(counter.PlusOnePlusOne) != 0 || initiate.Counters.Get(counter.Charge) != 1 ||
		bear.Counters.Get(counter.PlusOnePlusOne) != 1 || bear.Counters.Get(counter.Charge) != 1 {
		t.Fatalf("counters initiate=%v bear=%v, want one chosen +1/+1 counter removed from each and charge kept", initiate.Counters, bear.Counters)
	}
}

// TestMoonlitLamenterChoosesCounterKindToRemove proves registered Moonlit
// Lamenter's "Remove a counter from this creature" asks which kind to remove
// when it has more than one, and asks nothing when it has only one kind.
func TestMoonlitLamenterChoosesCounterKindToRemove(t *testing.T) {
	cases := []struct {
		name       string
		minus      int
		charge     int
		pick       string
		wantMinus  int
		wantCharge int
		wantPrompt bool
	}{
		{name: "minus", minus: 1, charge: 1, pick: "Moonlit Lamenter: -1/-1 counter", wantCharge: 1, wantPrompt: true},
		{name: "charge", minus: 1, charge: 1, pick: "Moonlit Lamenter: charge counter", wantMinus: 1, wantPrompt: true},
		{name: "single kind", minus: 2, wantMinus: 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := mainPhaseGame(game.Player1)
			setMainPhasePriority(g, game.Player1)
			e := choicesProgramEngine(t)
			def := registryCardDef(t, "Moonlit Lamenter")
			lamenter := addCombatPermanent(g, game.Player1, def)
			lamenter.Counters.Add(counter.MinusOneMinusOne, tc.minus)
			if tc.charge > 0 {
				lamenter.Counters.Add(counter.Charge, tc.charge)
			}
			g.Players[game.Player1].ManaPool.Add(mana.W, 2)
			agent := &counterLabelAgent{picks: [][]string{{tc.pick}}}
			activate := action.ActivateAbility(lamenter.ObjectID, def.ActivatedAbilityIndex(0), nil, 0)
			if !e.applyActionWithChoices(g, game.Player1, activate, counterKindCostAgents(agent), &TurnLog{}) {
				t.Fatal("registered Moonlit Lamenter activation rejected")
			}
			prompts := agent.paymentLabels()
			if tc.wantPrompt {
				want := [][]string{{"Moonlit Lamenter: -1/-1 counter", "Moonlit Lamenter: charge counter"}}
				if !slices.EqualFunc(prompts, want, slices.Equal[[]string]) {
					t.Fatalf("payment prompts = %q, want %q", prompts, want)
				}
			} else if len(prompts) != 0 {
				t.Fatalf("payment prompts = %q, want none for a single kind", prompts)
			}
			if lamenter.Counters.Get(counter.MinusOneMinusOne) != tc.wantMinus || lamenter.Counters.Get(counter.Charge) != tc.wantCharge {
				t.Fatalf("counters = %v, want -1/-1=%d charge=%d", lamenter.Counters, tc.wantMinus, tc.wantCharge)
			}
			if obj, ok := g.Stack.Peek(); !ok || obj.SourceID != lamenter.ObjectID {
				t.Fatal("Moonlit Lamenter's ability isn't on the stack")
			}
		})
	}
}

// TestGnarlbarkElmRemovesTwoChosenCounters proves registered Gnarlbark Elm's
// "Remove two counters from this creature" removes the two counters chosen,
// which may be of different kinds.
func TestGnarlbarkElmRemovesTwoChosenCounters(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	setMainPhasePriority(g, game.Player1)
	e := choicesProgramEngine(t)
	def := registryCardDef(t, "Gnarlbark Elm")
	elm := addCombatPermanent(g, game.Player1, def)
	elm.Counters.Add(counter.MinusOneMinusOne, 2)
	elm.Counters.Add(counter.Charge, 1)
	victim := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Victim",
		Types: []types.Card{types.Creature}, Power: opt.Val(game.PT{Value: 3}), Toughness: opt.Val(game.PT{Value: 3})}})
	g.Players[game.Player1].ManaPool.Add(mana.B, 3)
	agent := &counterLabelAgent{picks: [][]string{{"Gnarlbark Elm: charge counter", "Gnarlbark Elm: -1/-1 counter"}}}
	activate := action.ActivateAbility(elm.ObjectID, def.ActivatedAbilityIndex(0), []game.Target{game.PermanentTarget(victim.ObjectID)}, 0)
	if !e.applyActionWithChoices(g, game.Player1, activate, counterKindCostAgents(agent), &TurnLog{}) {
		t.Fatal("registered Gnarlbark Elm activation rejected")
	}
	if len(agent.paymentLabels()) != 1 || len(agent.paymentLabels()[0]) != 3 {
		t.Fatalf("payment prompts = %q, want one prompt offering all three counters", agent.paymentLabels())
	}
	if elm.Counters.Get(counter.MinusOneMinusOne) != 1 || elm.Counters.Get(counter.Charge) != 0 {
		t.Fatalf("counters = %v, want the charge counter and one -1/-1 counter removed", elm.Counters)
	}
}

// TestSourceKindSpecificCostReservesCountersForAnyKindCost proves an any-kind
// "from this permanent" cost is offered only the counters a kind-specific cost
// of the same ability doesn't claim, prompting only for a real choice.
func TestSourceKindSpecificCostReservesCountersForAnyKindCost(t *testing.T) {
	ability := &game.ActivatedAbility{
		Text: "Remove a charge counter and a counter from this creature: Draw a card.",
		AdditionalCosts: []cost.Additional{
			{Kind: cost.AdditionalRemoveCounter, Text: "Remove a charge counter from this creature", Amount: 1, CounterKind: counter.Charge},
			{Kind: cost.AdditionalRemoveCounter, Text: "Remove a counter from this creature", Amount: 1, AnyCounterKind: true},
		},
		Content: game.Mode{}.Ability(),
	}
	cases := []struct {
		name        string
		charge      int
		pick        string
		wantPrompt  []string
		wantCharge  int
		wantDefense int
	}{
		{name: "forced", charge: 1},
		{name: "chosen", charge: 2, pick: "Activated Creature: defense counter",
			wantPrompt: []string{"Activated Creature: charge counter", "Activated Creature: defense counter"}, wantCharge: 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := mainPhaseGame(game.Player1)
			setMainPhasePriority(g, game.Player1)
			e := choicesProgramEngine(t)
			def := activatedAbilityPermanent(ability)
			source := addCombatPermanent(g, game.Player1, def)
			source.Counters.Add(counter.Charge, tc.charge)
			source.Counters.Add(counter.Defense, 1)
			agent := &counterLabelAgent{picks: [][]string{{tc.pick}}}
			if !e.applyActionWithChoices(g, game.Player1, action.ActivateAbility(source.ObjectID, def.ActivatedAbilityIndex(0), nil, 0), counterKindCostAgents(agent), &TurnLog{}) {
				t.Fatal("activation rejected")
			}
			prompts := agent.paymentLabels()
			if tc.wantPrompt == nil && len(prompts) != 0 || tc.wantPrompt != nil && (len(prompts) != 1 || !slices.Equal(prompts[0], tc.wantPrompt)) {
				t.Fatalf("payment prompts = %q, want %q", prompts, tc.wantPrompt)
			}
			if source.Counters.Get(counter.Charge) != tc.wantCharge || source.Counters.Get(counter.Defense) != tc.wantDefense {
				t.Fatalf("counters = %v, want charge=%d defense=%d", source.Counters, tc.wantCharge, tc.wantDefense)
			}
		})
	}
}

// TestManaAbilityAnyCounterCostChoosesKind proves a mana ability's any-kind
// "from this permanent" removal cost removes the chosen kind.
func TestManaAbilityAnyCounterCostChoosesKind(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	setMainPhasePriority(g, game.Player1)
	e := choicesProgramEngine(t)
	def := &game.CardDef{CardFace: game.CardFace{Name: "Counter Battery", Types: []types.Card{types.Artifact},
		ManaAbilities: []game.ManaAbility{{
			Text:            "Remove a counter from this artifact: Add {G}.",
			AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalRemoveCounter, Text: "Remove a counter from this artifact", Amount: 1, AnyCounterKind: true}},
			Content:         game.Mode{Sequence: []game.Instruction{{Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.G}}}}.Ability(),
		}}}}
	battery := addCombatPermanent(g, game.Player1, def)
	battery.Counters.Add(counter.Charge, 1)
	battery.Counters.Add(counter.Defense, 1)
	agent := &counterLabelAgent{picks: [][]string{{"Counter Battery: defense counter"}}}
	if !e.applyActionWithChoices(g, game.Player1, action.ActivateAbility(battery.ObjectID, def.ManaAbilityIndex(0), nil, 0), counterKindCostAgents(agent), &TurnLog{}) {
		t.Fatal("mana ability activation rejected")
	}
	if battery.Counters.Get(counter.Defense) != 0 || battery.Counters.Get(counter.Charge) != 1 {
		t.Fatalf("counters = %v, want the chosen defense counter removed", battery.Counters)
	}
	if g.Players[game.Player1].ManaPool.Total() != 1 {
		t.Fatalf("mana pool = %v, want {G} added", g.Players[game.Player1].ManaPool)
	}
}

// TestChiseiUpkeepPaymentRemovesChosenSiegeCounter proves registered Chisei's
// resolution payment "remove a counter from a permanent you control" lets the
// player remove a Siege's defense counter instead of its charge counter,
// keeping Chisei and triggering the Siege.
func TestChiseiUpkeepPaymentRemovesChosenSiegeCounter(t *testing.T) {
	g := mainPhaseGame(game.Player1)
	g.Turn.Phase = game.PhaseBeginning
	g.Turn.Step = game.StepUpkeep
	e := choicesProgramEngine(t)
	chisei := addCombatPermanent(g, game.Player1, registryCardDef(t, "Chisei, Heart of Oceans"))
	siege := chargeAndDefenseSiege(t, g)
	agent := &counterLabelAgent{picks: [][]string{{"Invasion of Segovia: defense counter"}}}
	agents := counterKindCostAgents(agent)
	log := TurnLog{}
	emitBeginningOfStepEvent(g, game.StepUpkeep)
	if !e.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log) {
		t.Fatal("Chisei's upkeep trigger wasn't put on the stack")
	}
	e.resolveTopOfStackWithChoices(g, agents, &log)
	if _, ok := permanentByObjectID(g, chisei.ObjectID); !ok {
		t.Fatal("Chisei was sacrificed although the counter was removed")
	}
	if siege.Counters.Get(counter.Defense) != 0 || siege.Counters.Get(counter.Charge) != 1 {
		t.Fatalf("Siege counters = %v, want the chosen defense counter removed", siege.Counters)
	}
	defeatedOrWaitingSiege(t, g, e, siege)
	siegeDefeatAbilityOnTop(t, g, siege)
}

// TestSpellAnyCounterAmongCostChoosesKind proves a spell's any-kind "from
// among" additional cost removes the chosen Siege defense counter when cast,
// triggering the Siege in the game and a clone, and that an unaffordable cast
// removes nothing.
func TestSpellAnyCounterAmongCostChoosesKind(t *testing.T) {
	for _, canPay := range []bool{false, true} {
		t.Run(map[bool]string{false: "unpayable", true: "payable"}[canPay], func(t *testing.T) {
			g := mainPhaseGame(game.Player1)
			setMainPhasePriority(g, game.Player1)
			e := choicesProgramEngine(t)
			siege := chargeAndDefenseSiege(t, g)
			def := &game.CardDef{CardFace: game.CardFace{Name: "Counter Toll", Types: []types.Card{types.Instant},
				ManaCost: opt.Val(cost.Mana{cost.O(1)}),
				AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalRemoveCounterAmong,
					Text: "remove a counter from a permanent you control", Amount: 1, AnyCounterKind: true}},
				SpellAbility: opt.Val(game.AbilityContent{})}}
			spell := addCardToHand(g, game.Player1, def)
			if canPay {
				g.Players[game.Player1].ManaPool.Add(mana.C, 1)
			}
			agent := &counterLabelAgent{picks: [][]string{{"Invasion of Segovia: defense counter"}}}
			accepted := e.applyActionWithChoices(g, game.Player1, action.CastSpell(spell, nil, 0, nil), counterKindCostAgents(agent), &TurnLog{})
			if accepted != canPay {
				t.Fatalf("cast accepted = %v, want %v", accepted, canPay)
			}
			if !canPay {
				if siege.Counters.Get(counter.Defense) != 1 || siege.Counters.Get(counter.Charge) != 1 || len(countersRemovedEvents(g)) != 0 {
					t.Fatal("rejected cast removed counters")
				}
				return
			}
			if siege.Counters.Get(counter.Defense) != 0 || siege.Counters.Get(counter.Charge) != 1 {
				t.Fatalf("Siege counters = %v, want the chosen defense counter removed", siege.Counters)
			}
			clone := g.Clone()
			cloneSiege, ok := permanentByObjectID(clone, siege.ObjectID)
			if !ok {
				t.Fatal("Siege missing from the clone")
			}
			defeatedOrWaitingSiege(t, clone, e, cloneSiege)
			siegeDefeatAbilityOnTop(t, clone, cloneSiege)
			defeatedOrWaitingSiege(t, g, e, siege)
			siegeDefeatAbilityOnTop(t, g, siege)
		})
	}
}
