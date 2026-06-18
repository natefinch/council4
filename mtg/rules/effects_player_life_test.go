package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestDrawEffectDrawsRequestedCards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sourceID := addEffectSpellToStack(g, game.Player1, game.Draw{
		Amount: game.Fixed(2),
		Player: game.ControllerReference(),
	}, nil)
	firstDraw := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "First"}})
	secondDraw := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Second"}})
	log := TurnLog{}

	engine.resolveTopOfStack(g, &log)

	if !g.Players[game.Player1].Hand.Contains(firstDraw) {
		t.Fatal("first card was not drawn")
	}
	if !g.Players[game.Player1].Hand.Contains(secondDraw) {
		t.Fatal("second card was not drawn")
	}
	if len(log.Draws) != 2 {
		t.Fatalf("draw logs = %d, want 2", len(log.Draws))
	}
	if log.Resolves[0].SourceID != sourceID {
		t.Fatalf("resolve source = %v, want %v", log.Resolves[0].SourceID, sourceID)
	}
}

func TestGainLifeEffectIncreasesTargetLife(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addEffectSpellToStack(g, game.Player1, game.GainLife{
		Amount: game.Fixed(3),
		Player: game.TargetPlayerReference(0),
	}, []game.Target{game.PlayerTarget(game.Player2)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player2].Life != 43 {
		t.Fatalf("player 2 life = %d, want 43", g.Players[game.Player2].Life)
	}
}

func TestGainLifeGroupEffectAffectsAllOpponents(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addEffectSpellToStack(g, game.Player1, game.GainLife{
		Amount:      game.Fixed(2),
		PlayerGroup: game.OpponentsReference(),
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	for _, playerID := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		if got := g.Players[playerID].Life; got != 42 {
			t.Fatalf("player %d life = %d, want 42", playerID, got)
		}
	}
	if g.Players[game.Player1].Life != 40 {
		t.Fatalf("controller life changed unexpectedly: %d", g.Players[game.Player1].Life)
	}
}

func TestApplyContinuousGroupSnapshotsPermanentsAtResolution(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	affected := makeCreaturePermanent(g, game.Player1, "Affected")
	opponent := makeCreaturePermanent(g, game.Player2, "Opponent")
	addEffectSpellToStack(g, game.Player1, game.ApplyContinuous{
		ContinuousEffects: []game.ContinuousEffect{{
			Layer: game.LayerPowerToughnessModify,
			Group: game.BattlefieldGroup(game.Selection{
				RequiredTypes: []types.Card{types.Creature},
				Controller:    game.ControllerYou,
			}),
			PowerDelta:     1,
			ToughnessDelta: 1,
		}},
		Duration: game.DurationUntilEndOfTurn,
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := effectivePower(g, affected); got != 3 {
		t.Fatalf("affected creature power = %d, want 3", got)
	}
	if got := effectivePower(g, opponent); got != 2 {
		t.Fatalf("opponent creature power = %d, want 2", got)
	}
	affected.Controller = game.Player2
	if got := effectivePower(g, affected); got != 3 {
		t.Fatalf("affected creature power after control change = %d, want 3", got)
	}
	later := makeCreaturePermanent(g, game.Player1, "Later")
	if got := effectivePower(g, later); got != 2 {
		t.Fatalf("later creature power = %d, want 2", got)
	}
	if len(g.ContinuousEffects) != 1 ||
		g.ContinuousEffects[0].AffectedObjectID != affected.ObjectID ||
		g.ContinuousEffects[0].Group.Valid() {
		t.Fatalf("continuous effects = %+v, want one snapshotted object effect", g.ContinuousEffects)
	}
}

func TestLoseLifeGroupEffectAffectsAllOpponents(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addEffectSpellToStack(g, game.Player1, game.LoseLife{
		Amount:      game.Fixed(3),
		PlayerGroup: game.OpponentsReference(),
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	for _, playerID := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		if got := g.Players[playerID].Life; got != 37 {
			t.Fatalf("player %d life = %d, want 37", playerID, got)
		}
	}
	if g.Players[game.Player1].Life != 40 {
		t.Fatalf("controller life changed unexpectedly: %d", g.Players[game.Player1].Life)
	}
}

func TestLinkedLoseLifeGroupGainAmountIsTotal(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// 3 opponents each lose 1 life; controller gains the total (3).
	addInstructionSpellToStack(g, []game.Instruction{
		{
			Primitive:     game.LoseLife{PlayerGroup: game.OpponentsReference(), Amount: game.Fixed(1)},
			PublishResult: "life-change",
		},
		{
			Primitive: game.GainLife{
				Player: game.ControllerReference(),
				Amount: game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountPreviousEffectResult, ResultKey: "life-change"}),
			},
		},
	})

	engine.resolveTopOfStack(g, &TurnLog{})

	for _, playerID := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		if got := g.Players[playerID].Life; got != 39 {
			t.Fatalf("opponent %d life = %d, want 39", playerID, got)
		}
	}
	if got := g.Players[game.Player1].Life; got != 43 {
		t.Fatalf("controller life = %d, want 43 (gained total 3)", got)
	}
}

func TestCantGainLifeBlocksGroupGainLifePerPlayer(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// CantGainLife blocks all players.
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "No Lifegain",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:           game.RuleEffectCantGainLife,
				AffectedPlayer: game.PlayerAny,
			}},
		}}},
	})
	addEffectSpellToStack(g, game.Player1, game.GainLife{
		Amount:      game.Fixed(5),
		PlayerGroup: game.AllPlayersReference(),
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	for _, playerID := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
		if got := g.Players[playerID].Life; got != 40 {
			t.Fatalf("player %d life = %d, want gain prevented (40)", playerID, got)
		}
	}
}

