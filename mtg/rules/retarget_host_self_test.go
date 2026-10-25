package rules

// Source review: CR 2026-09-25 (official MagicCompRules 20260925.txt from
// magic.wizards.com/en/rules) 115.5 (a spell or ability on the stack is an
// illegal target for itself), 115.7a/d/e (changed or new targets must be legal;
// unchanged targets may stay; only the final set is evaluated), 707.10/707.10c
// (a copy is a new stack object whose new targets must be legal before it is
// put onto the stack) and 702.47 (spliced text is part of the host spell).
// Oracle: Redirect "You may choose new targets for target spell."; Willbender
// "When this creature is turned face up, change the target of target spell or
// ability with a single target."; Fork "Copy target instant or sorcery spell,
// except that the copy is red. You may choose new targets for the copy.";
// Counterspell "Counter target spell." Engine gap #3251.

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// retargetRecordingAgent answers every may prompt with yes and every "Choose
// new targets" request with the option whose targets equal want, recording each
// target request it saw.
type retargetRecordingAgent struct {
	simPassPolicy
	t        *testing.T
	want     []game.Target
	requests []game.ChoiceRequest
}

func (a *retargetRecordingAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceMay {
		return []int{1}
	}
	a.requests = append(a.requests, request)
	return chooseAuthoredTargets(a.t, request, a.want...)
}

func pushRetargetHostCard(g *game.Game, controller game.PlayerID, def *game.CardDef, targets ...game.Target) *game.StackObject {
	addImplementationSpellToStack(g, controller, def, targets)
	obj, _ := g.Stack.Peek()
	if len(targets) > 0 {
		obj.TargetCounts = []int{len(targets)}
	}
	return obj
}

func assertRetargetNeverOffers(t *testing.T, requests []game.ChoiceRequest, forbidden id.ID) {
	t.Helper()
	if len(requests) == 0 {
		t.Fatal("no new-target choice was offered")
	}
	for _, request := range requests {
		for _, option := range request.Options {
			if slices.Contains(option.Targets, game.StackObjectTarget(forbidden)) {
				t.Fatalf("retarget offered stack object %v as its own target: %+v", forbidden, option)
			}
		}
	}
}

func retargetOfferedTargets(requests []game.ChoiceRequest) [][]game.Target {
	var offered [][]game.Target
	for _, request := range requests {
		for _, option := range request.Options {
			if len(option.Targets) > 0 {
				offered = append(offered, option.Targets)
			}
		}
	}
	return offered
}

