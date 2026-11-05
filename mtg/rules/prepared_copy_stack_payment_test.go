package rules

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: Engine gap #3634. Default-registry cards checked against the
// pinned corpus 6da9ced0…a1b4 (oracle_id): Carnivorous Cultivator // Enroot
// db655c5a-06f6-42bc-8180-d9ca1c1d939a "This creature enters prepared." // {G}
// sorcery "Search your library for a land card, put it into your graveyard,
// then shuffle."; Thalia, Guardian of Thraben
// 9b7f1d05-707c-4ed3-9f0e-8ced1232c2ee "Noncreature spells cost {1} more to
// cast."; Waterlogged Grove 70fa2eba-565e-4fed-adc9-7f5d9fcbf1fa "{T}, Pay 1
// life: Add {G} or {U}. …"; Nurturing Peatland
// 8ed932ff-986c-4592-ad70-53b3fac80d69 "{T}, Pay 1 life: Add {B} or {G}. …";
// Young Pyromancer 5fac139a-07d3-4e6c-98e3-d98b199f7a6f "Whenever you cast an
// instant or sorcery spell, create a 1/1 red Elemental creature token.";
// Gitaxian Probe 1d67f5ff-1fce-45e5-b6a1-416c569351e2 {U/P} "Look at target
// player's hand. Draw a card."; Mental Misstep
// 1a0770e6-b093-4439-baff-6889a50ba12e {U/P} instant "Counter target spell
// with mana value 1."; Lightning Bolt 4457ed35-7c10-48c8-9776-456485fdf070
// {R}. The Probe and Misstep spells are used as the spell faces of test
// prepared tokens.
// Scryfall rulings for Enroot (2026-08-21): the prepared creature's controller
// creates the copy in exile and "may cast that copy as long as it remains in
// exile. As they cast it, that creature stops being prepared." Comprehensive
// Rules effective 2026-09-25 (verified official copy, retained under
// .cardwork/cr-20260925.txt): 722.3c the copy is created in exile, may be cast
// while it remains there, and the permanent "loses the prepared designation at
// the time the spell becomes cast (see rule 601.2i)"; 601.2a proposing a cast
// first moves "that card (or that copy of a card) from where it is to the
// stack. It becomes the topmost object on the stack."; 601.2b-c Phyrexian
// payments are announced and then targets chosen, after that move; 115.5 "A
// spell or ability on the stack is an illegal target for itself."; 601.2f-h determine,
// lock and pay the total cost, with mana abilities activated before paying;
// 601.2i the spell becomes cast and cast triggers trigger only then; 733.1 an
// action that can't be completed is reversed, the spell returns to the zone it
// came from, no abilities trigger, and mana abilities activated while making
// it may be reversed; 733.2 the player keeps priority and may redo it.

// preparedStackWindowAgent pays like priorityCastWindowAgent and, each time it
// is offered a payment-time mana ability, records what it observes: the stack,
// whether the prepared permanent is still prepared, and the spell-cast events
// so far.
type preparedStackWindowAgent struct {
	*priorityCastWindowAgent
	g         *game.Game
	permanent *game.Permanent
	stacks    [][]StackObjectView
	prepared  []bool
	castsSeen []int
}

func (a *preparedStackWindowAgent) ChooseChoice(obs PlayerObservation, request game.ChoiceRequest) []int {
	if request.Prompt == "Activate a mana ability while paying?" {
		a.stacks = append(a.stacks, obs.Stack())
		a.prepared = append(a.prepared, a.permanent.Prepared)
		a.castsSeen = append(a.castsSeen, spellCastEventCount(a.g))
	}
	return a.priorityCastWindowAgent.ChooseChoice(obs, request)
}

