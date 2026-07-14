package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// anyOpponentLifeLostCondition builds Bloodchief Ascension's first-ability gate:
// an opponent lost threshold or more life this turn.
func anyOpponentLifeLostCondition(threshold int) opt.V[game.Condition] {
	return opt.Val(game.Condition{
		Aggregates: []game.AggregateComparison{{
			Aggregate: game.AggregateAnyOpponentLifeLostThisTurn,
			Op:        compare.GreaterOrEqual,
			Value:     threshold,
		}},
	})
}

// TestConditionAnyOpponentLifeLostThisTurnCumulative proves the gate sums a
// single opponent's life loss across the whole turn and ignores prior turns, so
// two separate 1-life losses this turn satisfy the >= 2 gate.
func TestConditionAnyOpponentLifeLostThisTurnCumulative(t *testing.T) {
	g := buildTwoTurnEventLog(
		[]game.Event{
			// Prior-turn loss must not count toward this turn's total.
			{Kind: game.EventLifeLost, Player: game.Player2, Amount: 9},
		},
		[]game.Event{
			{Kind: game.EventLifeLost, Player: game.Player2, Amount: 1},
			{Kind: game.EventLifeLost, Player: game.Player2, Amount: 1},
		},
	)
	ctx := conditionContext{controller: game.Player1}

	if !conditionSatisfied(g, ctx, anyOpponentLifeLostCondition(2)) {
		t.Fatal("opponent lost 2 this turn should satisfy the >= 2 gate")
	}
	if conditionSatisfied(g, ctx, anyOpponentLifeLostCondition(3)) {
		t.Fatal("opponent lost only 2 this turn must not satisfy the >= 3 gate")
	}
}

// TestConditionAnyOpponentLifeLostThisTurnAllLossKinds proves the gate counts
// every kind of life loss uniformly. Combat damage, noncombat damage, life paid
// as a cost, and direct "loses life" effects all reach the window as per-player
// EventLifeLost events, so their amounts accumulate together.
func TestConditionAnyOpponentLifeLostThisTurnAllLossKinds(t *testing.T) {
	g := buildTwoTurnEventLog(nil, []game.Event{
		{Kind: game.EventLifeLost, Player: game.Player2, Amount: 1}, // e.g. combat damage
		{Kind: game.EventLifeLost, Player: game.Player2, Amount: 1}, // e.g. life paid as a cost
		// Life gained this turn must not net against the loss total.
		{Kind: game.EventLifeGained, Player: game.Player2, Amount: 100},
	})
	if !conditionSatisfied(g, conditionContext{controller: game.Player1}, anyOpponentLifeLostCondition(2)) {
		t.Fatal("mixed life-loss kinds should accumulate and satisfy the >= 2 gate despite later life gain")
	}
}

// TestConditionAnyOpponentLifeLostThisTurnIgnoresController confirms the gate is
// existential over opponents only: the controller's own life loss never unlocks
// it.
func TestConditionAnyOpponentLifeLostThisTurnIgnoresController(t *testing.T) {
	g := buildTwoTurnEventLog(nil, []game.Event{
		{Kind: game.EventLifeLost, Player: game.Player1, Amount: 15},
	})
	if conditionSatisfied(g, conditionContext{controller: game.Player1}, anyOpponentLifeLostCondition(2)) {
		t.Fatal("the controller's own life loss must not satisfy an opponent-life-lost gate")
	}
}

// TestConditionAnyOpponentLifeLostThisTurnPerOpponent proves the loss is
// measured per single opponent, not summed across opponents (Bloodchief ruling:
// two opponents each losing 1 life do not trigger; one opponent must lose 2).
func TestConditionAnyOpponentLifeLostThisTurnPerOpponent(t *testing.T) {
	// Three opponents each lose 1 life: the maximum single-opponent total is 1.
	spread := buildTwoTurnEventLog(nil, []game.Event{
		{Kind: game.EventLifeLost, Player: game.Player2, Amount: 1},
		{Kind: game.EventLifeLost, Player: game.Player3, Amount: 1},
		{Kind: game.EventLifeLost, Player: game.Player4, Amount: 1},
	})
	if conditionSatisfied(spread, conditionContext{controller: game.Player1}, anyOpponentLifeLostCondition(2)) {
		t.Fatal("three opponents each losing 1 life must not satisfy the >= 2 gate")
	}

	// A single opponent losing 2 life does satisfy the gate.
	concentrated := buildTwoTurnEventLog(nil, []game.Event{
		{Kind: game.EventLifeLost, Player: game.Player3, Amount: 2},
	})
	if !conditionSatisfied(concentrated, conditionContext{controller: game.Player1}, anyOpponentLifeLostCondition(2)) {
		t.Fatal("a single opponent losing 2 life should satisfy the >= 2 gate")
	}
}

