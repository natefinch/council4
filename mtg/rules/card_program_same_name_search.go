package rules

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

// sameNameCandidate is one card a same-name search can find.
type sameNameCandidate struct {
	cardID id.ID
	from   zone.Type
}

// SearchSameName has the source's controller search one player's graveyard,
// hand and library for cards sharing a name with the reference (CR 201.2a,
// 701.23). The searcher looks at every card in the searched hand and library
// throughout (CR 701.23a), privately, and always gets the search decision, even
// when nothing optional matches. The found cards stay where they are,
// unrevealed (CR 701.23e), and are returned as origin-tagged handles for one
// later MoveCards placement.
//
// A search of the searcher's own zones is a search of their own library: it
// fires the library-searched event and an opponent's Opposition Agent controls
// the whole multi-zone search and exiles every find (Opposition Agent ruling
// 2020-11-10), so those finds are not returned. Searching another player's
// library is not that player searching their own library, so it fires no
// library-searched event and is not subject to their opponents' search control.
func (c *programContext) SearchSameName(search cardprogram.SameNameSearch) (cardprogram.SameNameResult, error) {
	if err := c.valid(); err != nil {
		return cardprogram.SameNameResult{}, err
	}
	if err := validateSameNameSearch(search); err != nil {
		return cardprogram.SameNameResult{}, err
	}
	if problems := search.Player.Validate(); len(problems) != 0 {
		return cardprogram.SameNameResult{}, fmt.Errorf("invalid same-name search player: %s", problems[0])
	}
	player, available := resolvePlayerReference(c.game, c.obj, search.Player)
	if !available || !isPlayerAlive(c.game, player) {
		return cardprogram.SameNameResult{}, nil
	}
	names, err := c.sameNameReferenceNames(search.Name)
	if err != nil {
		return cardprogram.SameNameResult{}, err
	}
	result := cardprogram.SameNameResult{Available: true, Player: player}
	searcher := c.obj.Controller
	own := searcher == player
	endLook := beginHiddenZoneSearch(c.game, searcher, player)
	defer endLook()
	control := searchControl{searcher: searcher, decisionMaker: searcher}
	if own {
		control = resolveSearchControl(c.game, player)
		endSearch := beginSearchControl(c.game, control)
		defer endSearch()
		emitEvent(c.game, game.Event{Kind: game.EventLibrarySearched, Controller: player, Player: player})
		c.engine.offerSearchCasts(c.game, player, c.agents, c.log)
	}
	var forced, optional []sameNameCandidate
	for _, from := range []zone.Type{zone.Graveyard, zone.Hand, zone.Library} {
		source, _ := destinationZone(c.game, player, from)
		for _, cardID := range source.All() {
			card, exists := c.game.GetCardInstance(cardID)
			if !exists || !zoneCardHasAnyName(card, names) {
				continue
			}
			candidate := sameNameCandidate{cardID: cardID, from: from}
			if from == zone.Graveyard && search.Quantity == cardprogram.SameNameAll {
				forced = append(forced, candidate)
			} else {
				optional = append(optional, candidate)
			}
		}
	}
	found := append([]sameNameCandidate(nil), forced...)
	found = append(found, c.chooseSameNameFinds(search, searcher, optional)...)
	if control.exileFinds {
		for _, candidate := range found {
			_ = c.engine.exileFoundCard(c.game, c.obj, c.agents, c.log, control, player, candidate.cardID, candidate.from)
		}
		return result, nil
	}
	for _, candidate := range found {
		card, _ := c.game.GetCardInstance(candidate.cardID)
		result.Cards = append(result.Cards, c.mintFoundCard(card, candidate.from))
	}
	return result, nil
}

func validateSameNameSearch(search cardprogram.SameNameSearch) error {
	references := 0
	for _, set := range []bool{search.Name.Snapshot.Exists, search.Name.Spell.Exists, search.Name.Card.Exists, search.Name.Name != ""} {
		if set {
			references++
		}
	}
	if references != 1 {
		return errors.New("same-name search requires exactly one name reference")
	}
	switch search.Quantity {
	case cardprogram.SameNameAll, cardprogram.SameNameAnyNumber:
		if search.Max != 0 {
			return errors.New("same-name search Max applies only to a counted quantity")
		}
	case cardprogram.SameNameUpTo, cardprogram.SameNameCards:
		if search.Max <= 0 {
			return errors.New("counted same-name search requires a positive Max")
		}
	default:
		return errors.New("same-name search has an unknown quantity")
	}
	return nil
}

