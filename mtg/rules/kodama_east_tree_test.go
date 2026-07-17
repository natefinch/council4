package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

// kodamaEastTreeFilter is the choose-from-hand filter the generated Kodama of the
// East Tree carries: any permanent card whose mana value is equal to or less than
// the permanent that just entered.
func kodamaEastTreeFilter() game.Selection {
	return game.Selection{
		RequiredTypesAny:                   []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Land, types.Planeswalker, types.Battle},
		ManaValueLessOrEqualEventPermanent: true,
	}
}

// registerCardInstance registers a bare card instance in no zone and returns its
// id, so a triggering enter event can name an entering permanent by CardID while
// the mana-value comparison reads that card's printed definition (CR 608.2h),
// independent of whether the permanent is still on the battlefield.
func registerCardInstance(g *game.Game, owner game.PlayerID, def *game.CardDef) id.ID {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{ID: cardID, Def: def, Owner: owner}
	return cardID
}

// TestKodamaEastTreeManaValueBoundComparesToEventPermanent proves the event-
// relative "equal or lesser mana value" bound compares each hand card to the
// permanent named by the triggering event, inclusively: a card with mana value
// strictly greater than the entering permanent is rejected, while equal and
// lesser cards qualify. The entering card is looked up by its printed definition
// rather than a live permanent, so the comparison is last-known-information
// correct even when the permanent has already left.
func TestKodamaEastTreeManaValueBoundComparesToEventPermanent(t *testing.T) {
	cases := []struct {
		name   string
		handMV int
		want   bool
	}{
		{"lesser mana value qualifies", 2, true},
		{"equal mana value qualifies inclusively", 3, true},
		{"greater mana value is rejected", 4, false},
		{"zero mana value qualifies", 0, true},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			enteringID := registerCardInstance(g, game.Player1, chooseFromZoneCardDef("Entering", 3, []types.Card{types.Creature}, nil))
			event := game.Event{Kind: game.EventPermanentEnteredBattlefield, CardID: enteringID}

			handID := addCardToHand(g, game.Player1, chooseFromZoneCardDef("Hand", test.handMV, []types.Card{types.Creature}, nil))
			handCard, ok := g.GetCardInstance(handID)
			if !ok {
				t.Fatal("hand card not registered")
			}
			if got := handCardMatchesSelectionWithEvent(g, handCard, kodamaEastTreeFilter(), game.Player1, event); got != test.want {
				t.Fatalf("handCardMatchesSelectionWithEvent() = %v, want %v", got, test.want)
			}
		})
	}
}

// TestKodamaEastTreeManaValueBoundReadsEnteringTokenDefinition proves a token
// entering the battlefield supplies the comparison bound from its token
// definition. A zero mana value token (the common case) admits only zero mana
// value hand cards, exercising both the token path and the MV0 boundary.
func TestKodamaEastTreeManaValueBoundReadsEnteringTokenDefinition(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	tokenEvent := game.Event{
		Kind:     game.EventPermanentEnteredBattlefield,
		TokenDef: chooseFromZoneCardDef("Spirit Token", 0, []types.Card{types.Creature}, nil),
	}

	zeroID := addCardToHand(g, game.Player1, chooseFromZoneCardDef("Zero", 0, []types.Card{types.Creature}, nil))
	oneID := addCardToHand(g, game.Player1, chooseFromZoneCardDef("One", 1, []types.Card{types.Creature}, nil))
	zero, _ := g.GetCardInstance(zeroID)
	one, _ := g.GetCardInstance(oneID)

	if !handCardMatchesSelectionWithEvent(g, zero, kodamaEastTreeFilter(), game.Player1, tokenEvent) {
		t.Fatal("zero mana value card rejected against a zero mana value token")
	}
	if handCardMatchesSelectionWithEvent(g, one, kodamaEastTreeFilter(), game.Player1, tokenEvent) {
		t.Fatal("nonzero card admitted against a zero mana value token")
	}
}

// TestKodamaEastTreeManaValueBoundFailsClosedWithoutEvent proves the bound reads
// no value when the resolution carries no triggering event, so it rejects every
// hand card rather than silently admitting all of them. The choose-from-zone
// resolution passes the zero event in that case.
func TestKodamaEastTreeManaValueBoundFailsClosedWithoutEvent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	handID := addCardToHand(g, game.Player1, chooseFromZoneCardDef("Hand", 0, []types.Card{types.Creature}, nil))
	handCard, _ := g.GetCardInstance(handID)
	if handCardMatchesSelection(g, handCard, kodamaEastTreeFilter(), game.Player1) {
		t.Fatal("event-relative bound admitted a card with no triggering event")
	}
}

