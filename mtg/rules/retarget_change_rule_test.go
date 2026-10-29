package rules

// Source review: Engine gap #3382. CR 2026-09-25 (MagicCompRules 20260925.txt,
// "effective as of September 25, 2026", verified local copy) 115.7/115.7a
// ("change the target(s)": each target can be changed only to another legal
// target; if one can't, the original is unchanged even if illegal; if all
// aren't changed, none are), 115.7d ("choose new targets": any number may stay
// unchanged even if illegal; changed targets must be legal), 115.7e (only the
// final set is evaluated, so targets may swap), 115.5 (a stack object can't
// target itself), 707.10c (copies choose new targets like 115.7d) and 702.47
// (spliced text is part of the host spell). Pinned Oracle: Willbender
// 0aae277e-e58e-4115-b5fd-0459451e17ec "When this creature is turned face up,
// change the target of target spell or ability with a single target."
// (Scryfall rulings 2018-03-16: "You must change the target if possible ... If
// there are no legal targets to choose from, the target isn't changed.");
// Goblin Flectomancer 7c3171da-7a95-4f8e-ba30-c31af21c1828 "Sacrifice this
// creature: You may change the targets of target instant or sorcery spell."
// (rulings 2006-02-01: change all the targets or none, each to a different
// legal target; if changing one is impossible, none change); Redirect
// 74a7d880-5a49-412b-af4d-0b4cdb8453ee "You may choose new targets for target
// spell." (rulings 2010-08-15: change any number, including none; a target
// that can't change remains, even if illegal); Counterspell
// cc187110-1148-4090-bbb8-e205694a39f5 "Counter target spell."

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// retargetRuleAgent picks hostTarget for a retarget host's own target, accepts
// "you may" prompts, and answers each new-target prompt with want, recording
// those prompts. A want missing from the offered options fails the test.
type retargetRuleAgent struct {
	simPassPolicy
	t          *testing.T
	hostTarget game.Target
	want       []game.Target
	requests   []game.ChoiceRequest
}

func (a *retargetRuleAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	switch {
	case request.Kind == game.ChoiceMay:
		return []int{1}
	case request.Prompt == "Change targets" || request.Prompt == "Choose new targets":
		a.requests = append(a.requests, request)
		return chooseAuthoredTargets(a.t, request, a.want...)
	case request.Kind == game.ChoiceTarget:
		return chooseAuthoredTargets(a.t, request, a.hostTarget)
	case len(request.DefaultSelection) > 0:
		return request.DefaultSelection
	default:
		return []int{request.Options[0].Index}
	}
}

func retargetCreatureSpec(selection game.Selection) game.TargetSpec {
	selection.RequiredTypes = []types.Card{types.Creature}
	return game.TargetSpec{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent, Selection: opt.Val(selection)}
}

// pushCreatureTargetSpell puts an instant on the stack with one independent
// "target creature" word per selection, each targeting the matching target.
func pushCreatureTargetSpell(g *game.Game, controller game.PlayerID, selections []game.Selection, targets ...game.Target) *game.StackObject {
	specs := make([]game.TargetSpec, len(selections))
	counts := make([]int, len(selections))
	for i, selection := range selections {
		specs[i], counts[i] = retargetCreatureSpec(selection), 1
	}
	def := &game.CardDef{CardFace: game.CardFace{
		Name:         "Creature Target Spell",
		Types:        []types.Card{types.Instant},
		SpellAbility: opt.Val(game.Mode{Targets: specs}.Ability()),
	}}
	addImplementationSpellToStack(g, controller, def, targets)
	obj, _ := g.Stack.Peek()
	obj.TargetCounts = counts
	return obj
}

