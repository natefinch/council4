package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: Battle at the Helvault printing
// 0de0eaeb-9b62-427d-9873-d84de1740ddc (oracle
// 076584c8-ec58-4711-8f59-4839a4fa88ad), chapters I, II: "For each player,
// exile up to one target non-Saga, nonland permanent that player controls
// until this Saga leaves the battlefield." Its three WotC rulings dated
// 2023-07-28 say the target permanents aren't exiled if the Saga leaves before
// the chapter resolves and exiled tokens never return. Vault 13: Dweller's
// Journey printing a22ad436-215a-4da1-b12e-1faac474f948 (oracle
// 586fb987-cd77-420f-978c-2ce140bcdcfb), chapter I: "For each player, exile up
// to one other target enchantment or creature that player controls until this
// Saga leaves the battlefield." Its two WotC rulings dated 2024-03-08 repeat
// the leaves-first ruling and cover chapter III's library order. CR 2026-09-25
// from the verified official copy: 714.2b (a chapter ability triggers when lore
// counters reach its number), 115.1d and 603.3d with 601.2c (a triggered
// ability's targets are announced as it is put on the stack), 115.3 (each
// instance of "target" is chosen separately), 702.11b (hexproof only stops
// opponents' targeting), 608.2b (illegal targets are not affected while the
// legal ones still are), and 610.3a (nothing is exiled if the source already
// left).

// slotTargetAgent answers each per-player target slot by choosing the offered
// permanent in want, or the empty choice, and records every slot request.
type slotTargetAgent struct {
	want     []id.ID
	requests []game.ChoiceRequest
}

func (*slotTargetAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *slotTargetAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind != game.ChoiceTarget {
		return nil
	}
	a.requests = append(a.requests, request)
	empty := -1
	for _, option := range request.Options {
		if len(option.Targets) == 0 {
			empty = option.Index
			continue
		}
		if slices.Contains(a.want, option.Targets[0].PermanentID) {
			return []int{option.Index}
		}
	}
	return []int{empty}
}

// offered reports the permanents each recorded slot request offered.
func (a *slotTargetAgent) offered() [][]id.ID {
	var offered [][]id.ID
	for _, request := range a.requests {
		var slot []id.ID
		for _, option := range request.Options {
			for _, target := range option.Targets {
				slot = append(slot, target.PermanentID)
			}
		}
		offered = append(offered, slot)
	}
	return offered
}

// announceSagaChapterOne gives the Saga its first lore counter and puts the
// real chapter I trigger on the stack, announcing its targets through agent.
func announceSagaChapterOne(t *testing.T, engine *Engine, g *game.Game, saga *game.Permanent, agent *slotTargetAgent) *game.StackObject {
	t.Helper()
	addCountersToPermanent(g, saga, counter.Lore, 1)
	var agents [game.NumPlayers]PlayerAgent
	agents[saga.Controller] = agent
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("chapter I was not put on the stack")
	}
	object, ok := g.Stack.Peek()
	if !ok || !object.SagaChapter || object.SourceID != saga.ObjectID {
		t.Fatalf("stack top = %+v, want the Saga's chapter I", object)
	}
	return object
}

func requireAnnouncedTargets(t *testing.T, object *game.StackObject, want ...*game.Permanent) {
	t.Helper()
	got := make([]id.ID, 0, len(object.Targets))
	for _, target := range object.Targets {
		got = append(got, target.PermanentID)
	}
	expected := make([]id.ID, 0, len(want))
	for _, permanent := range want {
		expected = append(expected, permanent.ObjectID)
	}
	if !slices.Equal(got, expected) {
		t.Fatalf("announced targets = %v, want %v", got, expected)
	}
}

func requireOnBattlefield(t *testing.T, g *game.Game, permanent *game.Permanent) {
	t.Helper()
	if _, ok := permanentByObjectID(g, permanent.ObjectID); !ok {
		t.Fatal("permanent left the battlefield, want it untouched")
	}
}

