package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: Engine #3742 (reflexive triggers carry "that many" / X).
// CR 2026-11-06 from the official copy linked at magic.wizards.com/en/rules:
// 107.3c and 107.3n (an ability's X is fixed by the effect that created it),
// 601.2c (number of targets fixed as announced), 603.3c and 700.2b (a modal
// triggered ability announces modes as it is put on the stack; with no mode it
// is removed), 603.3d, 603.7 and 603.12 (reflexive triggers follow delayed
// trigger rules and trigger once even when a cost is paid several times),
// 700.2g and 707.10 (copies keep modes and X). Scryfall rulings: Spectral
// Adversary 7c67fe04 (9 rulings, 2021-09-24: targets are chosen as the second
// ability triggers and players may respond), Tainted Adversary e6b9b8fc (5,
// 2021-09-24), Intrepid Adversary 773199f9 (1, 2021-09-24: players may respond
// to the reflexive ability), Tranquil Frillback 5b647377 (1, 2023-05-12: modes
// and targets are chosen when the reflexive ability triggers), Hawkeye, Master
// Marksman 9991b684 (no rulings). The scenarios below use test-local
// definitions with the cards' pay-repeatedly / "that many" shape.

const (
	reflexiveXPaidKey  = game.ResultKey("reflexive-x-paid")
	reflexiveXCountKey = game.ResultKey("reflexive-x-count")
)

// reflexiveXEnablingContent is "you may pay <payment> any number of times (up
// to maxCount when positive). When you pay this cost one or more times,
// <reflexive>", with the payment count carried as the reflexive ability's X.
func reflexiveXEnablingContent(payment cost.Mana, maxCount int, reflexive game.AbilityContent) game.AbilityContent {
	pay := game.PayRepeatedly{
		Payment:      game.ResolutionPayment{ManaCost: opt.Val(payment)},
		PublishCount: reflexiveXCountKey,
		Prompt:       "Pay?",
	}
	if maxCount > 0 {
		pay.MaxCount = opt.Val(&game.DynamicAmount{Kind: game.DynamicAmountConstant, Constant: maxCount, Multiplier: 1})
	}
	return game.Mode{Sequence: []game.Instruction{
		{Primitive: pay, PublishResult: reflexiveXPaidKey},
		{
			Primitive: game.CreateReflexiveTrigger{Trigger: game.ReflexiveTriggerDef{
				Content: reflexive,
				X: opt.Val(game.Dynamic(game.DynamicAmount{
					Kind:      game.DynamicAmountChosenNumber,
					ResultKey: reflexiveXCountKey,
				})),
			}},
			ResultGate: opt.Val(game.InstructionResultGate{Key: reflexiveXPaidKey, Succeeded: game.TriTrue}),
		},
	}}.Ability()
}

func reflexiveX() game.Quantity {
	return game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX, Multiplier: 1})
}

func pushReflexiveXEnablingTrigger(g *game.Game, source *game.Permanent, content game.AbilityContent) {
	trigger := game.TriggeredAbility{Content: content}
	g.Stack.Push(&game.StackObject{
		ID:            g.IDGen.Next(),
		Kind:          game.StackTriggeredAbility,
		SourceID:      source.ObjectID,
		SourceCardID:  source.CardInstanceID,
		Controller:    source.Controller,
		InlineTrigger: &trigger,
	})
}

// reflexiveXAgent pays up to pays times, picks modes, and picks a target set
// with pickTarget (the largest offered set by default). It records every
// choice request so tests can inspect announced ranges.
type reflexiveXAgent struct {
	simPassPolicy

	pays       int
	paid       int
	modes      []int
	pickTarget func([]game.ChoiceOption) int
	requests   []game.ChoiceRequest
}