// TestRedirectCannotRetargetCounterspellToItself proves a spell's new targets
// exclude the spell itself while every other legal spell remains available.
func TestRedirectCannotRetargetCounterspellToItself(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := removalTokenEngine(t, authored.Counterspell)
	victim := pushStackSpell(g, game.Player2, "Victim", types.Instant)
	bystander := pushStackSpell(g, game.Player2, "Bystander", types.Sorcery)
	counter := pushRetargetHostCard(g, game.Player1, authored.Counterspell.New(), game.StackObjectTarget(victim.ID))
	pushRetargetHostCard(g, game.Player2, productionChoiceCard(t, "Redirect"), game.StackObjectTarget(counter.ID))

	agent := &retargetRecordingAgent{t: t, want: []game.Target{game.StackObjectTarget(bystander.ID)}}
	agents := [game.NumPlayers]PlayerAgent{game.Player2: agent}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	assertRetargetNeverOffers(t, agent.requests, counter.ID)
	offered := retargetOfferedTargets(agent.requests)
	for _, want := range []id.ID{victim.ID, bystander.ID} {
		if !slices.ContainsFunc(offered, func(targets []game.Target) bool {
			return slices.Equal(targets, []game.Target{game.StackObjectTarget(want)})
		}) {
			t.Fatalf("legal spell %v missing from new-target options %v", want, offered)
		}
	}
	if !slices.Equal(counter.Targets, []game.Target{game.StackObjectTarget(bystander.ID)}) {
		t.Fatalf("Counterspell targets = %v, want bystander", counter.Targets)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := stackObjectByID(g, bystander.ID); ok {
		t.Fatal("redirected Counterspell did not counter its new target")
	}
	if _, ok := stackObjectByID(g, victim.ID); !ok {
		t.Fatal("redirected Counterspell countered its old target")
	}
}

// TestWillbenderKeepsTargetWhenOnlyHostIsAlternative proves the no-valid
// alternative case: the host is never offered, so changing the target leaves the
// original target unchanged (CR 115.7a).
func TestWillbenderKeepsTargetWhenOnlyHostIsAlternative(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := removalTokenEngine(t, authored.Counterspell)
	victim := pushStackSpell(g, game.Player2, "Victim", types.Instant)
	counter := pushRetargetHostCard(g, game.Player1, authored.Counterspell.New(), game.StackObjectTarget(victim.ID))
	willbenderDef := productionChoiceCard(t, "Willbender")
	willbender := addCombatPermanent(g, game.Player2, willbenderDef)
	g.Stack.Push(&game.StackObject{
		ID:            g.IDGen.Next(),
		Kind:          game.StackTriggeredAbility,
		SourceID:      willbender.ObjectID,
		SourceCardID:  willbender.CardInstanceID,
		Controller:    game.Player2,
		InlineTrigger: &willbenderDef.TriggeredAbilities[0],
		Targets:       []game.Target{game.StackObjectTarget(counter.ID)},
		TargetCounts:  []int{1},
	})

	agent := &retargetRecordingAgent{t: t, want: []game.Target{game.StackObjectTarget(victim.ID)}}
	agents := [game.NumPlayers]PlayerAgent{game.Player2: agent}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	assertRetargetNeverOffers(t, agent.requests, counter.ID)
	if offered := retargetOfferedTargets(agent.requests); len(offered) != 1 {
		t.Fatalf("new-target options = %v, want only the unchanged victim", offered)
	}
	if !slices.Equal(counter.Targets, []game.Target{game.StackObjectTarget(victim.ID)}) {
		t.Fatalf("Counterspell targets = %v, want unchanged victim", counter.Targets)
	}
}

// TestForkCopyOfCounterspellMayTargetOriginal proves a copy is a distinct stack
// object: its new targets may include the spell it copies (CR 707.10c).
func TestForkCopyOfCounterspellMayTargetOriginal(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := removalTokenEngine(t, authored.Counterspell)
	victim := pushStackSpell(g, game.Player2, "Victim", types.Instant)
	counter := pushRetargetHostCard(g, game.Player1, authored.Counterspell.New(), game.StackObjectTarget(victim.ID))
	pushRetargetHostCard(g, game.Player2, productionChoiceCard(t, "Fork"), game.StackObjectTarget(counter.ID))

	agent := &retargetRecordingAgent{t: t, want: []game.Target{game.StackObjectTarget(counter.ID)}}
	agents := [game.NumPlayers]PlayerAgent{game.Player2: agent}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	copyObj, ok := g.Stack.Peek()
	if !ok || !copyObj.Copy {
		t.Fatal("Fork did not put a Counterspell copy on the stack")
	}
	assertRetargetNeverOffers(t, agent.requests, copyObj.ID)
	if !slices.Equal(copyObj.Targets, []game.Target{game.StackObjectTarget(counter.ID)}) {
		t.Fatalf("copy targets = %v, want original Counterspell", copyObj.Targets)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := stackObjectByID(g, counter.ID); ok {
		t.Fatal("Fork copy did not counter the original Counterspell")
	}
	if _, ok := stackObjectByID(g, victim.ID); !ok {
		t.Fatal("victim left the stack although its counter was countered")
	}
}

// TestChooseNewTargetsExcludesHostAbilityAndKeepsSourceIdentity proves an
// ability cannot be retargeted to itself while its source-relative "another
// target creature" restriction still excludes its source permanent.
func TestChooseNewTargetsExcludesHostAbilityAndKeepsSourceIdentity(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creatureA := addCreaturePermanent(g, game.Player1)
	creatureB := addCreaturePermanent(g, game.Player1)
	source := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Ability Source", Types: []types.Card{types.Creature},
	}})
	other := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Other Source", Types: []types.Card{types.Artifact},
	}})
	activatedKinds := game.TargetPredicate{StackObjectKinds: []game.StackObjectKind{game.StackActivatedAbility}}
	otherAbility := &game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackActivatedAbility, SourceID: other.ObjectID,
		SourceCardID: other.CardInstanceID, Controller: game.Player2,
		InlineActivated: &game.ActivatedAbility{},
	}
	g.Stack.Push(otherAbility)
	host := &game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackActivatedAbility, SourceID: source.ObjectID,
		SourceCardID: source.CardInstanceID, Controller: game.Player2,
		InlineActivated: &game.ActivatedAbility{Content: game.Mode{Targets: []game.TargetSpec{
			{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowStackObject, Predicate: activatedKinds},
			{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent, Selection: opt.Val(game.Selection{
				RequiredTypes: []types.Card{types.Creature}, ExcludeSource: true,
			})},
		}}.Ability()},
		Targets:      []game.Target{game.StackObjectTarget(otherAbility.ID), game.PermanentTarget(creatureA.ObjectID)},
		TargetCounts: []int{1, 1},
	}
	g.Stack.Push(host)
	addEffectSpellToStack(g, game.Player1, game.ChooseNewTargets{Object: game.TargetStackObjectReference(0)},
		[]game.Target{game.StackObjectTarget(host.ID)})

	want := []game.Target{game.StackObjectTarget(otherAbility.ID), game.PermanentTarget(creatureB.ObjectID)}
	agent := &retargetRecordingAgent{t: t, want: want}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	assertRetargetNeverOffers(t, agent.requests, host.ID)
	for _, offered := range retargetOfferedTargets(agent.requests) {
		if slices.Contains(offered, game.PermanentTarget(source.ObjectID)) {
			t.Fatalf("another-target restriction offered the ability source: %v", offered)
		}
	}
	if offered := retargetOfferedTargets(agent.requests); len(offered) != 2 {
		t.Fatalf("new-target options = %v, want the other ability with creature A or B", offered)
	}
	if !slices.Equal(host.Targets, want) {
		t.Fatalf("ability targets = %v, want %v", host.Targets, want)
	}
}

