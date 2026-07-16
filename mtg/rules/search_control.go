package rules

import (
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
	searcher       game.PlayerID
	decisionMaker  game.PlayerID
	exileFinds     bool
	beneficiary    game.PlayerID
	sourceCardID   id.ID
	sourceObjectID id.ID
}

// resolveSearchControl inspects the active rule effects for an opponent of
// searcher who controls searcher's library searches
// (RuleEffectControlOpponentSearches) or exiles what searcher finds
// (RuleEffectExileOpponentSearchFinds), and reports how searcher's search is
// directed. When more than one opponent would control the search, the effect
// whose source most recently entered the battlefield wins (CR: the controller of
// the Opposition Agent that most recently entered the battlefield controls that
// player during its search). A permanent's battlefield-entry timestamp is its
// ObjectID, which increases monotonically, so the greatest SourceObjectID marks
// the newest source; selecting one fixed source keeps the result deterministic
// and reproducible across identical states. The control and exile roles are
// resolved independently, so distinct sources can supply each, though a single
// Opposition Agent supplies both. It returns the identity control — decisionMaker
// equal to searcher and no redirection — when no such effect applies, which is the
// case for every search not opposed by an Opposition Agent. Effects controlled by
// the searcher never match, so the controller searching their own library is
// unaffected.
func resolveSearchControl(g *game.Game, searcher game.PlayerID) searchControl {
	control := searchControl{searcher: searcher, decisionMaker: searcher}
	effects := activeRuleEffects(g)
	var bestControl, bestExile id.ID
	haveControl, haveExile := false, false
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
			if !haveExile || effect.SourceObjectID > bestExile {
				haveExile = true
				bestExile = effect.SourceObjectID
				control.exileFinds = true
				control.beneficiary = effect.Controller
				control.sourceCardID = effect.SourceCardID
				control.sourceObjectID = effect.SourceObjectID
			}
		default:
			// Other rule effects never direct a library search.
		}
	}
	return control
}

// exileFoundCard redirects a found library card to its owner's exile instead of
// the search's normal destination and grants the controlling beneficiary the
// lasting permission to play it, backing Opposition Agent's "they exile each card
// they find. You may play those cards ...". The card has already been removed
// from its owner's library; owner is the searching player, who owns every card in
// their own library, so the redirected card lands in the searching player's exile
// bucket.
func exileFoundCard(g *game.Game, obj *game.StackObject, control searchControl, owner *game.Player, ownerID game.PlayerID, cardID id.ID, fromZone zone.Type) {
	owner.Exile.Add(cardID)
	emitZoneChangeEvent(g, game.Event{
		SourceID:      stackObjectSourceID(obj),
		StackObjectID: stackObjectID(obj),
		Controller:    stackObjectController(obj),
		Player:        ownerID,
		CardID:        cardID,
		FromZone:      fromZone,
		ToZone:        zone.Exile,
		Amount:        1,
	})
	grantSearchExilePlayPermission(g, control, cardID)
}

// grantSearchExilePlayPermission registers the lasting per-card permission that
// lets a searchControl's beneficiary play a card exiled from a controlled search
// for as long as it remains exiled, spending mana as though it were mana of any
// color ("You may play those cards for as long as they remain exiled, and you may
// spend mana as though it were mana of any color to cast them.", Opposition
// Agent). It is a DurationPermanent RuleEffectPlayFromZone whose provenance points
// at the Opposition Agent source; persistsWhileCardExiled keeps it active while
// the card stays in exile even after that source leaves the battlefield or its
// controller changes (CR 610.3b). The permission is beneficiary-relative rather
// than owner-scoped, so foreignExileCastableCards surfaces the card — resting in
// its owner's exile bucket — to the beneficiary.
func grantSearchExilePlayPermission(g *game.Game, control searchControl, cardID id.ID) {
	if cardID == 0 {
		return
	}
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		ID:             g.IDGen.Next(),
		Kind:           game.RuleEffectPlayFromZone,
		Controller:     control.beneficiary,
		SourceCardID:   control.sourceCardID,
		SourceObjectID: control.sourceObjectID,
		AffectedPlayer: game.PlayerYou,
		Duration:       game.DurationPermanent,
		CreatedTurn:    g.Turn.TurnNumber,
		CastFromZone:   zone.Exile,
		AffectedCardID: cardID,
		SpendAnyMana:   true,
		ExpiresFor:     control.beneficiary,
	})
}
