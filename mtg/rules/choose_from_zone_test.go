package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// sequencedChoiceAgent answers each successive ChooseChoice with the next scripted
// selection, falling back to the offered default once the script is exhausted.
type sequencedChoiceAgent struct {
	choices [][]int
	calls   int
}

func (*sequencedChoiceAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Action{}
}

func (a *sequencedChoiceAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if a.calls >= len(a.choices) {
		return request.DefaultSelection
	}
	selection := a.choices[a.calls]
	a.calls++
	return selection
}

func chooseFromZoneCardDef(name string, manaValue int, cardTypes []types.Card, subtypes []types.Sub) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		ManaCost: opt.Val(cost.Mana{cost.O(manaValue)}),
		Types:    cardTypes,
		Subtypes: subtypes,
	}}
}

func chooseFromZoneSource(g *game.Game) *game.StackObject {
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Source",
		Types: []types.Card{types.Artifact},
	}})
	return triggeredObjFor(source)
}

func addCfzGraveyardCard(g *game.Game, playerID game.PlayerID, def *game.CardDef) id.ID {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{ID: cardID, Def: def, Owner: playerID}
	g.Players[playerID].Graveyard.AddToBottom(cardID)
	return cardID
}

func addCfzLibraryCard(g *game.Game, playerID game.PlayerID, def *game.CardDef) id.ID {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{ID: cardID, Def: def, Owner: playerID}
	g.Players[playerID].Library.AddToBottom(cardID)
	return cardID
}

func resolveChoose(g *game.Game, obj *game.StackObject, agent PlayerAgent, env game.ChooseFromZone) effectResolved {
	engine := NewEngine(nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
	resolver := newEffectResolver(engine, g, obj, agents, &TurnLog{})
	return resolver.resolveChooseFromZone(env)
}

// TestChooseFromZoneAcrossSetReturnsChosenCard verifies the ordinary multi-card
// choice moves the chosen matching card to its destination while a non-matching
// card stays in the source zone.
func TestChooseFromZoneAcrossSetReturnsChosenCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obj := chooseFromZoneSource(g)
	creature := addCfzGraveyardCard(g, game.Player1, chooseFromZoneCardDef("Bear", 2, []types.Card{types.Creature}, nil))
	instant := addCfzGraveyardCard(g, game.Player1, chooseFromZoneCardDef("Bolt", 1, []types.Card{types.Instant}, nil))

	resolveChoose(g, obj, defaultChoiceAgent{}, game.ChooseFromZone{
		Player:      game.ControllerReference(),
		SourceZone:  zone.Graveyard,
		Filter:      game.Selection{RequiredTypes: []types.Card{types.Creature}},
		Quantity:    game.Fixed(1),
		Count:       game.ChooseExactly,
		Destination: game.ChooseDestination{Zone: zone.Hand},
	})

	if !g.Players[game.Player1].Hand.Contains(creature) {
		t.Fatal("chosen creature was not moved to hand")
	}
	if g.Players[game.Player1].Graveyard.Contains(creature) {
		t.Fatal("chosen creature still in graveyard")
	}
	if !g.Players[game.Player1].Graveyard.Contains(instant) {
		t.Fatal("non-matching instant was moved")
	}
}

// TestChooseFromZoneNameFilterSelectsNamedCard verifies the Selection.Name
// equality filter restricts the candidate pool to cards of that name.
func TestChooseFromZoneNameFilterSelectsNamedCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obj := chooseFromZoneSource(g)
	wanted := addCfzLibraryCard(g, game.Player1, chooseFromZoneCardDef("Llanowar Elves", 1, []types.Card{types.Creature}, nil))
	other := addCfzLibraryCard(g, game.Player1, chooseFromZoneCardDef("Grizzly Bears", 2, []types.Card{types.Creature}, nil))

	resolveChoose(g, obj, defaultChoiceAgent{}, game.ChooseFromZone{
		Player:      game.ControllerReference(),
		SourceZone:  zone.Library,
		Filter:      game.Selection{Name: "Llanowar Elves"},
		Quantity:    game.Fixed(1),
		Count:       game.ChooseExactly,
		Destination: game.ChooseDestination{Zone: zone.Hand},
	})

	if !g.Players[game.Player1].Hand.Contains(wanted) {
		t.Fatal("named card was not chosen")
	}
	if !g.Players[game.Player1].Library.Contains(other) {
		t.Fatal("differently named card was chosen")
	}
}

