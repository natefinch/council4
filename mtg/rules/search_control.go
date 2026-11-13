package rules

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

// searchControl separates the four roles the library-search subsystem must keep
// distinct when a player's search is directed by an opponent's static ability
// (Opposition Agent): the searcher whose library is searched, the decisionMaker
// who chooses which cards are found, whether every found card is redirected to
// its owner's exile, and the beneficiary who may afterward play those exiled
// cards. Without such an effect the decisionMaker is the searcher and no
// redirection applies, so the struct is inert for every ordinary search.
type searchControl struct {
	searcher      game.PlayerID
	decisionMaker game.PlayerID
	exileFinds    bool
	// exileSources lists every applicable exile-and-play effect, newest source
	// first. Each found card's owner chooses which one applies (see
	// chooseSearchExileSource); the newest is the default.
	exileSources []searchExileSource
}

// searchExileSource is one opponent's "they exile each card they find. You may
// play those cards ..." effect (Opposition Agent) that could apply to a find:
// the beneficiary who gains the play permission and the effect's source.
type searchExileSource struct {
	beneficiary    game.PlayerID
	sourceCardID   id.ID
	sourceObjectID id.ID
}

// Search control is a scoped override of the common decision policy, not an
// agent wrapper with a captured live observation. Ordinary searches still share
// the searcher's library visibility with their turn controller.
func beginSearchControl(g *game.Game, control searchControl) func() {
	depth := len(g.SearchControlScopes)
	g.SearchControlScopes = append(g.SearchControlScopes, game.SearchControlScope{
		Player: control.searcher, Controller: control.decisionMaker,
		Override: control.searcher != control.decisionMaker,
	})
	return func() { g.SearchControlScopes = g.SearchControlScopes[:depth] }
}

// resolveSearchControl inspects the active rule effects for an opponent of
// searcher who controls searcher's library searches
// (RuleEffectControlOpponentSearches) or exiles what searcher finds
// (RuleEffectExileOpponentSearchFinds), and reports how searcher's search is
// directed. When more than one opponent would control the search, the effect
// whose source most recently entered the battlefield wins (Opposition Agent
// ruling: the controller of the Agent that most recently entered the
// battlefield controls that player during its search). A permanent's
// battlefield-entry timestamp is its ObjectID, which increases monotonically, so
// the greatest SourceObjectID marks the newest source. Competing exile effects
// are not resolved here: each applies to every find, so every one is retained,
// newest first, and the found card's owner — controlled by the search's
// decision maker — chooses which applies to each card (exileFoundCard). It
// returns the identity control — decisionMaker equal to searcher and no
// redirection — when no such effect applies, which is the case for every search
// not opposed by an Opposition Agent. Effects controlled by the searcher never
// match, so the controller searching their own library is unaffected.
func resolveSearchControl(g *game.Game, searcher game.PlayerID) searchControl {
	control := searchControl{searcher: searcher, decisionMaker: searcher}
	effects := activeRuleEffects(g)
	var bestControl id.ID
	haveControl := false
	for i := range effects {
		effect := &effects[i]
		// An effect controlled by the searcher is their own Opposition Agent,
		// which only reaches opponents' searches, so it never directs this one.
		if effect.Controller == searcher {
			continue
		}
		switch effect.Kind {
		case game.RuleEffectControlOpponentSearches:
			if !haveControl || effect.SourceObjectID > bestControl {
				haveControl = true
				bestControl = effect.SourceObjectID
				control.decisionMaker = effect.Controller
			}
		case game.RuleEffectExileOpponentSearchFinds:
			source := searchExileSource{
				beneficiary:    effect.Controller,
				sourceCardID:   effect.SourceCardID,
				sourceObjectID: effect.SourceObjectID,
			}
			if !slices.Contains(control.exileSources, source) {
				control.exileSources = append(control.exileSources, source)
			}
		default:
			// Other rule effects never direct a library search.
		}
	}
	slices.SortStableFunc(control.exileSources, func(a, b searchExileSource) int {
		return cmp.Compare(b.sourceObjectID, a.sourceObjectID)
	})
	control.exileFinds = len(control.exileSources) > 0
	return control
}

