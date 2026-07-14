package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

// distributiveExileSagaDef is a minimal stand-in for the Vault 13: Dweller's
// Journey source permanent: a legendary Saga enchantment whose chapters drive
// the distributive exile and the linked partial return.
func distributiveExileSagaDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       "Dweller's Journey",
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Enchantment},
		Subtypes:   []types.Sub{types.Sub("Saga")},
	}}
}

// TestExileForEachPlayerExilesOnePerPlayerUnderLink verifies the distributive
// Saga exile: each player's one matching permanent is exiled under the
// exile-until-leaves link, with no prompt when a player controls a single
// eligible permanent.
func TestExileForEachPlayerExilesOnePerPlayerUnderLink(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	mine := addCombatCreaturePermanent(g, game.Player1)
	theirs := addCombatCreaturePermanent(g, game.Player2)
	source := addCombatPermanent(g, game.Player1, distributiveExileSagaDef())
	obj := linkedSourceObject(source)

	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.ExileForEachPlayer{
		Chooser:   game.ControllerReference(),
		Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludeSource: true},
		LinkedKey: game.LinkedKey("exile-until-leaves"),
	}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if permanentByCardID(g, mine.CardInstanceID) != nil || permanentByCardID(g, theirs.CardInstanceID) != nil {
		t.Fatal("a chosen permanent remained on the battlefield after distributive exile")
	}
	if !g.Players[game.Player1].Exile.Contains(mine.CardInstanceID) {
		t.Fatal("Player1's creature did not reach its owner's exile zone")
	}
	if !g.Players[game.Player2].Exile.Contains(theirs.CardInstanceID) {
		t.Fatal("Player2's creature did not reach its owner's exile zone")
	}
	key := linkedObjectSourceKey(g, obj, "exile-until-leaves")
	if got := len(linkedObjects(g, key)); got != 2 {
		t.Fatalf("linked exiled objects = %d, want 2 (one per player)", got)
	}
}

// TestReturnLinkedExiledCardsToBattlefieldPartial verifies the chapter payoff:
// the controller returns exactly the chosen count of linked exiled cards to the
// battlefield under their owners' control and the remainder goes to the bottom
// of its owner's library, clearing the link.
func TestReturnLinkedExiledCardsToBattlefieldPartial(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatCreaturePermanent(g, game.Player1)
	addCombatCreaturePermanent(g, game.Player2)
	source := addCombatPermanent(g, game.Player1, distributiveExileSagaDef())
	obj := linkedSourceObject(source)

	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.ExileForEachPlayer{
		Chooser:   game.ControllerReference(),
		Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludeSource: true},
		LinkedKey: game.LinkedKey("exile-until-leaves"),
	}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	key := linkedObjectSourceKey(g, obj, "exile-until-leaves")
	live := liveLinkedExiledCards(g, key)
	if len(live) != 2 {
		t.Fatalf("live linked exiled cards = %d, want 2", len(live))
	}
	returnedCard := live[0].card.ID
	bottomedCard := live[1].card.ID
	bottomedOwner := live[1].owner

	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}},
	}
	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.ReturnLinkedExiledCardsToBattlefield{
		Chooser:             game.ControllerReference(),
		LinkedKey:           game.LinkedKey("exile-until-leaves"),
		Amount:              game.Fixed(1),
		RestToLibraryBottom: true,
	}}, agents, &TurnLog{})

	if permanentByCardID(g, returnedCard) == nil {
		t.Fatal("the chosen card was not returned to the battlefield")
	}
	if !bottomedOwner.Library.Contains(bottomedCard) {
		t.Fatal("the unreturned card was not put on the bottom of its owner's library")
	}
	if bottomedOwner.Exile.Contains(bottomedCard) {
		t.Fatal("the unreturned card remained in exile")
	}
	if got := len(linkedObjects(g, key)); got != 0 {
		t.Fatalf("linked objects after return = %d, want 0 (link cleared)", got)
	}
}

// TestManifestForEachLinkedCloaksPerRemovedController verifies the generic
// linked payoff used by Unexplained Absence: each removed permanent's last-known
// controller cloaks one card, including when the removed permanent was a token.
func TestManifestForEachLinkedCloaksPerRemovedController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTokenCreaturePermanent(g, game.Player1, "Citizen")
	addCombatCreaturePermanent(g, game.Player2)
	source := addCombatPermanent(g, game.Player1, distributiveExileSagaDef())
	obj := linkedSourceObject(source)
	player1Top := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Player 1 Top"}})
	player2Top := addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Player 2 Top"}})
	const link = game.LinkedKey("exiled-for-cloak")

	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.ExileForEachPlayer{
		Chooser:   game.ControllerReference(),
		Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludeSource: true},
		LinkedKey: link,
	}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.ManifestForEachLinked{
		Cloak:     true,
		LinkedKey: link,
	}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	for _, test := range []struct {
		cardID     id.ID
		controller game.PlayerID
	}{
		{player1Top, game.Player1},
		{player2Top, game.Player2},
	} {
		permanent := permanentByCardID(g, test.cardID)
		if permanent == nil ||
			permanent.Controller != test.controller ||
			!permanent.FaceDown ||
			permanent.FaceDownKind != game.FaceDownCloak {
			t.Fatalf("cloaked permanent for %v = %#v", test.controller, permanent)
		}
	}
	key := linkedObjectSourceKey(g, obj, string(link))
	if got := len(linkedObjects(g, key)); got != 0 {
		t.Fatalf("linked objects after cloak = %d, want 0", got)
	}
}
