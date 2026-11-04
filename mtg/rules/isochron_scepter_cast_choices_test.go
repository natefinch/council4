package rules

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Source review: pinned corpus Oracle text for Isochron Scepter (oracle
// ddd649f4-dbaf-48ec-8ff7-95581257772d; five WotC rulings dated 2020-08-07 from
// https://api.scryfall.com/cards/2aa24fe0-e275-4307-b26c-2a656068a451/rulings),
// Vampire's Bite (oracle 9e88796b-f37f-4f9f-9a6f-695db2801f0e; six 2024-11-08
// kicker rulings), Izzet Charm (oracle a07698f6-5ad5-49a3-9da2-f82d407f5cd7; one
// 2020-08-07 ruling) and Fire // Ice (oracle
// ae92942b-919c-4ea9-b693-85fcef765d5a; seven 2022-12-08 split-card rulings),
// with registered Young Pyromancer as a "whenever you cast" observer.
// CR 2026-09-25 from the official copy: casting a copy of a card follows
// 601.2a-h (707.12), so the caster announces modes, optional additional costs
// such as kicker and targets (601.2b-c, 702.33a/d, 700.2a) and chooses a split
// half (709.3); "without paying its mana cost" still allows additional costs
// (118.9d, Scepter ruling) with X as 0 (107.3b); a copy of a spell keeps its
// kicked status (707.10, Bite ruling); a copy of a card off the stack ceases to
// exist (707.10a), leaving the imprinted card in exile.

// scepterCopyPolicy consents to copying and casting the imprinted card and
// picks the first offered way to cast the copy that pick accepts, recording
// every cast-choice request.
type scepterCopyPolicy struct {
	simPassPolicy

	pick     func(game.ChoiceOption) bool
	requests []game.ChoiceRequest
}

func (p *scepterCopyPolicy) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	switch {
	case request.Kind == game.ChoiceMay:
		return []int{1}
	case request.Kind == game.ChoiceResolution && strings.HasPrefix(request.Prompt, "Choose how to cast"):
		p.requests = append(p.requests, request)
		for _, option := range request.Options {
			if p.pick != nil && p.pick(option) {
				return []int{option.Index}
			}
		}
	}
	return request.DefaultSelection
}

// onlyRequest returns the single cast-choice request the copy cast offered.
func (p *scepterCopyPolicy) onlyRequest(t *testing.T) game.ChoiceRequest {
	t.Helper()
	if len(p.requests) != 1 {
		t.Fatalf("cast-choice requests = %d, want 1", len(p.requests))
	}
	return p.requests[0]
}

func scepterAgents(policy *scepterCopyPolicy) [game.NumPlayers]PlayerAgent {
	agents := simPassPolicies()
	agents[game.Player1] = policy
	return agents
}

// pickCast accepts the cast option with or without kicker targeting want.
func pickCast(kicked bool, want ...game.Target) func(game.ChoiceOption) bool {
	return func(option game.ChoiceOption) bool {
		return strings.Contains(option.Label, "with kicker") == kicked && sameTargetObjects(option.Targets, want)
	}
}

type scepterScenario struct {
	g         *game.Game
	engine    *Engine
	scepter   *game.Permanent
	imprinted id.ID
}

// lookupRegisteredCard returns the single default-registry definition named
// name. These legacy cards are registered by name, not by authored Oracle
// identity.
func lookupRegisteredCard(t *testing.T, registry *cards.Registry, name string) *game.CardDef {
	t.Helper()
	defs := registry.LookupAll(name)
	if len(defs) != 1 {
		t.Fatalf("default registry has %d definitions named %s, want 1", len(defs), name)
	}
	return defs[0]
}