func slotTargetPermanent(g *game.Game, controller game.PlayerID, name string, cardTypes ...types.Card) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{Name: name, Types: cardTypes}})
}

// TestBattleAtTheHelvaultAnnouncesPerPlayerTargetsAndSurvivesResponse proves
// chapter I announces up to one target per player as it goes on the stack,
// offers only that player's legal permanents (no lands, no Sagas, no opponent
// hexproof, but the controller's own hexproof permanent), and still exiles the
// remaining legal target after a response removes another one.
func TestBattleAtTheHelvaultAnnouncesPerPlayerTargetsAndSurvivesResponse(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	saga := addCombatPermanent(g, game.Player1, untilLeavesRegistryDef(t, "Battle at the Helvault"))
	ownHexproof := addCombatCreaturePermanent(g, game.Player1, game.Hexproof)
	first := addCombatCreaturePermanent(g, game.Player2)
	alternative := slotTargetPermanent(g, game.Player2, "Test Artifact", types.Artifact)
	addCombatCreaturePermanent(g, game.Player3, game.Hexproof)
	addLandPermanent(g, game.Player3, "Forest", types.Forest)
	addCombatPermanent(g, game.Player4, untilLeavesRegistryDef(t, "Vault 13: Dweller's Journey"))
	second := slotTargetPermanent(g, game.Player4, "Test Enchantment", types.Enchantment)
	agent := &slotTargetAgent{want: []id.ID{first.ObjectID, second.ObjectID}}

	chapter := announceSagaChapterOne(t, engine, g, saga, agent)

	requireAnnouncedTargets(t, chapter, first, second)
	wantOffered := [][]id.ID{
		{ownHexproof.ObjectID},
		{first.ObjectID, alternative.ObjectID},
		{second.ObjectID},
	}
	if got := agent.offered(); !slices.EqualFunc(got, wantOffered, slices.Equal) {
		t.Fatalf("offered slots = %v, want %v (Player3's slot has no legal target)", got, wantOffered)
	}
	requireOnBattlefield(t, g, first)
	requireOnBattlefield(t, g, second)

	resolveOpponentDestroy(t, engine, g, game.Player2, first)
	engine.resolveTopOfStack(g, &TurnLog{})

	if zoneOf, ok := cardZone(g, first.CardInstanceID); !ok || zoneOf != zone.Graveyard {
		t.Fatalf("destroyed target zone = %v, want graveyard", zoneOf)
	}
	requireExiled(t, g, second.CardInstanceID)
	requireOnBattlefield(t, g, alternative)
	requireOnBattlefield(t, g, ownHexproof)

	resolveOpponentDestroy(t, engine, g, game.Player2, saga)
	requireReturned(t, g, second.CardInstanceID, second.ObjectID, game.Player4)
}

// TestBattleAtTheHelvaultSkipsTargetThatChangedControl proves a target whose
// controller changes before resolution no longer fits its player's slot and
// is not exiled, while the still-legal targets are exiled simultaneously.
func TestBattleAtTheHelvaultSkipsTargetThatChangedControl(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	saga := addCombatPermanent(g, game.Player1, untilLeavesRegistryDef(t, "Battle at the Helvault"))
	stolen := addCombatCreaturePermanent(g, game.Player2)
	third := addCombatCreaturePermanent(g, game.Player3)
	fourth := addCombatCreaturePermanent(g, game.Player4)
	agent := &slotTargetAgent{want: []id.ID{stolen.ObjectID, third.ObjectID, fourth.ObjectID}}

	chapter := announceSagaChapterOne(t, engine, g, saga, agent)
	requireAnnouncedTargets(t, chapter, stolen, third, fourth)

	// A control-changing response hands Player2's target to Player4, so it no
	// longer is a permanent "that player controls" for Player2's slot.
	stolen.Controller = game.Player4
	before := len(g.Events)
	engine.resolveTopOfStack(g, &TurnLog{})

	requireOnBattlefield(t, g, stolen)
	requireExiled(t, g, third.CardInstanceID)
	requireExiled(t, g, fourth.CardInstanceID)
	var simultaneous []id.ID
	for _, event := range g.Events[before:] {
		if event.Kind == game.EventZoneChanged && event.FromZone == zone.Battlefield && event.ToZone == zone.Exile {
			simultaneous = append(simultaneous, event.SimultaneousID)
		}
	}
	if len(simultaneous) != 2 || simultaneous[0] == 0 || simultaneous[0] != simultaneous[1] {
		t.Fatalf("exile SimultaneousIDs = %v, want one simultaneous exile of both legal targets", simultaneous)
	}

	resolveOpponentDestroy(t, engine, g, game.Player2, saga)
	requireReturned(t, g, third.CardInstanceID, third.ObjectID, game.Player3)
	requireReturned(t, g, fourth.CardInstanceID, fourth.ObjectID, game.Player4)
}

