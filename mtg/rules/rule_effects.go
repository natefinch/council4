package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

func createRuleEffectTemplates(g *game.Game, obj *game.StackObject, object opt.V[game.ObjectReference], templates []game.RuleEffect, duration game.EffectDuration) bool {
	if len(templates) == 0 {
		return false
	}
	sourceID, sourceObjectID := damageSourceIDs(g, obj)
	for i := range templates {
		ruleEffect := templates[i]
		ruleEffect.ID = g.IDGen.Next()
		ruleEffect.Controller = obj.Controller
		ruleEffect.SourceCardID = sourceID
		ruleEffect.SourceObjectID = sourceObjectID
		if ruleEffect.AffectedSource {
			ruleEffect.AffectedObjectID = sourceObjectID
		} else if ruleEffect.AffectedObjectID == 0 {
			if object.Exists {
				if resolved, ok := resolveObjectReference(g, obj, object.Val); ok && resolved.permanent != nil {
					ruleEffect.AffectedObjectID = resolved.permanent.ObjectID
				}
			}
		}
		ruleEffect.CreatedTurn = g.Turn.TurnNumber
		if duration != game.DurationPermanent {
			ruleEffect.Duration = duration
		}
		if ruleEffect.Duration == game.DurationUntilYourNextTurn ||
			ruleEffect.Duration == game.DurationUntilEndOfYourNextTurn {
			ruleEffect.ExpiresFor = obj.Controller
		}
		g.RuleEffects = append(g.RuleEffects, ruleEffect)
	}
	return true
}

func activeRuleEffects(g *game.Game) []game.RuleEffect {
	effects := make([]game.RuleEffect, 0, len(g.RuleEffects))
	for i := range g.RuleEffects {
		if ruleEffectSourceIsActive(g, &g.RuleEffects[i]) {
			effects = append(effects, g.RuleEffects[i])
		}
	}
	effects = append(effects, staticRuleEffects(g)...)
	effects = append(effects, stackStaticRuleEffects(g)...)
	return effects
}

func staticRuleEffects(g *game.Game) []game.RuleEffect {
	var effects []game.RuleEffect
	for _, source := range g.Battlefield {
		if source.PhasedOut {
			continue
		}
		visitPermanentStaticAbilityComponents(g, source, func(component permanentAbilityComponent) {
			for i := range component.card.StaticAbilities {
				body := &component.card.StaticAbilities[i]
				if len(body.RuleEffects) == 0 || !bodyFunctionsOnBattlefield(body) {
					continue
				}
				controller := effectiveController(g, source)
				if !conditionSatisfied(g, conditionContext{
					controller: controller,
					source:     source,
				}, body.Condition) {
					continue
				}
				for j := range body.RuleEffects {
					ruleEffect := body.RuleEffects[j]
					ruleEffect.Controller = controller
					ruleEffect.SourceObjectID = source.ObjectID
					ruleEffect.SourceCardID = component.cardID
					if ruleEffect.AffectedSource {
						ruleEffect.AffectedObjectID = source.ObjectID
					} else if ruleEffect.AffectedAttached {
						if !source.AttachedTo.Exists {
							continue
						}
						ruleEffect.AffectedObjectID = source.AttachedTo.Val
					}
					effects = append(effects, ruleEffect)
				}
			}
		})
	}
	return effects
}

func stackStaticRuleEffects(g *game.Game) []game.RuleEffect {
	var effects []game.RuleEffect
	for _, source := range g.Stack.Objects() {
		if source.Kind != game.StackSpell {
			continue
		}
		_, sourceDef, ok := cardInstanceFaceDef(g, source.SourceID, source.Face)
		if !ok {
			continue
		}
		for i := range sourceDef.StaticAbilities {
			body := &sourceDef.StaticAbilities[i]
			if body.ZoneOfFunction != zone.Stack {
				continue
			}
			if !conditionSatisfied(g, conditionContext{
				controller: source.Controller,
			}, body.Condition) {
				continue
			}
			for j := range body.RuleEffects {
				ruleEffect := body.RuleEffects[j]
				ruleEffect.Controller = source.Controller
				ruleEffect.SourceObjectID = source.ID
				ruleEffect.SourceCardID = source.SourceID
				if ruleEffect.AffectedSource {
					ruleEffect.AffectedObjectID = source.ID
				}
				effects = append(effects, ruleEffect)
			}
		}
	}
	return effects
}

