package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: Halfdane (oracle 2ee13155-a3b8-4cad-8e68-5d2c5aba3bb4,
// printing cac1e57a-0ef5-442b-a492-059fc6ff3dfc) from the pinned corpus;
// https://api.scryfall.com/cards/cac1e57a-0ef5-442b-a492-059fc6ff3dfc/rulings
// returned one WotC ruling (2009-10-01): when the next trigger fails to resolve,
// the previous effect lasts through that upkeep, then Halfdane reverts to 3/3.
// CR 2026-11-06 verified copy: 500.5 (until end of a step expires as the step
// ends), 503.1 (upkeep), 608.2b (all targets illegal), 611.2c-d (values fixed at
// resolution), 613.4b-c and 613.7 (layer 7b set by timestamp, 7c counters).
// The card stays unregistered here; this file proves the shared duration with
// its own text.
func halfdaneNextUpkeepTestCard() *game.CardDef {
	pt := game.PT{Value: 3}
	return &game.CardDef{CardFace: game.CardFace{
		Name:       "Halfdane",
		ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.W, cost.U, cost.B}),
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Shapeshifter},
		Power:      opt.Val(pt),
		Toughness:  opt.Val(pt),
		TriggeredAbilities: []game.TriggeredAbility{{
			Text: "At the beginning of your upkeep, Halfdane's base power and toughness become equal to the power and toughness of target creature other than Halfdane until the end of your next upkeep.",
			Trigger: game.TriggerCondition{
				Type: game.TriggerAt,
				Pattern: game.TriggerPattern{
					Event:      game.EventBeginningOfStep,
					Controller: game.TriggerControllerYou,
					Step:       game.StepUpkeep,
				},
			},
			Content: game.Mode{
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target creature other than Halfdane",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludeSource: true}),
				}},
				Sequence: []game.Instruction{{Primitive: game.ApplyContinuous{
					Object:   opt.Val(game.SourcePermanentReference()),
					Duration: game.DurationUntilEndOfYourNextUpkeep,
					ContinuousEffects: []game.ContinuousEffect{{
						Layer: game.LayerPowerToughnessSet,
						SetPowerDynamic: opt.Val(game.DynamicAmount{
							Kind:   game.DynamicAmountTargetPower,
							Object: game.TargetPermanentReference(0),
						}),
						SetToughnessDynamic: opt.Val(game.DynamicAmount{
							Kind:   game.DynamicAmountTargetToughness,
							Object: game.TargetPermanentReference(0),
						}),
					}},
				}}},
			}.Ability(),
		}},
	}}
}

func halfdaneTestCreature(name string, power, toughness int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: power}),
		Toughness: opt.Val(game.PT{Value: toughness}),
	}}
}

func halfdaneTestRemoval() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Halfdane Test Removal",
		ManaCost: opt.Val(cost.Mana{cost.O(0)}),
		Types:    []types.Card{types.Instant},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Constraint: "target creature",
				Allow:      game.TargetAllowPermanent,
				Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
			}},
			Sequence: []game.Instruction{{Primitive: game.Destroy{Object: game.TargetPermanentReference(0)}}},
		}.Ability()),
	}}
}

type halfdaneObservation struct {
	turn             int
	step             game.Step
	triggerOnStack   bool
	power, toughness int
}

// halfdaneScenario is shared by every seat's agent. Player1 aims Halfdane's
// trigger at target; Player2 casts removal at removalTarget in response to it.
type halfdaneScenario struct {
	g             *game.Game
	halfdane      id.ID
	target        game.Target
	removal       id.ID
	removalTarget game.Target
	observations  []halfdaneObservation
}

func (s *halfdaneScenario) halfdanePT(t *testing.T) (int, int) {
	t.Helper()
	permanent, ok := permanentByObjectID(s.g, s.halfdane)
	if !ok {
		t.Fatal("Halfdane left the battlefield")
	}
	toughness, _ := effectiveToughness(s.g, permanent)
	return effectivePower(s.g, permanent), toughness
}