func (a *reflexiveXAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	a.requests = append(a.requests, request)
	switch request.Kind {
	case game.ChoiceMay:
		if a.paid < a.pays {
			a.paid++
			return []int{1}
		}
		return []int{0}
	case game.ChoiceModal:
		return append([]int(nil), a.modes...)
	case game.ChoiceTarget:
		if a.pickTarget != nil {
			return []int{a.pickTarget(request.Options)}
		}
		best := 0
		for i, option := range request.Options {
			if len(option.Targets) > len(request.Options[best].Targets) {
				best = i
			}
		}
		return []int{best}
	default:
		return request.DefaultSelection
	}
}

func (a *reflexiveXAgent) requestsOfKind(kind game.ChoiceKind) []game.ChoiceRequest {
	var out []game.ChoiceRequest
	for _, request := range a.requests {
		if request.Kind == kind {
			out = append(out, request)
		}
	}
	return out
}

func reflexiveXCreature(name string, power int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: power}),
		Toughness: opt.Val(game.PT{Value: power}),
	}}
}

func reflexiveXAgents(controller *reflexiveXAgent) [game.NumPlayers]PlayerAgent {
	var agents [game.NumPlayers]PlayerAgent
	for i := range agents {
		agents[i] = &reflexiveXAgent{}
	}
	agents[game.Player1] = controller
	return agents
}

// intrepidLikeDef is "Creatures you control get +1/+1 for each valor counter
// on this creature."
func intrepidLikeDef() *game.CardDef {
	valor := game.DynamicAmount{
		Kind:        game.DynamicAmountObjectCounters,
		Multiplier:  1,
		CounterKind: counter.Valor,
		Object:      game.SourcePermanentReference(),
	}
	def := reflexiveXCreature("Intrepid Analogue", 1)
	def.StaticAbilities = []game.StaticAbility{{ContinuousEffects: []game.ContinuousEffect{{
		Layer:                 game.LayerPowerToughnessModify,
		Group:                 game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
		PowerDeltaDynamic:     opt.Val(valor),
		ToughnessDeltaDynamic: opt.Val(valor),
	}}}}
	return def
}

// intrepidLikeEnabling is "you may pay {1}{W} any number of times. When you pay
// this cost one or more times, put that many valor counters on this creature."
func intrepidLikeEnabling() game.AbilityContent {
	return reflexiveXEnablingContent(cost.Mana{cost.O(1), cost.W}, 0, game.Mode{Sequence: []game.Instruction{{
		Primitive: game.AddCounter{Amount: reflexiveX(), Object: game.SourcePermanentReference(), CounterKind: counter.Valor},
	}}}.Ability())
}

type reflexiveXPriorityRecord struct {
	player    game.PlayerID
	stackSize int
	topX      int
	counters  int
}

// reflexiveXPriorityRecorder records each priority grant while a reflexive
// trigger with a carried X is on top of the stack.
type reflexiveXPriorityRecorder struct {
	reflexiveXAgent

	g       *game.Game
	source  *game.Permanent
	records []reflexiveXPriorityRecord
}

func (r *reflexiveXPriorityRecorder) ChooseAction(obs PlayerObservation, legal []action.Action) action.Action {
	if top, ok := r.g.Stack.Peek(); ok && top.Kind == game.StackTriggeredAbility && top.XValue != 0 {
		r.records = append(r.records, reflexiveXPriorityRecord{
			player:    obs.Player,
			stackSize: r.g.Stack.Size(),
			topX:      top.XValue,
			counters:  r.source.Counters.Get(counter.Valor),
		})
	}
	return r.simPassPolicy.ChooseAction(obs, legal)
}

