package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/opt"
)

// Source review: pinned corpus Oracle text for Flusterstorm (oracle
// 86bf58f2-7f25-4e10-b797-25e0e8e67769, IMA 55) and Grapeshot (oracle
// ebd2d760-5ad8-4027-b124-171822f3edfe); Scryfall rulings for both returned
// the same five WotC rulings dated 2022-12-08: countered spells and spells cast
// from other zones are counted, copies may each choose new targets, copies are
// not cast or counted, the storm trigger can be countered (no copies), and
// countering the storm spell does not affect the copies. CR effective
// 2025-09-19: 702.40a/b (storm is a cast trigger; instances trigger
// separately), 707.10 and 707.10c (copies are not cast; new targets are chosen
// before each copy is put on the stack), 608.2h (last known information), 603.3b.

// stormRetargetAgent answers each "choose new targets" prompt for a storm copy
// with the next target in want, recording the prompts it was asked.
type stormRetargetAgent struct {
	t        *testing.T
	want     []game.Target
	requests []game.ChoiceRequest
}

func (*stormRetargetAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *stormRetargetAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind != game.ChoiceTarget {
		return request.DefaultSelection
	}
	a.requests = append(a.requests, request)
	if len(a.want) == 0 {
		a.t.Fatalf("unexpected retarget prompt %+v", request)
	}
	next := a.want[0]
	a.want = a.want[1:]
	for _, option := range request.Options {
		if slices.Equal(option.Targets, []game.Target{next}) {
			return []int{option.Index}
		}
	}
	a.t.Fatalf("retarget prompt does not offer %+v: %+v", next, request.Options)
	return nil
}

func castInstantFor(t *testing.T, engine *Engine, g *game.Game, player game.PlayerID, name string) id.ID {
	t.Helper()
	cardID := addCardToHand(g, player, simpleGainLifeInstant(name))
	g.Turn.PriorityPlayer = player
	if !engine.applyAction(g, player, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatalf("%s cast failed", name)
	}
	return cardID
}

func stormCopiesOf(g *game.Game, sourceID id.ID) []*game.StackObject {
	var copies []*game.StackObject
	for _, obj := range g.Stack.Objects() {
		if obj.Kind == game.StackSpell && obj.Copy && obj.SourceID == sourceID {
			copies = append(copies, obj)
		}
	}
	return copies
}

// TestStormCountsSpellsCastBeforeItButNotCopiesOrResponses proves storm counts
// every player's spells cast earlier in the turn, but not storm copies (which
// are not cast) or spells cast in response to the storm trigger.
func TestStormCountsSpellsCastBeforeItButNotCopiesOrResponses(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	castInstantFor(t, engine, g, game.Player2, "Opponent Spell")
	engine.resolveTopOfStack(g, &TurnLog{})

	firstStorm := addCardToHand(g, game.Player1, stormGainLifeInstant())
	g.Turn.PriorityPlayer = game.Player1
	if !engine.applyAction(g, game.Player1, action.CastSpell(firstStorm, nil, 0, nil)) {
		t.Fatal("first storm cast failed")
	}
	resolveStormTrigger(t, engine, g, nil)
	if got := len(stormCopiesOf(g, firstStorm)); got != 1 {
		t.Fatalf("first storm copies = %d, want 1 for the opponent's spell", got)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	engine.resolveTopOfStack(g, &TurnLog{})

	secondStorm := addCardToHand(g, game.Player1, stormGainLifeInstant())
	if !engine.applyAction(g, game.Player1, action.CastSpell(secondStorm, nil, 0, nil)) {
		t.Fatal("second storm cast failed")
	}
	engine.putTriggeredAbilitiesOnStack(g)
	trigger, _ := g.Stack.Peek()
	if !isStormTrigger(trigger) || trigger.TriggerEvent.SpellsCastBeforeThisTurn != opt.Val(2) {
		t.Fatalf("stack top = %+v, want storm trigger counting two cast spells and no copy", trigger)
	}
	castInstantFor(t, engine, g, game.Player3, "Response Spell")
	engine.resolveTopOfStack(g, &TurnLog{})
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := len(stormCopiesOf(g, secondStorm)); got != 2 {
		t.Fatalf("second storm copies = %d, want 2; the response was cast after it", got)
	}
}

// TestStormTriggerCounteredCreatesNoCopies proves countering the storm trigger
// stops the copies while the storm spell itself still resolves.
func TestStormTriggerCounteredCreatesNoCopies(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	castInstantFor(t, engine, g, game.Player1, "First Spell")
	engine.resolveTopOfStack(g, &TurnLog{})
	stormID := addCardToHand(g, game.Player1, stormGainLifeInstant())
	if !engine.applyAction(g, game.Player1, action.CastSpell(stormID, nil, 0, nil)) {
		t.Fatal("storm cast failed")
	}
	engine.putTriggeredAbilitiesOnStack(g)
	trigger, _ := g.Stack.Peek()
	if !isStormTrigger(trigger) || !counterStackObject(g, trigger.ID) {
		t.Fatal("could not counter the storm trigger")
	}
	if g.Stack.Size() != 1 || len(stormCopiesOf(g, stormID)) != 0 || spellCopiedEventCount(g) != 0 {
		t.Fatal("countered storm trigger still copied the spell")
	}
	life := g.Players[game.Player1].Life
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player1].Life != life+1 || !g.Players[game.Player1].Graveyard.Contains(stormID) {
		t.Fatal("storm spell did not resolve after its trigger was countered")
	}
}