func (s *halfdaneScenario) triggerOnStack() bool {
	top, ok := s.g.Stack.Peek()
	return ok && top.Kind == game.StackTriggeredAbility && top.SourceID == s.halfdane
}

// upkeep returns Player1's upkeep priority observations on turn: the first is
// taken with the new trigger on the stack, the last after the stack emptied.
func (s *halfdaneScenario) upkeep(t *testing.T, turn int) (first, last halfdaneObservation) {
	t.Helper()
	var matching []halfdaneObservation
	for _, observation := range s.observations {
		if observation.turn == turn && observation.step == game.StepUpkeep {
			matching = append(matching, observation)
		}
	}
	if len(matching) == 0 {
		t.Fatalf("no upkeep priority observed on turn %d", turn)
	}
	return matching[0], matching[len(matching)-1]
}

func (s *halfdaneScenario) agents() [game.NumPlayers]PlayerAgent {
	var agents [game.NumPlayers]PlayerAgent
	for seat := range agents {
		agents[seat] = halfdaneAgent{scenario: s, seat: game.PlayerID(seat)}
	}
	return agents
}

type halfdaneAgent struct {
	scenario *halfdaneScenario
	seat     game.PlayerID
}

func (a halfdaneAgent) ChooseAction(_ PlayerObservation, actions []action.Action) action.Action {
	s := a.scenario
	if permanent, ok := permanentByObjectID(s.g, s.halfdane); ok {
		toughness, _ := effectiveToughness(s.g, permanent)
		s.observations = append(s.observations, halfdaneObservation{
			turn:           s.g.Turn.TurnNumber,
			step:           s.g.Turn.Step,
			triggerOnStack: s.triggerOnStack(),
			power:          effectivePower(s.g, permanent),
			toughness:      toughness,
		})
	}
	if a.seat == game.Player2 && s.removal != 0 && s.triggerOnStack() {
		for _, candidate := range actions {
			payload, ok := candidate.CastSpellPayload()
			if ok && payload.CardID == s.removal &&
				(len(payload.Targets) == 0 || slices.Contains(payload.Targets, s.removalTarget)) {
				s.removal = 0
				return candidate
			}
		}
	}
	return action.Pass()
}

func (a halfdaneAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceTarget {
		want := a.scenario.target
		if a.seat == game.Player2 {
			want = a.scenario.removalTarget
		}
		for _, option := range request.Options {
			if slices.Contains(option.Targets, want) {
				return []int{option.Index}
			}
		}
	}
	return request.DefaultSelection
}

func newHalfdaneScenario(g *game.Game, halfdane *game.Permanent) *halfdaneScenario {
	for _, player := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
		stockLibrary(g, player, 12)
	}
	return &halfdaneScenario{g: g, halfdane: halfdane.ObjectID}
}

func countNextUpkeepEffects(g *game.Game) int {
	count := 0
	for i := range g.ContinuousEffects {
		if g.ContinuousEffects[i].Duration == game.DurationUntilEndOfYourNextUpkeep {
			count++
		}
	}
	return count
}

func wantHalfdanePT(t *testing.T, label string, power, toughness, wantPower, wantToughness int) {
	t.Helper()
	if power != wantPower || toughness != wantToughness {
		t.Fatalf("%s: Halfdane is %d/%d, want %d/%d", label, power, toughness, wantPower, wantToughness)
	}
}