// TestReflexiveTriggerCarriesPaymentCountAsSeparateRespondableAbility plays the
// Intrepid Adversary shape: paying {1}{W} twice creates one reflexive trigger
// whose X is 2. It is a separate stack object that every player, including the
// opponent, gets priority over before it resolves (CR 603.3, 603.12); it then
// puts two valor counters on the source, so creatures you control get +2/+2.
func TestReflexiveTriggerCarriesPaymentCountAsSeparateRespondableAbility(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, intrepidLikeDef())
	ally := addCombatPermanent(g, game.Player1, reflexiveXCreature("Ally", 2))
	for range 4 {
		addBasicLandPermanent(g, game.Player1, types.Plains)
	}
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	pushReflexiveXEnablingTrigger(g, source, intrepidLikeEnabling())

	recorder := &reflexiveXPriorityRecorder{reflexiveXAgent: reflexiveXAgent{pays: 99}, g: g, source: source}
	agents := [game.NumPlayers]PlayerAgent{recorder, recorder, recorder, recorder}
	engine.runPriorityLoop(g, agents, &TurnLog{})

	opponentResponded := false
	for _, record := range recorder.records {
		if record.stackSize != 1 || record.topX != 2 || record.counters != 0 {
			t.Fatalf("priority record %+v, want only the reflexive trigger (X=2) on the stack before counters", record)
		}
		if record.player == game.Player2 {
			opponentResponded = true
		}
	}
	if !opponentResponded {
		t.Fatalf("opponent never received priority with the reflexive trigger on the stack: %+v", recorder.records)
	}
	if got := recorder.paid; got != 2 {
		t.Fatalf("paid %d times, want 2 (four Plains fund two {1}{W} payments)", got)
	}
	if got := source.Counters.Get(counter.Valor); got != 2 {
		t.Fatalf("source has %d valor counters, want 2", got)
	}
	if got := effectivePower(g, ally); got != 4 {
		t.Fatalf("ally power = %d, want 4 (+2/+2 from two valor counters)", got)
	}
	if got := effectivePower(g, source); got != 3 {
		t.Fatalf("source power = %d, want 3", got)
	}
}

// TestReflexiveTriggerWithCarriedXNotCreatedWhenNothingPaid proves declining
// every payment creates no reflexive trigger and no counters.
func TestReflexiveTriggerWithCarriedXNotCreatedWhenNothingPaid(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, intrepidLikeDef())
	for range 4 {
		addBasicLandPermanent(g, game.Player1, types.Plains)
	}
	pushReflexiveXEnablingTrigger(g, source, intrepidLikeEnabling())

	agents := reflexiveXAgents(&reflexiveXAgent{pays: 0})
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if len(g.PendingReflexiveTriggers) != 0 {
		t.Fatalf("pending reflexive triggers = %+v, want none after declining every payment", g.PendingReflexiveTriggers)
	}
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want 0", g.Stack.Size())
	}
	if got := source.Counters.Get(counter.Valor); got != 0 {
		t.Fatalf("source has %d valor counters, want 0", got)
	}
}

// TestReflexiveTriggerCarriedXScalesCountersAndTwiceThatManyTokens plays the
// Tainted Adversary shape: paying {2}{B} twice puts two +1/+1 counters on the
// source and then creates twice that many 2/2 black Zombie tokens.
func TestReflexiveTriggerCarriedXScalesCountersAndTwiceThatManyTokens(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, reflexiveXCreature("Tainted Analogue", 2))
	for range 6 {
		addBasicLandPermanent(g, game.Player1, types.Swamp)
	}
	zombie := &game.CardDef{CardFace: game.CardFace{
		Name:      "Zombie",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Zombie},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}}
	twiceX := game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX, Multiplier: 2})
	pushReflexiveXEnablingTrigger(g, source, reflexiveXEnablingContent(cost.Mana{cost.O(2), cost.B}, 0, game.Mode{Sequence: []game.Instruction{
		{Primitive: game.AddCounter{Amount: reflexiveX(), Object: game.SourcePermanentReference(), CounterKind: counter.PlusOnePlusOne}},
		{Primitive: game.CreateToken{Amount: twiceX, Source: game.TokenDef(zombie)}},
	}}.Ability()))

	resolveStackWithTriggers(engine, g, reflexiveXAgents(&reflexiveXAgent{pays: 99}))

	if got := source.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("source has %d +1/+1 counters, want 2", got)
	}
	zombies := 0
	for _, permanent := range g.Battlefield {
		if permanent.Controller == game.Player1 && permanent.Token && permanentHasSubtype(g, permanent, types.Zombie) {
			zombies++
		}
	}
	if zombies != 4 {
		t.Fatalf("created %d Zombie tokens, want 4", zombies)
	}
}