// TestBloodchiefEndStepTriggerFiresOnOpponentEndStep proves ability 1 triggers
// at each end step, not only the controller's. The pattern uses
// TriggerControllerAny, so on an opponent's end step, with an opponent having
// lost 2 or more life this turn, the intervening-if holds and the quest-counter
// trigger goes on the stack; without the life loss it does not.
func TestBloodchiefEndStepTriggerFiresOnOpponentEndStep(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:      game.EventBeginningOfStep,
		Step:       game.StepEnd,
		Controller: game.TriggerControllerAny,
	}, []game.Instruction{{
		Primitive: game.AddCounter{Amount: game.Fixed(1), Object: game.SourcePermanentReference(), CounterKind: counter.Quest},
		Optional:  true,
	}}, nil)
	card, ok := g.GetCardInstance(source.CardInstanceID)
	if !ok {
		t.Fatal("triggered permanent card instance not found")
	}
	card.Def.TriggeredAbilities[0].Trigger.InterveningCondition = anyOpponentLifeLostCondition(2)

	// An opponent lost 2 life this turn; it is a different opponent's end step.
	emitEvent(g, game.Event{Kind: game.EventLifeLost, Player: game.Player2, Amount: 2})
	g.Turn.ActivePlayer = game.Player3
	emitBeginningOfStepEvent(g, game.StepEnd)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("quest-counter trigger did not fire on an opponent's end step despite the opponent life loss")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want 1 after an opponent's end step with the condition met", got)
	}

	// Fresh turn with no opponent life loss: the intervening-if fails, so the
	// each-end-step trigger does not fire.
	g.Stack = game.Stack{}
	g.Events = nil
	g.TriggerEventCursor = 0
	g.Turn.ActivePlayer = game.Player4
	emitBeginningOfStepEvent(g, game.StepEnd)
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("quest-counter trigger fired with no opponent life loss this turn")
	}
}

// trigger: a card is put into an opponent's graveyard from anywhere. Because a
// token is not a card (CR 108.1), the subject carries a non-token requirement.
func bloodchiefGraveyardPattern() *game.TriggerPattern {
	return &game.TriggerPattern{
		Event:            game.EventZoneChanged,
		Player:           game.TriggerPlayerOpponent,
		MatchToZone:      true,
		ToZone:           zone.Graveyard,
		SubjectSelection: game.Selection{NonToken: true},
	}
}

// TestBloodchiefGraveyardTriggerFiresForOpponentCard proves a real card put into
// an opponent's graveyard fires the trigger, and that each card put in produces
// its own trigger (a mill or batch of multiple cards yields multiple triggers).
func TestBloodchiefGraveyardTriggerFiresForOpponentCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, bloodchiefGraveyardPattern(),
		[]game.Instruction{{Primitive: game.LoseLife{Amount: game.Fixed(2), Player: game.EventPlayerReference()}}}, nil)

	first := addCombatPermanent(g, game.Player2, &game.CardDef{
		CardFace: game.CardFace{Name: "Opponent Card One", Types: []types.Card{types.Creature}},
	})
	second := addCombatPermanent(g, game.Player2, &game.CardDef{
		CardFace: game.CardFace{Name: "Opponent Card Two", Types: []types.Card{types.Creature}},
	})
	destroyPermanent(g, first.ObjectID)
	destroyPermanent(g, second.ObjectID)

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("graveyard trigger did not fire for cards entering an opponent's graveyard")
	}
	if got := g.Stack.Size(); got != 2 {
		t.Fatalf("stack size = %d, want 2 (one trigger per card put into the graveyard)", got)
	}
}

// TestBloodchiefGraveyardTriggerIgnoresToken proves a token put into an
// opponent's graveyard does not fire the trigger: a token is not a card, so the
// non-token subject requirement excludes it (matching Bloodchief Ascension and
// The Haunt of Hightower, whose second ability reads "a card").
func TestBloodchiefGraveyardTriggerIgnoresToken(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, bloodchiefGraveyardPattern(),
		[]game.Instruction{{Primitive: game.LoseLife{Amount: game.Fixed(2), Player: game.EventPlayerReference()}}}, nil)

	token, ok := createTokenPermanent(g, game.Player2, &game.CardDef{
		CardFace: game.CardFace{Name: "Zombie Token", Types: []types.Card{types.Creature}},
	})
	if !ok {
		t.Fatal("token permanent was not created")
	}
	destroyPermanent(g, token.ObjectID)

	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("graveyard trigger fired for a token, but a token is not a card")
	}
	if got := g.Stack.Size(); got != 0 {
		t.Fatalf("stack size = %d, want 0 (a token must not trigger a \"a card\" graveyard trigger)", got)
	}
}