func ruleEffectSourceStillApplies(g *game.Game, effect *game.RuleEffect) bool {
	if effect == nil {
		return false
	}
	if effect.Duration != game.DurationPermanent || effect.SourceObjectID == 0 {
		return true
	}
	_, ok := permanentByObjectID(g, effect.SourceObjectID)
	return ok
}

func ruleEffectSourceIsActive(g *game.Game, effect *game.RuleEffect) bool {
	if !ruleEffectSourceStillApplies(g, effect) {
		return false
	}
	if effect.Duration != game.DurationPermanent || effect.SourceObjectID == 0 {
		return true
	}
	source, ok := permanentByObjectID(g, effect.SourceObjectID)
	return ok && activeBattlefieldPermanent(source)
}

func expireRuleEffects(g *game.Game) {
	if len(g.RuleEffects) == 0 {
		return
	}
	kept := g.RuleEffects[:0]
	for i := range g.RuleEffects {
		effect := &g.RuleEffects[i]
		if effect.Duration == game.DurationUntilEndOfYourNextTurn &&
			effect.ExpiresFor == g.Turn.ActivePlayer &&
			effect.CreatedTurn < g.Turn.TurnNumber {
			continue
		}
		if effect.Duration == game.DurationUntilEndOfTurn || effect.Duration == game.DurationThisTurn {
			continue
		}
		if !ruleEffectSourceStillApplies(g, effect) {
			continue
		}
		kept = append(kept, *effect)
	}
	g.RuleEffects = kept
}

func canGainLife(g *game.Game, playerID game.PlayerID) bool {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind != game.RuleEffectCantGainLife {
			continue
		}
		if playerRelationMatches(effect.Controller, playerID, effect.AffectedPlayer) {
			return false
		}
	}
	return true
}

// playerCanCastAsThoughFlash reports whether an active rule effect lets playerID
// cast spells as though they had flash, i.e. at instant speed ("You may cast
// spells this turn as though they had flash.", Borne Upon a Wind, Emergence
// Zone; CR 702.8 / 601.3e).
func playerCanCastAsThoughFlash(g *game.Game, playerID game.PlayerID) bool {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind != game.RuleEffectCastSpellsAsThoughFlash {
			continue
		}
		if playerRelationMatches(effect.Controller, playerID, effect.AffectedPlayer) {
			return true
		}
	}
	return false
}

// additionalLandPlaysFor returns the number of extra land plays granted to
// playerID by active RuleEffectAdditionalLandPlays effects (Explore, Exploration,
// Azusa, etc.), summed across all such effects. It is added to the
// one-land-per-turn baseline when checking whether the player may play a land.
func additionalLandPlaysFor(g *game.Game, playerID game.PlayerID) int {
	total := 0
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind != game.RuleEffectAdditionalLandPlays {
			continue
		}
		if playerRelationMatches(effect.Controller, playerID, effect.AffectedPlayer) {
			total += effect.AdditionalLandPlays
		}
	}
	return total
}

// playerHasNoMaximumHandSize reports whether an active rule effect removes the
// maximum hand size of playerID, so that player skips discarding down to a
// hand-size limit during their cleanup step (CR 402.2).
func playerHasNoMaximumHandSize(g *game.Game, playerID game.PlayerID) bool {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind != game.RuleEffectNoMaximumHandSize {
			continue
		}
		if playerRelationMatches(effect.Controller, playerID, effect.AffectedPlayer) {
			return true
		}
	}
	return false
}

// spellCastProhibited reports whether an active RuleEffectCantCastSpells effect
// forbids playerID from casting spellDef ("Your opponents can't cast spells.",
// Grand Abolisher's "During your turn, your opponents can't cast spells ...").
// The prohibition honors its affected-player relation, optional controller-turn
// scope, and optional spell-type filter.
func spellCastProhibited(g *game.Game, playerID game.PlayerID, spellDef *game.CardDef) bool {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind != game.RuleEffectCantCastSpells ||
			!playerRelationMatches(effect.Controller, playerID, effect.AffectedPlayer) ||
			!actionRestrictionTurnActive(g, effect) {
			continue
		}
		if len(effect.SpellTypes) > 0 && !cardDefHasAnyType(spellDef, effect.SpellTypes) {
			continue
		}
		return true
	}
	return false
}