// TestHalfdaneBasePTLastsUntilEndOfNextUpkeep plays Halfdane's upkeep trigger
// at a 5/4 with a +1/+1 counter. Halfdane is 6/5 for the rest of the turn and
// through every opponent's turn, unaffected by later changes to that creature,
// and during Player1's next upkeep until the new trigger (aimed at a 2/2)
// resolves; the newer 2/2 effect wins by timestamp and survives as the first
// effect ends with that upkeep.
func TestHalfdaneBasePTLastsUntilEndOfNextUpkeep(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	halfdane := addCombatPermanent(g, game.Player1, halfdaneNextUpkeepTestCard())
	copied := addCombatPermanent(g, game.Player2, halfdaneTestCreature("Copied Brute", 5, 4))
	addCountersToPermanent(g, copied, counter.PlusOnePlusOne, 1)
	small := addCombatPermanent(g, game.Player3, halfdaneTestCreature("Small Bear", 2, 2))
	s := newHalfdaneScenario(g, halfdane)
	s.target = game.PermanentTarget(copied.ObjectID)
	agents := s.agents()

	engine.runTurn(g, agents)
	first, last := s.upkeep(t, 1)
	if !first.triggerOnStack || last.triggerOnStack {
		t.Fatalf("turn 1 upkeep observations = %+v / %+v, want trigger then empty stack", first, last)
	}
	wantHalfdanePT(t, "before the first trigger resolves", first.power, first.toughness, 3, 3)
	wantHalfdanePT(t, "after the first trigger resolves", last.power, last.toughness, 6, 5)
	if count := countNextUpkeepEffects(g); count != 1 {
		t.Fatalf("effects after turn 1 = %d, want 1 surviving the creating upkeep and cleanup", count)
	}
	power, toughness := s.halfdanePT(t)
	wantHalfdanePT(t, "after Player1's cleanup", power, toughness, 6, 5)

	// CR 611.2d: the copied values were locked when the trigger resolved.
	removeCountersFromPermanent(g, game.Player2, copied, counter.PlusOnePlusOne, 1)
	addCountersToPermanent(g, copied, counter.PlusOnePlusOne, 3)
	for turn := 2; turn <= 4; turn++ {
		engine.runTurn(g, agents)
		power, toughness = s.halfdanePT(t)
		wantHalfdanePT(t, "after an opponent's turn", power, toughness, 6, 5)
	}
	if g.Turn.ActivePlayer != game.Player1 || g.Turn.TurnNumber != 5 {
		t.Fatalf("turn = %d for %v, want Player1's turn 5", g.Turn.TurnNumber, g.Turn.ActivePlayer)
	}

	s.target = game.PermanentTarget(small.ObjectID)
	engine.runTurn(g, agents)
	first, last = s.upkeep(t, 5)
	if !first.triggerOnStack {
		t.Fatal("Player1's next upkeep did not put Halfdane's trigger on the stack")
	}
	wantHalfdanePT(t, "next upkeep before the new trigger resolves", first.power, first.toughness, 6, 5)
	wantHalfdanePT(t, "next upkeep after the 2/2 copy", last.power, last.toughness, 2, 2)
	for _, observation := range s.observations {
		if observation.turn == 5 && observation.step != game.StepUpkeep && (observation.power != 2 || observation.toughness != 2) {
			t.Fatalf("turn 5 %v: Halfdane is %d/%d after the first effect ended, want 2/2",
				observation.step, observation.power, observation.toughness)
		}
	}
	if count := countNextUpkeepEffects(g); count != 1 {
		t.Fatalf("effects after Player1's next upkeep = %d, want only the newer effect", count)
	}
	power, toughness = s.halfdanePT(t)
	wantHalfdanePT(t, "after turn 5", power, toughness, 2, 2)
}