// newScepterScenario puts registered Isochron Scepter onto the battlefield and
// resolves its imprint trigger, exiling the registered card name from
// Player1's hand.
func newScepterScenario(t *testing.T, name string) scepterScenario {
	t.Helper()
	g, engine, scepter, scepterDef, imprintDef := newScepterGame(t, name)
	cardID := addCardToHand(g, game.Player1, imprintDef)
	trigger := &game.StackObject{
		Kind:         game.StackTriggeredAbility,
		SourceID:     scepter.ObjectID,
		SourceCardID: scepter.CardInstanceID,
		Controller:   game.Player1,
	}
	engine.resolveAbilityContentWithChoices(g, trigger, scepterDef.TriggeredAbilities[0].Content, simPassPolicies(), &TurnLog{})
	if !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatalf("Isochron Scepter did not imprint %s", name)
	}
	return scepterScenario{g: g, engine: engine, scepter: scepter, imprinted: cardID}
}

func newScepterGame(t *testing.T, name string) (*game.Game, *Engine, *game.Permanent, *game.CardDef, *game.CardDef) {
	t.Helper()
	registry := cards.NewDefaultRegistry()
	scepterDef := lookupRegisteredCard(t, registry, "Isochron Scepter")
	imprintDef := lookupRegisteredCard(t, registry, name)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	return g, NewEngine(nil), addCombatPermanent(g, game.Player1, scepterDef), scepterDef, imprintDef
}

// newLinkedScepterScenario links the registered card name, resting in
// Player1's exile, to registered Isochron Scepter's imprint. It stands in for an
// imprint the Scepter's own filter would not allow (a split card's combined
// mana value, CR 709.4), exercising the same linked-exile copy route.
func newLinkedScepterScenario(t *testing.T, name string) scepterScenario {
	t.Helper()
	g, engine, scepter, _, imprintDef := newScepterGame(t, name)
	cardID := addExiledInstant(g, game.Player1, imprintDef)
	rememberLinkedObject(g, linkedObjectByObjectKey(g, activatedObjFor(scepter), "imprint"), game.LinkedObjectRef{CardID: cardID})
	return scepterScenario{g: g, engine: engine, scepter: scepter, imprinted: cardID}
}

// activate activates the Scepter's {2}, {T} ability with colorless mana and
// leaves it on the stack.
func (s scepterScenario) activate(t *testing.T) {
	t.Helper()
	setMainPhasePriority(s.g, game.Player1)
	s.g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	if !s.engine.applyAction(s.g, game.Player1, action.ActivateAbility(s.scepter.ObjectID, 0, nil, 0)) {
		t.Fatal("Isochron Scepter activation failed")
	}
	if s.g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("activation did not spend {2}")
	}
}

// resolveCopyCast resolves the activated ability and returns the cast copy.
func (s scepterScenario) resolveCopyCast(t *testing.T, policy *scepterCopyPolicy) *game.StackObject {
	t.Helper()
	casts := spellCastEventCount(s.g)
	s.engine.resolveTopOfStackWithChoices(s.g, scepterAgents(policy), &TurnLog{})
	copyObj, ok := s.g.Stack.Peek()
	if !ok || copyObj.Kind != game.StackSpell || !copyObj.Copy || !copyObj.CastCopy || copyObj.SourceCardID != s.imprinted {
		t.Fatalf("stack top = %+v, want a cast copy of the imprinted card", copyObj)
	}
	if got := spellCastEventCount(s.g) - casts; got != 1 {
		t.Fatalf("spell cast events = %d, want 1 for the cast copy", got)
	}
	if !s.g.Players[game.Player1].Exile.Contains(s.imprinted) {
		t.Fatal("imprinted card left exile when its copy was cast")
	}
	return copyObj
}

func biteScenario(t *testing.T) (scepterScenario, *game.Permanent, *game.Permanent) {
	t.Helper()
	s := newScepterScenario(t, "Vampire's Bite")
	first := addCreatureWithPT(s.g, game.Player1, 2, 2)
	second := addCreatureWithPT(s.g, game.Player1, 1, 1)
	s.activate(t)
	return s, first, second
}

