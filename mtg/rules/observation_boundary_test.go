package rules

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

// addLibraryCard registers a card and places it on top of a player's library.
func addLibraryCard(g *game.Game, owner game.PlayerID, def *game.CardDef) id.ID {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{ID: cardID, Def: def, Owner: owner}
	g.Players[owner].Library.Add(cardID)
	return cardID
}

// reachableNames collects every card/permanent/stack name surfaced by the
// observation's public accessors, so a leak of a hidden card's identity shows up
// as its distinctive name appearing in the set.
func reachableNames(obs PlayerObservation) map[string]bool {
	names := make(map[string]bool)
	for _, card := range obs.Hand() {
		names[card.Name] = true
	}
	for i := range game.NumPlayers {
		playerID := game.PlayerID(i)
		for _, card := range obs.Graveyard(playerID) {
			names[card.Name] = true
		}
		for _, card := range obs.Exile(playerID) {
			names[card.Name] = true
		}
		for _, card := range obs.CommandZone(playerID) {
			names[card.Name] = true
		}
	}
	for _, permanent := range obs.Battlefield() {
		names[permanent.Name] = true
	}
	for _, stackObject := range obs.Stack() {
		names[stackObject.Name] = true
	}
	return names
}

func TestObservationOwnHandFullyVisible(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addHandCard(g, game.Player1, vanillaCreature("Own Card A", 1, 1))
	addHandCard(g, game.Player1, vanillaCreature("Own Card B", 2, 2))

	obs := observe(g, game.Player1)
	names := make(map[string]bool)
	for _, card := range obs.Hand() {
		names[card.Name] = true
	}
	if !names["Own Card A"] || !names["Own Card B"] {
		t.Errorf("own hand = %v, want both Own Card A and Own Card B visible", names)
	}
}

func TestObservationOpponentHandContentsNeverReachable(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addHandCard(g, game.Player2, vanillaCreature("Opponent Secret", 5, 5))
	addHandCard(g, game.Player3, vanillaCreature("Another Secret", 6, 6))

	obs := observe(g, game.Player1)
	names := reachableNames(obs)
	if names["Opponent Secret"] || names["Another Secret"] {
		t.Errorf("opponent hand contents leaked through an accessor: %v", names)
	}
	// The opponents' hands are still observable as counts.
	if obs.PlayerState(game.Player2).HandSize != 1 || obs.PlayerState(game.Player3).HandSize != 1 {
		t.Error("opponent hand sizes should be observable")
	}
}

func TestObservationLibraryContentsAndOrderHidden(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// Distinctively named cards in libraries — including the observer's own —
	// must never surface through any accessor; only sizes are public.
	addLibraryCard(g, game.Player1, vanillaCreature("Own Library Top", 1, 1))
	addLibraryCard(g, game.Player1, vanillaCreature("Own Library Next", 1, 1))
	addLibraryCard(g, game.Player2, vanillaCreature("Opponent Library Card", 1, 1))

	obs := observe(g, game.Player1)
	names := reachableNames(obs)
	for _, hidden := range []string{"Own Library Top", "Own Library Next", "Opponent Library Card"} {
		if names[hidden] {
			t.Errorf("library card %q leaked through an accessor", hidden)
		}
	}
	if obs.PlayerState(game.Player1).LibrarySize != 2 {
		t.Errorf("own LibrarySize = %d, want 2", obs.PlayerState(game.Player1).LibrarySize)
	}
	if obs.PlayerState(game.Player2).LibrarySize != 1 {
		t.Errorf("opponent LibrarySize = %d, want 1", obs.PlayerState(game.Player2).LibrarySize)
	}
}

func TestObservationFaceDownPermanentExposesOnlyPublicInfo(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	morph := addFaceDownPermanent(g, game.Player2, vanillaCreature("Secret Identity", 9, 9, game.Flying, game.Trample), game.FaceDownMorph)

	obs := observe(g, game.Player1)
	if reachableNames(obs)["Secret Identity"] {
		t.Error("face-down permanent leaked its hidden name")
	}
	view := findPermanentView(t, obs, morph.ObjectID)
	if view.Power != 2 || view.Toughness != 2 {
		t.Errorf("face-down P/T = %d/%d, want public 2/2", view.Power, view.Toughness)
	}
	if view.HasKeyword(game.Flying) || view.HasKeyword(game.Trample) {
		t.Error("face-down permanent leaked hidden keywords")
	}
}

func TestObservationFaceDownExiledCardHidden(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// A face-down card in exile (e.g. foretold/suspended) is registered but its
	// identity is hidden.
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{ID: cardID, Def: vanillaCreature("Foretold Secret", 4, 4), Owner: game.Player2}
	g.Players[game.Player2].Exile.Add(cardID)
	g.Players[game.Player2].Exile.SetFaceDown(cardID, true)

	obs := observe(g, game.Player1)
	if reachableNames(obs)["Foretold Secret"] {
		t.Error("face-down exiled card leaked its hidden name")
	}
	// It is still present as an opaque entry (instance ID only, no name).
	exile := obs.Exile(game.Player2)
	if len(exile) != 1 || exile[0].Name != "" || exile[0].CardInstanceID != cardID {
		t.Errorf("face-down exile entry = %+v, want one opaque entry with the instance ID", exile)
	}
}

func TestObservationReturnedDataIsCopied(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addHandCard(g, game.Player1, vanillaCreature("My Card", 1, 1))
	addCombatPermanent(g, game.Player1, vanillaCreature("My Creature", 2, 2))

	obs := observe(g, game.Player1)

	// Mutating returned view slices must not affect game state.
	hand := obs.Hand()
	hand[0].Name = "Tampered"
	if obs.Hand()[0].Name != "My Card" {
		t.Error("mutating a returned Hand() view changed observed state")
	}

	battlefield := obs.Battlefield()
	if len(battlefield[0].Types) > 0 {
		battlefield[0].Types[0] = "Tampered"
		if obs.Battlefield()[0].Types[0] == "Tampered" {
			t.Error("mutating a returned PermanentView.Types changed observed state")
		}
	}
}

// TestObservationPublicAPINeverExposesRawGame is a regression guard: no exported
// field or method of PlayerObservation may hand back the live *game.Game, which
// would let an agent bypass the fog-of-war boundary entirely.
func TestObservationPublicAPINeverExposesRawGame(t *testing.T) {
	gameType := reflect.TypeFor[game.Game]()
	obsType := reflect.TypeFor[PlayerObservation]()

	for field := range obsType.Fields() {
		if field.IsExported() && pointsToGame(field.Type, gameType) {
			t.Errorf("exported field %q exposes *game.Game", field.Name)
		}
	}
	// Inspect the pointer type's method set, which is a superset that includes
	// both value- and pointer-receiver methods, so a pointer-receiver leak is
	// caught too.
	for method := range reflect.TypeFor[*PlayerObservation]().Methods() {
		for out := range method.Func.Type().Outs() {
			if pointsToGame(out, gameType) {
				t.Errorf("exported method %q returns *game.Game", method.Name)
			}
		}
	}
}

func pointsToGame(t, gameType reflect.Type) bool {
	return t.Kind() == reflect.Pointer && t.Elem() == gameType
}
