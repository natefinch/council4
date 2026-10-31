package rules

import (
	"math/rand/v2"
	"slices"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// cascadeTriggeredAbility is the "When you cast this spell" triggered ability a
// cascade instance creates (CR 702.85a). It carries the Cascade keyword so
// resolution dispatches to the shared cascade operation, and it has no targets,
// so it can be countered (Stifle) or copied like any triggered ability.
func cascadeTriggeredAbility() *game.TriggeredAbility {
	return &game.TriggeredAbility{
		Text:             "Cascade",
		KeywordAbilities: game.SimpleKeywords(game.Cascade),
	}
}

// spellCascadeInstancesAtCast counts the cascade instances a spell has at the
// moment it becomes cast (CR 601.2i), each of which triggers separately (CR
// 702.85c):
//
//   - each cascade keyword on a static ability of the spell's face that works
//     on the stack (no ZoneOfFunction, or zone.Stack) and whose Condition holds
//     for the caster now ("This spell has cascade as long as ...", Bloodbraid
//     Marauder), so "Cascade, cascade" yields two;
//   - each active RuleEffectGrantSpellKeyword granting Cascade that matches the
//     spell, its caster, its card's commander designation and the zone it was
//     cast from ("Sliver spells you cast have cascade.", The First Sliver).
//
// A face-down spell has no printed abilities (CR 708.2) but can still be granted
// cascade, matched against its face-down characteristics (CR 708.4). The count
// is recorded once on the spell-cast event (Event.CascadeInstances) rather than
// on the stack object, so stack copies never inherit it and later changes to the
// grants do not alter it.
func spellCascadeInstancesAtCast(g *game.Game, event game.Event) int {
	return spellKeywordInstancesAtCast(g, event, game.Cascade)
}

// spellKeywordInstancesAtCast counts the instances of a "when you cast this
// spell" keyword trigger (cascade, storm) the cast spell has as it becomes cast,
// from the spell's own statics that work on the stack and from active spell
// keyword grants, as documented on spellCascadeInstancesAtCast.
func spellKeywordInstancesAtCast(g *game.Game, event game.Event, keyword game.Keyword) int {
	if event.Kind != game.EventSpellCast {
		return 0
	}
	var def *game.CardDef
	cardID := event.CardID
	if event.CardID == 0 && event.TokenDef != nil {
		def, _ = event.TokenDef.FaceDef(event.Face)
	} else if card, ok := g.GetCardInstance(event.CardID); ok {
		def, _ = cardFaceDef(card, event.Face)
	}
	if def == nil {
		return 0
	}
	obj, onStack := stackObjectByID(g, event.StackObjectID)
	if !onStack {
		obj = nil
	} else if obj.Copy {
		// A cast copy is not the commander card (CR 903.3, 707.12).
		cardID = 0
	}
	count := 0
	grantDef := def
	if obj != nil && obj.FaceDown {
		grantDef = faceDownSpellDef()
	} else {
		ctx := conditionContext{controller: event.Controller, sourceObjectID: event.StackObjectID, obj: obj}
		for i := range def.StaticAbilities {
			static := &def.StaticAbilities[i]
			if static.ZoneOfFunction != zone.None && static.ZoneOfFunction != zone.Stack {
				continue
			}
			instances := 0
			for _, ability := range game.BodyKeywordAbilities(static) {
				if game.KeywordAbilityKind(ability) == keyword {
					instances++
				}
			}
			if instances > 0 && conditionSatisfied(g, ctx, static.Condition) {
				count += instances
			}
		}
	}
	for _, effect := range activeRuleEffects(g) {
		if effect.Kind == game.RuleEffectGrantSpellKeyword &&
			effect.GrantedKeyword == keyword &&
			spellKeywordGrantMatches(g, &effect, event.Controller, grantDef, cardID, event.FromZone) {
			count++
		}
	}
	return count
}

// eventCascadeInstances returns the cascade instances a spell had when it
// became cast. Emitted spell-cast events carry the recorded count; a raw event
// that bypassed emission is evaluated now. Any other event, including a
// non-cast spell copy, has none.
func eventCascadeInstances(g *game.Game, event game.Event) int {
	if event.Kind != game.EventSpellCast {
		return 0
	}
	if event.CascadeInstances.Exists {
		return event.CascadeInstances.Val
	}
	return spellCascadeInstancesAtCast(g, event)
}

// cascadeCastTriggers returns one pending cascade trigger for each cascade
// instance a genuinely cast spell had as it became cast (CR 702.85a/c). Only
// EventSpellCast reaches here, so non-cast stack and storm copies never cascade
// (CR 707.10), while a cast copy of a card does (CR 707.12). The trigger keeps
// the cast event, whose ManaValue snapshots the spell as it was cast on the
// stack, including X; resolution compares against that value even if the spell
// has since left the stack (CR 608.2h).
func cascadeCastTriggers(g *game.Game, source *game.Permanent, event game.Event) []pendingTriggeredAbility {
	count := eventCascadeInstances(g, event)
	if count == 0 {
		return nil
	}
	pending := make([]pendingTriggeredAbility, 0, count)
	for range count {
		pending = append(pending, pendingTriggeredAbility{
			controller:      event.Controller,
			sourceID:        source.ObjectID,
			sourceCardID:    source.CardInstanceID,
			sourceToken:     source.TokenDef,
			face:            event.Face,
			inline:          cascadeTriggeredAbility(),
			event:           event,
			hasEvent:        true,
			ordinaryTrigger: true,
		})
	}
	return pending
}

// resolveCascadeTriggeredAbility resolves one cascade trigger with the spell's
// mana value captured when it was cast.
func (e *Engine) resolveCascadeTriggeredAbility(g *game.Game, obj *game.StackObject, agents [game.NumPlayers]PlayerAgent, log *TurnLog) string {
	if !obj.HasTriggerEvent || !obj.TriggerEvent.ManaValue.Exists {
		return "missing source"
	}
	e.cascade(g, obj, obj.Controller, obj.TriggerEvent.ManaValue.Val, agents, log)
	return "resolved"
}

// cascade performs the cascade process for playerID against a spell of
// manaValue (CR 702.85a): exile cards from the top of the library until a
// nonland card with lesser mana value is exiled, judged by the card's
// characteristics in exile, so a split card's halves are combined (CR 709.4),
// take any "as you cascade" actions (CR 702.85b), offer to cast the hit without
// paying its mana cost if the resulting spell's mana value is also less (only
// the chosen half of a split card), then put every exiled card that
// wasn't cast or otherwise moved on the bottom in a random order as one batch.
func (e *Engine) cascade(g *game.Game, obj *game.StackObject, playerID game.PlayerID, manaValue int, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	player, ok := playerByID(g, playerID)
	if !ok {
		return
	}
	var revealed []game.ZoneCardSnapshot
	var found id.ID
	for {
		cardID, ok := player.Library.Top()
		if !ok {
			break
		}
		epoch := player.Library.ReorderVersion()
		destination, moved := moveCardBetweenZonesWithEvent(g, game.Event{
			Controller: playerID,
			Player:     playerID,
			CardID:     cardID,
			FromZone:   zone.Library,
			ToZone:     zone.Exile,
		}, zoneMoveOptions{})
		if !moved {
			break
		}
		if destination == zone.Exile {
			card, _ := g.GetCardInstance(cardID)
			revealed = append(revealed, game.ZoneCardSnapshot{CardID: cardID, ZoneVersion: card.ZoneVersion})
		}
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			continue
		}
		if destination == zone.Exile && !graveyardCardHasType(g, card, types.Land) && offStackCardManaValue(card) < manaValue {
			found = cardID
			break
		}
		requireLibraryAdvance(player, cardID, epoch)
	}
	e.cascadeLandOptions(g, obj, player, revealed, agents, log)
	if found != 0 {
		// "You may cast it": the cast is optional, so the choice offers declining
		// and a declined (or uncastable) hit is bottomed with the rest.
		e.chooseAndCastFreeSpellFromSourceMay(g, player, playerID, found, zone.Exile, game.Selection{}, opt.Val(manaValue-1), false, true, agents, log)
	}
	bottomExiledCards(g, player, revealed, e.rng)
}