// TestHalfdaneFizzledTriggerKeepsOldEffectThroughUpkeep follows the 2009-10-01
// ruling: when the next trigger's only target is removed in response, the old
// effect lasts through the rest of that upkeep and then Halfdane reverts to its
// printed 3/3. Its own +1/+1 counter applies over either base (layer 7c).
func TestHalfdaneFizzledTriggerKeepsOldEffectThroughUpkeep(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	halfdane := addCombatPermanent(g, game.Player1, halfdaneNextUpkeepTestCard())
	addCountersToPermanent(g, halfdane, counter.PlusOnePlusOne, 1)
	copied := addCombatPermanent(g, game.Player2, halfdaneTestCreature("Copied Brute", 5, 4))
	doomed := addCombatPermanent(g, game.Player3, halfdaneTestCreature("Doomed Bear", 2, 2))
	s := newHalfdaneScenario(g, halfdane)
	s.target = game.PermanentTarget(copied.ObjectID)
	agents := s.agents()

	for range 4 {
		engine.runTurn(g, agents)
	}
	power, toughness := s.halfdanePT(t)
	wantHalfdanePT(t, "before Player1's next turn", power, toughness, 6, 5)

	s.target = game.PermanentTarget(doomed.ObjectID)
	s.removal = addCardToHand(g, game.Player2, halfdaneTestRemoval())
	s.removalTarget = game.PermanentTarget(doomed.ObjectID)
	engine.runTurn(g, agents)

	if s.removal != 0 {
		t.Fatal("Player2 never cast removal in response to Halfdane's trigger")
	}
	if _, ok := permanentByObjectID(g, doomed.ObjectID); ok {
		t.Fatal("the trigger's target was not removed in response")
	}
	first, last := s.upkeep(t, 5)
	if !first.triggerOnStack || last.triggerOnStack {
		t.Fatalf("turn 5 upkeep observations = %+v / %+v, want trigger then empty stack", first, last)
	}
	wantHalfdanePT(t, "upkeep after the trigger did nothing", last.power, last.toughness, 6, 5)
	for _, observation := range s.observations {
		if observation.turn == 5 && observation.step != game.StepUpkeep && (observation.power != 4 || observation.toughness != 4) {
			t.Fatalf("turn 5 %v: Halfdane is %d/%d, want printed 3/3 plus its counter",
				observation.step, observation.power, observation.toughness)
		}
	}
	if count := countNextUpkeepEffects(g); count != 0 {
		t.Fatalf("effects after the fizzled upkeep = %d, want 0", count)
	}
	power, toughness = s.halfdanePT(t)
	wantHalfdanePT(t, "after turn 5", power, toughness, 4, 4)
}

// TestHalfdaneNextUpkeepDurationSurvivesClone clones the game during an
// opponent's turn. The clone carries the creation point and expires the
// effect as Player1's next upkeep ends, exactly as the original does.
func TestHalfdaneNextUpkeepDurationSurvivesClone(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	halfdane := addCombatPermanent(g, game.Player1, halfdaneNextUpkeepTestCard())
	copied := addCombatPermanent(g, game.Player2, halfdaneTestCreature("Copied Brute", 5, 4))
	s := newHalfdaneScenario(g, halfdane)
	s.target = game.PermanentTarget(copied.ObjectID)
	agents := s.agents()
	engine.runTurn(g, agents)
	engine.runTurn(g, agents)

	clone := g.Clone()
	cloned := &halfdaneScenario{g: clone, halfdane: s.halfdane}
	for _, run := range []*halfdaneScenario{s, cloned} {
		// Destroy the only legal target so neither game creates a newer effect.
		target, ok := permanentByObjectID(run.g, copied.ObjectID)
		if !ok {
			t.Fatal("copied creature is missing")
		}
		addEffectSpellToStack(run.g, game.Player2, game.Destroy{Object: game.TargetPermanentReference(0)},
			[]game.Target{game.PermanentTarget(target.ObjectID)})
		engine.resolveTopOfStack(run.g, &TurnLog{})
		runAgents := run.agents()
		for run.g.Turn.TurnNumber < 6 {
			engine.runTurn(run.g, runAgents)
		}
		first, last := run.upkeep(t, 5)
		if first.triggerOnStack {
			t.Fatal("Halfdane's trigger went on the stack without a legal target")
		}
		wantHalfdanePT(t, "during Player1's next upkeep", last.power, last.toughness, 5, 4)
		if count := countNextUpkeepEffects(run.g); count != 0 {
			t.Fatalf("effects after Player1's next upkeep = %d, want 0", count)
		}
		power, toughness := run.halfdanePT(t)
		wantHalfdanePT(t, "after Player1's next upkeep", power, toughness, 3, 3)
	}
	if !slices.Equal(s.observations[len(s.observations)-len(cloned.observations):], cloned.observations) {
		t.Fatal("clone and original observed different Halfdane values after the clone point")
	}
}
