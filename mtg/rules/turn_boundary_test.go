package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

type upkeepManaBoundaryPolicy struct {
	activated int
	before    []int
	draw      []int
}

func (p *upkeepManaBoundaryPolicy) ChooseAction(obs PlayerObservation, legal []action.Action) action.Action {
	if obs.Turn.Step == game.StepDraw {
		p.draw = append(p.draw, obs.g.Players[obs.Player].ManaPool.Total())
	}
	if obs.Turn.Step == game.StepUpkeep && obs.g.Turn.UpkeepStepsThisTurn > p.activated {
		p.before = append(p.before, obs.g.Players[obs.Player].ManaPool.Total())
		p.activated = obs.g.Turn.UpkeepStepsThisTurn
		for _, candidate := range legal {
			if _, ok := candidate.ActivateAbilityPayload(); ok {
				return candidate
			}
		}
	}
	return action.Pass()
}

func TestTurnBoundaryUpkeepManaExpiresBeforeExtraUpkeepAndDraw(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	g.Turn.ExtraUpkeepSteps = 1
	for range 2 {
		addCardToHand(g, game.Player1, spiritGuideHandCard("Boundary Spirit Guide", mana.R))
	}
	stockPlayForwardLibrary(g, game.Player1, 2)
	policy := &upkeepManaBoundaryPolicy{}
	policies := simPassPolicies()
	policies[game.Player1] = policy
	NewEngine(nil).runBeginningPhase(g, policies, &TurnLog{})
	if len(policy.before) != 2 || len(policy.draw) == 0 {
		t.Fatal("fixture did not execute two real upkeep activations and draw priority")
	}
	for _, amount := range append(policy.before, policy.draw...) {
		if amount != 0 {
			t.Fatalf("ordinary mana survived a real upkeep boundary: %v / %v", policy.before, policy.draw)
		}
	}
}

type mainManaBoundaryPolicy struct {
	amounts []int
	spell   id.ID
	funded  bool
}

func (p *mainManaBoundaryPolicy) ChooseAction(obs PlayerObservation, legal []action.Action) action.Action {
	if obs.Turn.Step == game.StepBeginningOfCombat || obs.Turn.Step == game.StepEnd {
		p.amounts = append(p.amounts, obs.g.Players[obs.Player].ManaPool.Total())
		for _, candidate := range legal {
			if cast, ok := candidate.CastSpellPayload(); ok && cast.CardID == p.spell {
				p.funded = true
			}
		}
	}
	if len(legal) > 0 {
		if _, declares := legal[0].DeclareAttackersPayload(); declares {
			return legal[len(legal)-1]
		}
	}
	return action.Pass()
}

func TestTurnBoundarySimulatorCompletesResolvedMainLikeLive(t *testing.T) {
	for _, phase := range []game.Phase{game.PhasePrecombatMain, game.PhasePostcombatMain} {
		for _, extra := range []bool{false, true} {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			g.Turn.ActivePlayer = game.Player1
			g.Turn.Phase = phase
			g.Turn.Step = game.StepNone
			g.Players[game.Player1].ManaPool.Add(mana.R, 3)
			if extra {
				g.Turn.ExtraPhases = []game.Phase{game.PhaseCombat}
			}
			engine := NewEngine(nil)
			probe := &mainManaBoundaryPolicy{}
			probe.spell = addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
				Name: "Three mana boundary spell", Types: []types.Card{types.Instant},
				ManaCost: opt.Val(cost.Mana{cost.O(2), cost.R}),
				SpellAbility: opt.Val(game.Mode{
					Sequence: []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
				}.Ability()),
			}})
			policies := simPassPolicies()
			policies[game.Player1] = probe
			stable := engine.Simulator().ResolvePriority(g, policies)
			if stable.Players[game.Player1].ManaPool.Total() != 3 {
				t.Fatal("ResolvePriority crossed the current window's boundary")
			}
			engine.Simulator().PlayForward(stable, policies, 0)
			if len(probe.amounts) == 0 {
				t.Fatal("fixture never reached a later priority window")
			}
			for _, amount := range probe.amounts {
				if amount != 0 {
					t.Fatalf("Simulator carried main-phase mana into a later window: phase=%v extra=%v amounts=%v", phase, extra, probe.amounts)
				}
				if probe.funded {
					t.Fatal("stale main-phase mana funded a spell in a later priority window")
				}
			}
			live := g.Clone()
			probe.amounts = nil
			probe.funded = false
			engine.runMainPhase(live, policies, phase, &TurnLog{})
			engine.runTurnFromPhase(live, policies, &TurnLog{}, phase+1)
			for _, amount := range probe.amounts {
				if amount != 0 {
					t.Fatal("live and resumed main completion disagreed")
				}
			}
			if g.Players[game.Player1].ManaPool.Total() != 3 || stable.Players[game.Player1].ManaPool.Total() != 3 {
				t.Fatal("Simulator mutated an input mana pool")
			}
		}
	}
}

func TestTurnBoundaryRetainedManaKeepsItsRider(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	unit := mana.Unit{Color: mana.R, Snow: true, FromCreature: true, Persistent: true}
	g.Players[game.Player1].ManaPool.AddPersistentUnit(unit, 1)
	g.Players[game.Player1].ManaRiders = []game.ManaRiderInstance{{Unit: unit, SourceID: id.ID(100)}}
	emptyManaPools(g)
	if g.Players[game.Player1].ManaPool.Total() != 1 || len(g.Players[game.Player1].ManaRiders) != 1 {
		t.Fatal("retained exact mana lost its spend rider at a boundary")
	}
}

