package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// searchLibraryAndGraveyardChoice resolves the general "Search your library
// and/or graveyard for <filter> and put it onto the battlefield" search (Finale
// of Devastation). The searching player chooses which of their library and
// graveyard to search, finds a single card matching the spec's filter in a
// searched zone, and puts it at the spec's destination under controllerID's
// control. Unlike the folded named-tutor searchLibraryAndGraveyard it does not
// shuffle the library itself — it reports whether the library was searched so a
// following ShuffleLibrary gated on that result performs the "If you search your
// library this way, shuffle." step. It returns whether a card was placed and
// whether the library was among the searched zones.
func (e *Engine) searchLibraryAndGraveyardChoice(g *game.Game, obj *game.StackObject, agents [game.NumPlayers]PlayerAgent, log *TurnLog, playerID, controllerID game.PlayerID, spec game.SearchSpec) (placed, searchedLibrary bool) {
	player, ok := playerByID(g, playerID)
	if !ok {
		return false, false
	}
	control := resolveSearchControl(g, playerID)
	if spec.MaxManaValueFromX {
		// "with mana value X or less" bounds the search by the spell's chosen X,
		// resolved from the resolving stack object as the search runs.
		spec.Filter.ManaValue = opt.Val(compare.Int{Op: compare.LessOrEqual, Value: obj.XValue})
		spec.MaxManaValueFromX = false
	}
	searchLibrary, searchGraveyard := e.chooseSearchZones(g, agents, log, control.decisionMaker)
	if searchLibrary {
		// Searching the library fires the search event once (CR 701.19a); the
		// closing shuffle is a separate instruction gated on the reported result.
		emitEvent(g, game.Event{
			Kind:       game.EventLibrarySearched,
			Controller: playerID,
			Player:     playerID,
		})
	}
	type multiZoneCandidate struct {
		cardID   id.ID
		fromZone zone.Type
	}
	var candidates []multiZoneCandidate
	if searchLibrary {
		for _, cardID := range player.Library.All() {
			if searchSpecMatches(g, obj, cardID, spec) {
				candidates = append(candidates, multiZoneCandidate{cardID: cardID, fromZone: zone.Library})
			}
		}
	}
	if searchGraveyard {
		for _, cardID := range player.Graveyard.All() {
			if searchSpecMatches(g, obj, cardID, spec) {
				candidates = append(candidates, multiZoneCandidate{cardID: cardID, fromZone: zone.Graveyard})
			}
		}
	}
	// Searching a hidden zone (the library) permits declining to find even when
	// matches exist (CR 701.19e); a public-zone-only search (the graveyard) must
	// find a legal card when one exists.
	minChoices := 0
	if !searchLibrary && len(candidates) > 0 {
		minChoices = 1
	}
	candidateIDs := make([]id.ID, len(candidates))
	for i := range candidates {
		candidateIDs[i] = candidates[i].cardID
	}
	found := e.chooseSearchMatches(g, agents, log, control.decisionMaker, candidateIDs, 1, minChoices)
	if len(found) == 0 {
		return false, searchLibrary
	}
	cardID := found[0]
	fromZone := zone.Library
	for i := range candidates {
		if candidates[i].cardID == cardID {
			fromZone = candidates[i].fromZone
			break
		}
	}
	switch fromZone {
	case zone.Graveyard:
		if !player.Graveyard.Remove(cardID) {
			return false, searchLibrary
		}
	default:
		if !player.Library.Remove(cardID) {
			return false, searchLibrary
		}
	}
	dest := game.SearchDestination{Zone: spec.Destination, EntersTapped: spec.EntersTapped}
	_, placed = e.placeFoundCard(g, obj, playerID, controllerID, player, cardID, dest, fromZone, control)
	return placed, searchLibrary
}

// chooseSearchZones asks the deciding player which of the searcher's library and
// graveyard to search for the "search your library and/or graveyard" form. The
// decider is the searcher normally, or the opponent controlling the search
// (Opposition Agent). At least one zone must be chosen; the deterministic
// fallback searches both.
func (e *Engine) chooseSearchZones(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog, decider game.PlayerID) (searchLibrary, searchGraveyard bool) {
	const (
		libraryOption   = 0
		graveyardOption = 1
	)
	request := game.ChoiceRequest{
		Kind:   game.ChoiceZoneSelection,
		Player: decider,
		Prompt: "Choose which zones to search.",
		Options: []game.ChoiceOption{
			{Index: libraryOption, Label: "Your library"},
			{Index: graveyardOption, Label: "Your graveyard"},
		},
		MinChoices:       1,
		MaxChoices:       2,
		DefaultSelection: []int{libraryOption, graveyardOption},
	}
	selected := e.chooseChoice(g, agents, request, log)
	for _, idx := range selected {
		switch idx {
		case libraryOption:
			searchLibrary = true
		case graveyardOption:
			searchGraveyard = true
		default:
			// The request only offers the library and graveyard options, so any
			// other index is impossible; ignore it defensively.
		}
	}
	if !searchLibrary && !searchGraveyard {
		// A valid selection always names at least one zone; guard a degenerate
		// answer by searching both rather than searching nothing.
		return true, true
	}
	return searchLibrary, searchGraveyard
}