// TestKodamaEastTreeProvenanceInterveningIf proves the "if it wasn't put onto the
// battlefield with this ability" intervening condition fires for every entry
// except one this same ability source put onto the battlefield. A normal entry
// and an entry another source put both fire — so two Kodamas (or a Kodama and a
// copy, which have distinct object identities) keep chaining off each other —
// while an entry this source put does not, stopping a single ability instance
// from recursing on its own placement.
func TestKodamaEastTreeProvenanceInterveningIf(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Kodama of the East Tree",
		Types: []types.Card{types.Creature},
	}})
	otherKodama := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Other Kodama",
		Types: []types.Card{types.Creature},
	}})
	trigger := game.TriggerCondition{InterveningIfEventPermanentWasNotPutByThisAbilitySource: true}

	cases := []struct {
		name   string
		source *game.Permanent
		event  *game.Event
		want   bool
	}{
		{"ordinary entry fires", source, &game.Event{Kind: game.EventPermanentEnteredBattlefield}, true},
		{"entry put by another source fires", source, &game.Event{Kind: game.EventPermanentEnteredBattlefield, EnterPutByAbilitySource: otherKodama.ObjectID}, true},
		{"entry put by this source does not fire", source, &game.Event{Kind: game.EventPermanentEnteredBattlefield, EnterPutByAbilitySource: source.ObjectID}, false},
		{"nil event fires", source, nil, true},
		{"nil source fires", nil, &game.Event{Kind: game.EventPermanentEnteredBattlefield, EnterPutByAbilitySource: source.ObjectID}, true},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := triggerInterveningIf(g, test.source, game.Player1, &trigger, test.event); got != test.want {
				t.Fatalf("triggerInterveningIf() = %v, want %v", got, test.want)
			}
		})
	}
}

// TestKodamaEastTreePutFromHandRespectsManaValueBound proves the end-to-end
// resolution offers only mana-value-eligible hand cards: with the triggering
// event bound to the resolving object, a lesser card is put onto the battlefield
// while a greater card is never a candidate and stays in hand.
func TestKodamaEastTreePutFromHandRespectsManaValueBound(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Kodama of the East Tree",
		Types: []types.Card{types.Creature},
	}})
	obj := triggeredObjFor(source)
	enteringID := registerCardInstance(g, game.Player1, chooseFromZoneCardDef("Entering", 3, []types.Card{types.Creature}, nil))
	obj.TriggerEvent = game.Event{Kind: game.EventPermanentEnteredBattlefield, CardID: enteringID}
	obj.HasTriggerEvent = true

	eligible := addCardToHand(g, game.Player1, chooseFromZoneCardDef("Lesser", 2, []types.Card{types.Creature}, nil))
	tooExpensive := addCardToHand(g, game.Player1, chooseFromZoneCardDef("Greater", 4, []types.Card{types.Creature}, nil))

	instruction := &game.Instruction{Primitive: game.PutFromHandChoice(
		game.ControllerReference(),
		kodamaEastTreeFilter(),
		game.Fixed(1),
		false,
		false,
		false,
	)}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: defaultChoiceAgent{}}
	engine.resolveInstructionWithChoices(g, obj, instruction, agents, &TurnLog{})

	if g.Players[game.Player1].Hand.Contains(eligible) {
		t.Fatal("eligible lesser card was not put onto the battlefield")
	}
	if _, ok := reanimatedPermanent(g, eligible); !ok {
		t.Fatal("eligible lesser card is not on the battlefield")
	}
	if !g.Players[game.Player1].Hand.Contains(tooExpensive) {
		t.Fatal("greater mana value card was put despite exceeding the bound")
	}
}

// TestKodamaEastTreePutFromHandStampsProvenance proves the put-from-hand
// resolution records the resolving ability source on the entering permanent's
// event, which is exactly the fact the provenance intervening-if reads to avoid
// re-triggering the same ability instance on its own placement.
func TestKodamaEastTreePutFromHandStampsProvenance(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Kodama of the East Tree",
		Types: []types.Card{types.Creature},
	}})
	obj := triggeredObjFor(source)
	put := addCardToHand(g, game.Player1, chooseFromZoneCardDef("Put", 1, []types.Card{types.Creature}, nil))

	instruction := &game.Instruction{Primitive: game.PutFromHandChoice(
		game.ControllerReference(),
		game.Selection{RequiredTypesAny: []types.Card{types.Creature}},
		game.Fixed(1),
		false,
		false,
		false,
	)}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: defaultChoiceAgent{}}
	engine.resolveInstructionWithChoices(g, obj, instruction, agents, &TurnLog{})

	if g.Players[game.Player1].Hand.Contains(put) {
		t.Fatal("card was not put onto the battlefield")
	}
	assertEvent(t, g.Events, game.EventPermanentEnteredBattlefield, func(event game.Event) bool {
		return event.EnterPutByAbilitySource == source.ObjectID
	})
}