// bloodchiefDrainDef builds Bloodchief Ascension's complete second ability:
// "Whenever a card is put into an opponent's graveyard from anywhere, if this
// enchantment has three or more quest counters on it, you may have that player
// lose 2 life. If you do, you gain 2 life." The intervening-if reads the source
// permanent's quest counters; the drain is an optional LoseLife on the graveyard
// owner ("that player" = the zone-change event player) that publishes its result,
// and the GainLife is gated on the loss actually happening.
func bloodchiefDrainDef() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Bloodchief Ascension",
			Types: []types.Card{types.Enchantment},
			TriggeredAbilities: []game.TriggeredAbility{{
				Trigger: game.TriggerCondition{
					Type:          game.TriggerWhenever,
					Pattern:       *bloodchiefGraveyardPattern(),
					InterveningIf: "if this enchantment has three or more quest counters on it",
					InterveningCondition: opt.Val(game.Condition{
						Object: opt.Val(game.SourcePermanentReference()),
						ObjectMatches: opt.Val(game.Selection{
							RequiredCounterCount: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 3}),
							RequiredCounter:      counter.Quest,
						}),
					}),
				},
				Content: game.Mode{Sequence: []game.Instruction{
					{
						Primitive:     game.LoseLife{Amount: game.Fixed(2), Player: game.EventPlayerReference()},
						Optional:      true,
						PublishResult: game.ResultKey("may-have-action"),
					},
					{
						Primitive:  game.GainLife{Amount: game.Fixed(2), Player: game.ControllerReference()},
						ResultGate: opt.Val(game.InstructionResultGate{Key: "may-have-action", Succeeded: game.TriTrue}),
					},
				}}.Ability(),
			}},
		},
	}
}

// bloodchiefDrainSetup builds a Player1 Bloodchief Ascension with questCounters
// quest counters, sends a Player2-owned card into Player2's graveyard, and puts
// the resulting drain trigger on the stack. It returns the source permanent (so
// a test can mutate its counters before resolution) and the engine.
func bloodchiefDrainSetup(t *testing.T, questCounters int) (*game.Game, *Engine, *game.Permanent) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, bloodchiefDrainDef())
	if questCounters > 0 {
		source.Counters.Add(counter.Quest, questCounters)
	}
	victim := addCombatPermanent(g, game.Player2, &game.CardDef{
		CardFace: game.CardFace{Name: "Opponent Creature", Types: []types.Card{types.Creature}},
	})
	destroyPermanent(g, victim.ObjectID)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("drain trigger did not fire for a card entering an opponent's graveyard with three quest counters")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want 1 (one drain trigger)", got)
	}
	return g, engine, source
}

// TestBloodchiefDrainAcceptedDrainsAndGains proves the end-to-end resolution the
// stack-only tests never exercised: when the controller accepts the optional
// drain, "that player" (the opponent whose graveyard received the card, carried
// by the zone-change event's player) loses 2 life and the controller gains 2.
// This is the regression guard for the EventPlayerReference/EventZoneChanged bug:
// before triggeringEventPlayer handled EventZoneChanged, "that player" resolved
// to no one, so no life changed and the gain gate stayed off.
func TestBloodchiefDrainAcceptedDrainsAndGains(t *testing.T) {
	g, engine, _ := bloodchiefDrainSetup(t, 3)
	controllerLife := g.Players[game.Player1].Life
	opponentLife := g.Players[game.Player2].Life

	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}}}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := g.Players[game.Player2].Life; got != opponentLife-2 {
		t.Fatalf("opponent life = %d, want %d (that player must lose 2)", got, opponentLife-2)
	}
	if got := g.Players[game.Player1].Life; got != controllerLife+2 {
		t.Fatalf("controller life = %d, want %d (controller must gain 2 after the drain)", got, controllerLife+2)
	}
}

// TestBloodchiefDrainDeclinedDoesNothing proves the optional gate: declining the
// "you may have that player lose 2 life" leaves both players' life totals
// unchanged, because the gain is gated on the loss actually happening.
func TestBloodchiefDrainDeclinedDoesNothing(t *testing.T) {
	g, engine, _ := bloodchiefDrainSetup(t, 3)
	controllerLife := g.Players[game.Player1].Life
	opponentLife := g.Players[game.Player2].Life

	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}}}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := g.Players[game.Player2].Life; got != opponentLife {
		t.Fatalf("opponent life = %d, want %d (declining must not drain)", got, opponentLife)
	}
	if got := g.Players[game.Player1].Life; got != controllerLife {
		t.Fatalf("controller life = %d, want %d (declining must not gain)", got, controllerLife)
	}
}