// preparedEnrootPaymentGame has Player1's prepared Carnivorous Cultivator,
// Young Pyromancer, Waterlogged Grove and Nurturing Peatland, with Player2's
// Thalia making Enroot's locked total {1}{G} and no mana in the pool.
func preparedEnrootPaymentGame(t *testing.T) (*game.Game, *Engine, id.ID, *game.Permanent, *game.Permanent) {
	t.Helper()
	g, e, cardID, permanent := preparedExileCastGame(t, "Carnivorous Cultivator", mana.G, 2)
	pyromancer := addCombatPermanent(g, game.Player1, registryCardDef(t, "Young Pyromancer"))
	addCombatPermanent(g, game.Player2, registryCardDef(t, "Thalia, Guardian of Thraben"))
	addCombatPermanent(g, game.Player1, registryCardDef(t, "Waterlogged Grove"))
	addCombatPermanent(g, game.Player1, registryCardDef(t, "Nurturing Peatland"))
	g.Players[game.Player1].Life = 20
	return g, e, cardID, permanent, pyromancer
}

// requireEnrootOnStackWhilePaying checks that every payment window the agent
// saw had the Enroot copy alone on the stack as the topmost object, before the
// spell became cast: its permanent still prepared and no spell-cast event yet.
func requireEnrootOnStackWhilePaying(t *testing.T, agent *preparedStackWindowAgent, castsBefore int) id.ID {
	t.Helper()
	return requirePreparedCopyOnStackWhilePaying(t, agent.stacks, agent.prepared, agent.castsSeen, "Enroot", castsBefore)
}

func requirePreparedCopyOnStackWhilePaying(t *testing.T, stacks [][]StackObjectView, prepared []bool, castsSeen []int, name string, castsBefore int) id.ID {
	t.Helper()
	if len(stacks) == 0 {
		t.Fatal("no payment-time mana window was offered")
	}
	var copyID id.ID
	for i, stack := range stacks {
		if len(stack) != 1 || stack[0].Kind != game.StackSpell || stack[0].Name != name || stack[0].Controller != game.Player1 {
			t.Fatalf("window %d stack = %+v, want Player1's %s copy as the only stack object", i, stack, name)
		}
		if copyID != 0 && stack[0].ID != copyID {
			t.Fatalf("window %d stack object = %v, want the same %s copy %v throughout payment", i, stack[0].ID, name, copyID)
		}
		copyID = stack[0].ID
		if !prepared[i] || castsSeen[i] != castsBefore {
			t.Fatalf("window %d prepared = %t, cast events = %d; want the permanent prepared and no cast event (%d) before the spell becomes cast",
				i, prepared[i], castsSeen[i], castsBefore)
		}
	}
	return copyID
}

// TestPreparedCopyOnStackWhilePaying casts Enroot's prepared copy with a locked
// {1}{G} total and no mana in the pool, paying it by activating Waterlogged
// Grove and Nurturing Peatland while paying (CR 601.2g). Throughout payment the
// copy is already the topmost stack object (CR 601.2a) and the Cultivator is
// still prepared, with no spell-cast event (CR 722.3c, 601.2i). Once paid, that
// same object is the cast spell, the Cultivator is unprepared, exactly one cast
// event names it, and Young Pyromancer triggers once.
func TestPreparedCopyOnStackWhilePaying(t *testing.T) {
	t.Parallel()
	g, e, cardID, permanent, pyromancer := preparedEnrootPaymentGame(t)
	castsBefore := spellCastEventCount(g)
	cast := action.CastSpellFaceFromZone(cardID, zone.Battlefield, game.FaceAlternate, nil, 0, nil)
	agent := &preparedStackWindowAgent{priorityCastWindowAgent: &priorityCastWindowAgent{limit: -1}, g: g, permanent: permanent}

	outcome, _ := e.attemptActionWithChoices(g, game.Player1, cast, reboundKickerAgents(agent), &TurnLog{}, nil)

	if outcome != actionApplied || agent.activated != 2 {
		t.Fatalf("outcome = %v, activated = %d; want the cast applied with both lands", outcome, agent.activated)
	}
	copyID := requireEnrootOnStackWhilePaying(t, agent, castsBefore)
	obj, ok := g.Stack.Peek()
	if !ok || g.Stack.Size() != 1 || obj.ID != copyID || !obj.Copy || !obj.CastCopy || obj.SourceCardID != cardID ||
		obj.SourceZone != zone.Exile || obj.ManaSpentToCast != 2 {
		t.Fatalf("stack top = %+v (size %d), want the copy seen while paying (%v) cast from exile with two mana",
			obj, g.Stack.Size(), copyID)
	}
	if permanent.Prepared {
		t.Fatal("Carnivorous Cultivator is still prepared after its copy became cast")
	}
	var castEvents []game.Event
	for _, event := range g.Events {
		if event.Kind == game.EventSpellCast && event.StackObjectID == copyID {
			castEvents = append(castEvents, event)
		}
	}
	if len(castEvents) != 1 || spellCastEventCount(g) != castsBefore+1 || castEvents[0].FromZone != zone.Exile ||
		castEvents[0].PermanentID != permanent.ObjectID {
		t.Fatalf("cast events = %+v, want exactly one for the copy, from exile and linked to the Cultivator", castEvents)
	}
	e.putTriggeredAbilitiesOnStack(g)
	if got := triggeredAbilitiesFrom(g, pyromancer.ObjectID); got != 1 {
		t.Fatalf("Young Pyromancer triggers = %d, want 1 for the cast sorcery", got)
	}
}

