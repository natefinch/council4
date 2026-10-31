package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: issue #3372. CR 2026-09-25 (verified copy, effective
// September 25, 2026): 702.37c turns a morph card face down as it is cast, a
// 2/2 face-down creature card with no text, name, subtypes or mana cost, and
// puts it onto the stack as a face-down spell with the same characteristics;
// 708.2 says those are its only characteristics and 708.4 that objects cast
// face down are turned face down before they are put onto the stack, so effects
// that care about the spell's characteristics see only the face-down ones.
// 603.10a makes "leaves a graveyard" triggers look back in time, and the object
// they look back at was already face down. 400.7 makes the moved object new.
// Default-registry cards checked against the pinned corpus: Birchlore Rangers
// oracle 667c5676-7a18-4fa9-b965-b26797538e23 (Creature — Elf Druid Ranger,
// green, Morph {G}; 2004-10-04 ruling only about its mana ability), Aveline de
// Grandpré oracle 5c8c4d53-f364-40da-989b-cfe1e3f79b7d (Legendary Creature —
// Human Assassin, green, Disguise {B}{G}; 2024-07-05 rulings: the face-down
// spell is a colorless 2/2 with ward {2} and no name, mana cost or creature
// types), Zoetic Cavern oracle 3763de30-28e1-4689-a71c-07d2fea3a466 (Land,
// Morph {2}), Syr Konrad, the Grim oracle 14c3ff84-1e82-4606-a433-869fc52cc382
// (legacy name-only default-registry entry; "... or a creature card leaves your
// graveyard, Syr Konrad deals 1 damage to each opponent") and Muldrotha, the Gravetide oracle
// e4625704-1d52-44e4-804f-2f45644d76ac (cast a permanent spell of each
// permanent type from your graveyard; 2020-11-10 ruling: an alternative cost
// may be used).

func castZoneChangeEvent(t *testing.T, g *game.Game, cardID id.ID, from zone.Type) game.Event {
	t.Helper()
	for _, event := range g.Events {
		if event.Kind == game.EventZoneChanged && event.CardID == cardID && event.FromZone == from && event.ToZone == zone.Stack {
			return event
		}
	}
	t.Fatalf("no %v -> stack zone-change event for card %v", from, cardID)
	return game.Event{}
}

func zoneChangeMatches(g *game.Game, watcher *game.Permanent, event game.Event, card, subject game.Selection) bool {
	pattern := &game.TriggerPattern{Event: game.EventZoneChanged, CardSelection: card, SubjectSelection: subject}
	return triggerMatchesEventForController(g, watcher, watcher.Controller, pattern, event)
}

func TestFaceDownCastZoneChangeHidesPrintedCharacteristics(t *testing.T) {
	t.Parallel()
	g, _, obj := castFaceDownForOpponentResponse(t, "Birchlore Rangers", game.FaceDownMorph)
	event := castZoneChangeEvent(t, g, obj.SourceID, zone.Hand)
	if !event.FaceDown || !event.CardCharacteristicsKnown {
		t.Fatalf("face-down cast zone change FaceDown=%v known=%v, want both", event.FaceDown, event.CardCharacteristicsKnown)
	}
	for _, state := range []*game.Game{g, g.Clone()} {
		watcher := addCreaturePermanent(state, game.Player2)
		creature := game.Selection{RequiredTypes: []types.Card{types.Creature}}
		if !zoneChangeMatches(state, watcher, event, creature, game.Selection{}) ||
			!zoneChangeMatches(state, watcher, event, game.Selection{}, creature) {
			t.Fatal("face-down cast zone change must match a creature filter")
		}
		for _, sel := range []game.Selection{
			{SubtypesAny: []types.Sub{types.Elf}},
			{SubtypesAny: []types.Sub{types.Druid, types.Ranger}},
			{ColorsAny: []color.Color{color.Green}},
			{Colored: true},
			{Name: "Birchlore Rangers"},
		} {
			if zoneChangeMatches(state, watcher, event, sel, game.Selection{}) {
				t.Fatalf("card selection %+v matched the hidden printed Birchlore Rangers", sel)
			}
			if zoneChangeMatches(state, watcher, event, game.Selection{}, sel) {
				t.Fatalf("subject selection %+v matched the hidden printed Birchlore Rangers", sel)
			}
		}
		if !zoneChangeMatches(state, watcher, event, game.Selection{Colorless: true}, game.Selection{}) {
			t.Fatal("face-down cast zone change must be colorless")
		}
	}
}

