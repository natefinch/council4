package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: Oracle text from the pinned corpus for Zektar Shrine
// Expedition (f12d2c01-0ef2-40e6-8c7d-8512e2c8c074): "Remove three quest
// counters from this enchantment and sacrifice it: Create a 7/1 red Elemental
// creature token with trample and haste. Exile it at the beginning of the next
// end step."; Feldon of the Third Path (0360f1ff-15c9-48e3-89eb-fbc4bf140c55):
// "{2}{R}, {T}: Create a token that's a copy of target creature card in your
// graveyard, except it's an artifact in addition to its other types. It gains
// haste. Sacrifice it at the beginning of the next end step."; and Ojer Taq,
// Deepest Foundation (1ca79dd4-67fc-496c-96fc-489b039c4932): "If one or more
// creature tokens would be created under your control, three times that many of
// those tokens are created instead." https://api.scryfall.com/cards/<id>/rulings
// returned three 2024-11-08 landfall rulings for Zektar and, for Feldon, the
// 2021-03-19 ruling "If Feldon's ability creates multiple tokens due to a
// replacement effect (such as the one Doubling Season creates), each of those
// tokens will gain haste and you'll sacrifice each of them." CR 2026-09-25 from
// the official TXT
// (https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt):
// 603.7c (a delayed trigger affects the particular objects it refers to), 608.2c
// (read the whole text: "it" refers to every token the replaced creation made),
// 616.1 (replacement effects modify the creation event), 111.1 (tokens) and
// 702.26b (a phased-out permanent is treated as though it does not exist). Each
// resolution's delayed trigger binds every token that resolution created, acts
// on them as one simultaneous group, and leaves another resolution's tokens alone.

func playerTokensNamed(g *game.Game, playerID game.PlayerID, name string) []id.ID {
	var ids []id.ID
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanent.Controller == playerID && permanentTokenName(permanent) == name {
			ids = append(ids, permanent.ObjectID)
		}
	}
	return ids
}

func newTokens(before, after []id.ID) []id.ID {
	var created []id.ID
	for _, objectID := range after {
		if !slices.Contains(before, objectID) {
			created = append(created, objectID)
		}
	}
	return created
}

func onBattlefield(g *game.Game, ids []id.ID) []id.ID {
	var present []id.ID
	for _, objectID := range ids {
		if _, ok := g.PermanentByID(objectID); ok {
			present = append(present, objectID)
		}
	}
	return present
}

func sameIDs(got, want []id.ID) bool {
	return len(got) == len(want) && !slices.ContainsFunc(want, func(objectID id.ID) bool {
		return !slices.Contains(got, objectID)
	})
}

// activateZektarUnderOjerTaq activates a real Zektar Shrine Expedition and
// resolves it, returning the Elementals that resolution created.
func activateZektarUnderOjerTaq(t *testing.T, g *game.Game, engine *Engine) []id.ID {
	t.Helper()
	agents := [game.NumPlayers]PlayerAgent{}
	before := playerTokensNamed(g, game.Player1, "Elemental")
	expedition := addCombatPermanent(g, game.Player1, registryCard(t, "Zektar Shrine Expedition"))
	expedition.Counters.Add(counter.Quest, 3)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyActionWithChoices(g, game.Player1, action.ActivateAbility(expedition.ObjectID, 0, nil, 0), agents, &TurnLog{}) {
		t.Fatal("could not activate Zektar Shrine Expedition")
	}
	resolveStackWithTriggers(engine, g, agents)
	created := newTokens(before, playerTokensNamed(g, game.Player1, "Elemental"))
	if len(created) != 3 {
		t.Fatalf("Elementals created under Ojer Taq = %d, want 3", len(created))
	}
	return created
}

func zektarOjerTaqGame(t *testing.T) (*game.Game, *Engine) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addReplacementPermanent(t, g, game.Player1, registryCard(t, "Ojer Taq, Deepest Foundation"))
	return g, NewEngine(nil)
}

// Ojer Taq triples Zektar Shrine Expedition's Elemental; the delayed "Exile it"
// is bound to all three while pending and on the stack, and exiles all three
// together. A clone taken before the end step resolves the same way without
// touching the original game.
func TestZektarShrineExpeditionExilesEveryElementalOjerTaqCreates(t *testing.T) {
	g, engine := zektarOjerTaqGame(t)
	elementals := activateZektarUnderOjerTaq(t, g, engine)

	obs := observe(g, game.Player2)
	views := obs.DelayedTriggers()
	if len(views) != 1 || !sameIDs(views[0].BoundObjectIDs, elementals) {
		t.Fatalf("pending delayed triggers = %+v, want one exile bound to Elementals %v", views, elementals)
	}
	for _, elemental := range elementals {
		if delayedViewBoundTo(findPermanentView(t, obs, elemental).PendingDelayedTriggers, elemental) != 1 {
			t.Fatalf("Elemental %d does not show the pending exile", elemental)
		}
	}

	clone := g.Clone()
	for _, run := range []struct {
		name string
		g    *game.Game
	}{{"clone", clone}, {"original", g}} {
		fireEndStep(t, run.g, engine)
		stack := observe(run.g, game.Player2).Stack()
		if len(stack) != 1 || !sameIDs(stack[0].BoundObjectIDs, elementals) {
			t.Fatalf("%s stack = %+v, want one exile bound to Elementals %v", run.name, stack, elementals)
		}
		if run.name == "clone" && len(onBattlefield(g, elementals)) != 3 {
			t.Fatal("resolving the clone's end step changed the original game")
		}
		engine.resolveTopOfStack(run.g, &TurnLog{})
		if left := onBattlefield(run.g, elementals); len(left) != 0 {
			t.Fatalf("%s: Elementals %v remained after the next end step's exile", run.name, left)
		}
	}
}