// TestBloodchiefDrainLifeLockedFailsAndNoGain proves the "if you do" result gate:
// when the opponent's life total can't change, accepting the drain still fails to
// remove any life, so the loss did not happen and the controller does not gain.
func TestBloodchiefDrainLifeLockedFailsAndNoGain(t *testing.T) {
	g, engine, _ := bloodchiefDrainSetup(t, 3)
	controllerLife := g.Players[game.Player1].Life
	opponentLife := g.Players[game.Player2].Life
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		ID:             g.IDGen.Next(),
		Kind:           game.RuleEffectLifeTotalCantChange,
		Controller:     game.Player2,
		AffectedPlayer: game.PlayerYou,
	})

	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}}}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := g.Players[game.Player2].Life; got != opponentLife {
		t.Fatalf("opponent life = %d, want %d (a locked life total cannot be drained)", got, opponentLife)
	}
	if got := g.Players[game.Player1].Life; got != controllerLife {
		t.Fatalf("controller life = %d, want %d (no gain when the drain removed no life)", got, controllerLife)
	}
}

// TestBloodchiefDrainCounterThresholdRecheck proves the intervening-if is checked
// again on resolution (CR 603.4): a trigger that went on the stack while the
// source had three quest counters does nothing if the counters drop below three
// before it resolves.
func TestBloodchiefDrainCounterThresholdRecheck(t *testing.T) {
	g, engine, source := bloodchiefDrainSetup(t, 3)
	controllerLife := g.Players[game.Player1].Life
	opponentLife := g.Players[game.Player2].Life

	// Remove a counter so only two remain: below the three-counter threshold.
	source.Counters.Remove(counter.Quest, 1)

	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}}}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := g.Players[game.Player2].Life; got != opponentLife {
		t.Fatalf("opponent life = %d, want %d (resolution recheck must fail below three counters)", got, opponentLife)
	}
	if got := g.Players[game.Player1].Life; got != controllerLife {
		t.Fatalf("controller life = %d, want %d (no drain, so no gain)", got, controllerLife)
	}
}

// TestBloodchiefDrainBelowThresholdNeverTriggers proves the intervening-if also
// gates the trigger itself: with only two quest counters, the drain trigger never
// goes on the stack when a card enters an opponent's graveyard.
func TestBloodchiefDrainBelowThresholdNeverTriggers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, bloodchiefDrainDef())
	source.Counters.Add(counter.Quest, 2)
	victim := addCombatPermanent(g, game.Player2, &game.CardDef{
		CardFace: game.CardFace{Name: "Opponent Creature", Types: []types.Card{types.Creature}},
	})
	destroyPermanent(g, victim.ObjectID)

	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("drain trigger fired with only two quest counters (below the three-counter threshold)")
	}
	if got := g.Stack.Size(); got != 0 {
		t.Fatalf("stack size = %d, want 0 (intervening-if must gate the trigger)", got)
	}
}

// TestTriggeringEventPlayerResolvesZoneChangeOwner is the direct reference-resolver
// guard for the EventZoneChanged fix: "that player" (EventPlayerReference) on a
// captured zone-change trigger must resolve to the event's player — the owner of
// the destination graveyard. Before EventZoneChanged was added to
// triggeringEventPlayer this returned (0, false), leaving Bloodchief's drain
// bound to no player.
func TestTriggeringEventPlayerResolvesZoneChangeOwner(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obj := &game.StackObject{
		Kind:            game.StackTriggeredAbility,
		Controller:      game.Player1,
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:   game.EventZoneChanged,
			Player: game.Player2,
			ToZone: zone.Graveyard,
		},
	}

	got, ok := resolvePlayerReference(g, obj, game.EventPlayerReference())
	if !ok {
		t.Fatal("EventPlayerReference did not resolve for a zone-change trigger event")
	}
	if got != game.Player2 {
		t.Fatalf("event player = %d, want %d (the graveyard owner)", got, game.Player2)
	}

	// The same helper reads the event player directly, independent of the
	// resolver, for every zone destination (here a hand move).
	handMove := game.Event{Kind: game.EventZoneChanged, Player: game.Player3, ToZone: zone.Hand}
	if player, ok := triggeringEventPlayer(handMove); !ok || player != game.Player3 {
		t.Fatalf("triggeringEventPlayer(hand move) = (%d, %v), want (%d, true)", player, ok, game.Player3)
	}
}