// cascadeLandOptions applies each active RuleEffectCascadeLandOption the
// cascading player controls ("As you cascade, you may put a land card from among
// the exiled cards onto the battlefield tapped.", Averna, the Chaos Bloom). CR
// 702.85b times these actions after the last card is exiled and before the cast
// decision, or before bottoming when nothing was hit. The applicable instances
// are fixed as the step begins; each lets the player put up to one land card
// still in exile from this cascade onto the battlefield tapped under their
// control through the shared battlefield-entry path, so zone-change
// replacements, "as enters" choices and enters triggers apply. An opponent's
// effect never applies, and the land is put, not played (CR 305.4).
func (e *Engine) cascadeLandOptions(g *game.Game, obj *game.StackObject, player *game.Player, revealed []game.ZoneCardSnapshot, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	var sources []id.ID
	for _, effect := range activeRuleEffects(g) {
		if effect.Kind == game.RuleEffectCascadeLandOption && effect.Controller == player.ID {
			sources = append(sources, effect.SourceCardID)
		}
	}
	for _, sourceCardID := range sources {
		lands := exiledCascadeLands(g, player, revealed)
		if len(lands) == 0 {
			return
		}
		options := make([]game.ChoiceOption, len(lands))
		for i, cardID := range lands {
			options[i] = game.ChoiceOption{Index: i, Label: cardChoiceLabel(g, cardID), Card: cardChoiceInfo(g, cardID)}
		}
		selected := e.chooseChoice(g, agents, game.ChoiceRequest{
			Kind:             game.ChoiceResolution,
			Player:           player.ID,
			Prompt:           "As you cascade, you may put a land card from among the exiled cards onto the battlefield tapped",
			Options:          options,
			MinChoices:       0,
			MaxChoices:       1,
			DefaultSelection: []int{0},
			Subject:          cardChoiceInfo(g, sourceCardID),
		}, log)
		if len(selected) != 1 || selected[0] < 0 || selected[0] >= len(lands) {
			continue
		}
		card, ok := g.GetCardInstance(lands[selected[0]])
		if !ok {
			continue
		}
		newEffectResolver(e, g, obj, agents, log).putResolvedCardOnBattlefieldValue(
			card, zone.Exile, player.ID, nil, permanentCreationOptions{ForceTapped: true})
	}
}