// TestReversedPreparedCopyLeavesStack offers Enroot's prepared copy because
// both lands could pay its locked {1}{G}, but leaves the total unpaid while
// paying: activating only Waterlogged Grove, or declining every window. The
// copy was on the stack while paying, and the unpaid cast is reversed (CR
// 733.1): the copy leaves the stack, the Cultivator stays prepared, the
// Grove's activation is undone, no cast event is emitted and Young Pyromancer
// doesn't trigger. The player keeps priority and a redo paying with both lands
// casts it (CR 733.2).
func TestReversedPreparedCopyLeavesStack(t *testing.T) {
	t.Parallel()
	for _, limit := range []int{1, 0} {
		t.Run(fmt.Sprintf("activating %d", limit), func(t *testing.T) {
			t.Parallel()
			g, e, cardID, permanent, pyromancer := preparedEnrootPaymentGame(t)
			castsBefore := spellCastEventCount(g)
			life := g.Players[game.Player1].Life
			cast := action.CastSpellFaceFromZone(cardID, zone.Battlefield, game.FaceAlternate, nil, 0, nil)
			attempts := &castAttempts{}

			agent := &preparedStackWindowAgent{priorityCastWindowAgent: &priorityCastWindowAgent{limit: limit}, g: g, permanent: permanent}
			outcome, _ := e.attemptActionWithChoices(g, game.Player1, cast, reboundKickerAgents(agent), &TurnLog{}, attempts)
			if outcome != actionReversed || agent.activated != limit {
				t.Fatalf("outcome = %v, activated = %d; want the cast reversed", outcome, agent.activated)
			}
			requireEnrootOnStackWhilePaying(t, agent, castsBefore)
			if !g.Stack.IsEmpty() || !permanent.Prepared || spellCastEventCount(g) != castsBefore {
				t.Fatalf("stack = %d, prepared = %t, cast events = %d; want the copy off the stack, the Cultivator prepared and no cast",
					g.Stack.Size(), permanent.Prepared, spellCastEventCount(g))
			}
			requireLandsUnused(t, g, life, "Waterlogged Grove", "Nurturing Peatland")
			e.putTriggeredAbilitiesOnStack(g)
			if got := triggeredAbilitiesFrom(g, pyromancer.ObjectID); got != 0 || !g.Stack.IsEmpty() {
				t.Fatalf("Young Pyromancer triggers = %d, stack = %d; want no trigger from a reversed cast", got, g.Stack.Size())
			}

			agent = &preparedStackWindowAgent{priorityCastWindowAgent: &priorityCastWindowAgent{limit: -1}, g: g, permanent: permanent}
			outcome, _ = e.attemptActionWithChoices(g, game.Player1, cast, reboundKickerAgents(agent), &TurnLog{}, attempts)
			if outcome != actionApplied || agent.activated != 2 {
				t.Fatalf("redo outcome = %v, activated = %d; want the cast applied with both lands", outcome, agent.activated)
			}
			copyID := requireEnrootOnStackWhilePaying(t, agent, castsBefore)
			if obj, ok := g.Stack.Peek(); !ok || g.Stack.Size() != 1 || obj.ID != copyID || permanent.Prepared ||
				spellCastEventCount(g) != castsBefore+1 || g.Players[game.Player1].Life != life-2 {
				t.Fatalf("stack top = %+v (size %d), prepared = %t, life = %d; want only the redone copy cast and paid",
					obj, g.Stack.Size(), permanent.Prepared, g.Players[game.Player1].Life)
			}
			e.putTriggeredAbilitiesOnStack(g)
			if got := triggeredAbilitiesFrom(g, pyromancer.ObjectID); got != 1 {
				t.Fatalf("Young Pyromancer triggers = %d, want 1 for the redone cast", got)
			}
		})
	}
}