// TestVault13ChapterOneTargetsOtherEnchantmentsOrCreatures proves Vault 13's
// slots exclude the Saga itself and non-enchantment, noncreature permanents,
// and that the announced captives feed chapter III's linked return.
func TestVault13ChapterOneTargetsOtherEnchantmentsOrCreatures(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	def := untilLeavesRegistryDef(t, "Vault 13: Dweller's Journey")
	saga := addCombatPermanent(g, game.Player1, def)
	ownEnchantment := slotTargetPermanent(g, game.Player1, "Test Enchantment", types.Enchantment)
	slotTargetPermanent(g, game.Player2, "Test Artifact", types.Artifact)
	creature := addCombatCreaturePermanent(g, game.Player2)
	enchantment := slotTargetPermanent(g, game.Player3, "Test Enchantment", types.Enchantment)
	agent := &slotTargetAgent{want: []id.ID{creature.ObjectID, enchantment.ObjectID}}

	chapter := announceSagaChapterOne(t, engine, g, saga, agent)

	requireAnnouncedTargets(t, chapter, creature, enchantment)
	wantOffered := [][]id.ID{{ownEnchantment.ObjectID}, {creature.ObjectID}, {enchantment.ObjectID}}
	if got := agent.offered(); !slices.EqualFunc(got, wantOffered, slices.Equal) {
		t.Fatalf("offered slots = %v, want %v", got, wantOffered)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	requireExiled(t, g, creature.CardInstanceID)
	requireExiled(t, g, enchantment.CardInstanceID)
	requireOnBattlefield(t, g, ownEnchantment)

	chapterThree := def.ChapterAbilities[2].Content.Modes[0].Sequence[0].Primitive
	engine.resolveInstructionWithChoices(g, untilLeavesChapterObject(saga), &game.Instruction{Primitive: chapterThree}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if permanentByCardID(g, creature.CardInstanceID) == nil || permanentByCardID(g, enchantment.CardInstanceID) == nil {
		t.Fatal("chapter III did not return both announced captives")
	}
}

// TestVault13LeavingInResponseToChapterOneExilesNothing proves the leaves-first
// ruling on the real announced chapter: targets were chosen, but the Saga left
// before resolution, so nothing is exiled (CR 610.3a).
func TestVault13LeavingInResponseToChapterOneExilesNothing(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	saga := addCombatPermanent(g, game.Player1, untilLeavesRegistryDef(t, "Vault 13: Dweller's Journey"))
	victim := addCombatCreaturePermanent(g, game.Player2)
	agent := &slotTargetAgent{want: []id.ID{victim.ObjectID}}

	chapter := announceSagaChapterOne(t, engine, g, saga, agent)
	requireAnnouncedTargets(t, chapter, victim)
	resolveOpponentDestroy(t, engine, g, game.Player2, saga)
	engine.resolveTopOfStack(g, &TurnLog{})

	requireOnBattlefield(t, g, victim)
	if len(g.UntilSourceLeavesExiles) != 0 {
		t.Fatalf("until-source-leaves records = %+v, want none", g.UntilSourceLeavesExiles)
	}
}