// spectralLikeReflexive is "put that many +1/+1 counters on this creature, then
// up to that many other target artifacts, creatures, and/or enchantments phase
// out."
func spectralLikeReflexive() game.AbilityContent {
	return game.Mode{
		Targets: []game.TargetSpec{{
			MinTargets:      0,
			MaxTargets:      99,
			MaxTargetsFromX: true,
			Constraint:      "up to that many other target artifacts, creatures, and/or enchantments",
			Allow:           game.TargetAllowPermanent,
			Selection: opt.Val(game.Selection{
				RequiredTypesAny: []types.Card{types.Artifact, types.Creature, types.Enchantment},
				ExcludeSource:    true,
			}),
		}},
		Sequence: []game.Instruction{
			{Primitive: game.AddCounter{Amount: reflexiveX(), Object: game.SourcePermanentReference(), CounterKind: counter.PlusOnePlusOne}},
			{Primitive: game.PhaseOut{Object: game.AllTargetPermanentsReference(0)}},
		},
	}.Ability()
}

func spectralLikeBoard(t *testing.T) (*game.Game, *Engine, *game.Permanent, []*game.Permanent) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, reflexiveXCreature("Spectral Analogue", 1))
	for range 4 {
		addBasicLandPermanent(g, game.Player1, types.Island)
	}
	opponents := []*game.Permanent{
		addCombatPermanent(g, game.Player2, reflexiveXCreature("Opposing Bear", 2)),
		addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Opposing Relic", Types: []types.Card{types.Artifact}}}),
		addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Opposing Aura", Types: []types.Card{types.Enchantment}}}),
	}
	pushReflexiveXEnablingTrigger(g, source, reflexiveXEnablingContent(cost.Mana{cost.O(1), cost.U}, 0, spectralLikeReflexive()))
	return g, engine, source, opponents
}

// TestReflexiveTriggerCarriedXLimitsUpToThatManyTargets plays the Spectral
// Adversary shape: after paying {1}{U} twice, the reflexive trigger is put on
// the stack offering at most two other targets (CR 601.2c via 603.3d), a
// three-target announcement is rejected, and the two chosen permanents phase
// out while the source gets two +1/+1 counters.
func TestReflexiveTriggerCarriedXLimitsUpToThatManyTargets(t *testing.T) {
	g, engine, source, opponents := spectralLikeBoard(t)
	agent := &reflexiveXAgent{pays: 99}
	agents := reflexiveXAgents(agent)
	log := &TurnLog{}

	engine.resolveTopOfStackWithChoices(g, agents, log)
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, log)
	reflexive, ok := g.Stack.Peek()
	if !ok || reflexive.XValue != 2 {
		t.Fatalf("reflexive stack object = %+v, want X=2", reflexive)
	}
	targetRequests := agent.requestsOfKind(game.ChoiceTarget)
	if len(targetRequests) != 1 {
		t.Fatalf("target requests = %d, want 1", len(targetRequests))
	}
	most := 0
	for _, option := range targetRequests[0].Options {
		most = max(most, len(option.Targets))
		for _, target := range option.Targets {
			if target.PermanentID == source.ObjectID {
				t.Fatalf("offered the source itself as an \"other\" target: %+v", option.Targets)
			}
		}
	}
	if most != 2 {
		t.Fatalf("largest offered target set = %d, want 2 (that many)", most)
	}
	ability := reflexive.InlineTrigger
	sourceDef, _ := permanentCardDef(g, source)
	all := []game.Target{
		game.PermanentTarget(opponents[0].ObjectID),
		game.PermanentTarget(opponents[1].ObjectID),
		game.PermanentTarget(opponents[2].ObjectID),
	}
	if _, legal := bodyTargetCountsWithModes(g, game.Player1, sourceDef, source.ObjectID, game.Event{}, ability, nil, all, reflexive.XValue); legal {
		t.Fatal("three targets accepted for an X=2 \"up to that many\" announcement")
	}
	if len(reflexive.Targets) != 2 {
		t.Fatalf("reflexive targets = %+v, want 2", reflexive.Targets)
	}

	resolveStackWithTriggers(engine, g, agents)

	if got := source.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("source has %d +1/+1 counters, want 2", got)
	}
	phased := 0
	for _, permanent := range opponents {
		if permanent.PhasedOut {
			phased++
		}
	}
	if phased != 2 || source.PhasedOut {
		t.Fatalf("phased out %d opponents (source phased=%v), want exactly the two chosen targets", phased, source.PhasedOut)
	}
}