// preparedPhyrexianWindowAgent announces mana for a Phyrexian symbol, activates
// Waterlogged Grove for {U} while paying, and records the stack and the
// prepared token at the Phyrexian announcement, the target prompt and each
// payment window.
type preparedPhyrexianWindowAgent struct {
	*phyrexianWindowAgent
	g         *game.Game
	token     *game.Permanent
	stacks    [][]StackObjectView
	prepared  []bool
	castsSeen []int
}

func (a *preparedPhyrexianWindowAgent) ChooseChoice(obs PlayerObservation, request game.ChoiceRequest) []int {
	phyrexian := request.Kind == game.ChoicePayment && strings.HasPrefix(request.Prompt, "Pay {") && strings.HasSuffix(request.Prompt, "/P}")
	if phyrexian || request.Prompt == "Choose spell targets." || request.Prompt == "Activate a mana ability while paying?" {
		a.stacks = append(a.stacks, obs.Stack())
		a.prepared = append(a.prepared, a.token.Prepared)
		a.castsSeen = append(a.castsSeen, spellCastEventCount(a.g))
	}
	for _, option := range request.Options {
		if option.Label == "U" {
			return []int{option.Index}
		}
	}
	return a.phyrexianWindowAgent.ChooseChoice(obs, request)
}

// TestPreparedCopyPhyrexianAnnouncementOnStackWhilePaying casts the spell copy
// of a prepared token whose spell face is Gitaxian Probe's {U/P} "Look at
// target player's hand. Draw a card.", announcing {U} for the {U/P} before
// choosing Player2 as its target (CR 601.2b-c). The {U} is paid by activating
// Waterlogged Grove while paying (CR 601.2g). At the announcement, the target
// prompt and the payment window the same copy is already the topmost stack
// object (CR 601.2a), the token still prepared and the spell not yet cast (CR
// 722.3c, 601.2i). That same object becomes the cast spell.
func TestPreparedCopyPhyrexianAnnouncementOnStackWhilePaying(t *testing.T) {
	t.Parallel()
	e, g := newEffectiveCascadeGame()
	token := phyrexianPreparedToken(t, g)
	addCombatPermanent(g, game.Player1, registryCardDef(t, "Waterlogged Grove"))
	life := g.Players[game.Player1].Life
	castsBefore := spellCastEventCount(g)
	agent := &preparedPhyrexianWindowAgent{
		phyrexianWindowAgent: &phyrexianWindowAgent{payMana: true, ability: "Pay 1 life"},
		g:                    g, token: token,
	}
	cast := action.CastSpellAction{CardID: token.ObjectID, SourceZone: zone.Battlefield, Face: game.FaceAlternate,
		Targets: []game.Target{game.PlayerTarget(game.Player2)}}

	outcome, _ := e.attemptCastSpellWithChoices(g, game.Player1, cast, reboundKickerAgents(agent), &TurnLog{}, nil)

	if outcome != actionApplied || agent.activated != 1 ||
		!slices.Equal(agent.steps, []string{"phyrexian", "target", "window"}) {
		t.Fatalf("outcome = %v, activated = %d, steps = %v; want {U} announced before the target and paid by the Grove",
			outcome, agent.activated, agent.steps)
	}
	copyID := requirePreparedCopyOnStackWhilePaying(t, agent.stacks, agent.prepared, agent.castsSeen, "Prepared Spell", castsBefore)
	obj, ok := g.Stack.Peek()
	if !ok || g.Stack.Size() != 1 || obj.ID != copyID || !obj.CastCopy || obj.SourceID != token.ObjectID ||
		!slices.Equal(obj.Targets, []game.Target{game.PlayerTarget(game.Player2)}) || obj.ManaSpentToCast != 1 ||
		token.Prepared || spellCastEventCount(g) != castsBefore+1 || g.Players[game.Player1].Life != life-1 {
		t.Fatalf("stack top = %+v (size %d), prepared = %t, life = %d; want the copy seen while paying cast at Player2 with {U}",
			obj, g.Stack.Size(), token.Prepared, g.Players[game.Player1].Life)
	}
}

