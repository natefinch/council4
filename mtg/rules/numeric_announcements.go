package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/zone"
)

func castGiftRecipientValid(g *game.Game, player game.PlayerID, cast action.CastSpellAction) bool {
	return !cast.GiftPromised || cast.GiftRecipient != player && isPlayerAlive(g, cast.GiftRecipient)
}

func abilityCanAnnounceX(body *game.ActivatedAbility) bool {
	if body == nil {
		return false
	}
	if costHasVariableMana(manaCostPtr(body.ManaCost)) || additionalCostsUseX(abilityAdditionalCosts(body.AdditionalCosts)) {
		return true
	}
	for _, alternative := range body.AlternativeCosts {
		if costHasVariableMana(manaCostPtr(alternative.ManaCost)) || additionalCostsUseX(alternative.AdditionalCosts) {
			return true
		}
	}
	return false
}

func (e *Engine) announcementLegal(g *game.Game, player game.PlayerID, announcement action.Action) bool {
	if announcement.Validate() != nil || !canAct(g, player) || player != g.Turn.PriorityPlayer || splitSecondOnStack(g) {
		return false
	}

	if cast, ok := announcement.CastSpellPayload(); ok {
		if !castGiftRecipientValid(g, player, cast) {
			return false
		}
		if cast.Mutate {
			return false
		}
		source := normalizedCastSourceZone(cast)
		if source == zone.Battlefield {
			return cast.Face == game.FaceAlternate && cast.Permission == (game.PlayPermission{}) &&
				!cast.KickerPaid && cast.KickerCount == 0 && !cast.Overloaded && !cast.GiftPromised &&
				!cast.Bargained && !cast.Bestowed && !cast.Offspring &&
				canCastPreparedCopy(g, player, preparedSourcePermanent(g, cast.CardID), cast.Targets, cast.XValue, cast.ChosenModes)
		}
		var permission []game.PlayPermission
		if cast.Permission != (game.PlayPermission{}) {
			permission = []game.PlayPermission{cast.Permission}
		}
		return e.canCastSpellFaceFromZoneWithOptions(g, player, cast.CardID, source, cast.Face,
			cast.Targets, cast.XValue, cast.ChosenModes, effectiveKickerCount(cast.KickerPaid, cast.KickerCount),
			cast.Overloaded, cast.GiftPromised, cast.Bargained, cast.Bestowed, cast.Offspring, permission...)
	}
	activate, ok := announcement.ActivateAbilityPayload()
	if !ok {
		return false
	}
	if permanent, ability, present := activatedAbilitySource(g, player, activate.SourceID, activate.AbilityIndex); present {
		def, exists := permanentCardDef(g, permanent)
		if !exists {
			return false
		}
		if _, valid := bodyTargetCountsWithModesAndRecorded(g, player, def, permanent.ObjectID, ability,
			activate.ChosenModes, activate.TargetCounts, activate.Targets, activate.XValue); !valid {
			return false
		}
		if body, general := ability.(*game.ActivatedAbility); general {
			return canActivateEquipAbilityWithModes(g, player, permanent, body, activate.AbilityIndex,
				activate.Targets, activate.XValue, activate.ChosenModes) ||
				canActivateGeneralAbilityWithModes(g, player, permanent, body, activate.AbilityIndex,
					activate.Targets, activate.XValue, activate.ChosenModes)
		}
		return false
	}
	if card, body, present := handActivatedAbilitySource(g, player, activate.SourceID, activate.AbilityIndex); present {
		if _, valid := bodyTargetCountsWithModesAndRecorded(g, player, cardFaceOrDefault(card, game.FaceFront), 0, &body,
			activate.ChosenModes, activate.TargetCounts, activate.Targets, activate.XValue); !valid {
			return false
		}
		return canActivateHandAbilityWithModes(g, player, activate.SourceID, &body, activate.AbilityIndex,
			activate.Targets, activate.XValue, activate.ChosenModes)
	}
	if card, body, present := graveyardAbilitySource(g, player, activate.SourceID, activate.AbilityIndex); present {
		if _, valid := bodyTargetCountsWithModesAndRecorded(g, player, cardFaceOrDefault(card, game.FaceFront), 0, &body,
			activate.ChosenModes, activate.TargetCounts, activate.Targets, activate.XValue); !valid {
			return false
		}
		return canActivateGraveyardAbilityWithModes(g, player, activate.SourceID, &body, activate.AbilityIndex,
			activate.Targets, activate.XValue, activate.ChosenModes)
	}
	return false
}