// turnUpWillbender turns a face-down Willbender face up for its morph cost and
// puts its triggered ability on the stack targeting agent.hostTarget.
func turnUpWillbender(t *testing.T, g *game.Game, engine *Engine, agent *retargetRuleAgent) {
	t.Helper()
	willbender := addCombatPermanent(g, game.Player2, productionChoiceCard(t, "Willbender"))
	willbender.FaceDown, willbender.FaceDownKind, willbender.FaceDownFace = true, game.FaceDownMorph, game.FaceFront
	g.Players[game.Player2].ManaPool.Add(mana.U, 2)
	g.Turn.PriorityPlayer = game.Player2
	agents := [game.NumPlayers]PlayerAgent{game.Player2: agent}
	if !engine.applyTurnFaceUpWithChoices(g, game.Player2, willbender.ObjectID, agents, &TurnLog{}) {
		t.Fatal("Willbender was not turned face up for its morph cost")
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("Willbender's turned-face-up trigger was not put on the stack")
	}
	trigger, _ := g.Stack.Peek()
	if trigger.Kind != game.StackTriggeredAbility || !slices.Equal(trigger.Targets, []game.Target{agent.hostTarget}) {
		t.Fatalf("Willbender trigger = %+v, want target %v", trigger, agent.hostTarget)
	}
}

// TestWillbenderMustChangeTargetWhenAnotherIsLegal proves change-target is
// mandatory: the legal original target isn't offered and another legal target
// is selected (CR 115.7a).
func TestWillbenderMustChangeTargetWhenAnotherIsLegal(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := removalTokenEngine(t, authored.Counterspell)
	victim := pushStackSpell(g, game.Player2, "Victim", types.Instant)
	bystander := pushStackSpell(g, game.Player2, "Bystander", types.Sorcery)
	counter := pushRetargetHostCard(g, game.Player1, authored.Counterspell.New(), game.StackObjectTarget(victim.ID))
	agent := &retargetRuleAgent{t: t, hostTarget: game.StackObjectTarget(counter.ID), want: []game.Target{game.StackObjectTarget(bystander.ID)}}
	turnUpWillbender(t, g, engine, agent)

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player2: agent}, &TurnLog{})

	if offered := retargetOfferedTargets(agent.requests); !slices.EqualFunc(offered, [][]game.Target{{game.StackObjectTarget(bystander.ID)}}, slices.Equal) {
		t.Fatalf("Willbender change options = %v, want only the other legal spell", offered)
	}
	if !slices.Equal(counter.Targets, []game.Target{game.StackObjectTarget(bystander.ID)}) {
		t.Fatalf("Counterspell targets = %v, want bystander", counter.Targets)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := stackObjectByID(g, bystander.ID); ok {
		t.Fatal("changed Counterspell did not counter its new target")
	}
	if _, ok := stackObjectByID(g, victim.ID); !ok {
		t.Fatal("changed Counterspell countered its original target")
	}
}

// TestWillbenderLeavesIllegalTargetWithoutAlternative proves an original target
// that became illegal remains unchanged when no other legal target exists (CR
// 115.7a), so the spell then fails to resolve against it.
func TestWillbenderLeavesIllegalTargetWithoutAlternative(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := removalTokenEngine(t, authored.Counterspell)
	creature := addCreaturePermanent(g, game.Player1)
	spell := pushCreatureTargetSpell(g, game.Player1, []game.Selection{{Controller: game.ControllerYou}}, game.PermanentTarget(creature.ObjectID))
	// Leaving the battlefield makes the original target illegal, and Player1
	// controls no other creature to change it to (Willbender is Player2's).
	g.Battlefield = slices.DeleteFunc(g.Battlefield, func(p *game.Permanent) bool { return p.ObjectID == creature.ObjectID })
	agent := &retargetRuleAgent{t: t, hostTarget: game.StackObjectTarget(spell.ID)}
	turnUpWillbender(t, g, engine, agent)

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player2: agent}, &TurnLog{})

	if len(agent.requests) != 0 {
		t.Fatalf("Willbender offered a change with no legal alternative: %+v", agent.requests)
	}
	if !slices.Equal(spell.Targets, []game.Target{game.PermanentTarget(creature.ObjectID)}) {
		t.Fatalf("spell targets = %v, want the unchanged illegal original", spell.Targets)
	}
}