// TestStormCopiesCounteredSpellFromLastKnownInformation proves countering the
// storm spell in response does not stop its trigger: the real Grapeshot is
// copied from last known information, with its original target, in the game
// and in a clone taken while the trigger waits on the stack.
func TestStormCopiesCounteredSpellFromLastKnownInformation(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	castInstantFor(t, engine, g, game.Player1, "First Spell")
	engine.resolveTopOfStack(g, &TurnLog{})
	grapeshot := addCardToHand(g, game.Player1, registryCard(t, "Grapeshot"))
	g.Players[game.Player1].ManaPool.Add(mana.R, 2)
	if !engine.applyAction(g, game.Player1, action.CastSpell(grapeshot,
		[]game.Target{game.PlayerTarget(game.Player2)}, 0, nil)) {
		t.Fatal("Grapeshot cast failed")
	}
	original, _ := g.Stack.Peek()
	engine.putTriggeredAbilitiesOnStack(g)
	if top, _ := g.Stack.Peek(); !isStormTrigger(top) {
		t.Fatalf("stack top = %+v, want Grapeshot's storm trigger", top)
	}
	if !counterStackObject(g, original.ID) || !g.Players[game.Player1].Graveyard.Contains(grapeshot) {
		t.Fatal("could not counter Grapeshot in response to its storm trigger")
	}
	for _, branch := range []*game.Game{g, g.Clone()} {
		engine.resolveTopOfStack(branch, &TurnLog{})
		copies := stormCopiesOf(branch, grapeshot)
		if len(copies) != 1 || copies[0].Controller != game.Player1 ||
			!slices.Equal(copies[0].Targets, []game.Target{game.PlayerTarget(game.Player2)}) {
			t.Fatalf("storm copies = %+v, want one Grapeshot copy targeting Player2", copies)
		}
		engine.resolveTopOfStack(branch, &TurnLog{})
		if branch.Players[game.Player2].Life != 39 || branch.Stack.Size() != 0 ||
			!branch.Players[game.Player1].Graveyard.Contains(grapeshot) {
			t.Fatal("last-known-information storm copy did not deal its damage cleanly")
		}
	}
}

// TestStormInstancesTriggerSeparately proves each storm instance triggers and
// copies the spell on its own (CR 702.40b).
func TestStormInstancesTriggerSeparately(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	castInstantFor(t, engine, g, game.Player2, "Opponent Spell")
	engine.resolveTopOfStack(g, &TurnLog{})
	def := stormGainLifeInstant()
	def.StaticAbilities = append(def.StaticAbilities, game.StormStaticBody)
	stormID := addCardToHand(g, game.Player1, def)
	g.Turn.PriorityPlayer = game.Player1
	if !engine.applyAction(g, game.Player1, action.CastSpell(stormID, nil, 0, nil)) {
		t.Fatal("double-storm cast failed")
	}
	engine.putTriggeredAbilitiesOnStack(g)
	triggers := 0
	for _, obj := range g.Stack.Objects() {
		if isStormTrigger(obj) {
			triggers++
		}
	}
	if triggers != 2 {
		t.Fatalf("storm triggers = %d, want 2", triggers)
	}
	life := g.Players[game.Player1].Life
	for range 5 {
		engine.resolveTopOfStack(g, &TurnLog{})
	}
	if g.Stack.Size() != 0 || spellCopiedEventCount(g) != 2 || g.Players[game.Player1].Life != life+3 {
		t.Fatalf("stack=%d copies=%d life=%d, want two copies plus the original resolved",
			g.Stack.Size(), spellCopiedEventCount(g), g.Players[game.Player1].Life)
	}
}