func TestFaceDownDisguiseCastZoneChangeHidesLegendaryHuman(t *testing.T) {
	t.Parallel()
	g, _, obj := castFaceDownForOpponentResponse(t, "Aveline de Grandpré", game.FaceDownDisguise)
	event := castZoneChangeEvent(t, g, obj.SourceID, zone.Hand)
	watcher := addCreaturePermanent(g, game.Player2)
	for _, sel := range []game.Selection{
		{Supertypes: []types.Super{types.Legendary}},
		{SubtypesAny: []types.Sub{types.Human}},
		{ColorsAny: []color.Color{color.Green}},
	} {
		if zoneChangeMatches(g, watcher, event, sel, game.Selection{}) ||
			zoneChangeMatches(g, watcher, event, game.Selection{}, sel) {
			t.Fatalf("selection %+v matched the hidden printed Aveline de Grandpré", sel)
		}
	}
}

func TestFaceUpCastAndDiscardZoneChangesKeepPrintedCharacteristics(t *testing.T) {
	t.Parallel()
	g, engine, cardID := faceDownCostGame(t, "Birchlore Rangers", mana.G, 1)
	if !engine.applyAction(g, game.Player1, actionBuild.castSpell(cardID, zone.Hand, game.FaceFront, nil, 0, nil)) {
		t.Fatal("face-up Birchlore Rangers cast failed")
	}
	cast := castZoneChangeEvent(t, g, cardID, zone.Hand)
	if cast.FaceDown {
		t.Fatal("face-up cast zone change marked face down")
	}
	discardedID := addCardToHand(g, game.Player1, productionChoiceCard(t, "Birchlore Rangers"))
	if !discardCardFromHand(g, game.Player1, discardedID) {
		t.Fatal("discard failed")
	}
	var discard game.Event
	for _, event := range g.Events {
		if event.Kind == game.EventZoneChanged && event.CardID == discardedID && event.FromZone == zone.Hand {
			discard = event
		}
	}
	if discard.CardID != discardedID || discard.ToZone != zone.Graveyard {
		t.Fatalf("discard zone change = %+v, want hand -> graveyard", discard)
	}
	watcher := addCreaturePermanent(g, game.Player2)
	for _, event := range []game.Event{cast, discard} {
		for _, sel := range []game.Selection{
			{SubtypesAny: []types.Sub{types.Elf}},
			{ColorsAny: []color.Color{color.Green}},
		} {
			if !zoneChangeMatches(g, watcher, event, sel, game.Selection{}) {
				t.Fatalf("%v -> %v card selection %+v lost the printed characteristics", event.FromZone, event.ToZone, sel)
			}
			if !zoneChangeMatches(g, watcher, event, game.Selection{}, sel) {
				t.Fatalf("%v -> %v subject selection %+v lost the printed characteristics", event.FromZone, event.ToZone, sel)
			}
		}
	}
}

// A face-down cast from the graveyard looks back at an object that was already
// turned face down (CR 603.10a, 708.4): a face-down Zoetic Cavern leaving the
// graveyard is a creature card for Syr Konrad, and not a land card.
func TestFaceDownGraveyardCastIsCreatureCardLeavingGraveyard(t *testing.T) {
	g, e, registry := lifetimeGame(t)
	source := addCombatPermanent(g, game.Player1, lifetimeDefinition(t, registry, muldrothaID))
	konrad := addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Syr Konrad, the Grim"))
	cavern := addCardToGraveyard(g, game.Player1, productionChoiceCard(t, "Zoetic Cavern"))
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	cast := action.CastFaceDownFromZone(cavern, zone.Graveyard, game.FaceFront, game.FaceDownMorph).
		WithPlayPermission(muldrothaPermission(source, types.Creature))
	if !e.applyAction(g, game.Player1, cast) {
		t.Fatal("face-down Zoetic Cavern graveyard cast failed")
	}
	event := castZoneChangeEvent(t, g, cavern, zone.Graveyard)
	land := game.Selection{RequiredTypes: []types.Card{types.Land}}
	if zoneChangeMatches(g, konrad, event, land, game.Selection{}) ||
		zoneChangeMatches(g, konrad, event, game.Selection{}, land) {
		t.Fatal("face-down Zoetic Cavern left the graveyard as a land card")
	}
	if !e.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Syr Konrad did not trigger for the face-down creature card leaving the graveyard")
	}
	top, ok := g.Stack.Peek()
	if !ok || top.Kind != game.StackTriggeredAbility || top.SourceID != konrad.ObjectID {
		t.Fatalf("stack top = %+v, want Syr Konrad's trigger", top)
	}
	opponentLife := g.Players[game.Player2].Life
	e.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player2].Life; got != opponentLife-1 {
		t.Fatalf("opponent life = %d, want %d", got, opponentLife-1)
	}
}