// abilityActivationProhibited reports whether an active
// RuleEffectCantActivateAbilities effect forbids playerID from activating an
// ability of permanent ("... activate abilities of artifacts, creatures, or
// enchantments."). The prohibition honors its affected-player relation, optional
// controller-turn scope, and the permanent-type filter.
func abilityActivationProhibited(g *game.Game, playerID game.PlayerID, permanent *game.Permanent) bool {
	card, ok := permanentCardDef(g, permanent)
	if !ok {
		return false
	}
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind != game.RuleEffectCantActivateAbilities ||
			!playerRelationMatches(effect.Controller, playerID, effect.AffectedPlayer) ||
			!actionRestrictionTurnActive(g, effect) {
			continue
		}
		if len(effect.PermanentTypes) > 0 && !cardDefHasAnyType(card, effect.PermanentTypes) {
			continue
		}
		return true
	}
	return false
}

// actionRestrictionTurnActive reports whether a turn-scoped action restriction
// is in force right now: an effect scoped to the controller's turn applies only
// while that controller is the active player.
func actionRestrictionTurnActive(g *game.Game, effect *game.RuleEffect) bool {
	return !effect.RestrictedDuringControllerTurn || g.Turn.ActivePlayer == effect.Controller
}

// cardDefHasAnyType reports whether def has at least one of the given card types.
func cardDefHasAnyType(def *game.CardDef, cardTypes []types.Card) bool {
	return slices.ContainsFunc(cardTypes, def.HasType)
}

func gainLife(g *game.Game, playerID game.PlayerID, amount int) int {
	if amount <= 0 || !canGainLife(g, playerID) ||
		playerRuleEffectActive(g, playerID, game.RuleEffectLifeTotalCantChange) {
		return 0
	}
	amount = replacementLifeGainAmount(g, playerID, amount)
	if amount <= 0 {
		return 0
	}
	player, ok := playerByID(g, playerID)
	if !ok || player.Eliminated {
		return 0
	}
	player.Life += amount
	emitEvent(g, game.Event{
		Kind:                       game.EventLifeGained,
		Player:                     playerID,
		Amount:                     amount,
		PlayerEventOrdinalThisTurn: nextPlayerEventOrdinalThisTurn(g, game.EventLifeGained, playerID),
	})
	return amount
}

func loseLife(g *game.Game, playerID game.PlayerID, amount int) int {
	if amount <= 0 || playerRuleEffectActive(g, playerID, game.RuleEffectLifeTotalCantChange) {
		return 0
	}
	player, ok := playerByID(g, playerID)
	if !ok || player.Eliminated {
		return 0
	}
	player.Life -= amount
	increaseActivePlayerSpeedForOpponentLifeLoss(g, playerID)
	emitEvent(g, game.Event{
		Kind:                       game.EventLifeLost,
		Player:                     playerID,
		Amount:                     amount,
		PlayerEventOrdinalThisTurn: nextPlayerEventOrdinalThisTurn(g, game.EventLifeLost, playerID),
	})
	return amount
}

func startEngines(g *game.Game, playerID game.PlayerID) bool {
	player, ok := playerByID(g, playerID)
	if !ok || player.Eliminated {
		return false
	}
	if player.Speed == 0 {
		player.Speed = 1
	}
	return true
}

func increaseActivePlayerSpeedForOpponentLifeLoss(g *game.Game, losingPlayer game.PlayerID) {
	active := g.Turn.ActivePlayer
	if active == losingPlayer || active < 0 || active >= game.NumPlayers {
		return
	}
	player, ok := playerByID(g, active)
	if !ok || player.Eliminated || player.Speed <= 0 || player.Speed >= 4 || player.SpeedIncreasedTurn == g.Turn.TurnNumber {
		return
	}
	player.Speed++
	player.SpeedIncreasedTurn = g.Turn.TurnNumber
}

func ruleEffectProhibitsAttack(g *game.Game, attacker *game.Permanent, target *game.AttackTarget) bool {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind != game.RuleEffectCantAttack {
			continue
		}
		if !ruleEffectMatchesPermanent(g, effect, attacker) {
			continue
		}
		if effect.DefendingPlayer != game.PlayerAny {
			if target == nil {
				continue
			}
			if !playerRelationMatches(effect.Controller, target.Player, effect.DefendingPlayer) {
				continue
			}
		}
		return true
	}
	return false
}

func ruleEffectRequiresAttack(g *game.Game, attacker *game.Permanent) bool {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind == game.RuleEffectMustAttack && ruleEffectMatchesPermanent(g, effect, attacker) {
			return true
		}
	}
	return false
}