func TestIsochronScepterCopyOffersKickerAndTargetChoice(t *testing.T) {
	s, first, second := biteScenario(t)
	s.g.Players[game.Player1].ManaPool.Add(mana.B, 1)
	s.g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	want := game.PermanentTarget(second.ObjectID)
	policy := &scepterCopyPolicy{pick: pickCast(true, want)}

	copyObj := s.resolveCopyCast(t, policy)

	var kicked, unkicked int
	for _, option := range policy.onlyRequest(t).Options {
		if strings.Contains(option.Label, "with kicker") {
			kicked++
		} else {
			unkicked++
		}
	}
	if kicked != 2 || unkicked != 2 {
		t.Fatalf("offered %d kicked and %d unkicked casts, want 2 of each (two targets)", kicked, unkicked)
	}
	if !copyObj.KickerPaid || !sameTargetObjects(copyObj.Targets, []game.Target{want}) {
		t.Fatalf("copy kicked=%v targets=%v, want kicked targeting the second creature", copyObj.KickerPaid, copyObj.Targets)
	}
	if s.g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("pool after kicked copy = %d, want {2}{B} kicker paid", s.g.Players[game.Player1].ManaPool.Total())
	}
	if event := lastSpellCastEvent(t, s.g); !event.KickerPaid || event.CardID != 0 || event.SourceID != s.imprinted {
		t.Fatalf("cast event = %+v, want a kicked cast copy of the imprinted card", event)
	}

	s.engine.resolveTopOfStack(s.g, &TurnLog{})
	if s.g.Stack.Size() != 0 {
		t.Fatal("resolved copy did not leave the stack")
	}
	if effectivePower(s.g, second) != 4 || !hasKeyword(s.g, second, game.Lifelink) {
		t.Fatalf("chosen creature power=%d lifelink=%v, want 4 with lifelink", effectivePower(s.g, second), hasKeyword(s.g, second, game.Lifelink))
	}
	if effectivePower(s.g, first) != 2 || hasKeyword(s.g, first, game.Lifelink) {
		t.Fatal("unchosen creature was affected")
	}
	if !s.g.Players[game.Player1].Exile.Contains(s.imprinted) || s.g.Players[game.Player1].Graveyard.Contains(s.imprinted) {
		t.Fatal("resolved copy moved the imprinted card")
	}
}

func TestIsochronScepterCopyDeclinedKickerLeavesManaAndIsUnkicked(t *testing.T) {
	s, first, second := biteScenario(t)
	s.g.Players[game.Player1].ManaPool.Add(mana.B, 1)
	s.g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	want := game.PermanentTarget(first.ObjectID)

	copyObj := s.resolveCopyCast(t, &scepterCopyPolicy{pick: pickCast(false, want)})

	if copyObj.KickerPaid || !sameTargetObjects(copyObj.Targets, []game.Target{want}) {
		t.Fatalf("copy kicked=%v targets=%v, want unkicked targeting the first creature", copyObj.KickerPaid, copyObj.Targets)
	}
	if s.g.Players[game.Player1].ManaPool.Total() != 3 {
		t.Fatalf("pool after unkicked copy = %d, want 3 unspent", s.g.Players[game.Player1].ManaPool.Total())
	}
	s.engine.resolveTopOfStack(s.g, &TurnLog{})
	if effectivePower(s.g, first) != 5 || hasKeyword(s.g, first, game.Lifelink) {
		t.Fatalf("chosen creature power=%d lifelink=%v, want 5 without lifelink", effectivePower(s.g, first), hasKeyword(s.g, first, game.Lifelink))
	}
	if effectivePower(s.g, second) != 1 {
		t.Fatal("unchosen creature was affected")
	}
}

func TestIsochronScepterCopyUnaffordableKickerIsNotOffered(t *testing.T) {
	s, _, second := biteScenario(t)
	s.g.Players[game.Player1].ManaPool.Add(mana.B, 1)
	want := game.PermanentTarget(second.ObjectID)
	policy := &scepterCopyPolicy{pick: pickCast(false, want)}

	copyObj := s.resolveCopyCast(t, policy)

	for _, option := range policy.onlyRequest(t).Options {
		if strings.Contains(option.Label, "with kicker") {
			t.Fatalf("unaffordable kicker was offered: %q", option.Label)
		}
	}
	if copyObj.KickerPaid || !sameTargetObjects(copyObj.Targets, []game.Target{want}) {
		t.Fatalf("copy kicked=%v targets=%v, want unkicked targeting the second creature", copyObj.KickerPaid, copyObj.Targets)
	}
	if s.g.Players[game.Player1].ManaPool.Total() != 1 {
		t.Fatal("unkicked free copy spent mana")
	}
}