// exiledCascadeLands returns the land cards among a cascade's exiled cards that
// are still in that exile incarnation, in exile order.
func exiledCascadeLands(g *game.Game, player *game.Player, revealed []game.ZoneCardSnapshot) []id.ID {
	var lands []id.ID
	for _, snapshot := range revealed {
		card, ok := g.GetCardInstance(snapshot.CardID)
		if !ok || card.ZoneVersion != snapshot.ZoneVersion || !player.Exile.Contains(snapshot.CardID) {
			continue
		}
		if cardFaceOrDefault(card, game.FaceFront).HasType(types.Land) {
			lands = append(lands, snapshot.CardID)
		}
	}
	return lands
}

func (e *Engine) resolveDiscover(g *game.Game, obj *game.StackObject, manaValue int, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	if manaValue < 0 {
		return false
	}
	player, ok := playerByID(g, obj.Controller)
	if !ok {
		return false
	}
	revealed, found := exileUntilDiscoverHit(g, player, obj.Controller, manaValue)
	if found == 0 {
		bottomExiledCards(g, player, revealed, e.rng)
		return len(revealed) > 0
	}
	if e.chooseMay(g, agents, obj.Controller, "Cast discovered card without paying its mana cost?", log) &&
		e.castFreeSpellFromExileWithMaxManaValue(g, obj.Controller, found, manaValue, agents, log) {
		bottomExiledCards(g, player, revealed, e.rng)
		return true
	}
	moveCardBetweenZonesInBatch(g, obj.Controller, found, zone.Exile, zone.Hand, false, 0)
	bottomExiledCards(g, player, revealed, e.rng)
	return true
}