// TestGrapeshotStormCopiesChooseNewTargetsIndependently proves the default
// registry Grapeshot's controller chooses a new target for each storm copy
// separately, while the original keeps its target.
func TestGrapeshotStormCopiesChooseNewTargetsIndependently(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	castInstantFor(t, engine, g, game.Player2, "Opponent Spell")
	engine.resolveTopOfStack(g, &TurnLog{})
	castInstantFor(t, engine, g, game.Player1, "Warmup Spell")
	engine.resolveTopOfStack(g, &TurnLog{})
	grapeshot := addCardToHand(g, game.Player1, registryCard(t, "Grapeshot"))
	g.Players[game.Player1].ManaPool.Add(mana.R, 2)
	if !engine.applyAction(g, game.Player1, action.CastSpell(grapeshot,
		[]game.Target{game.PlayerTarget(game.Player2)}, 0, nil)) {
		t.Fatal("Grapeshot cast failed")
	}
	agent := &stormRetargetAgent{t: t, want: []game.Target{
		game.PlayerTarget(game.Player3), game.PlayerTarget(game.Player4),
	}}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
	resolveStormTrigger(t, engine, g, &agents)
	if len(agent.requests) != 2 || agent.requests[0].Player != game.Player1 || agent.requests[1].Player != game.Player1 {
		t.Fatalf("retarget prompts = %+v, want one per copy for Grapeshot's controller", agent.requests)
	}
	copies := stormCopiesOf(g, grapeshot)
	if len(copies) != 2 ||
		!slices.Equal(copies[0].Targets, []game.Target{game.PlayerTarget(game.Player3)}) ||
		!slices.Equal(copies[1].Targets, []game.Target{game.PlayerTarget(game.Player4)}) {
		t.Fatalf("storm copies = %+v, want copies targeting Player3 and Player4", copies)
	}
	for range 3 {
		engine.resolveTopOfStack(g, &TurnLog{})
	}
	// Player1's warmup and Player2's spell each gained their caster 1 life.
	for player, want := range map[game.PlayerID]int{game.Player2: 40, game.Player3: 39, game.Player4: 39} {
		if g.Players[player].Life != want {
			t.Fatalf("player %v life = %d, want %d", player, g.Players[player].Life, want)
		}
	}
	if g.Players[game.Player1].Life != 41 || g.Stack.Size() != 0 {
		t.Fatal("Grapeshot or its copies hit the wrong player")
	}
}

// TestFlusterstormStormCopiesRetargetOtherSpells proves the default registry
// Flusterstorm's storm copies may target other spells, including another copy's
// choice and the original Flusterstorm itself.
func TestFlusterstormStormCopiesRetargetOtherSpells(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	castInstantFor(t, engine, g, game.Player2, "Opponent Spell A")
	spellA, _ := g.Stack.Peek()
	castInstantFor(t, engine, g, game.Player2, "Opponent Spell B")
	spellB, _ := g.Stack.Peek()
	flusterstorm := addCardToHand(g, game.Player1, registryCard(t, "Flusterstorm"))
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	g.Turn.PriorityPlayer = game.Player1
	if !engine.applyAction(g, game.Player1, action.CastSpell(flusterstorm,
		[]game.Target{game.StackObjectTarget(spellA.ID)}, 0, nil)) {
		t.Fatal("Flusterstorm cast failed")
	}
	original, _ := g.Stack.Peek()
	agent := &stormRetargetAgent{t: t, want: []game.Target{
		game.StackObjectTarget(spellB.ID), game.StackObjectTarget(original.ID),
	}}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
	resolveStormTrigger(t, engine, g, &agents)
	copies := stormCopiesOf(g, flusterstorm)
	if len(copies) != 2 ||
		!slices.Equal(copies[0].Targets, []game.Target{game.StackObjectTarget(spellB.ID)}) ||
		!slices.Equal(copies[1].Targets, []game.Target{game.StackObjectTarget(original.ID)}) ||
		!slices.Equal(original.Targets, []game.Target{game.StackObjectTarget(spellA.ID)}) {
		t.Fatalf("Flusterstorm copies = %+v, want copies targeting spell B and the original", copies)
	}
	// The top copy counters the original Flusterstorm, then the other copy
	// counters spell B; neither controller has mana to pay {1}.
	for range 2 {
		engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	}
	if g.Stack.Size() != 1 || g.Stack.Objects()[0].ID != spellA.ID ||
		!g.Players[game.Player1].Graveyard.Contains(flusterstorm) ||
		!g.Players[game.Player2].Graveyard.Contains(spellB.SourceID) {
		t.Fatalf("stack = %+v, want only spell A left after the copies countered spell B and Flusterstorm", g.Stack.Objects())
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player2].Life != 41 {
		t.Fatalf("Player2 life = %d, want 41 from spell A only", g.Players[game.Player2].Life)
	}
}