// Two separate Zektar Shrine Expedition resolutions keep independent bindings:
// each delayed exile acts on exactly the three Elementals its own resolution
// created, though both publish under the same link key.
func TestZektarShrineExpeditionSeparateResolutionsKeepTheirOwnElementals(t *testing.T) {
	g, engine := zektarOjerTaqGame(t)
	first := activateZektarUnderOjerTaq(t, g, engine)
	second := activateZektarUnderOjerTaq(t, g, engine)

	views := observe(g, game.Player1).DelayedTriggers()
	if len(views) != 2 || !sameIDs(views[0].BoundObjectIDs, first) || !sameIDs(views[1].BoundObjectIDs, second) {
		t.Fatalf("pending delayed triggers = %+v, want exiles bound to %v and %v", views, first, second)
	}
	fireEndStep(t, g, engine)
	if g.Stack.Size() != 2 {
		t.Fatalf("stack size = %d, want both delayed exiles", g.Stack.Size())
	}
	top, _ := g.Stack.Peek()
	topGroup, otherGroup := first, second
	if !sameIDs(top.CapturedObjectIDs, first) {
		topGroup, otherGroup = second, first
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if left := onBattlefield(g, topGroup); len(left) != 0 {
		t.Fatalf("Elementals %v remained after their own exile resolved", left)
	}
	if left := onBattlefield(g, otherGroup); len(left) != 3 {
		t.Fatalf("the other resolution's Elementals on the battlefield = %v, want all of %v", left, otherGroup)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if left := onBattlefield(g, append(slices.Clone(first), second...)); len(left) != 0 {
		t.Fatalf("Elementals %v remained after both exiles resolved", left)
	}
}

// A captured Elemental that is phased out when the exile resolves is treated as
// though it does not exist (CR 702.26b), so only the phased-in Elementals are
// exiled, as with a single captured object.
func TestZektarShrineExpeditionSkipsPhasedOutElemental(t *testing.T) {
	g, engine := zektarOjerTaqGame(t)
	elementals := activateZektarUnderOjerTaq(t, g, engine)
	fireEndStep(t, g, engine)
	phased, _ := g.PermanentByID(elementals[0])
	phased.PhasedOut = true
	engine.resolveTopOfStack(g, &TurnLog{})
	if left := onBattlefield(g, elementals); !sameIDs(left, elementals[:1]) {
		t.Fatalf("Elementals on the battlefield = %v, want only the phased-out %d", left, elementals[0])
	}
}

// Feldon of the Third Path's official ruling: when a replacement effect makes
// its ability create several tokens, each is sacrificed at the next end step.
func TestFeldonOfTheThirdPathSacrificesEveryTokenOjerTaqCreates(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	agents := [game.NumPlayers]PlayerAgent{}
	addReplacementPermanent(t, g, game.Player1, registryCard(t, "Ojer Taq, Deepest Foundation"))
	feldon := addCombatPermanent(g, game.Player1, registryCard(t, "Feldon of the Third Path"))
	blueprint := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Graveyard Bear",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	g.Players[game.Player1].ManaPool.Add(mana.R, 3)
	setMainPhasePriority(g, game.Player1)
	targets := []game.Target{{Kind: game.TargetCard, CardID: blueprint}}
	if !engine.applyActionWithChoices(g, game.Player1, action.ActivateAbility(feldon.ObjectID, 0, targets, 0), agents, &TurnLog{}) {
		t.Fatal("could not activate Feldon of the Third Path")
	}
	resolveStackWithTriggers(engine, g, agents)
	copies := playerTokensNamed(g, game.Player1, "Graveyard Bear")
	if len(copies) != 3 {
		t.Fatalf("copy tokens created under Ojer Taq = %d, want 3", len(copies))
	}
	for _, tokenID := range copies {
		token, _ := g.PermanentByID(tokenID)
		if !hasKeyword(g, token, game.Haste) || !permanentHasType(g, token, types.Artifact) {
			t.Fatalf("Feldon token %d lacks haste or the artifact type", tokenID)
		}
	}
	views := observe(g, game.Player2).DelayedTriggers()
	if len(views) != 1 || !sameIDs(views[0].BoundObjectIDs, copies) {
		t.Fatalf("pending delayed triggers = %+v, want one sacrifice bound to %v", views, copies)
	}
	fireEndStep(t, g, engine)
	engine.resolveTopOfStack(g, &TurnLog{})
	if left := onBattlefield(g, copies); len(left) != 0 {
		t.Fatalf("copy tokens %v remained after the next end step's sacrifice", left)
	}
}