// exileFoundCard redirects a found library card to its owner's exile instead of
// the search's normal destination and grants the chosen effect's beneficiary
// the lasting permission to play it, backing Opposition Agent's "they exile each
// card they find. You may play those cards ...". Only an actual exile arrival
// receives the play permission.
func (e *Engine) exileFoundCard(g *game.Game, obj *game.StackObject, agents [game.NumPlayers]PlayerAgent, log *TurnLog, control searchControl, ownerID game.PlayerID, cardID id.ID, fromZone zone.Type) bool {
	source := e.chooseSearchExileSource(g, agents, log, control, ownerID, cardID)
	destination, placed := moveSearchCard(g, obj, ownerID, cardID, fromZone, zone.Exile)
	if !placed {
		return false
	}
	if destination == zone.Exile {
		grantSearchExilePlayPermission(g, source, cardID)
	}
	return true
}

// chooseSearchExileSource picks which of several competing exile-and-play
// effects applies to one found card. Each Opposition Agent tries to exile the
// card and grant its controller permission to play it, so the card's owner
// chooses which effect wins; that owner is controlled during the search, so the
// request is the owner's and the search control scope routes it to the player
// controlling them (Opposition Agent ruling, 2020-11-10). The choice is made
// separately for every find and defaults to the newest source. The request
// names the found card, which is still in its hidden zone, so like a
// replacement choice for a card leaving a hidden zone (zoneChangeCardHidden) it
// is private: the log keeps the selection but withholds the card and labels.
func (e *Engine) chooseSearchExileSource(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog, control searchControl, ownerID game.PlayerID, cardID id.ID) searchExileSource {
	if len(control.exileSources) == 1 {
		return control.exileSources[0]
	}
	options := make([]game.ChoiceOption, len(control.exileSources))
	for i, source := range control.exileSources {
		option := game.ChoiceOption{Index: i}
		name := "Opposition effect"
		if permanent, ok := permanentByObjectID(g, source.sourceObjectID); ok {
			option.Card = permanentChoiceInfo(g, permanent)
			if option.Card.Exists {
				name = option.Card.Val.Name
			}
		}
		option.Label = fmt.Sprintf("%s controlled by Player %d", name, source.beneficiary+1)
		options[i] = option
	}
	selected := e.chooseChoice(g, agents, game.ChoiceRequest{
		Kind:             game.ChoiceReplacement,
		Player:           ownerID,
		Prompt:           "Choose which effect exiles the found card and lets its controller play it.",
		Options:          options,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{0},
		Subject:          cardChoiceInfo(g, cardID),
		Private:          true,
	}, log)
	if len(selected) == 1 && selected[0] >= 0 && selected[0] < len(control.exileSources) {
		return control.exileSources[selected[0]]
	}
	return control.exileSources[0]
}

// grantSearchExilePlayPermission registers the lasting per-card permission that
// lets the chosen exile source's beneficiary play a card exiled from a
// controlled search for as long as it remains exiled, spending mana as though it
// were mana of any color ("You may play those cards for as long as they remain
// exiled, and you may spend mana as though it were mana of any color to cast
// them.", Opposition Agent). It is a DurationPermanent RuleEffectPlayFromZone
// whose provenance points at the Opposition Agent source;
// persistsWhileCardExiled keeps it active while the card stays in exile even
// after that source leaves the battlefield or its controller changes (CR
// 610.3b). The permission is beneficiary-relative rather than owner-scoped, so
// foreignExileCastableCards surfaces the card — resting in its owner's exile
// bucket — to the beneficiary.
func grantSearchExilePlayPermission(g *game.Game, source searchExileSource, cardID id.ID) {
	if cardID == 0 {
		return
	}
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		ID:             g.IDGen.Next(),
		Kind:           game.RuleEffectPlayFromZone,
		Controller:     source.beneficiary,
		SourceCardID:   source.sourceCardID,
		SourceObjectID: source.sourceObjectID,
		AffectedPlayer: game.PlayerYou,
		Duration:       game.DurationPermanent,
		CreatedTurn:    g.Turn.TurnNumber,
		CastFromZone:   zone.Exile,
		AffectedCardID: cardID,
		SpendAnyMana:   game.ManaSpendAsAnyColor,
		ExpiresFor:     source.beneficiary,
	})
}