// TestChooseFromZoneUpToWithTotalManaValueCap verifies the up-to count with a
// MaxTotalManaValue rider lets the player choose a subset within the cap and
// makes choosing none legal.
func TestChooseFromZoneUpToWithTotalManaValueCap(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obj := chooseFromZoneSource(g)
	cheap := addCfzGraveyardCard(g, game.Player1, chooseFromZoneCardDef("Bear", 2, []types.Card{types.Creature}, nil))
	expensive := addCfzGraveyardCard(g, game.Player1, chooseFromZoneCardDef("Dragon", 6, []types.Card{types.Creature}, nil))

	agent := &sequencedChoiceAgent{choices: [][]int{{0}}}
	resolveChoose(g, obj, agent, game.ChooseFromZone{
		Player:      game.ControllerReference(),
		SourceZone:  zone.Graveyard,
		Filter:      game.Selection{RequiredTypes: []types.Card{types.Creature}},
		Quantity:    game.Fixed(2),
		Count:       game.ChooseUpTo,
		Destination: game.ChooseDestination{Zone: zone.Hand},
		Riders:      game.ChooseRiders{MaxTotalManaValue: opt.Val(4)},
	})

	if !g.Players[game.Player1].Hand.Contains(cheap) {
		t.Fatal("the affordable creature was not chosen")
	}
	if !g.Players[game.Player1].Graveyard.Contains(expensive) {
		t.Fatal("a creature over the mana-value cap was chosen")
	}
}

// TestChooseFromZoneAnyNumberChoosesWholePool verifies the any-number count moves
// every matching card by default.
func TestChooseFromZoneAnyNumberChoosesWholePool(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obj := chooseFromZoneSource(g)
	bear := addCfzGraveyardCard(g, game.Player1, chooseFromZoneCardDef("Bear", 2, []types.Card{types.Creature}, nil))
	ox := addCfzGraveyardCard(g, game.Player1, chooseFromZoneCardDef("Ox", 3, []types.Card{types.Creature}, nil))

	agent := &sequencedChoiceAgent{choices: [][]int{{0, 1}}}
	resolveChoose(g, obj, agent, game.ChooseFromZone{
		Player:      game.ControllerReference(),
		SourceZone:  zone.Graveyard,
		Filter:      game.Selection{RequiredTypes: []types.Card{types.Creature}},
		Count:       game.ChooseAnyNumber,
		Destination: game.ChooseDestination{Zone: zone.Hand},
	})
	for _, cardID := range []id.ID{bear, ox} {
		if !g.Players[game.Player1].Hand.Contains(cardID) {
			t.Fatalf("card %v was not chosen", cardID)
		}
	}
}

// TestChooseFromZoneSharedSubtypeRejectsIncompatiblePair verifies the
// shared-subtype grouping only offers later picks that share a subtype with the
// cards already chosen, so a non-sharing card is never added.
func TestChooseFromZoneSharedSubtypeRejectsIncompatiblePair(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obj := chooseFromZoneSource(g)
	forestA := addCfzLibraryCard(g, game.Player1, chooseFromZoneCardDef("Forest A", 0, []types.Card{types.Land}, []types.Sub{types.Forest}))
	island := addCfzLibraryCard(g, game.Player1, chooseFromZoneCardDef("Island", 0, []types.Card{types.Land}, []types.Sub{types.Island}))
	forestB := addCfzLibraryCard(g, game.Player1, chooseFromZoneCardDef("Forest B", 0, []types.Card{types.Land}, []types.Sub{types.Forest}))

	// First pick the first Forest (index 0). The second-pick pool then drops the
	// Island, leaving only the other Forest at index 0.
	agent := &sequencedChoiceAgent{choices: [][]int{{0}, {0}, {}}}
	resolveChoose(g, obj, agent, game.ChooseFromZone{
		Player:      game.ControllerReference(),
		SourceZone:  zone.Library,
		Filter:      game.Selection{RequiredTypes: []types.Card{types.Land}},
		Quantity:    game.Fixed(2),
		Count:       game.ChooseExactly,
		Grouping:    game.ChooseSharedSubtype,
		Destination: game.ChooseDestination{Zone: zone.Hand},
	})

	if !g.Players[game.Player1].Hand.Contains(forestA) || !g.Players[game.Player1].Hand.Contains(forestB) {
		t.Fatal("both Forests sharing a subtype should have been chosen")
	}
	if !g.Players[game.Player1].Library.Contains(island) {
		t.Fatal("the Island does not share a subtype and must not be chosen")
	}
}