// sameNameReferenceNames resolves the names a same-name search compares
// against. An empty result is a known nameless reference, which matches
// nothing (CR 201.2a).
func (c *programContext) sameNameReferenceNames(reference cardprogram.SameName) ([]string, error) {
	switch {
	case reference.Snapshot.Exists:
		return splitCardNames(reference.Snapshot.Val.Name), nil
	case reference.Spell.Exists:
		if reference.Spell.Val.Kind() != game.ObjectReferenceTargetStackObject {
			return nil, errors.New("same-name spell reference requires a target stack object")
		}
		if problems := reference.Spell.Val.Validate(); len(problems) != 0 {
			return nil, errors.New(problems[0])
		}
		// A spell that was countered was captured with its cast-face name; one
		// that could not be countered is still on the stack.
		name, _ := iterativeStopReferenceName(c.game, c.obj, reference.Spell.Val)
		return splitCardNames(name), nil
	case reference.Card.Exists:
		card, exists := c.game.GetCardInstance(reference.Card.Val.Identity.CardID)
		if !exists {
			return nil, errors.New("same-name card reference names an unknown card")
		}
		return zoneCardNames(card), nil
	default:
		return splitCardNames(reference.Name), nil
	}
}

// splitCardNames returns the names in a single object name. A fused split
// spell's name joins both halves' names (game.CardDef.fusedFace), and that spell
// has both names.
func splitCardNames(name string) []string {
	var names []string
	for _, part := range strings.Split(name, " // ") {
		if part != "" {
			names = append(names, part)
		}
	}
	return names
}

// zoneCardNames returns a card's names in a zone other than the battlefield
// or stack: both halves of a split card (CR 709.4, 709.4a), and otherwise only
// the front face's name (a double-faced card, CR 712.6a; an adventurer card,
// CR 716.4).
func zoneCardNames(card *game.CardInstance) []string {
	if card.Def.Layout == game.LayoutSplit && card.Def.Alternate.Exists {
		return splitCardNames(card.Def.Name + " // " + card.Def.Alternate.Val.Name)
	}
	return splitCardNames(cardFaceOrDefault(card, game.FaceFront).Name)
}

func zoneCardHasAnyName(card *game.CardInstance, names []string) bool {
	return slices.ContainsFunc(zoneCardNames(card), func(name string) bool {
		return slices.Contains(names, name)
	})
}

// chooseSameNameFinds asks the searcher which optional matches to find, with
// no options when none match: the searcher still looks at the searched zones.
// The options show the cards in hidden zones, so the decision stays out of the
// public log entirely (CR 400.2): how many options it had would reveal the
// searched hand and library.
func (c *programContext) chooseSameNameFinds(search cardprogram.SameNameSearch, searcher game.PlayerID, candidates []sameNameCandidate) []sameNameCandidate {
	maxChoices, minChoices := len(candidates), 0
	switch search.Quantity {
	case cardprogram.SameNameUpTo:
		maxChoices = min(search.Max, len(candidates))
	case cardprogram.SameNameCards:
		maxChoices = min(search.Max, len(candidates))
		public := 0
		for _, candidate := range candidates {
			if candidate.from == zone.Graveyard {
				public++
			}
		}
		minChoices = min(maxChoices, public)
	case cardprogram.SameNameAll, cardprogram.SameNameAnyNumber:
	}
	options := make([]game.ChoiceOption, len(candidates))
	for index, candidate := range candidates {
		info := cardChoiceInfo(c.game, candidate.cardID)
		options[index] = game.ChoiceOption{
			Index: index,
			Label: fmt.Sprintf("%s (%s)", info.Val.Name, strings.ToLower(candidate.from.String())),
			Card:  info,
		}
	}
	// Candidates are in graveyard, hand, library order, so the default
	// satisfies a graveyard minimum.
	request := game.ChoiceRequest{
		Kind:             game.ChoiceSearch,
		Player:           searcher,
		Prompt:           "Search the graveyard, hand and library: choose cards with the same name to find.",
		Options:          options,
		MinChoices:       minChoices,
		MaxChoices:       maxChoices,
		DefaultSelection: firstChoiceIndices(maxChoices),
		Private:          true,
	}
	selected := c.engine.chooseUnloggedChoice(c.game, c.agents, request, c.log)
	found := make([]sameNameCandidate, 0, len(selected))
	for _, index := range selected {
		found = append(found, candidates[index])
	}
	return found
}

// beginHiddenZoneSearch lets searcher look at every card in player's hand and
// library until the returned function ends the search (CR 701.23a). It grants
// no control of player's decisions.
func beginHiddenZoneSearch(g *game.Game, searcher, player game.PlayerID) func() {
	depth := len(g.HiddenZoneSearchScopes)
	g.HiddenZoneSearchScopes = append(g.HiddenZoneSearchScopes, game.HiddenZoneSearchScope{Searcher: searcher, Player: player})
	return func() { g.HiddenZoneSearchScopes = g.HiddenZoneSearchScopes[:depth] }
}