// TestIsochronScepterKickedCopyIsKeptByStackCopyCloneAndReplay proves the kicked
// status and chosen target survive a copy of the cast copy (CR 707.10), a game
// clone and a replay of the same choices.
func TestIsochronScepterKickedCopyIsKeptByStackCopyCloneAndReplay(t *testing.T) {
	s, _, second := biteScenario(t)
	s.g.Players[game.Player1].ManaPool.Add(mana.B, 1)
	s.g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	want := game.PermanentTarget(second.ObjectID)
	replay := s.g.Clone()

	copyObj := s.resolveCopyCast(t, &scepterCopyPolicy{pick: pickCast(true, want)})
	replayed := scepterScenario{g: replay, engine: s.engine, scepter: s.scepter, imprinted: s.imprinted}.
		resolveCopyCast(t, &scepterCopyPolicy{pick: pickCast(true, want)})
	if replayed.KickerPaid != copyObj.KickerPaid || !sameTargetObjects(replayed.Targets, copyObj.Targets) ||
		len(replay.Events) != len(s.g.Events) {
		t.Fatalf("replayed copy = %+v, want same choices as %+v", replayed, copyObj)
	}

	branch := s.g.Clone()
	if top, _ := branch.Stack.Peek(); !top.KickerPaid || !sameTargetObjects(top.Targets, []game.Target{want}) {
		t.Fatal("clone lost the cast copy's kicker or target")
	}

	addEffectSpellToStack(s.g, game.Player1, game.CopyStackObject{Object: game.TargetStackObjectReference(0)},
		[]game.Target{game.StackObjectTarget(copyObj.ID)})
	s.engine.resolveTopOfStack(s.g, &TurnLog{})
	stackCopy, ok := s.g.Stack.Peek()
	if !ok || stackCopy.ID == copyObj.ID || !stackCopy.Copy || stackCopy.CastCopy || !stackCopy.KickerPaid {
		t.Fatalf("stack copy = %+v, want an uncast kicked copy of the cast copy", stackCopy)
	}
	s.engine.resolveTopOfStack(s.g, &TurnLog{})
	s.engine.resolveTopOfStack(s.g, &TurnLog{})
	if effectivePower(s.g, second) != 7 || !hasKeyword(s.g, second, game.Lifelink) {
		t.Fatalf("creature power=%d lifelink=%v after both kicked copies, want 7 with lifelink",
			effectivePower(s.g, second), hasKeyword(s.g, second, game.Lifelink))
	}

	branchSecond, ok := permanentByObjectID(branch, second.ObjectID)
	if !ok || effectivePower(branch, branchSecond) != 1 || hasKeyword(branch, branchSecond, game.Lifelink) {
		t.Fatal("resolving in the original game mutated the clone")
	}
	s.engine.resolveTopOfStack(branch, &TurnLog{})
	if effectivePower(branch, branchSecond) != 4 || !hasKeyword(branch, branchSecond, game.Lifelink) {
		t.Fatal("cloned kicked copy did not resolve kicked against its target")
	}
}

func TestIsochronScepterCopyOffersModes(t *testing.T) {
	s := newScepterScenario(t, "Izzet Charm")
	victim := addCreatureWithPT(s.g, game.Player2, 2, 2)
	s.activate(t)
	want := game.PermanentTarget(victim.ObjectID)
	policy := &scepterCopyPolicy{pick: func(option game.ChoiceOption) bool {
		return strings.Contains(option.Label, "modes [1]") && sameTargetObjects(option.Targets, []game.Target{want})
	}}
	hand := s.g.Players[game.Player1].Hand.Size()

	copyObj := s.resolveCopyCast(t, policy)

	if len(copyObj.ChosenModes) != 1 || copyObj.ChosenModes[0] != 1 {
		t.Fatalf("copy modes = %v, want the damage mode", copyObj.ChosenModes)
	}
	var drawMode bool
	for _, option := range policy.onlyRequest(t).Options {
		drawMode = drawMode || strings.Contains(option.Label, "modes [2]")
	}
	if !drawMode {
		t.Fatal("the targetless draw mode was not offered")
	}
	s.engine.resolveTopOfStack(s.g, &TurnLog{})
	if victim.MarkedDamage != 2 {
		t.Fatalf("chosen creature damage = %d, want 2 from the damage mode", victim.MarkedDamage)
	}
	if s.g.Players[game.Player1].Hand.Size() != hand {
		t.Fatal("unchosen draw mode resolved")
	}
}