// exileUntilDiscoverHit exiles cards from the top of player's library until a
// nonland card with mana value manaValue or less is exiled, judged by the
// card's characteristics in exile, so a split card's halves are combined (CR
// 701.57a, 709.4).
func exileUntilDiscoverHit(g *game.Game, player *game.Player, playerID game.PlayerID, manaValue int) (revealed []game.ZoneCardSnapshot, found id.ID) {
	for {
		cardID, ok := player.Library.Top()
		if !ok {
			break
		}
		epoch := player.Library.ReorderVersion()
		destination, moved := moveCardBetweenZonesWithEvent(g, game.Event{
			Controller: playerID,
			Player:     playerID,
			CardID:     cardID,
			FromZone:   zone.Library,
			ToZone:     zone.Exile,
		}, zoneMoveOptions{})
		if !moved {
			break
		}
		if destination == zone.Exile {
			card, _ := g.GetCardInstance(cardID)
			revealed = append(revealed, game.ZoneCardSnapshot{CardID: cardID, ZoneVersion: card.ZoneVersion})
		}
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			continue
		}
		if destination == zone.Exile && !graveyardCardHasType(g, card, types.Land) && offStackCardManaValue(card) <= manaValue {
			found = cardID
			break
		}
		requireLibraryAdvance(player, cardID, epoch)
	}
	return revealed, found
}

// bottomExiledCards puts the still-exiled cards on the bottom of player's library
// in a random order as one simultaneous move ("put all ... on the bottom"): every
// card shares one batch id, so a one-or-more zone-change trigger sees one event
// (CR 603.2c). Each card's zone-change replacements and commander choice are
// determined before any card moves (CR 400.6, 614, 903.9b), in exile order, and
// only the cards that actually arrive are then randomly ordered, so neither the
// choices nor the zone-change events reveal the hidden library order (CR 401.2,
// 702.85a, 701.57a). A snapshot whose card has left that exile incarnation is
// skipped.
func bottomExiledCards(g *game.Game, player *game.Player, cards []game.ZoneCardSnapshot, rng *rand.Rand) bool {
	var moves []simultaneousCardMove
	for _, snapshot := range cards {
		if player.Exile.Contains(snapshot.CardID) {
			moves = append(moves, simultaneousCardMove{card: snapshot, fromZone: zone.Exile})
		}
	}
	if len(moves) == 0 {
		return false
	}
	return moveCardsBetweenZonesSimultaneously(g, moves, zone.Library, true, g.IDGen.Next(), rng)
}

func (e *Engine) castFreeSpellFromExile(g *game.Game, playerID game.PlayerID, cardID id.ID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	return e.castFreeSpellFromZone(g, playerID, cardID, zone.Exile, agents, log)
}

func (e *Engine) castFreeSpellFromExileWithMaxManaValue(g *game.Game, playerID game.PlayerID, cardID id.ID, maxManaValue int, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	player, ok := playerByID(g, playerID)
	if !ok {
		return false
	}
	return e.chooseAndCastFreeSpellFromSource(
		g,
		player,
		playerID,
		cardID,
		zone.Exile,
		game.Selection{},
		opt.Val(maxManaValue),
		false,
		agents,
		log,
	)
}

// castAnyNumberFromExileForFree lets controllerID cast any number of the given
// exiled cards without paying their mana costs, one at a time in an order the
// controller chooses, during the resolution of the ability that exiled them
// (Etali, Primal Storm). Each cast card is put on the stack under controllerID's
// control wherever it currently rests, so a card exiled from an opponent's
// library is still cast by the controller. Only cards still in exile that have a
// legal cast choice are offered, so lands and uncastable cards are skipped and
// left exiled. It stops when the controller casts nothing more or no castable
// card remains; removing the chosen card each iteration guarantees termination.
func (e *Engine) castAnyNumberFromExileForFree(g *game.Game, controllerID game.PlayerID, cards []id.ID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	remaining := append([]id.ID(nil), cards...)
	for len(remaining) > 0 {
		var candidates []id.ID
		for _, cardID := range remaining {
			if e.castableExiledSpell(g, controllerID, cardID) {
				candidates = append(candidates, cardID)
			}
		}
		if len(candidates) == 0 {
			return
		}
		options := make([]game.ChoiceOption, len(candidates))
		for i, cardID := range candidates {
			options[i] = game.ChoiceOption{
				Index: i,
				Label: cardChoiceLabel(g, cardID),
				Card:  cardChoiceInfo(g, cardID),
			}
		}
		selected := e.chooseChoice(g, agents, game.ChoiceRequest{
			Kind:       game.ChoiceResolution,
			Player:     controllerID,
			Prompt:     "Choose a spell to cast without paying its mana cost, or none to stop",
			Options:    options,
			MinChoices: 0,
			MaxChoices: 1,
		}, log)
		if len(selected) == 0 || selected[0] < 0 || selected[0] >= len(candidates) {
			return
		}
		chosen := candidates[selected[0]]
		if i := slices.Index(remaining, chosen); i >= 0 {
			remaining = slices.Delete(remaining, i, i+1)
		}
		e.castFreeTargetedSpell(g, controllerID, chosen, zone.Exile, false, agents, log)
	}
}