// TestCopiedReflexiveTriggerKeepsCarriedXForNewTargets proves a copy of the
// reflexive ability keeps its X (CR 707.10), so new targets chosen for the copy
// stay limited to that many.
func TestCopiedReflexiveTriggerKeepsCarriedXForNewTargets(t *testing.T) {
	g, engine, _, _ := spectralLikeBoard(t)
	agent := &reflexiveXAgent{pays: 99}
	agents := reflexiveXAgents(agent)
	log := &TurnLog{}
	engine.resolveTopOfStackWithChoices(g, agents, log)
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, log)
	original, _ := g.Stack.Peek()

	addEffectSpellToStack(g, game.Player1,
		game.CopyStackObject{Object: game.TargetStackObjectReference(0), MayChooseNewTargets: true},
		[]game.Target{game.StackObjectTarget(original.ID)})
	engine.resolveTopOfStackWithChoices(g, agents, log)

	copied, ok := g.Stack.Peek()
	if !ok || !copied.Copy || copied.ID == original.ID {
		t.Fatalf("top of stack = %+v, want a copy of the reflexive trigger", copied)
	}
	if copied.XValue != 2 {
		t.Fatalf("copy X = %d, want 2", copied.XValue)
	}
	if len(copied.Targets) > 2 {
		t.Fatalf("copy targets = %+v, want at most 2", copied.Targets)
	}
}

// frillbackLikeReflexive is "choose up to that many — • Destroy target artifact
// or enchantment. • Exile target player's graveyard. • You gain 4 life."
func frillbackLikeReflexive() game.AbilityContent {
	return game.AbilityContent{
		Modes: []game.Mode{
			{
				Text: "Destroy target artifact or enchantment.",
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target artifact or enchantment",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}}),
				}},
				Sequence: []game.Instruction{{Primitive: game.Destroy{Object: game.TargetPermanentReference(0)}}},
			},
			{
				Text: "Exile target player's graveyard.",
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target player's graveyard",
					Allow:      game.TargetAllowPlayer,
				}},
				Sequence: []game.Instruction{{Primitive: game.MoveCard{
					Player:      game.TargetPlayerReference(0),
					FromZone:    zone.Graveyard,
					Destination: zone.Exile,
				}}},
			},
			{
				Text:     "You gain 4 life.",
				Sequence: []game.Instruction{{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(4)}}},
			},
		},
		MinModes:      0,
		MaxModes:      3,
		MaxModesFromX: true,
	}
}

