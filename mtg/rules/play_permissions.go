package rules

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
)

func permanentTypePlayPermissions(g *game.Game, player game.PlayerID, sourceZone zone.Type, def *game.CardDef) []game.PlayPermission {
	if sourceZone != zone.Graveyard || player != g.Turn.ActivePlayer || def == nil {
		return nil
	}
	var permissions []game.PlayPermission
	for _, source := range g.Battlefield {
		if source.PhasedOut || effectiveController(g, source) != player {
			continue
		}
		// Unlike layer-effect discovery, permissions depend on the finished
		// ability layer: copied/granted bodies count, removed bodies do not.
		for _, ability := range permanentEffectiveAbilitiesView(g, source) {
			provider, ok := ability.(game.StaticRuleEffectProvider)
			if !ok {
				continue
			}
			functionZone, condition, effects := provider.StaticRuleEffectDeclarations()
			if (functionZone != zone.None && functionZone != zone.Battlefield) ||
				!conditionSatisfied(g, conditionContext{controller: player, source: source}, condition) {
				continue
			}
			for _, effect := range effects {
				if effect.Kind != game.RuleEffectPlayFromGraveyardPerPermanentType || effect.PermissionKey == "" {
					continue
				}
				for _, cardType := range def.Types {
					if !cardType.IsPermanent() {
						continue
					}
					permission := game.PlayPermission{SourceID: source.ObjectID, Key: effect.PermissionKey, Type: cardType}
					if !g.PermanentTypePlayUsesThisTurn[game.PlayPermissionUse{Permission: permission, Player: player}] &&
						!slices.Contains(permissions, permission) {
						permissions = append(permissions, permission)
					}
				}
			}
		}
	}
	return permissions
}

func recordPermanentTypePlay(g *game.Game, player game.PlayerID, permission game.PlayPermission) {
	if permission == (game.PlayPermission{}) {
		return
	}
	if g.PermanentTypePlayUsesThisTurn == nil {
		g.PermanentTypePlayUsesThisTurn = make(map[game.PlayPermissionUse]bool)
	}
	g.PermanentTypePlayUsesThisTurn[game.PlayPermissionUse{Permission: permission, Player: player}] = true
}

func castPermissionsForSpell(g *game.Game, player game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex, def *game.CardDef) []payment.SpellCastPermission {
	permissions := independentCastPermissionsForZone(g, player, cardID, sourceZone, face)
	if len(permanentTypePlayPermissions(g, player, sourceZone, def)) > 0 &&
		!slices.Contains(permissions, payment.SpellCastPermissionRuleEffect) {
		permissions = append(permissions, payment.SpellCastPermissionRuleEffect)
	}

	return permissions
}

func selectedSpellPermissions(g *game.Game, player game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex, def *game.CardDef, announced []game.PlayPermission) ([]payment.SpellCastPermission, bool) {
	if len(announced) > 0 {
		if announced[0] == (game.PlayPermission{}) {
			permissions := independentCastPermissionsForZone(g, player, cardID, sourceZone, face)
			return permissions, len(permissions) > 0
		}
		if !slices.Contains(permanentTypePlayPermissions(g, player, sourceZone, def), announced[0]) {
			return nil, false
		}
		return []payment.SpellCastPermission{payment.SpellCastPermissionRuleEffect}, true
	}
	return castPermissionsForSpell(g, player, cardID, sourceZone, face, def), true
}

// selectPermanentTypePlay validates an announcement before any movement or
// payment. Direct callers without an action payload get the same choice through
// the ordinary agent choice seam; independent permissions never spend a quota.
func (e *Engine) selectPermanentTypePlay(g *game.Game, player game.PlayerID, announced game.PlayPermission, sourceZone zone.Type, def *game.CardDef, independent bool, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (game.PlayPermission, bool) {
	permissions := permanentTypePlayPermissions(g, player, sourceZone, def)
	if announced != (game.PlayPermission{}) {
		return announced, slices.Contains(permissions, announced)
	}
	if independent {
		return game.PlayPermission{}, true
	}
	if len(permissions) == 0 {
		return game.PlayPermission{}, false
	}
	options := make([]game.ChoiceOption, len(permissions))
	for i, permission := range permissions {
		options[i] = game.ChoiceOption{Index: i, Label: fmt.Sprintf("Use %s permission from source %d", permission.Type, permission.SourceID)}
	}
	selected := e.chooseChoice(g, agents, game.ChoiceRequest{
		Kind: game.ChoiceResolution, Player: player, Prompt: "Choose graveyard play permission",
		Options: options, MinChoices: 1, MaxChoices: 1, DefaultSelection: []int{0},
	}, log)
	return permissions[selected[0]], true
}

func (e *Engine) expandPermanentTypePlayActions(g *game.Game, player game.PlayerID, actions []action.Action) []action.Action {
	var expanded []action.Action
	for _, act := range actions {
		var cardID id.ID
		var sourceZone zone.Type
		var face game.FaceIndex
		var bestowed, independent bool
		if cast, ok := act.CastSpellPayload(); ok {
			cardID, sourceZone, face, bestowed = cast.CardID, normalizedCastSourceZone(cast), cast.Face, cast.Bestowed
			independent = len(independentCastPermissionsForZone(g, player, cardID, sourceZone, face)) > 0
		} else if land, ok := act.PlayLandPayload(); ok {
			cardID, sourceZone, face = land.CardID, land.SourceZone, land.Face
			_, independent = matchingPlayLandFromZoneEffect(g, player, cardID, sourceZone)
		} else {
			expanded = append(expanded, act)
			continue
		}
		if sourceZone != zone.Graveyard {
			expanded = append(expanded, act)
			continue
		}
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			continue
		}
		if independent {
			if cast, ok := act.CastSpellPayload(); !ok || e.castActionPermissionLegal(g, player, cast, game.PlayPermission{}) {
				expanded = append(expanded, act)
			}
		}
		for _, permission := range permanentTypePlayPermissions(g, player, sourceZone, castSelectionFace(cardFaceOrDefault(card, face), bestowed)) {
			if cast, ok := act.CastSpellPayload(); ok && !e.castActionPermissionLegal(g, player, cast, permission) {
				continue
			}
			expanded = append(expanded, act.WithPlayPermission(permission))
		}
	}
	return expanded
}

func (e *Engine) castActionPermissionLegal(g *game.Game, player game.PlayerID, cast action.CastSpellAction, permission game.PlayPermission) bool {
	sourceZone := normalizedCastSourceZone(cast)
	if cast.Mutate {
		return canCastMutateSpell(g, player, cast.CardID, sourceZone, cast.MutateTargetID, permission)
	}
	return e.canCastSpellFaceFromZoneWithOptions(g, player, cast.CardID, sourceZone, cast.Face, cast.Targets, cast.XValue, cast.ChosenModes, effectiveKickerCount(cast.KickerPaid, cast.KickerCount), cast.Overloaded, cast.GiftPromised, cast.Bargained, cast.Bestowed, cast.Offspring, permission)
}