func TestTurnBoundaryOrdinaryRiderDoesNotAttachToRetainedMana(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	ordinary := mana.Unit{Color: mana.R}
	retained := mana.Unit{Color: mana.R, Persistent: true}
	player := g.Players[game.Player1]
	player.ManaPool.AddUnit(ordinary, 1)
	player.ManaPool.AddPersistentUnit(retained, 1)
	player.ManaRiders = []game.ManaRiderInstance{
		{Unit: ordinary, SourceID: id.ID(100)},
		{Unit: retained, SourceID: id.ID(200)},
	}
	emptyManaPools(g)
	if len(player.ManaRiders) != 1 || player.ManaRiders[0].SourceID != id.ID(200) {
		t.Fatal("an ordinary mana rider was reassigned to unrelated retained mana")
	}
}

type cleanupManaBoundaryPolicy struct {
	spell   id.ID
	seen    []int
	cast    bool
	decider game.PlayerID
}

func (p *cleanupManaBoundaryPolicy) ChooseAction(obs PlayerObservation, legal []action.Action) action.Action {
	if obs.Turn.Step == game.StepCleanup {
		p.seen = append(p.seen, obs.g.Players[obs.Player].ManaPool.Total())
		p.decider = obs.DecisionMaker
		for _, candidate := range legal {
			if cast, ok := candidate.CastSpellPayload(); ok && cast.CardID == p.spell && !p.cast {
				p.cast = true
				return candidate
			}
		}
	}
	return action.Pass()
}

func TestTurnBoundaryCleanupRetainedManaCanBeSpentBeforeStepEnds(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	g.Turn.TurnNumber = 2
	g.PlayerControlEffects = []game.PlayerControlEffect{{ID: id.ID(100), Player: game.Player1, Controller: game.Player2, Turn: 2}}
	g.Players[game.Player1].ManaPool.AddPersistent(mana.R, 3)
	creature := addCombatPermanent(g, game.Player1, vanillaCreature("Cleanup casualty", 0, 0))
	creature.TemporaryToughnessModifier = 1
	spell := &game.CardDef{CardFace: game.CardFace{
		Name: "Cleanup mana spell", Types: []types.Card{types.Instant},
		ManaCost: opt.Val(cost.Mana{cost.O(2), cost.R}),
		SpellAbility: opt.Val(game.Mode{
			Sequence: []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(2), Player: game.ControllerReference()}}},
		}.Ability()),
	}}
	probe := &cleanupManaBoundaryPolicy{spell: addCardToHand(g, game.Player1, spell)}
	policies := simPassPolicies()
	policies[game.Player2] = probe
	NewEngine(nil).runEndingPhaseWithLog(g, policies, &TurnLog{})
	if !probe.cast || len(probe.seen) == 0 || probe.seen[0] != 3 || probe.decider != game.Player2 {
		t.Fatal("cleanup lost retained mana early or failed controlled-player casting")
	}
	if g.Players[game.Player1].Life != 42 || g.Players[game.Player1].ManaPool.Total() != 0 ||
		g.Players[game.Player1].ManaPool.Spent() != 3 {
		t.Fatal("cleanup payment, expiry or spent accounting disagreed")
	}
}

func TestTurnBoundarySimulatorRejectsUnsupportedResume(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.Phase = game.PhaseCombat
	defer func() {
		if recover() == nil {
			t.Fatal("Simulator silently skipped an unsupported combat resume")
		}
	}()
	NewEngine(nil).Simulator().PlayForward(g, simPassPolicies(), 0)
}

func TestTurnBoundarySimulatorCombatSharesPhaseExit(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID: g.IDGen.Next(), AffectedObjectID: attacker.ObjectID, Layer: game.LayerAbility,
		AddKeywords: []game.Keyword{game.Flying}, Duration: game.DurationUntilEndOfCombat,
	})
	declare := action.DeclareAttackersAction{Attackers: []game.AttackDeclaration{
		{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
	}}
	resolved, ok := NewEngine(nil).Simulator().ResolveCombatWithAttackers(g, game.Player1, declare, simPassPolicies())
	if !ok {
		t.Fatal("could not resolve the real attack")
	}
	if resolved.Combat != nil || len(resolved.ContinuousEffects) != 0 {
		t.Fatal("Simulator combat skipped its phase-exit obligations")
	}
	if g.Combat == nil || len(g.ContinuousEffects) != 1 {
		t.Fatal("Simulator phase teardown changed the original game")
	}
}

func TestTurnBoundarySkippedDrawHasNoEntryAction(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	g.SkippedSteps[game.Player1] = map[game.Step]int{game.StepDraw: 1}
	addCardToHand(g, game.Player1, spiritGuideHandCard("Skipped draw guide", mana.R))
	stockPlayForwardLibrary(g, game.Player1, 2)
	probe := &upkeepManaBoundaryPolicy{}
	policies := simPassPolicies()
	policies[game.Player1] = probe
	NewEngine(nil).runBeginningPhase(g, policies, &TurnLog{})
	if len(probe.before) != 1 || len(probe.draw) != 0 || g.Players[game.Player1].Library.Size() != 2 {
		t.Fatal("a skipped draw step performed its entry or priority actions")
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("skipping draw skipped the preceding real boundary")
	}
}