// TestReflexiveTriggerCarriedXLimitsChooseUpToThatManyModes plays the Tranquil
// Frillback shape: paying {G} twice (of up to three) announces the reflexive
// trigger's modes as it is put on the stack with at most two choices (CR 603.3c),
// and both chosen modes resolve.
func TestReflexiveTriggerCarriedXLimitsChooseUpToThatManyModes(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, reflexiveXCreature("Frillback Analogue", 3))
	for range 5 {
		addBasicLandPermanent(g, game.Player1, types.Forest)
	}
	relic := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Opposing Relic", Types: []types.Card{types.Artifact}}})
	pushReflexiveXEnablingTrigger(g, source, reflexiveXEnablingContent(cost.Mana{cost.G}, 3, frillbackLikeReflexive()))
	life := g.Players[game.Player1].Life

	agent := &reflexiveXAgent{pays: 2, modes: []int{0, 2}}
	agents := reflexiveXAgents(agent)
	log := &TurnLog{}
	engine.resolveTopOfStackWithChoices(g, agents, log)
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, log)

	if offers := len(agent.requestsOfKind(game.ChoiceMay)); offers != 3 {
		t.Fatalf("payment offers = %d, want 3 (two accepted, third declined)", offers)
	}
	modal := agent.requestsOfKind(game.ChoiceModal)
	if len(modal) != 1 || modal[0].MinChoices != 0 || modal[0].MaxChoices != 2 {
		t.Fatalf("mode requests = %+v, want one request choosing 0..2 modes", modal)
	}
	reflexive, ok := g.Stack.Peek()
	if !ok || reflexive.XValue != 2 || !slices.Equal(reflexive.ChosenModes, []int{0, 2}) {
		t.Fatalf("reflexive stack object = %+v, want X=2 with modes [0 2]", reflexive)
	}

	resolveStackWithTriggers(engine, g, agents)

	if _, present := permanentByObjectID(g, relic.ObjectID); present {
		t.Fatal("chosen destroy mode left the artifact on the battlefield")
	}
	if got := g.Players[game.Player1].Life; got != life+4 {
		t.Fatalf("life = %d, want %d", got, life+4)
	}
}

// TestReflexiveTriggerCarriedXOffersOnlyModesWithLegalTargets plays the
// Frillback shape with no artifact or enchantment on the battlefield: the
// destroy mode can't be chosen (CR 700.2b), so only the graveyard and life modes
// are offered, still up to X of them, and both chosen modes resolve.
func TestReflexiveTriggerCarriedXOffersOnlyModesWithLegalTargets(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, reflexiveXCreature("Frillback Analogue", 3))
	for range 2 {
		addBasicLandPermanent(g, game.Player1, types.Forest)
	}
	buried := addCardToGraveyard(g, game.Player2, reflexiveXCreature("Buried", 1))
	pushReflexiveXEnablingTrigger(g, source, reflexiveXEnablingContent(cost.Mana{cost.G}, 3, frillbackLikeReflexive()))
	life := g.Players[game.Player1].Life

	agent := &reflexiveXAgent{pays: 99, modes: []int{1, 2}, pickTarget: func(options []game.ChoiceOption) int {
		for i, option := range options {
			if len(option.Targets) == 1 && option.Targets[0].Kind == game.TargetPlayer && option.Targets[0].PlayerID == game.Player2 {
				return i
			}
		}
		t.Fatalf("target options %+v lack Player2's graveyard", options)
		return 0
	}}
	resolveStackWithTriggers(engine, g, reflexiveXAgents(agent))

	modal := agent.requestsOfKind(game.ChoiceModal)
	if len(modal) != 1 || modal[0].MaxChoices != 2 || len(modal[0].Options) != 2 ||
		modal[0].Options[0].Index != 1 || modal[0].Options[1].Index != 2 {
		t.Fatalf("mode requests = %+v, want up to two of modes [1 2]", modal)
	}
	if !g.Players[game.Player2].Exile.Contains(buried) {
		t.Fatal("chosen graveyard mode did not exile Player2's graveyard")
	}
	if got := g.Players[game.Player1].Life; got != life+4 {
		t.Fatalf("life = %d, want %d", got, life+4)
	}
}