// TestChooseFromZoneOneOfEachNamedType verifies the one-of-each-named-type
// grouping chooses at most one card matching each listed card type.
func TestChooseFromZoneOneOfEachNamedType(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obj := chooseFromZoneSource(g)
	creature := addCfzLibraryCard(g, game.Player1, chooseFromZoneCardDef("Bear", 2, []types.Card{types.Creature}, nil))
	land := addCfzLibraryCard(g, game.Player1, chooseFromZoneCardDef("Forest", 0, []types.Card{types.Land}, []types.Sub{types.Forest}))
	extraCreature := addCfzLibraryCard(g, game.Player1, chooseFromZoneCardDef("Ox", 3, []types.Card{types.Creature}, nil))

	resolveChoose(g, obj, defaultChoiceAgent{}, game.ChooseFromZone{
		Player:      game.ControllerReference(),
		SourceZone:  zone.Library,
		Filter:      game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Land}},
		Quantity:    game.Fixed(1),
		Count:       game.ChooseExactly,
		Grouping:    game.ChooseOneOfEachNamedType,
		Destination: game.ChooseDestination{Zone: zone.Hand},
	})

	if !g.Players[game.Player1].Hand.Contains(creature) {
		t.Fatal("a creature card should have been chosen")
	}
	if !g.Players[game.Player1].Hand.Contains(land) {
		t.Fatal("a land card should have been chosen")
	}
	if !g.Players[game.Player1].Library.Contains(extraCreature) {
		t.Fatal("only one card of each named type should be chosen")
	}
}

// TestChooseFromZoneSplitDestinationDistributesTwoCards verifies the split
// grouping sends one chosen card to the primary battlefield slot (tapped) and
// the other to the secondary hand slot.
func TestChooseFromZoneSplitDestinationDistributesTwoCards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obj := chooseFromZoneSource(g)
	first := addCfzLibraryCard(g, game.Player1, chooseFromZoneCardDef("Forest", 0, []types.Card{types.Land}, []types.Sub{types.Forest}))
	second := addCfzLibraryCard(g, game.Player1, chooseFromZoneCardDef("Mountain", 0, []types.Card{types.Land}, []types.Sub{types.Mountain}))

	// Choose both cards (indices 0,1), then assign index 0 as the primary card.
	agent := &sequencedChoiceAgent{choices: [][]int{{0, 1}, {0}}}
	resolveChoose(g, obj, agent, game.ChooseFromZone{
		Player:         game.ControllerReference(),
		SourceZone:     zone.Library,
		Filter:         game.Selection{RequiredTypes: []types.Card{types.Land}},
		Quantity:       game.Fixed(2),
		Count:          game.ChooseExactly,
		Grouping:       game.ChooseSplitDestination,
		Destination:    game.ChooseDestination{Zone: zone.Battlefield},
		SplitSecondary: opt.Val(game.ChooseSplitSlot{Destination: game.ChooseDestination{Zone: zone.Hand}}),
		Riders:         game.ChooseRiders{EntersTapped: true},
	})

	if !onBattlefieldByCard(g, first) {
		t.Fatal("the primary card was not put onto the battlefield")
	}
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == first && !permanent.Tapped {
			t.Fatal("the primary card should enter tapped")
		}
	}
	if !g.Players[game.Player1].Hand.Contains(second) {
		t.Fatal("the secondary card was not put into hand")
	}
}