func ruleEffectProhibitsBlock(g *game.Game, blocker *game.Permanent) bool {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind == game.RuleEffectCantBlock && ruleEffectMatchesPermanent(g, effect, blocker) {
			return true
		}
	}
	return false
}

func ruleEffectProhibitsBeingBlocked(g *game.Game, attacker *game.Permanent) bool {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind == game.RuleEffectCantBeBlocked && ruleEffectMatchesPermanent(g, effect, attacker) {
			return true
		}
	}
	return false
}

func ruleEffectRequiresBeingBlocked(g *game.Game, attacker *game.Permanent) bool {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind == game.RuleEffectMustBeBlocked && ruleEffectMatchesPermanent(g, effect, attacker) {
			return true
		}
	}
	return false
}

func ruleEffectLimitsBlockersToOne(g *game.Game, attacker *game.Permanent) bool {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind == game.RuleEffectCantBeBlockedByMoreThanOne && ruleEffectMatchesPermanent(g, effect, attacker) {
			return true
		}
	}
	return false
}

// ruleEffectRestrictsBlocker reports whether a restricted block prohibition
// ("can't be blocked by creatures with ...") on attacker stops the given blocker
// because the blocker matches the prohibition's BlockerRestriction.
func ruleEffectRestrictsBlocker(g *game.Game, attacker, blocker *game.Permanent) bool {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind != game.RuleEffectCantBeBlockedByCreaturesWith {
			continue
		}
		if !ruleEffectMatchesPermanent(g, effect, attacker) {
			continue
		}
		if blockerMatchesRestriction(g, blocker, effect.BlockerRestriction) {
			return true
		}
	}
	return false
}

func blockerMatchesRestriction(g *game.Game, blocker *game.Permanent, restriction game.BlockerRestriction) bool {
	switch restriction.Kind {
	case game.BlockerRestrictionFlying:
		return hasKeyword(g, blocker, game.Flying)
	case game.BlockerRestrictionPowerLessOrEqual:
		return effectivePower(g, blocker) <= restriction.Power
	case game.BlockerRestrictionPowerGreaterOrEqual:
		return effectivePower(g, blocker) >= restriction.Power
	case game.BlockerRestrictionColor:
		return slices.Contains(permanentEffectiveColors(g, blocker), restriction.Color)
	case game.BlockerRestrictionArtifact:
		return permanentHasType(g, blocker, types.Artifact)
	default:
		return false
	}
}

func ruleEffectPreventsUntap(g *game.Game, permanent *game.Permanent) bool {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind == game.RuleEffectDoesntUntap && ruleEffectMatchesPermanent(g, effect, permanent) {
			return true
		}
	}
	return false
}

func ruleEffectMatchesPermanent(g *game.Game, effect *game.RuleEffect, permanent *game.Permanent) bool {
	if effect == nil {
		return false
	}
	if permanent == nil {
		return false
	}
	if !controllerRelationMatches(effect.Controller, effectiveController(g, permanent), effect.AffectedController) {
		return false
	}
	if effect.AffectedObjectID != 0 && effect.AffectedObjectID != permanent.ObjectID {
		return false
	}
	for _, cardType := range effect.PermanentTypes {
		if !permanentHasType(g, permanent, cardType) {
			return false
		}
	}
	return true
}

func controllerRelationMatches(sourceController, candidate game.PlayerID, relation game.ControllerRelation) bool {
	switch relation {
	case game.ControllerYou:
		return candidate == sourceController
	case game.ControllerOpponent, game.ControllerNotYou:
		return candidate != sourceController && candidate >= 0 && candidate < game.NumPlayers
	default:
		return true
	}
}

func playerRelationMatches(sourceController, candidate game.PlayerID, relation game.PlayerRelation) bool {
	switch relation {
	case game.PlayerAny:
		return true
	case game.PlayerYou:
		return candidate == sourceController
	case game.PlayerOpponent, game.PlayerNotYou:
		return candidate != sourceController && candidate >= 0 && candidate < game.NumPlayers
	default:
		return false
	}
}

func playerRuleEffectActive(g *game.Game, playerID game.PlayerID, kind game.RuleEffectKind) bool {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind == kind &&
			playerRelationMatches(effect.Controller, playerID, effect.AffectedPlayer) {
			return true
		}
	}
	return false
}