// castableExiledSpell reports whether cardID still rests in some player's exile
// and has a legal free-cast choice for controllerID, the filter Etali's "cast
// any number of spells from among those cards" applies to the exiled pool.
func (e *Engine) castableExiledSpell(g *game.Game, controllerID game.PlayerID, cardID id.ID) bool {
	if _, ok := playerHoldingCastSource(g, cardID, zone.Exile); !ok {
		return false
	}
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return false
	}
	return len(e.freeCastOptionsForCard(g, controllerID, card, zone.Exile, game.Selection{}, opt.V[int]{})) > 0
}

// castFreeSpellFromZone casts cardID from fromZone for playerID without paying
// its mana cost, letting the player choose its legal face, modes, targets, and
// optional additional costs. It returns false (casting nothing) when the card is
// no longer in fromZone or has no legal cast choice.
func (e *Engine) castFreeSpellFromZone(g *game.Game, playerID game.PlayerID, cardID id.ID, fromZone zone.Type, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	player, ok := playerByID(g, playerID)
	if !ok {
		return false
	}
	return e.castFreeSpellFromSource(g, player, playerID, cardID, fromZone, false, agents, log)
}

// castFreeTargetedSpell casts a specific targeted card from fromZone for
// controllerID without paying its mana cost, locating whichever player's zone
// currently holds the card so a spell that targets a card in an opponent's
// graveyard (Memory Plunder) casts it under the controller's control. When
// exileOnResolution is set the cast spell moves to exile instead of its owner's
// graveyard after it resolves, modeling the "exile it instead" rider. It returns
// false when the card no longer rests in a player's fromZone or has no legal cast
// choice.
func (e *Engine) castFreeTargetedSpell(g *game.Game, controllerID game.PlayerID, cardID id.ID, fromZone zone.Type, exileOnResolution bool, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	source, ok := playerHoldingCastSource(g, cardID, fromZone)
	if !ok {
		return false
	}
	return e.castFreeSpellFromSource(g, source, controllerID, cardID, fromZone, exileOnResolution, agents, log)
}

// playerHoldingCastSource returns the player whose fromZone currently contains
// cardID, scanning every player in turn order. It backs targeted free casts,
// whose card may rest in any player's zone (an opponent's graveyard).
func playerHoldingCastSource(g *game.Game, cardID id.ID, fromZone zone.Type) (*game.Player, bool) {
	for i := range g.Players {
		player := g.Players[i]
		if castSourceContains(player, cardID, fromZone) {
			return player, true
		}
	}
	return nil, false
}

// castFreeSpellFromSource casts cardID out of sourcePlayer's fromZone for
// controllerID without paying its mana cost, letting the player choose its legal
// face, modes, targets, and optional additional costs before pushing it to the
// stack under controllerID's control. The source player and controller differ
// only for a targeted cast from another player's zone. exileOnResolution
// redirects the resolved spell to exile in place of its owner's graveyard.
func (e *Engine) castFreeSpellFromSource(g *game.Game, sourcePlayer *game.Player, controllerID game.PlayerID, cardID id.ID, fromZone zone.Type, exileOnResolution bool, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	return e.chooseAndCastFreeSpellFromSource(
		g,
		sourcePlayer,
		controllerID,
		cardID,
		fromZone,
		game.Selection{},
		opt.V[int]{},
		exileOnResolution,
		agents,
		log,
	)
}