// TestChooseFromZoneSplitDestinationSingleCardChoosesSlot verifies that with one
// card chosen the player selects which slot it fills.
func TestChooseFromZoneSplitDestinationSingleCardChoosesSlot(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obj := chooseFromZoneSource(g)
	only := addCfzLibraryCard(g, game.Player1, chooseFromZoneCardDef("Forest", 0, []types.Card{types.Land}, []types.Sub{types.Forest}))

	// Choose the single card (index 0), then choose the secondary slot (index 1).
	agent := &sequencedChoiceAgent{choices: [][]int{{0}, {1}}}
	resolveChoose(g, obj, agent, game.ChooseFromZone{
		Player:         game.ControllerReference(),
		SourceZone:     zone.Library,
		Filter:         game.Selection{RequiredTypes: []types.Card{types.Land}},
		Quantity:       game.Fixed(2),
		Count:          game.ChooseUpTo,
		Grouping:       game.ChooseSplitDestination,
		Destination:    game.ChooseDestination{Zone: zone.Battlefield},
		SplitSecondary: opt.Val(game.ChooseSplitSlot{Destination: game.ChooseDestination{Zone: zone.Hand}}),
	})

	if !g.Players[game.Player1].Hand.Contains(only) {
		t.Fatal("the lone card should have gone to the chosen secondary slot")
	}
	if onBattlefieldByCard(g, only) {
		t.Fatal("the lone card should not have entered the battlefield")
	}
}

// TestChooseFromZoneBattlefieldEntersTappedUnderOwnerControl verifies the
// battlefield entry applies the tapped and owner-control riders.
func TestChooseFromZoneBattlefieldEntersTappedUnderOwnerControl(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obj := chooseFromZoneSource(g)
	mine := addCfzGraveyardCard(g, game.Player1, chooseFromZoneCardDef("Bear", 2, []types.Card{types.Creature}, nil))

	resolveChoose(g, obj, defaultChoiceAgent{}, game.ChooseFromZone{
		Player:      game.ControllerReference(),
		SourceZone:  zone.Graveyard,
		Filter:      game.Selection{RequiredTypes: []types.Card{types.Creature}},
		Quantity:    game.Fixed(1),
		Count:       game.ChooseExactly,
		Destination: game.ChooseDestination{Zone: zone.Battlefield},
		Riders:      game.ChooseRiders{EntersTapped: true, UnderOwnerControl: true},
	})

	controller, ok := battlefieldControllerByCard(g, mine)
	if !ok {
		t.Fatal("the chosen creature did not enter the battlefield")
	}
	if controller != game.Player1 {
		t.Fatalf("controller = %v, want owner Player1", controller)
	}
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == mine && !permanent.Tapped {
			t.Fatal("the creature should enter tapped")
		}
	}
}

// TestChooseFromZoneEntryCounters verifies the entry-counter rider places
// counters on a card entering the battlefield.
func TestChooseFromZoneEntryCounters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obj := chooseFromZoneSource(g)
	bear := addCfzGraveyardCard(g, game.Player1, chooseFromZoneCardDef("Bear", 2, []types.Card{types.Creature}, nil))

	resolveChoose(g, obj, defaultChoiceAgent{}, game.ChooseFromZone{
		Player:      game.ControllerReference(),
		SourceZone:  zone.Graveyard,
		Filter:      game.Selection{RequiredTypes: []types.Card{types.Creature}},
		Quantity:    game.Fixed(1),
		Count:       game.ChooseExactly,
		Destination: game.ChooseDestination{Zone: zone.Battlefield},
		Riders:      game.ChooseRiders{EntryCounters: []game.CounterPlacement{{Kind: counter.PlusOnePlusOne, Amount: 2}}},
	})

	found := false
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == bear {
			found = true
			if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 2 {
				t.Fatalf("+1/+1 counters = %d, want 2", got)
			}
		}
	}
	if !found {
		t.Fatal("the chosen creature did not enter the battlefield")
	}
}