// TestReversedPreparedCopyAnnouncedOnStack announces {U} for the {U/P} of a
// prepared token's Gitaxian Probe copy and chooses Player2 as its target with
// the copy already on the stack (CR 601.2a-c), then declines to activate
// Waterlogged Grove, leaving the announced {U} unpaid. The cast is reversed
// (CR 733.1): the copy seen during announcement leaves the stack, the token
// stays prepared, no life is paid and no cast event is emitted. The player
// keeps priority and a redo casts a new copy from the same token (CR 733.2).
func TestReversedPreparedCopyAnnouncedOnStack(t *testing.T) {
	t.Parallel()
	e, g := newEffectiveCascadeGame()
	token := phyrexianPreparedToken(t, g)
	addCombatPermanent(g, game.Player1, registryCardDef(t, "Waterlogged Grove"))
	life := g.Players[game.Player1].Life
	castsBefore := spellCastEventCount(g)
	cast := action.CastSpellAction{CardID: token.ObjectID, SourceZone: zone.Battlefield, Face: game.FaceAlternate,
		Targets: []game.Target{game.PlayerTarget(game.Player2)}}
	attempts := &castAttempts{}
	agent := &preparedPhyrexianWindowAgent{phyrexianWindowAgent: &phyrexianWindowAgent{payMana: true}, g: g, token: token}

	outcome, _ := e.attemptCastSpellWithChoices(g, game.Player1, cast, reboundKickerAgents(agent), &TurnLog{}, attempts)

	if outcome != actionReversed || agent.activated != 0 ||
		!slices.Equal(agent.steps, []string{"phyrexian", "target", "window"}) {
		t.Fatalf("outcome = %v, activated = %d, steps = %v; want {U} announced, the target chosen and the cast reversed unpaid",
			outcome, agent.activated, agent.steps)
	}
	requirePreparedCopyOnStackWhilePaying(t, agent.stacks, agent.prepared, agent.castsSeen, "Prepared Spell", castsBefore)
	if !g.Stack.IsEmpty() || !token.Prepared || spellCastEventCount(g) != castsBefore || g.Players[game.Player1].Life != life {
		t.Fatalf("stack = %d, prepared = %t, cast events = %d, life = %d; want the copy off the stack, the token prepared and nothing paid",
			g.Stack.Size(), token.Prepared, spellCastEventCount(g), g.Players[game.Player1].Life)
	}
	requireLandsUnused(t, g, life, "Waterlogged Grove")

	agent = &preparedPhyrexianWindowAgent{phyrexianWindowAgent: &phyrexianWindowAgent{payLife: true}, g: g, token: token}
	outcome, _ = e.attemptCastSpellWithChoices(g, game.Player1, cast, reboundKickerAgents(agent), &TurnLog{}, attempts)
	if outcome != actionApplied {
		t.Fatalf("redo outcome = %v, want the cast applied paying 2 life", outcome)
	}
	copyID := requirePreparedCopyOnStackWhilePaying(t, agent.stacks, agent.prepared, agent.castsSeen, "Prepared Spell", castsBefore)
	if obj, ok := g.Stack.Peek(); !ok || g.Stack.Size() != 1 || obj.ID != copyID || token.Prepared ||
		spellCastEventCount(g) != castsBefore+1 || g.Players[game.Player1].Life != life-2 {
		t.Fatalf("stack top = %+v (size %d), prepared = %t, life = %d; want only the redone copy cast for 2 life",
			obj, g.Stack.Size(), token.Prepared, g.Players[game.Player1].Life)
	}
}

// cloningTargetAgent chooses like announcedTargetAgent and, at the spell
// target prompt, clones the game and records the clone's topmost stack object
// and the stack objects it offers as targets for spec.
type cloningTargetAgent struct {
	*announcedTargetAgent
	g          *game.Game
	def        *game.CardDef
	spec       game.TargetSpec
	cloneTop   *game.StackObject
	candidates map[id.ID]bool
}