// TestReflexiveTriggerCarriedXRejectsMoreModesThanThatMany proves a mode
// selection larger than X is illegal and the ability is removed (CR 700.2b),
// as is an empty selection.
func TestReflexiveTriggerCarriedXRejectsMoreModesThanThatMany(t *testing.T) {
	for name, modes := range map[string][]int{"three of two": {0, 1, 2}, "none": nil} {
		t.Run(name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			source := addCombatPermanent(g, game.Player1, reflexiveXCreature("Frillback Analogue", 3))
			for range 2 {
				addBasicLandPermanent(g, game.Player1, types.Forest)
			}
			addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Opposing Relic", Types: []types.Card{types.Artifact}}})
			pushReflexiveXEnablingTrigger(g, source, reflexiveXEnablingContent(cost.Mana{cost.G}, 3, frillbackLikeReflexive()))
			life := g.Players[game.Player1].Life

			agents := reflexiveXAgents(&reflexiveXAgent{pays: 99, modes: modes})
			resolveStackWithTriggers(engine, g, agents)

			if got := g.Players[game.Player1].Life; got != life {
				t.Fatalf("life = %d, want unchanged %d", got, life)
			}
			if g.Stack.Size() != 0 {
				t.Fatalf("stack size = %d, want the reflexive ability removed", g.Stack.Size())
			}
		})
	}
}

// TestReflexiveTriggerCarriedXFromCappedGenericPayments plays the Hawkeye,
// Master Marksman shape: "you may pay {1} up to three times. When you do,
// choose up to that many —". With five lands only three payments are offered,
// so up to three modes may be chosen; Explosive and Boomerang resolve.
func TestReflexiveTriggerCarriedXFromCappedGenericPayments(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, reflexiveXCreature("Hawkeye Analogue", 2))
	for range 5 {
		addBasicLandPermanent(g, game.Player1, types.Mountain)
	}
	addCombatPermanent(g, game.Player2, reflexiveXCreature("Opposing Bear", 2))
	discarded := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Discarded"}})
	drawn := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	reflexive := game.AbilityContent{
		Modes: []game.Mode{
			{
				Text: "Net — Target creature can't block this turn.",
				Targets: []game.TargetSpec{{
					MinTargets: 1, MaxTargets: 1, Constraint: "target creature", Allow: game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}},
				Sequence: []game.Instruction{{Primitive: game.ApplyRule{
					Object:      opt.Val(game.TargetPermanentReference(0)),
					RuleEffects: []game.RuleEffect{{Kind: game.RuleEffectCantBlock, Duration: game.DurationPermanent}},
					Duration:    game.DurationThisTurn,
				}}},
			},
			{
				Text:    "Explosive — Hawkeye deals 2 damage to target player.",
				Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "target player", Allow: game.TargetAllowPlayer}},
				Sequence: []game.Instruction{{Primitive: game.Damage{
					Amount:    game.Fixed(2),
					Recipient: game.PlayerDamageRecipient(game.TargetPlayerReference(0)),
				}}},
			},
			{
				Text: "Boomerang — Discard a card, then draw a card.",
				Sequence: []game.Instruction{
					{Primitive: game.Discard{Amount: game.Fixed(1), Player: game.ControllerReference()}},
					{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}},
				},
			},
		},
		MinModes:      0,
		MaxModes:      3,
		MaxModesFromX: true,
	}
	pushReflexiveXEnablingTrigger(g, source, reflexiveXEnablingContent(cost.Mana{cost.O(1)}, 3, reflexive))
	opponentLife := g.Players[game.Player2].Life

	agent := &reflexiveXAgent{pays: 99, modes: []int{1, 2}, pickTarget: func(options []game.ChoiceOption) int {
		for i, option := range options {
			if len(option.Targets) == 1 && option.Targets[0].Kind == game.TargetPlayer && option.Targets[0].PlayerID == game.Player2 {
				return i
			}
		}
		return 0
	}}
	resolveStackWithTriggers(engine, g, reflexiveXAgents(agent))

	if agent.paid != 3 {
		t.Fatalf("paid %d times, want 3 (capped)", agent.paid)
	}
	modal := agent.requestsOfKind(game.ChoiceModal)
	if len(modal) != 1 || modal[0].MaxChoices != 3 {
		t.Fatalf("mode requests = %+v, want one request with up to three modes", modal)
	}
	if got := g.Players[game.Player2].Life; got != opponentLife-2 {
		t.Fatalf("opponent life = %d, want %d", got, opponentLife-2)
	}
	if !g.Players[game.Player1].Graveyard.Contains(discarded) || !g.Players[game.Player1].Hand.Contains(drawn) {
		t.Fatal("Boomerang did not discard the hand card and draw the library card")
	}
}