func TestIsochronScepterCopyOffersSplitHalves(t *testing.T) {
	s := newLinkedScepterScenario(t, "Fire")
	victim := addCreatureWithPT(s.g, game.Player2, 2, 2)
	addCardToLibrary(s.g, game.Player1, plainInstantDef("Library Card"))
	s.activate(t)
	want := game.PermanentTarget(victim.ObjectID)
	policy := &scepterCopyPolicy{pick: func(option game.ChoiceOption) bool {
		return strings.HasPrefix(option.Label, "Ice") && sameTargetObjects(option.Targets, []game.Target{want})
	}}
	hand := s.g.Players[game.Player1].Hand.Size()

	copyObj := s.resolveCopyCast(t, policy)

	if copyObj.Face != game.FaceAlternate {
		t.Fatalf("copy face = %v, want the Ice half", copyObj.Face)
	}
	var fire bool
	for _, option := range policy.onlyRequest(t).Options {
		fire = fire || strings.HasPrefix(option.Label, "Fire")
	}
	if !fire {
		t.Fatal("the Fire half was not offered")
	}
	s.engine.resolveTopOfStack(s.g, &TurnLog{})
	if !victim.Tapped || s.g.Players[game.Player1].Hand.Size() != hand+1 {
		t.Fatalf("Ice copy tapped=%v drew=%d, want tapped and one card drawn", victim.Tapped, s.g.Players[game.Player1].Hand.Size()-hand)
	}
	if !s.g.Players[game.Player1].Exile.Contains(s.imprinted) || s.g.Players[game.Player1].Graveyard.Contains(s.imprinted) {
		t.Fatal("resolved split copy moved the imprinted card")
	}
}

// TestIsochronScepterKickedCopyTriggersWheneverYouCast proves the chosen copy
// cast is a genuine cast (CR 707.12): registered Young Pyromancer's "whenever
// you cast an instant or sorcery spell" triggers for the kicked copy.
func TestIsochronScepterKickedCopyTriggersWheneverYouCast(t *testing.T) {
	s, first, _ := biteScenario(t)
	pyromancer := addCombatPermanent(s.g, game.Player1, lookupRegisteredCard(t, cards.NewDefaultRegistry(), "Young Pyromancer"))
	s.g.Players[game.Player1].ManaPool.Add(mana.B, 1)
	s.g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	policy := &scepterCopyPolicy{pick: pickCast(true, game.PermanentTarget(first.ObjectID))}

	copyObj := s.resolveCopyCast(t, policy)
	s.engine.putTriggeredAbilitiesOnStackWithChoices(s.g, scepterAgents(policy), &TurnLog{})

	top, ok := s.g.Stack.Peek()
	if !ok || top.Kind != game.StackTriggeredAbility || top.SourceID != pyromancer.ObjectID {
		t.Fatalf("stack top = %+v, want Young Pyromancer's cast trigger above the copy", top)
	}
	if !copyObj.KickerPaid {
		t.Fatal("cast copy was not kicked")
	}
	s.engine.resolveTopOfStackWithChoices(s.g, scepterAgents(policy), &TurnLog{})
	var elementals int
	for _, permanent := range s.g.Battlefield {
		if permanent.Token && permanent.Controller == game.Player1 {
			elementals++
		}
	}
	if elementals != 1 {
		t.Fatalf("Elemental tokens = %d, want 1 from the copy's cast trigger", elementals)
	}
}
