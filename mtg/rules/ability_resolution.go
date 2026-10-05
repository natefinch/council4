package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

func permanentActivatedResolutionUse(g *game.Game, permanent *game.Permanent, body game.Ability, index int) opt.V[game.ActivatedAbilityResolutionUse] {
	activated, ok := body.(*game.ActivatedAbility)
	if !ok || !activated.CountsResolutionsThisTurn {
		return opt.V[game.ActivatedAbilityResolutionUse]{}
	}
	values := effectivePermanentValues(g, permanent)
	use, ok := values.resolutionUses[index]
	if !ok || use.Ability != activated {
		return opt.V[game.ActivatedAbilityResolutionUse]{}
	}
	return opt.Val(use)
}

func appendEffectiveAbility(values *permanentEffectiveValues, permanent *game.Permanent, body game.Ability, origin game.ActivatedAbilityResolutionUse) {
	index := len(values.abilities)
	values.abilities = append(values.abilities, body)
	activated, ok := body.(*game.ActivatedAbility)
	if !ok || !activated.CountsResolutionsThisTurn {
		return
	}
	origin.SourceID = permanent.ObjectID
	origin.Ability = activated
	if values.resolutionUses == nil {
		values.resolutionUses = make(map[int]game.ActivatedAbilityResolutionUse)
	}
	values.resolutionUses[index] = origin
}

func permanentAbilityOrigin(permanent *game.Permanent) game.ActivatedAbilityResolutionUse {
	origin := permanent.AbilityOriginID
	if origin == 0 {
		origin = permanent.CardInstanceID
	}
	if origin == 0 {
		origin = permanent.ObjectID
	}
	return game.ActivatedAbilityResolutionUse{OriginID: origin}
}

func cardActivatedResolutionUse(card *game.CardInstance, face *game.CardDef, index int) opt.V[game.ActivatedAbilityResolutionUse] {
	body, ok := face.BodyAt(index).(*game.ActivatedAbility)
	if !ok || !body.CountsResolutionsThisTurn {
		return opt.V[game.ActivatedAbilityResolutionUse]{}
	}
	return opt.Val(game.ActivatedAbilityResolutionUse{
		SourceID: card.ID, SourceZoneVersion: card.ZoneVersion, Ability: body, OriginID: card.ID,
	})
}

func recordActivatedAbilityResolution(g *game.Game, obj *game.StackObject, body *game.ActivatedAbility) bool {
	obj.ResolutionOrdinalThisTurn = 0
	if !body.CountsResolutionsThisTurn {
		return true
	}
	if !obj.ActivatedResolutionUse.Exists ||
		obj.SourceID == 0 || obj.AbilityIndex < 0 {
		return false
	}
	use := obj.ActivatedResolutionUse.Val
	if use.SourceID != obj.SourceID || use.Ability == nil ||
		!use.Ability.CountsResolutionsThisTurn || use.Occurrence < 0 ||
		(!use.Granted && use.OriginID == 0) ||
		(use.Granted && use.GrantID == 0 && use.GrantSourceID == 0) ||
		(obj.SourceID == obj.SourceCardID && use.SourceZoneVersion != obj.SourceZoneVersion) {
		return false
	}
	if g.ResolvedActivatedAbilitiesThisTurn == nil {
		g.ResolvedActivatedAbilitiesThisTurn = make(map[game.ActivatedAbilityResolutionUse]int)
	}
	g.ResolvedActivatedAbilitiesThisTurn[use]++
	obj.ResolutionOrdinalThisTurn = g.ResolvedActivatedAbilitiesThisTurn[use]
	return true
}