func staticCostModifiersForContext(g *game.Game, playerID game.PlayerID, card *game.CardDef) []game.CostModifier {
	var modifiers []game.CostModifier
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind != game.RuleEffectCostModifier {
			continue
		}
		if effect.AffectedSource {
			continue
		}
		if !playerRelationMatches(effect.Controller, playerID, effect.AffectedPlayer) {
			continue
		}
		modifier := effect.CostModifier
		if modifier.Kind != game.CostModifierSpell {
			continue
		}
		if !spellCostModifierEffectMatchesCard(g, effect, card) {
			continue
		}
		modifiers = append(modifiers, modifier)
	}
	return modifiers
}

// spellCostModifierMatchesCard reports whether a spell cost modifier's card-type
// and color filters admit the given spell card. A nil card fails any active
// filter. The colorless sentinel (MatchColor with an empty Color) matches spells
// that have no colors; otherwise the spell must carry the named color.
func spellCostModifierEffectMatchesCard(g *game.Game, effect *game.RuleEffect, card *game.CardDef) bool {
	modifier := effect.CostModifier
	if !spellCostModifierBaseMatchesCard(modifier, card) {
		return false
	}
	if !modifier.ChosenSubtypeFromEntryChoice {
		return true
	}
	source, ok := permanentByObjectID(g, effect.SourceObjectID)
	if !ok || card == nil {
		return false
	}
	choice, ok := source.EntryChoices[game.EntryTypeChoiceKey]
	return ok &&
		choice.Kind == game.ResolutionChoiceSubtype &&
		types.KnownSubtypeForType(types.Creature, choice.Subtype) &&
		card.HasSubtype(choice.Subtype)
}

func spellCostModifierMatchesCard(modifier game.CostModifier, card *game.CardDef) bool {
	return !modifier.ChosenSubtypeFromEntryChoice && spellCostModifierBaseMatchesCard(modifier, card)
}

func spellCostModifierBaseMatchesCard(modifier game.CostModifier, card *game.CardDef) bool {
	if modifier.MatchCardType && (card == nil || !card.HasType(modifier.CardType)) {
		return false
	}
	if modifier.MatchColor {
		if card == nil {
			return false
		}
		if modifier.Color == "" {
			if len(card.Colors) != 0 {
				return false
			}
		} else if !slices.Contains(card.Colors, modifier.Color) {
			return false
		}
	}
	if len(modifier.MatchColors) != 0 {
		if card == nil {
			return false
		}
		matched := false
		for _, c := range modifier.MatchColors {
			if slices.Contains(card.Colors, c) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	if len(modifier.MatchSubtypes) != 0 {
		if card == nil {
			return false
		}
		if !slices.ContainsFunc(modifier.MatchSubtypes, card.HasSubtype) {
			return false
		}
	}
	return true
}

func canCastFromZoneByRuleEffect(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex) bool {
	if sourceZone == zone.Graveyard {
		return len(castPermissionsForZone(g, playerID, cardID, sourceZone, face)) > 0
	}
	return hasCastFromZoneRuleEffect(g, playerID, cardID, sourceZone, face)
}

func castPermissionsForZone(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex) []payment.SpellCastPermission {
	if sourceZone != zone.Graveyard {
		return []payment.SpellCastPermission{payment.SpellCastPermissionDefault}
	}
	var permissions []payment.SpellCastPermission
	card, cardOK := g.GetCardInstance(cardID)
	if face == game.FaceFront && cardOK && cardHasFlashbackAlternative(card) {
		permissions = append(permissions, payment.SpellCastPermissionFlashback)
	}
	if hasCastFromZoneRuleEffect(g, playerID, cardID, sourceZone, face) {
		permissions = append(permissions, payment.SpellCastPermissionRuleEffect)
	}
	return permissions
}

func hasCastFromZoneRuleEffect(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex) bool {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if (effect.Kind != game.RuleEffectCastFromZone && effect.Kind != game.RuleEffectPlayFromZone) ||
			effect.CastFromZone != sourceZone {
			continue
		}
		if !playerRelationMatches(effect.Controller, playerID, effect.AffectedPlayer) {
			continue
		}
		if effect.AffectedCardID != 0 && effect.AffectedCardID != cardID {
			continue
		}
		if effect.CastFace.Exists {
			if effect.CastFace.Val != face {
				continue
			}
		} else if effect.Kind == game.RuleEffectCastFromZone && face != game.FaceFront {
			continue
		}
		return true
	}
	return false
}

func canPlayLandFromZoneByRuleEffect(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type) bool {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.CastFromZone != sourceZone ||
			!playerRelationMatches(effect.Controller, playerID, effect.AffectedPlayer) {
			continue
		}
		switch effect.Kind {
		case game.RuleEffectPlayFromZone:
			if effect.AffectedCardID == cardID {
				return true
			}
		case game.RuleEffectPlayLandsFromZone:
			if effect.TopCardOnly && !cardIsTopOfLibrary(g, playerID, cardID) {
				continue
			}
			return true
		default:
			continue
		}
	}
	return false
}