// TestChooseNewTargetsExcludesHostFromSplicedTargets proves spliced text, which
// is part of the host spell (CR 702.47), cannot be retargeted to the host.
func TestChooseNewTargetsExcludesHostFromSplicedTargets(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	victim := pushStackSpell(g, game.Player2, "Victim", types.Instant)
	bystander := pushStackSpell(g, game.Player2, "Bystander", types.Sorcery)
	host := pushStackSpell(g, game.Player1, "Arcane Host", types.Instant)
	host.SplicedContent = []game.AbilityContent{{MinModes: 1, MaxModes: 1, Modes: []game.Mode{{Targets: []game.TargetSpec{{
		MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowStackObject,
		Predicate: game.TargetPredicate{StackObjectKinds: []game.StackObjectKind{game.StackSpell}},
	}}}}}}
	host.SplicedTargets = [][]game.Target{{game.StackObjectTarget(victim.ID)}}
	host.SplicedTargetCounts = [][]int{{1}}
	addEffectSpellToStack(g, game.Player2, game.ChooseNewTargets{Object: game.TargetStackObjectReference(0)},
		[]game.Target{game.StackObjectTarget(host.ID)})

	agent := &retargetRecordingAgent{t: t, want: []game.Target{game.StackObjectTarget(bystander.ID)}}
	agents := [game.NumPlayers]PlayerAgent{game.Player2: agent}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	assertRetargetNeverOffers(t, agent.requests, host.ID)
	if offered := retargetOfferedTargets(agent.requests); len(offered) != 2 {
		t.Fatalf("spliced new-target options = %v, want victim and bystander", offered)
	}
	if !slices.Equal(host.SplicedTargets[0], []game.Target{game.StackObjectTarget(bystander.ID)}) {
		t.Fatalf("spliced targets = %v, want bystander", host.SplicedTargets)
	}
}