// TestChooseFromZoneDestinationBottom verifies the destination-bottom rider puts
// a card on the bottom of the library.
func TestChooseFromZoneDestinationBottom(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obj := chooseFromZoneSource(g)
	existing := addCfzLibraryCard(g, game.Player1, chooseFromZoneCardDef("Top Card", 1, []types.Card{types.Instant}, nil))
	bear := addCfzGraveyardCard(g, game.Player1, chooseFromZoneCardDef("Bear", 2, []types.Card{types.Creature}, nil))

	resolveChoose(g, obj, defaultChoiceAgent{}, game.ChooseFromZone{
		Player:      game.ControllerReference(),
		SourceZone:  zone.Graveyard,
		Filter:      game.Selection{RequiredTypes: []types.Card{types.Creature}},
		Quantity:    game.Fixed(1),
		Count:       game.ChooseExactly,
		Destination: game.ChooseDestination{Zone: zone.Library},
		Riders:      game.ChooseRiders{DestinationBottom: true},
	})

	library := g.Players[game.Player1].Library.All()
	if len(library) != 2 || library[0] != existing || library[len(library)-1] != bear {
		t.Fatalf("library = %v, want bear on the bottom after %v", library, existing)
	}
}

// TestChooseFromZoneRevealEmitsRevealEvent verifies the reveal rider emits a
// card-revealed event for the chosen card.
func TestChooseFromZoneRevealEmitsRevealEvent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obj := chooseFromZoneSource(g)
	bear := addCfzGraveyardCard(g, game.Player1, chooseFromZoneCardDef("Bear", 2, []types.Card{types.Creature}, nil))

	resolveChoose(g, obj, defaultChoiceAgent{}, game.ChooseFromZone{
		Player:      game.ControllerReference(),
		SourceZone:  zone.Graveyard,
		Filter:      game.Selection{RequiredTypes: []types.Card{types.Creature}},
		Quantity:    game.Fixed(1),
		Count:       game.ChooseExactly,
		Destination: game.ChooseDestination{Zone: zone.Hand},
		Riders:      game.ChooseRiders{Reveal: true},
	})

	assertEvent(t, g.Events, game.EventCardRevealed, func(event game.Event) bool {
		return event.CardID == bear
	})
}

// TestChooseFromZoneFromLinkedRestrictsPool verifies the from-linked rider limits
// candidates to the previously remembered set.
func TestChooseFromZoneFromLinkedRestrictsPool(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obj := chooseFromZoneSource(g)
	linked := addCfzGraveyardCard(g, game.Player1, chooseFromZoneCardDef("Bear", 2, []types.Card{types.Creature}, nil))
	unlinked := addCfzGraveyardCard(g, game.Player1, chooseFromZoneCardDef("Ox", 3, []types.Card{types.Creature}, nil))

	key := linkedObjectSourceKey(g, obj, "milled")
	rememberLinkedObject(g, key, game.LinkedObjectRef{CardID: linked})

	resolveChoose(g, obj, defaultChoiceAgent{}, game.ChooseFromZone{
		Player:      game.ControllerReference(),
		SourceZone:  zone.Graveyard,
		Filter:      game.Selection{RequiredTypes: []types.Card{types.Creature}},
		Quantity:    game.Fixed(1),
		Count:       game.ChooseExactly,
		Destination: game.ChooseDestination{Zone: zone.Hand},
		Riders:      game.ChooseRiders{FromLinked: "milled"},
	})

	if !g.Players[game.Player1].Hand.Contains(linked) {
		t.Fatal("the linked card should have been an eligible candidate")
	}
	if !g.Players[game.Player1].Graveyard.Contains(unlinked) {
		t.Fatal("an unlinked card must not be an eligible candidate")
	}
}

// TestChooseFromZonePublishLinkedRemembersChosen verifies the publish-linked
// rider records the chosen cards under its key.
func TestChooseFromZonePublishLinkedRemembersChosen(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obj := chooseFromZoneSource(g)
	bear := addCfzGraveyardCard(g, game.Player1, chooseFromZoneCardDef("Bear", 2, []types.Card{types.Creature}, nil))

	resolveChoose(g, obj, defaultChoiceAgent{}, game.ChooseFromZone{
		Player:      game.ControllerReference(),
		SourceZone:  zone.Graveyard,
		Filter:      game.Selection{RequiredTypes: []types.Card{types.Creature}},
		Quantity:    game.Fixed(1),
		Count:       game.ChooseExactly,
		Destination: game.ChooseDestination{Zone: zone.Hand},
		Riders:      game.ChooseRiders{PublishLinked: "chosen"},
	})

	refs := linkedObjects(g, linkedObjectSourceKey(g, obj, "chosen"))
	if len(refs) != 1 || refs[0].CardID != bear {
		t.Fatalf("linked objects = %v, want one ref to %v", refs, bear)
	}
}