func activateGoblinFlectomancer(t *testing.T, g *game.Game, engine *Engine, spell *game.StackObject, agent *retargetRuleAgent) {
	t.Helper()
	flectomancer := addCombatPermanent(g, game.Player3, productionChoiceCard(t, "Goblin Flectomancer"))
	g.Turn.PriorityPlayer = game.Player3
	agents := [game.NumPlayers]PlayerAgent{game.Player3: agent}
	if !engine.applyActivateAbilityWithChoices(g, game.Player3, action.ActivateAbilityAction{
		SourceID: flectomancer.ObjectID, Targets: []game.Target{game.StackObjectTarget(spell.ID)},
	}, agents, &TurnLog{}) {
		t.Fatal("Goblin Flectomancer activation failed")
	}
	if _, ok := permanentByObjectID(g, flectomancer.ObjectID); ok {
		t.Fatal("Goblin Flectomancer was not sacrificed to pay its cost")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
}

// TestGoblinFlectomancerChangesEveryTargetOrSwaps proves "change the targets"
// offers only sets where every target changes to another legal target, and
// that targets may swap because only the final set is evaluated (CR 115.7a,
// CR 115.7e).
func TestGoblinFlectomancerChangesEveryTargetOrSwaps(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	a := addCreaturePermanent(g, game.Player1)
	b := addCreaturePermanent(g, game.Player1)
	c := addCreaturePermanent(g, game.Player2)
	ta, tb, tc := game.PermanentTarget(a.ObjectID), game.PermanentTarget(b.ObjectID), game.PermanentTarget(c.ObjectID)
	spell := pushCreatureTargetSpell(g, game.Player1, []game.Selection{{}, {}}, ta, tb)
	agent := &retargetRuleAgent{t: t, want: []game.Target{tb, ta}}

	activateGoblinFlectomancer(t, g, engine, spell, agent)

	want := [][]game.Target{{tb, ta}, {tb, tc}, {tc, ta}, {tc, tc}}
	offered := retargetOfferedTargets(agent.requests)
	if len(offered) != len(want) {
		t.Fatalf("change options = %v, want %v", offered, want)
	}
	for _, option := range want {
		if !slices.ContainsFunc(offered, func(got []game.Target) bool { return slices.Equal(got, option) }) {
			t.Fatalf("change options = %v, missing %v", offered, option)
		}
	}
	if !slices.Equal(spell.Targets, []game.Target{tb, ta}) {
		t.Fatalf("spell targets = %v, want swapped targets", spell.Targets)
	}
}

// TestGoblinFlectomancerChangesNoTargetUnlessAllCan proves that when one target
// has no other legal target, no target changes, even one that could (CR 115.7a).
func TestGoblinFlectomancerChangesNoTargetUnlessAllCan(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	a := addCreaturePermanent(g, game.Player1)
	b := addCreaturePermanent(g, game.Player2)
	addCreaturePermanent(g, game.Player2)
	original := []game.Target{game.PermanentTarget(b.ObjectID), game.PermanentTarget(a.ObjectID)}
	// Player1 controls only creature A, so the second target can't change.
	spell := pushCreatureTargetSpell(g, game.Player1, []game.Selection{{}, {Controller: game.ControllerYou}}, original...)
	agent := &retargetRuleAgent{t: t}

	activateGoblinFlectomancer(t, g, engine, spell, agent)

	if len(agent.requests) != 0 {
		t.Fatalf("change offered although one target can't change: %+v", agent.requests)
	}
	if !slices.Equal(spell.Targets, original) {
		t.Fatalf("spell targets = %v, want all unchanged %v", spell.Targets, original)
	}
}

// TestRedirectMayLeaveIllegalTargetAndChangeAnother proves "choose new
// targets" may keep an illegal target while changing another (CR 115.7d).
func TestRedirectMayLeaveIllegalTargetAndChangeAnother(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	a := addCreaturePermanent(g, game.Player1)
	b := addCreaturePermanent(g, game.Player2)
	c := addCreaturePermanent(g, game.Player2)
	ta, tb, tc := game.PermanentTarget(a.ObjectID), game.PermanentTarget(b.ObjectID), game.PermanentTarget(c.ObjectID)
	spell := pushCreatureTargetSpell(g, game.Player1, []game.Selection{{}, {Controller: game.ControllerYou}}, tb, ta)
	// Player1 loses control of A, making the second target illegal with no
	// legal replacement.
	a.Controller = game.Player2
	pushRetargetHostCard(g, game.Player4, productionChoiceCard(t, "Redirect"), game.StackObjectTarget(spell.ID))
	agent := &retargetRuleAgent{t: t, want: []game.Target{tc, ta}}

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player4: agent}, &TurnLog{})

	offered := retargetOfferedTargets(agent.requests)
	for _, option := range [][]game.Target{{tb, ta}, {tc, ta}} {
		if !slices.ContainsFunc(offered, func(got []game.Target) bool { return slices.Equal(got, option) }) {
			t.Fatalf("new-target options = %v, missing %v", offered, option)
		}
	}
	if !slices.Equal(spell.Targets, []game.Target{tc, ta}) {
		t.Fatalf("spell targets = %v, want first changed and illegal second kept", spell.Targets)
	}
}

