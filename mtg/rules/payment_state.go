package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

// rulesPaymentState implements payment.State by delegating to the rules-engine
// helpers that operate on *game.Game. One instance is created per payment
// operation and holds the game pointer for the duration of that call.
type rulesPaymentState struct {
	g *game.Game
}

var _ payment.State = (*rulesPaymentState)(nil)

func (s *rulesPaymentState) Player(playerID game.PlayerID) (*game.Player, bool) {
	return playerByID(s.g, playerID)
}

func (s *rulesPaymentState) Battlefield() []*game.Permanent {
	return s.g.Battlefield
}

func (s *rulesPaymentState) EffectiveController(p *game.Permanent) game.PlayerID {
	return effectiveController(s.g, p)
}

func (s *rulesPaymentState) PermanentCardDef(p *game.Permanent) (*game.CardDef, bool) {
	return permanentCardDef(s.g, p)
}

func (s *rulesPaymentState) PermanentHasType(p *game.Permanent, t types.Card) bool {
	return permanentHasType(s.g, p, t)
}

func (s *rulesPaymentState) PermanentHasSupertype(p *game.Permanent, supertype types.Super) bool {
	return permanentHasSupertype(s.g, p, supertype)
}

func (s *rulesPaymentState) PermanentEffectiveColors(p *game.Permanent) []color.Color {
	return permanentEffectiveColors(s.g, p)
}

func (s *rulesPaymentState) ActivationConditionSatisfied(playerID game.PlayerID, permanent *game.Permanent, condition opt.V[game.Condition]) bool {
	return activationConditionSatisfied(s.g, playerID, permanent, condition)
}

func (s *rulesPaymentState) ManaAbilityTimingAllowed(playerID game.PlayerID, permanent *game.Permanent, abilityIndex int, timing game.TimingRestriction) bool {
	return activatedAbilityTimingAllows(s.g, playerID, timing) &&
		!activatedAbilityUsedThisTurn(s.g, permanent.ObjectID, abilityIndex, timing)
}

func (s *rulesPaymentState) PermanentByObjectID(objectID id.ID) (*game.Permanent, bool) {
	return permanentByObjectID(s.g, objectID)
}

func (s *rulesPaymentState) CardInstance(cardID id.ID) (*game.CardInstance, bool) {
	return s.g.GetCardInstance(cardID)
}

func (*rulesPaymentState) CardFace(card *game.CardInstance, face game.FaceIndex) *game.CardDef {
	return cardFaceOrDefault(card, face)
}

func (s *rulesPaymentState) CostModifiersForSpell(playerID game.PlayerID, card *game.CardDef, cardID id.ID, sourceZone zone.Type) []game.CostModifier {
	var modifiers []game.CostModifier
	for _, modifier := range s.g.CostModifiers {
		if modifier.Kind != game.CostModifierSpell {
			continue
		}
		if modifier.MatchCardType && (card == nil || !card.HasType(modifier.CardType)) {
			continue
		}
		modifiers = append(modifiers, modifier)
	}
	if sourceZone == zone.Command && cardID != 0 {
		player, ok := playerByID(s.g, playerID)
		if ok && player.CommanderInstanceID == cardID && player.CommanderTax() > 0 {
			modifiers = append(modifiers, game.CostModifier{
				Kind:            game.CostModifierSpell,
				GenericIncrease: player.CommanderTax(),
			})
		}
	}
	modifiers = append(modifiers, staticCostModifiersForContext(s.g, card)...)
	return modifiers
}

func (s *rulesPaymentState) SetTapped(p *game.Permanent, tapped bool) {
	setPermanentTapped(s.g, p, tapped)
}

func (s *rulesPaymentState) RecordManaAbilityUse(p *game.Permanent, abilityIndex int, timing game.TimingRestriction) {
	recordActivatedAbilityUse(s.g, p.ObjectID, abilityIndex, timing)
}

func (*rulesPaymentState) RemoveCounters(p *game.Permanent, kind counter.Kind, amount int) bool {
	return p != nil && p.Counters.Remove(kind, amount) == amount
}

func (s *rulesPaymentState) LoseLife(playerID game.PlayerID, amount int) {
	loseLife(s.g, playerID, amount)
}

func (s *rulesPaymentState) EmitZoneChange(event game.Event) {
	emitZoneChangeEvent(s.g, event)
}

func (s *rulesPaymentState) EmitCardReveal(playerID game.PlayerID, cardID id.ID, from zone.Type) {
	emitEvent(s.g, game.Event{
		Kind:     game.EventCardRevealed,
		Player:   playerID,
		CardID:   cardID,
		FromZone: from,
		Amount:   1,
	})
}

func (s *rulesPaymentState) MovePermanentToZone(p *game.Permanent, dest zone.Type) bool {
	return movePermanentToZone(s.g, p, dest)
}

func (s *rulesPaymentState) DiscardFromHand(playerID game.PlayerID, cardID id.ID) bool {
	return discardCardFromHand(s.g, playerID, cardID)
}

func (s *rulesPaymentState) MoveCard(playerID game.PlayerID, cardID id.ID, from, to zone.Type) bool {
	return moveCardBetweenZones(s.g, playerID, cardID, from, to)
}