// The object a face-down cast moves is still a card, a colorless 2/2 creature
// card with mana value 0 (CR 708.2, 708.4), so the zone change's card-subject
// predicates read those face-down characteristics: it is a permanent card, has power and toughness 2,
// and a disguised one has ward {2}; it has no name and no hidden mana value.
func TestFaceDownCastZoneChangeIsFaceDownCreatureCard(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		kind game.FaceDownKind
		ward bool
	}{
		{"Birchlore Rangers", game.FaceDownMorph, false},
		{"Aveline de Grandpré", game.FaceDownDisguise, true},
	} {
		g, _, obj := castFaceDownForOpponentResponse(t, tc.name, tc.kind)
		event := castZoneChangeEvent(t, g, obj.SourceID, zone.Hand)
		for _, state := range []*game.Game{g, g.Clone()} {
			watcher := addCreaturePermanent(state, game.Player2)
			two := opt.Val(compare.Int{Op: compare.Equal, Value: 2})
			zero := opt.Val(compare.Int{Op: compare.Equal, Value: 0})
			for _, sel := range []game.Selection{
				{RequirePermanentCard: true},
				{RequirePermanentCard: true, RequiredTypes: []types.Card{types.Creature}},
				{Power: two, Toughness: two},
				{ManaValue: zero},
			} {
				if !zoneChangeMatches(state, watcher, event, game.Selection{}, sel) {
					t.Fatalf("%s: face-down card subject selection %+v did not match", tc.name, sel)
				}
			}
			if !zoneChangeMatches(state, watcher, event, game.Selection{ManaValue: zero}, game.Selection{}) {
				t.Fatalf("%s: face-down cast card selection lost mana value 0", tc.name)
			}
			ward := game.Selection{Keyword: game.Ward}
			if got := zoneChangeMatches(state, watcher, event, game.Selection{}, ward); got != tc.ward {
				t.Fatalf("%s: ward selection = %v, want %v", tc.name, got, tc.ward)
			}
			for _, sel := range []game.Selection{
				{ManaValue: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 1})},
				{Power: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 3})},
				{Name: tc.name},
			} {
				if zoneChangeMatches(state, watcher, event, game.Selection{}, sel) {
					t.Fatalf("%s: selection %+v matched the hidden printed card", tc.name, sel)
				}
			}
		}
	}
}

// A face-down Zoetic Cavern leaving the graveyard is a permanent card (a
// creature card, CR 708.2), so "whenever a permanent card leaves your
// graveyard" triggers, while a nonpermanent-type filter for the printed land
// does not.
func TestFaceDownGraveyardCastTriggersPermanentCardLeavingGraveyard(t *testing.T) {
	g, e, registry := lifetimeGame(t)
	source := addCombatPermanent(g, game.Player1, lifetimeDefinition(t, registry, muldrothaID))
	konrad := productionChoiceCard(t, "Syr Konrad, the Grim")
	watcherDef := *konrad
	leaves := konrad.TriggeredAbilities[2]
	leaves.Trigger.Pattern.SubjectSelection = game.Selection{RequirePermanentCard: true, ExcludedTypes: []types.Card{types.Land}}
	watcherDef.TriggeredAbilities = []game.TriggeredAbility{leaves}
	watcher := addCombatPermanent(g, game.Player1, &watcherDef)
	cavern := addCardToGraveyard(g, game.Player1, productionChoiceCard(t, "Zoetic Cavern"))
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	cast := action.CastFaceDownFromZone(cavern, zone.Graveyard, game.FaceFront, game.FaceDownMorph).
		WithPlayPermission(muldrothaPermission(source, types.Creature))
	if !e.applyAction(g, game.Player1, cast) {
		t.Fatal("face-down Zoetic Cavern graveyard cast failed")
	}
	if !e.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("permanent-card watcher did not trigger for the face-down card leaving the graveyard")
	}
	top, ok := g.Stack.Peek()
	if !ok || top.Kind != game.StackTriggeredAbility || top.SourceID != watcher.ObjectID {
		t.Fatalf("stack top = %+v, want the permanent-card watcher's trigger", top)
	}
	opponentLife := g.Players[game.Player2].Life
	e.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player2].Life; got != opponentLife-1 {
		t.Fatalf("opponent life = %d, want %d", got, opponentLife-1)
	}
}