// TestChooseFromZoneMaxManaValueFromX verifies the X-bound rider restricts the
// candidate pool to cards whose mana value is at most the resolving spell's X.
func TestChooseFromZoneMaxManaValueFromX(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obj := chooseFromZoneSource(g)
	obj.XValue = 3
	withinX := addCfzLibraryCard(g, game.Player1, chooseFromZoneCardDef("Bear", 2, []types.Card{types.Creature}, nil))
	beyondX := addCfzLibraryCard(g, game.Player1, chooseFromZoneCardDef("Dragon", 6, []types.Card{types.Creature}, nil))

	resolveChoose(g, obj, defaultChoiceAgent{}, game.ChooseFromZone{
		Player:      game.ControllerReference(),
		SourceZone:  zone.Library,
		Filter:      game.Selection{RequiredTypes: []types.Card{types.Creature}},
		Quantity:    game.Fixed(1),
		Count:       game.ChooseExactly,
		Destination: game.ChooseDestination{Zone: zone.Hand},
		Riders:      game.ChooseRiders{MaxManaValueFromX: true},
	})

	if !g.Players[game.Player1].Hand.Contains(withinX) {
		t.Fatal("a card within the X bound should be eligible")
	}
	if !g.Players[game.Player1].Library.Contains(beyondX) {
		t.Fatal("a card above the X bound must not be eligible")
	}
}

// TestChooseFromZoneFaceDownEntersFaceDown verifies the face-down rider puts the
// chosen card onto the battlefield face down.
func TestChooseFromZoneFaceDownEntersFaceDown(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obj := chooseFromZoneSource(g)
	bear := addCfzLibraryCard(g, game.Player1, chooseFromZoneCardDef("Bear", 2, []types.Card{types.Creature}, nil))

	resolveChoose(g, obj, defaultChoiceAgent{}, game.ChooseFromZone{
		Player:      game.ControllerReference(),
		SourceZone:  zone.Library,
		Filter:      game.Selection{RequiredTypes: []types.Card{types.Creature}},
		Quantity:    game.Fixed(1),
		Count:       game.ChooseExactly,
		Destination: game.ChooseDestination{Zone: zone.Battlefield},
		Riders:      game.ChooseRiders{FaceDown: true, FaceDownKind: game.FaceDownManifest},
	})

	found := false
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == bear {
			found = true
			if !permanent.FaceDown {
				t.Fatal("the chosen card should have entered face down")
			}
		}
	}
	if !found {
		t.Fatal("the chosen card did not enter the battlefield")
	}
}

// TestChooseFromZoneNoMatchIsNoOp verifies an empty candidate pool leaves the
// source zone untouched.
func TestChooseFromZoneNoMatchIsNoOp(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obj := chooseFromZoneSource(g)
	instant := addCfzGraveyardCard(g, game.Player1, chooseFromZoneCardDef("Bolt", 1, []types.Card{types.Instant}, nil))

	res := resolveChoose(g, obj, defaultChoiceAgent{}, game.ChooseFromZone{
		Player:      game.ControllerReference(),
		SourceZone:  zone.Graveyard,
		Filter:      game.Selection{RequiredTypes: []types.Card{types.Creature}},
		Quantity:    game.Fixed(1),
		Count:       game.ChooseExactly,
		Destination: game.ChooseDestination{Zone: zone.Hand},
	})

	if res.succeeded {
		t.Fatal("an empty pool should not report success")
	}
	if !g.Players[game.Player1].Graveyard.Contains(instant) {
		t.Fatal("the non-matching card was disturbed")
	}
}