// canCastSpellsFromZoneByRuleEffect reports whether a continuous
// RuleEffectCastSpellsFromZone permission lets playerID cast the face of cardID
// from sourceZone ("You may cast spells from the top of your library.", Future
// Sight). A non-empty SpellTypes filter requires the cast face to have at least
// one of the listed card types; TopCardOnly requires the card to be on top of
// the player's library.
func canCastSpellsFromZoneByRuleEffect(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex) bool {
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return false
	}
	faceTypes := cardFaceOrDefault(card, face).Types
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind != game.RuleEffectCastSpellsFromZone ||
			effect.CastFromZone != sourceZone ||
			!playerRelationMatches(effect.Controller, playerID, effect.AffectedPlayer) {
			continue
		}
		if effect.TopCardOnly && !cardIsTopOfLibrary(g, playerID, cardID) {
			continue
		}
		if len(effect.SpellTypes) > 0 && !slices.ContainsFunc(effect.SpellTypes, func(t types.Card) bool {
			return slices.Contains(faceTypes, t)
		}) {
			continue
		}
		return true
	}
	return false
}

// cardIsTopOfLibrary reports whether cardID is the top card of playerID's
// library.
func cardIsTopOfLibrary(g *game.Game, playerID game.PlayerID, cardID id.ID) bool {
	player, ok := playerByID(g, playerID)
	if !ok {
		return false
	}
	top, ok := player.Library.Top()
	return ok && top == cardID
}

// playerPlaysWithTopCardRevealed reports whether playerID plays with the top card
// of their library revealed to all players (a visibility static).
func playerPlaysWithTopCardRevealed(g *game.Game, playerID game.PlayerID) bool {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind != game.RuleEffectPlayWithTopCardRevealed {
			continue
		}
		if playerRelationMatches(effect.Controller, playerID, effect.AffectedPlayer) {
			return true
		}
	}
	return false
}

func castableZonesForPlayer(g *game.Game, playerID game.PlayerID) []zone.Type {
	zones := []zone.Type{zone.Hand}
	if player, ok := playerByID(g, playerID); ok {
		for _, cardID := range player.Graveyard.All() {
			card, ok := g.GetCardInstance(cardID)
			if !ok {
				continue
			}
			for _, face := range card.Def.LegalCastFaces() {
				if !canCastFromZoneByRuleEffect(g, playerID, cardID, zone.Graveyard, face) {
					continue
				}
				zones = append(zones, zone.Graveyard)
				break
			}
		}
		for _, cardID := range player.Exile.All() {
			card, ok := g.GetCardInstance(cardID)
			if !ok {
				continue
			}
			if g.AdventureCards[cardID] || slices.ContainsFunc(card.Def.LegalCastFaces(), func(face game.FaceIndex) bool {
				return canCastFromZoneByRuleEffect(g, playerID, cardID, zone.Exile, face)
			}) {
				zones = append(zones, zone.Exile)
				break
			}
		}
		if topID, ok := player.Library.Top(); ok {
			if card, cardOK := g.GetCardInstance(topID); cardOK &&
				slices.ContainsFunc(card.Def.LegalCastFaces(), func(face game.FaceIndex) bool {
					return canCastSpellsFromZoneByRuleEffect(g, playerID, topID, zone.Library, face)
				}) {
				zones = append(zones, zone.Library)
			}
		}
	}
	return slices.Compact(zones)
}

func cardHasFlashbackAlternative(card *game.CardInstance) bool {
	frontDef := cardFaceOrDefault(card, game.FaceFront)
	if !frontDef.HasKeyword(game.Flashback) {
		return false
	}
	if flashbackCost, ok := frontDef.FlashbackCost(); ok && len(flashbackCost) > 0 {
		return true
	}
	return slices.ContainsFunc(frontDef.AlternativeCosts, isFlashbackAlternative)
}