func TestAddPlayerCounterEffectUpdatesCountersAndEvents(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addEffectSpellToStack(g, game.Player1, game.AddPlayerCounter{
		Amount:      game.Fixed(3),
		Player:      game.TargetPlayerReference(0),
		CounterKind: counter.Poison,
	}, []game.Target{game.PlayerTarget(game.Player2)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player2].PoisonCounters; got != 3 {
		t.Fatalf("poison counters = %d, want 3", got)
	}
	var event game.Event
	for _, candidate := range g.Events {
		if candidate.Kind == game.EventCountersAdded {
			event = candidate
		}
	}
	if event.Kind != game.EventCountersAdded ||
		event.Player != game.Player2 ||
		event.CounterKind != counter.Poison ||
		event.PreviousCounterAmount != 0 ||
		event.Amount != 3 {
		t.Fatalf("counter event = %+v", event)
	}
}

func TestAddPlayerCounterEffectDynamicAndNonpositiveAmounts(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addEffectSpellToStack(g, game.Player1, game.AddPlayerCounter{
		Amount: game.Dynamic(game.DynamicAmount{
			Kind:       game.DynamicAmountOpponentCount,
			Multiplier: 2,
		}),
		Player:      game.TargetPlayerReference(0),
		CounterKind: counter.Energy,
	}, []game.Target{game.PlayerTarget(game.Player2)})
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player2].EnergyCounters; got != 6 {
		t.Fatalf("energy counters = %d, want 6", got)
	}

	for _, amount := range []int{0, -1} {
		addEffectSpellToStack(g, game.Player1, game.AddPlayerCounter{
			Amount:      game.Fixed(amount),
			Player:      game.TargetPlayerReference(0),
			CounterKind: counter.Experience,
		}, []game.Target{game.PlayerTarget(game.Player2)})
		engine.resolveTopOfStack(g, &TurnLog{})
	}
	if got := g.Players[game.Player2].ExperienceCounters; got != 0 {
		t.Fatalf("experience counters = %d, want 0", got)
	}
}

func TestAddPlayerPoisonCounterCausesStateBasedLoss(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Players[game.Player2].PoisonCounters = 9
	engine := NewEngine(nil)
	addEffectSpellToStack(g, game.Player1, game.AddPlayerCounter{
		Amount:      game.Fixed(1),
		Player:      game.TargetPlayerReference(0),
		CounterKind: counter.Poison,
	}, []game.Target{game.PlayerTarget(game.Player2)})
	engine.resolveTopOfStack(g, &TurnLog{})

	losses := engine.applyStateBasedActions(g)
	if !g.Players[game.Player2].Eliminated ||
		len(losses) != 1 ||
		losses[0].Reason != LossReasonPoisonCounters {
		t.Fatalf("player = %+v, losses = %+v", g.Players[game.Player2], losses)
	}
}

// TestDrainDrawLoseLifeSequenceResolvesInOrder proves the lowered drain spell
// "Target player draws two cards and loses 2 life" resolves both instructions,
// in Oracle order, against the single chosen target player: the target draws
// two cards and then loses 2 life.
func TestDrainDrawLoseLifeSequenceResolvesInOrder(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sourceID := addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{Primitive: game.Draw{Amount: game.Fixed(2), Player: game.TargetPlayerReference(0)}},
		{Primitive: game.LoseLife{Amount: game.Fixed(2), Player: game.TargetPlayerReference(0)}},
	}, []game.Target{game.PlayerTarget(game.Player2)})
	first := addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "First"}})
	second := addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Second"}})
	log := TurnLog{}

	engine.resolveTopOfStack(g, &log)

	if !g.Players[game.Player2].Hand.Contains(first) || !g.Players[game.Player2].Hand.Contains(second) {
		t.Fatal("target player did not draw both cards")
	}
	if g.Players[game.Player2].Life != 38 {
		t.Fatalf("target player life = %d, want 38 (40 - 2)", g.Players[game.Player2].Life)
	}
	if g.Players[game.Player1].Life != 40 {
		t.Fatalf("controller life changed unexpectedly: %d", g.Players[game.Player1].Life)
	}
	if len(log.Draws) != 2 {
		t.Fatalf("draw logs = %d, want 2", len(log.Draws))
	}
	if len(log.Resolves) == 0 || log.Resolves[0].SourceID != sourceID {
		t.Fatalf("resolve source = %#v, want %v", log.Resolves, sourceID)
	}
}