// TestChangeTargetsIsAllOrNoneAcrossSplicedText proves a spliced card's targets
// are part of the "change the targets" set: if a spliced target can't change,
// the host target doesn't change either (CR 115.7a, CR 702.47).
func TestChangeTargetsIsAllOrNoneAcrossSplicedText(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	victim := pushStackSpell(g, game.Player2, "Victim", types.Instant)
	bystander := pushStackSpell(g, game.Player2, "Bystander", types.Sorcery)
	host := pushStackSpell(g, game.Player1, "Arcane Host", types.Instant)
	g.CardInstances[host.SourceID].Def.SpellAbility = opt.Val(game.Mode{Targets: []game.TargetSpec{{
		MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowStackObject,
		Predicate: game.TargetPredicate{StackObjectKinds: []game.StackObjectKind{game.StackSpell}},
	}}}.Ability())
	host.Targets, host.TargetCounts = []game.Target{game.StackObjectTarget(victim.ID)}, []int{1}
	host.SplicedContent = []game.AbilityContent{{MinModes: 1, MaxModes: 1, Modes: []game.Mode{{Targets: []game.TargetSpec{{
		MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowStackObject,
		Predicate: game.TargetPredicate{StackObjectKinds: []game.StackObjectKind{game.StackActivatedAbility}},
	}}}}}}
	// The spliced text targets an activated ability that has left the stack,
	// and no other activated ability exists.
	gone := game.StackObjectTarget(g.IDGen.Next())
	host.SplicedTargets, host.SplicedTargetCounts = [][]game.Target{{gone}}, [][]int{{1}}
	addEffectSpellToStack(g, game.Player2, game.ChooseNewTargets{Object: game.TargetStackObjectReference(0)},
		[]game.Target{game.StackObjectTarget(host.ID)})
	agent := &retargetRuleAgent{t: t, want: []game.Target{game.StackObjectTarget(bystander.ID)}}

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player2: agent}, &TurnLog{})

	if len(agent.requests) != 0 {
		t.Fatalf("change offered although the spliced target can't change: %+v", agent.requests)
	}
	if !slices.Equal(host.Targets, []game.Target{game.StackObjectTarget(victim.ID)}) ||
		!slices.Equal(host.SplicedTargets[0], []game.Target{gone}) {
		t.Fatalf("targets = %v spliced %v, want both unchanged", host.Targets, host.SplicedTargets)
	}
}