func (a *cloningTargetAgent) ChooseChoice(obs PlayerObservation, request game.ChoiceRequest) []int {
	if request.Prompt == "Choose spell targets." {
		clone := a.g.Clone()
		a.cloneTop, _ = clone.Stack.Peek()
		a.candidates = stackTargetCandidateIDs(clone, game.Player1, a.def, a.spec)
	}
	return a.announcedTargetAgent.ChooseChoice(obs, request)
}

// TestPreparedCopyNotItsOwnTarget casts the spell copy of a prepared token
// whose spell face is Mental Misstep's {U/P} instant "Counter target spell
// with mana value 1." while Player2's Lightning Bolt is on the stack. The copy
// is on the stack while its targets are chosen (CR 601.2a-c) and also has mana
// value 1, but a spell can't target itself (CR 115.5): only Lightning Bolt is
// offered, also in a clone of the game taken at the target prompt, and the
// copy is cast targeting it. Once cast, the copy is a legal target for another
// spell with mana value 1.
func TestPreparedCopyNotItsOwnTarget(t *testing.T) {
	t.Parallel()
	e, g := newEffectiveCascadeGame()
	token := preparedSpellToken(t, g, "Mental Misstep")
	spellFace := token.TokenDef.Alternate.Val
	spellFace.Types = []types.Card{types.Instant}
	token.TokenDef.Alternate = opt.Val(spellFace)
	bolt := pushFaceUpCardSpell(g, game.Player2, registryCardDef(t, "Lightning Bolt"))
	bolt.Targets = []game.Target{game.PlayerTarget(game.Player1)}
	bolt.TargetCounts = []int{1}
	life := g.Players[game.Player1].Life
	misstep := registryCardDef(t, "Mental Misstep")
	spec := misstep.SpellAbility.Val.Modes[0].Targets[0]
	agent := &cloningTargetAgent{
		announcedTargetAgent: &announcedTargetAgent{phyrexianWindowAgent: &phyrexianWindowAgent{payLife: true}},
		g:                    g, def: misstep, spec: spec,
	}
	cast := action.CastSpellAction{CardID: token.ObjectID, SourceZone: zone.Battlefield, Face: game.FaceAlternate,
		Targets: []game.Target{game.StackObjectTarget(bolt.ID)}}

	outcome, _ := e.attemptCastSpellWithChoices(g, game.Player1, cast, reboundKickerAgents(agent), &TurnLog{}, nil)

	if outcome != actionApplied || !slices.Equal(agent.steps, []string{"phyrexian", "target"}) {
		t.Fatalf("outcome = %v, steps = %v; want 2 life announced before the target", outcome, agent.steps)
	}
	if len(agent.offered) != 1 || !slices.Equal(agent.offered[0], []game.Target{game.StackObjectTarget(bolt.ID)}) {
		t.Fatalf("offered targets = %v, want only Lightning Bolt %v", agent.offered, bolt.ID)
	}
	obj, ok := g.Stack.Peek()
	if !ok || g.Stack.Size() != 2 || !obj.CastCopy || obj.Casting || token.Prepared ||
		!slices.Equal(obj.Targets, []game.Target{game.StackObjectTarget(bolt.ID)}) || g.Players[game.Player1].Life != life-2 {
		t.Fatalf("stack top = %+v (size %d), prepared = %t, life = %d; want the copy cast at Lightning Bolt for 2 life",
			obj, g.Stack.Size(), token.Prepared, g.Players[game.Player1].Life)
	}
	if agent.cloneTop == nil || agent.cloneTop.ID != obj.ID || !agent.cloneTop.Casting ||
		!maps.Equal(agent.candidates, map[id.ID]bool{bolt.ID: true}) {
		t.Fatalf("clone at the target prompt: top = %+v, candidates = %v; want the casting copy %v and only Lightning Bolt",
			agent.cloneTop, agent.candidates, obj.ID)
	}
	if got := stackTargetCandidateIDs(g, game.Player2, misstep, spec); !got[obj.ID] {
		t.Fatalf("candidates after the cast = %v, want the cast copy %v", got, obj.ID)
	}
}