// TestPendingReflexiveTriggerCarriedXIsCloneIndependent proves the carried X is
// game-owned state copied by value when a game with a waiting reflexive
// trigger is cloned.
func TestPendingReflexiveTriggerCarriedXIsCloneIndependent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, intrepidLikeDef())
	for range 4 {
		addBasicLandPermanent(g, game.Player1, types.Plains)
	}
	pushReflexiveXEnablingTrigger(g, source, intrepidLikeEnabling())
	agents := reflexiveXAgents(&reflexiveXAgent{pays: 99})
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if len(g.PendingReflexiveTriggers) != 1 || g.PendingReflexiveTriggers[0].XValue != opt.Val(2) {
		t.Fatalf("pending reflexive triggers = %+v, want one with X=2", g.PendingReflexiveTriggers)
	}

	clone := g.Clone()
	clone.PendingReflexiveTriggers[0].XValue = opt.Val(5)
	if g.PendingReflexiveTriggers[0].XValue != opt.Val(2) {
		t.Fatalf("original X changed to %+v after mutating the clone", g.PendingReflexiveTriggers[0].XValue)
	}
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
	if top, ok := g.Stack.Peek(); !ok || top.XValue != 2 {
		t.Fatalf("original reflexive stack object = %+v, want X=2", top)
	}
}

// TestReflexiveXLimitsValidation proves X-limited modes and targets are
// accepted on a reflexive trigger that carries X and rejected where no X is
// known as modes and targets are announced.
func TestReflexiveXLimitsValidation(t *testing.T) {
	etb := game.TriggerCondition{Type: game.TriggerWhen, Pattern: game.TriggerPattern{
		Event:  game.EventPermanentEnteredBattlefield,
		Source: game.TriggerSourceSelf,
	}}
	withTrigger := func(content game.AbilityContent) *game.CardDef {
		def := reflexiveXCreature("Validated Analogue", 1)
		def.TriggeredAbilities = []game.TriggeredAbility{{Trigger: etb, Content: content}}
		return def
	}
	for name, content := range map[string]game.AbilityContent{
		"modes":   reflexiveXEnablingContent(cost.Mana{cost.G}, 3, frillbackLikeReflexive()),
		"targets": reflexiveXEnablingContent(cost.Mana{cost.O(1), cost.U}, 0, spectralLikeReflexive()),
	} {
		if issues := game.ValidateCardDef(withTrigger(content)); len(issues) != 0 {
			t.Fatalf("%s: valid reflexive X shape rejected: %+v", name, issues)
		}
	}

	noX := reflexiveXEnablingContent(cost.Mana{cost.G}, 3, frillbackLikeReflexive())
	create := noX.Modes[0].Sequence[1].Primitive.(game.CreateReflexiveTrigger)
	create.Trigger.X = opt.V[game.Quantity]{}
	noX.Modes[0].Sequence[1].Primitive = create
	if issues := game.ValidateCardDef(withTrigger(noX)); len(issues) == 0 {
		t.Fatal("X-limited modes accepted on a reflexive trigger without X")
	}

	activated := reflexiveXCreature("Activated Analogue", 1)
	activated.ActivatedAbilities = []game.ActivatedAbility{{Content: frillbackLikeReflexive()}}
	if issues := game.ValidateCardDef(activated); len(issues) == 0 {
		t.Fatal("X-limited modes accepted on an activated ability")
	}
}
