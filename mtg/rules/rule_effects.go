package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

func createRuleEffectTemplates(g *game.Game, obj *game.StackObject, object opt.V[game.ObjectReference], templates []game.RuleEffect, duration game.EffectDuration) bool {
	if len(templates) == 0 {
		return false
	}
	sourceID, sourceObjectID := damageSourceIDs(g, obj)
	appended := false
	for i := range templates {
		ruleEffect := templates[i]
		ruleEffect.ID = g.IDGen.Next()
		ruleEffect.Controller = obj.Controller
		ruleEffect.SourceCardID = sourceID
		ruleEffect.SourceObjectID = sourceObjectID
		if ruleEffect.AffectedSource {
			ruleEffect.AffectedObjectID = sourceObjectID
		} else if ruleEffect.AffectedObjectID == 0 && object.Exists {
			// An object-scoped rule effect that names a target slot must apply
			// only to that resolved permanent. When the slot is unfilled (an "up
			// to N" target the controller declined) or its target became
			// illegal, the reference does not resolve; skip the template rather
			// than appending an AffectedObjectID==0 effect, which would otherwise
			// match every permanent on the battlefield.
			resolved, ok := resolveObjectReference(g, obj, object.Val)
			if !ok || resolved.permanent == nil {
				continue
			}
			ruleEffect.AffectedObjectID = resolved.permanent.ObjectID
		}
		if ruleEffect.AffectedPlayerRef.Kind() != game.PlayerReferenceNone {
			affected, ok := resolvePlayerReference(g, obj, ruleEffect.AffectedPlayerRef)
			if !ok {
				continue
			}
			ruleEffect.AffectedSpecificPlayer = opt.Val(affected)
		}
		if ruleEffect.RequiredAttackTargetRef.Kind() != game.PlayerReferenceNone {
			required, ok := resolvePlayerReference(g, obj, ruleEffect.RequiredAttackTargetRef)
			if !ok {
				continue
			}
			ruleEffect.RequiredAttackTarget = opt.Val(required)
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
		appended = true
	}
	return appended
}

func activeRuleEffects(g *game.Game) []game.RuleEffect {
	fc := frameCacheFor(g)
	if fc != nil && fc.ruleEffectsBuilt {
		return fc.ruleEffects
	}
	effects := make([]game.RuleEffect, 0, len(g.RuleEffects))
	for i := range g.RuleEffects {
		if ruleEffectSourceIsActive(g, &g.RuleEffects[i]) {
			effects = append(effects, g.RuleEffects[i])
		}
	}
	effects = append(effects, staticRuleEffects(g)...)
	effects = append(effects, stackStaticRuleEffects(g)...)
	effects = append(effects, graveyardStaticRuleEffects(g)...)
	effects = append(effects, exileStaticRuleEffects(g)...)
	if fc != nil {
		// Clip so a caller that appends reallocates instead of writing into the
		// shared backing array.
		fc.ruleEffects = effects[:len(effects):len(effects)]
		fc.ruleEffectsBuilt = true
		return fc.ruleEffects
	}
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

// graveyardStaticRuleEffects gathers the rule effects of static abilities that
// function while their source card is in a graveyard ("You may cast this card
// from your graveyard ...", Gravecrawler, Hogaak). Each effect is scoped to its
// graveyard owner; a source-affecting permission self-scopes to the graveyard
// card itself so it grants permission only to cast that card.
func graveyardStaticRuleEffects(g *game.Game) []game.RuleEffect {
	var effects []game.RuleEffect
	for owner := range game.PlayerID(game.NumPlayers) {
		player := g.Players[owner]
		for _, cardID := range player.Graveyard.All() {
			_, def, ok := cardInstanceFaceDef(g, cardID, game.FaceFront)
			if !ok {
				continue
			}
			for i := range def.StaticAbilities {
				body := &def.StaticAbilities[i]
				if body.ZoneOfFunction != zone.Graveyard || len(body.RuleEffects) == 0 {
					continue
				}
				if !conditionSatisfied(g, conditionContext{controller: owner}, body.Condition) {
					continue
				}
				for j := range body.RuleEffects {
					ruleEffect := body.RuleEffects[j]
					ruleEffect.Controller = owner
					ruleEffect.SourceCardID = cardID
					if ruleEffect.AffectedSource {
						ruleEffect.AffectedCardID = cardID
					}
					effects = append(effects, ruleEffect)
				}
			}
		}
	}
	return effects
}

// exileStaticRuleEffects gathers the rule effects of static abilities that
// function while their source card is in exile ("You may cast this card from
// exile.", Misthollow Griffin, Eternal Scourge). Each effect is scoped to its
// exile owner; a source-affecting permission self-scopes to the exiled card
// itself so it grants permission only to cast that card.
func exileStaticRuleEffects(g *game.Game) []game.RuleEffect {
	var effects []game.RuleEffect
	for owner := range game.PlayerID(game.NumPlayers) {
		player := g.Players[owner]
		for _, cardID := range player.Exile.All() {
			_, def, ok := cardInstanceFaceDef(g, cardID, game.FaceFront)
			if !ok {
				continue
			}
			for i := range def.StaticAbilities {
				body := &def.StaticAbilities[i]
				if body.ZoneOfFunction != zone.Exile || len(body.RuleEffects) == 0 {
					continue
				}
				if !conditionSatisfied(g, conditionContext{controller: owner}, body.Condition) {
					continue
				}
				for j := range body.RuleEffects {
					ruleEffect := body.RuleEffects[j]
					ruleEffect.Controller = owner
					ruleEffect.SourceCardID = cardID
					if ruleEffect.AffectedSource {
						ruleEffect.AffectedCardID = cardID
					}
					effects = append(effects, ruleEffect)
				}
			}
		}
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

// persistsWhileCardExiled reports whether effect is a lasting permission to play
// or cast a specific card "for as long as it remains exiled" (Court of
// Locthwain). Such a permission is created by a resolved ability, not a static
// ability, so it persists independent of its source permanent: it remains active
// as long as the affected card is still in exile, even if the source leaves the
// battlefield or changes controllers (CR 610.3b). Ordinary DurationPermanent
// effects, by contrast, end when their source leaves the battlefield.
func persistsWhileCardExiled(g *game.Game, effect *game.RuleEffect) bool {
	if effect == nil ||
		effect.Duration != game.DurationPermanent ||
		effect.AffectedCardID == 0 ||
		effect.CastFromZone != zone.Exile {
		return false
	}
	if effect.Kind != game.RuleEffectPlayFromZone && effect.Kind != game.RuleEffectCastFromZone {
		return false
	}
	z, ok := cardZone(g, effect.AffectedCardID)
	return ok && z == zone.Exile
}

func ruleEffectSourceStillApplies(g *game.Game, effect *game.RuleEffect) bool {
	if effect == nil {
		return false
	}
	if persistsWhileCardExiled(g, effect) {
		return true
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
	if persistsWhileCardExiled(g, effect) {
		return true
	}
	if effect.Duration != game.DurationPermanent || effect.SourceObjectID == 0 {
		return true
	}
	source, ok := permanentByObjectID(g, effect.SourceObjectID)
	return ok && activeBattlefieldPermanent(source)
}

// expireEndOfCombatRuleEffects removes rule effects that last only until the end
// of combat (DurationUntilEndOfCombat, "this combat" — Canal Courier). It is
// called as the combat phase is torn down, before the following phases.
func expireEndOfCombatRuleEffects(g *game.Game) {
	if len(g.RuleEffects) == 0 {
		return
	}
	kept := g.RuleEffects[:0]
	for i := range g.RuleEffects {
		if g.RuleEffects[i].Duration == game.DurationUntilEndOfCombat {
			continue
		}
		kept = append(kept, g.RuleEffects[i])
	}
	g.RuleEffects = kept
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
		// A "this combat" effect (DurationUntilEndOfCombat) is normally removed as
		// the combat phase tears down; drop any that survive to the turn's cleanup
		// as a backstop so one never outlives its turn.
		if effect.Duration == game.DurationUntilEndOfCombat {
			continue
		}
		// "Until your next end step" expires at the cleanup that follows the
		// controller's next end step. expireRuleEffects runs at every cleanup, so
		// the effect is removed on the first cleanup whose turn belongs to the
		// player it expires for: the creating turn when made during that player's
		// own turn, otherwise their next turn.
		if effect.Duration == game.DurationUntilYourNextEndStep &&
			effect.ExpiresFor == g.Turn.ActivePlayer {
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
// cast card as though it had flash, i.e. at instant speed ("You may cast spells
// this turn as though they had flash.", Borne Upon a Wind; "You may cast spells
// as though they had flash.", Vedalken Orrery; CR 702.8 / 601.3e). An effect's
// optional SpellTypes/SpellSubtypes filters narrow the grant to spells of those
// card types ("sorcery spells") or subtypes ("Aura and Equipment spells").
func playerCanCastAsThoughFlash(g *game.Game, playerID game.PlayerID, card *game.CardDef) bool {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind != game.RuleEffectCastSpellsAsThoughFlash {
			continue
		}
		if !playerRelationMatches(effect.Controller, playerID, effect.AffectedPlayer) {
			continue
		}
		if len(effect.SpellTypes) != 0 && !cardDefHasAnyType(card, effect.SpellTypes) {
			continue
		}
		if len(effect.SpellSubtypes) != 0 && !card.HasAnySubtype(effect.SpellSubtypes...) {
			continue
		}
		return true
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

// manaProductionMultiplierFor returns the factor by which mana produced when
// playerID taps a permanent for mana is multiplied, the product of all active
// RuleEffectManaProductionMultiplier effects that player controls ("If you tap a
// permanent for mana, it produces twice as much of that mana instead.", Mana
// Reflection; "... three times as much ...", Nyxbloom Ancient). It returns 1 when
// no such effect applies, so the common case is unchanged. Multiple multipliers
// compound multiplicatively (CR 616 lets the affected player order overlapping
// replacement effects; the product is order-independent).
func manaProductionMultiplierFor(g *game.Game, playerID game.PlayerID) int {
	multiplier := 1
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind != game.RuleEffectManaProductionMultiplier {
			continue
		}
		if effect.Controller != playerID {
			continue
		}
		if effect.ManaProductionMultiplier > 1 {
			multiplier *= effect.ManaProductionMultiplier
		}
	}
	return multiplier
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

// playerSkipsDrawStep reports whether an active rule effect makes playerID skip
// their draw step ("Skip your draw step.", Necropotence, Yawgmoth's Bargain).
// When it does, the draw step does not happen at all: no beginning-of-step
// triggers, no turn-based draw, and no priority during that step (CR 500.8).
func playerSkipsDrawStep(g *game.Game, playerID game.PlayerID) bool {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind != game.RuleEffectSkipDrawStep {
			continue
		}
		if playerRelationMatches(effect.Controller, playerID, effect.AffectedPlayer) {
			return true
		}
	}
	return false
}

// entryFromZoneProhibited reports whether an active RuleEffectCantEnterFromZones
// effect forbids a card with the given definition from entering the battlefield
// out of sourceZone ("Creature cards in graveyards and libraries can't enter the
// battlefield.", Grafdigger's Cage; "Permanent cards in graveyards can't enter
// the battlefield.", Soulless Jailer). The restriction is global. An empty
// PermanentTypes restricts every permanent card; EnterExcludeLandCards exempts
// land cards for the "nonland permanent" forms.
func entryFromZoneProhibited(g *game.Game, def *game.CardDef, sourceZone zone.Type) bool {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind != game.RuleEffectCantEnterFromZones ||
			!slices.Contains(effect.EnterFromZones, sourceZone) {
			continue
		}
		if effect.EnterExcludeLandCards && def.HasType(types.Land) {
			continue
		}
		if len(effect.PermanentTypes) == 0 || cardDefHasAnyType(def, effect.PermanentTypes) {
			return true
		}
	}
	return false
}

// castFromZoneProhibited reports whether an active RuleEffectCantCastFromZones
// effect forbids playerID from casting a spell out of sourceZone ("Your
// opponents can't cast spells from anywhere other than their hands.", Drannith
// Magistrate; "Players can't cast spells from graveyards or libraries.",
// Grafdigger's Cage). A "can't" restriction overrides any casting permission.
func castFromZoneProhibited(g *game.Game, playerID game.PlayerID, sourceZone zone.Type) bool {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind != game.RuleEffectCantCastFromZones ||
			!playerRelationMatches(effect.Controller, playerID, effect.AffectedPlayer) ||
			!actionRestrictionTurnActive(g, effect) {
			continue
		}
		if slices.Contains(effect.CantCastFromZones, sourceZone) {
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
		if len(effect.ExcludedSpellTypes) > 0 && cardDefHasAnyType(spellDef, effect.ExcludedSpellTypes) {
			continue
		}
		return true
	}
	return false
}

// abilityActivationProhibited reports whether an active activation-prohibition
// rule effect forbids playerID from activating an ability of permanent. The
// player-scoped RuleEffectCantActivateAbilities ("... activate abilities of
// artifacts, creatures, or enchantments.") honors its affected-player relation,
// optional controller-turn scope, and permanent-type filter. The
// permanent-scoped RuleEffectCantActivateAbilitiesOfPermanent ("Enchanted
// creature ... its activated abilities can't be activated.", Arrest) forbids any
// player from activating the matched permanent's abilities, sparing mana
// abilities when isManaAbility and ExemptManaAbilities both hold (Faith's
// Fetters).
func abilityActivationProhibited(g *game.Game, playerID game.PlayerID, permanent *game.Permanent, isManaAbility bool) bool {
	card, ok := permanentCardDef(g, permanent)
	if !ok {
		return false
	}
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		switch effect.Kind {
		case game.RuleEffectCantActivateAbilities:
			if !playerRelationMatches(effect.Controller, playerID, effect.AffectedPlayer) ||
				!actionRestrictionTurnActive(g, effect) {
				continue
			}
			if len(effect.PermanentTypes) > 0 && !cardDefHasAnyType(card, effect.PermanentTypes) {
				continue
			}
			return true
		case game.RuleEffectCantActivateAbilitiesOfPermanent:
			if !ruleEffectMatchesPermanent(g, effect, permanent) {
				continue
			}
			if isManaAbility && effect.ExemptManaAbilities {
				continue
			}
			return true
		default:
			continue
		}
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
	amount = replacementLifeLossAmount(g, playerID, amount)
	if amount <= 0 {
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
			if effect.DefendingPlayerDirectOnly && !target.IsPlayerAttack() {
				// "Can't attack you" restricts direct attacks on the defending
				// player only; a planeswalker or battle that player controls is a
				// distinct attack target (CR 508.1) and stays attackable.
				continue
			}
			if !playerRelationMatches(effect.Controller, target.Player, effect.DefendingPlayer) {
				continue
			}
		}
		if !effect.AttackDefenderControlsSelection.Empty() {
			// "Can't attack unless defending player controls ...": the attacker may
			// attack only a defending player who controls a matching permanent. With
			// no specific target the restriction cannot rule out every defender, so
			// the attacker remains able to attack someone.
			if target == nil {
				continue
			}
			if defendingPlayerControlsSelection(g, effect, target.Player) {
				continue
			}
		}
		if effect.AttackDefenderIsMonarch {
			// "Can't attack unless defending player is the monarch": the attacker
			// may attack only a defending player who currently holds the monarch
			// designation. With no specific target the restriction cannot rule out
			// every defender, so the attacker remains able to attack someone.
			if target == nil {
				continue
			}
			if player, ok := playerByID(g, target.Player); ok && player.IsMonarch {
				continue
			}
		}
		return true
	}
	return false
}

// defendingPlayerControlsSelection reports whether the defending player controls
// at least one active battlefield permanent matching the effect's
// AttackDefenderControlsSelection, gating a conditional "can't attack unless
// defending player controls ..." restriction.
func defendingPlayerControlsSelection(g *game.Game, effect *game.RuleEffect, defender game.PlayerID) bool {
	sel := &effect.AttackDefenderControlsSelection
	for _, permanent := range g.Battlefield {
		if !activeBattlefieldPermanent(permanent) {
			continue
		}
		if effectiveController(g, permanent) != defender {
			continue
		}
		values := effectivePermanentValues(g, permanent)
		subject := selectionSubject{
			kind:           subjectPermanent,
			g:              g,
			permanent:      permanent,
			values:         &values,
			viewer:         effect.Controller,
			sourceObjectID: effect.SourceObjectID,
		}
		if sel.Controller != game.ControllerAny {
			subject.controller = effectiveController(g, permanent)
		}
		if matchSelection(&subject, sel) {
			return true
		}
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

// ruleEffectPermitsAttackDespiteDefender reports whether attacker carries a
// "can attack ... as though it didn't have defender" permission
// (RuleEffectCanAttackAsThoughDefender): a defender creature with this active
// permission may be declared as an attacker (CR 508.1a).
func ruleEffectPermitsAttackDespiteDefender(g *game.Game, attacker *game.Permanent) bool {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind == game.RuleEffectCanAttackAsThoughDefender && ruleEffectMatchesPermanent(g, effect, attacker) {
			return true
		}
	}
	return false
}

// ruleEffectProhibitsAttackingAlone reports whether attacker carries a
// "can't attack alone" restriction (RuleEffectCantAttackAlone): it may not be
// declared as an attacker unless at least one other creature also attacks.
func ruleEffectProhibitsAttackingAlone(g *game.Game, attacker *game.Permanent) bool {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind == game.RuleEffectCantAttackAlone && ruleEffectMatchesPermanent(g, effect, attacker) {
			return true
		}
	}
	return false
}

// ruleEffectProhibitsBlockingAlone reports whether blocker carries a
// "can't block alone" restriction (RuleEffectCantBlockAlone): it may not be
// declared as a blocker unless at least one other creature also blocks.
func ruleEffectProhibitsBlockingAlone(g *game.Game, blocker *game.Permanent) bool {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind == game.RuleEffectCantBlockAlone && ruleEffectMatchesPermanent(g, effect, blocker) {
			return true
		}
	}
	return false
}

func ruleEffectProhibitsBlock(g *game.Game, blocker *game.Permanent) bool {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind != game.RuleEffectCantBlock {
			continue
		}
		if effect.BlockedSource || !effect.BlockedSelection.Empty() {
			continue
		}
		if ruleEffectMatchesPermanent(g, effect, blocker) {
			return true
		}
	}
	return false
}

// ruleEffectProhibitsBlockingAttacker reports whether a conditional
// RuleEffectCantBlock restriction stops blocker from blocking attacker because
// blocker matches the affected (restricted) group and attacker is the protected
// object the restriction shields ("Creatures with power less than this
// creature's power can't block it.", "... can't block creatures you control.").
func ruleEffectProhibitsBlockingAttacker(g *game.Game, blocker, attacker *game.Permanent) bool {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind != game.RuleEffectCantBlock {
			continue
		}
		if !effect.BlockedSource && effect.BlockedSelection.Empty() {
			continue
		}
		if !ruleEffectMatchesPermanent(g, effect, blocker) {
			continue
		}
		if effect.BlockedSource {
			if attacker != nil && attacker.ObjectID == effect.SourceObjectID {
				return true
			}
			continue
		}
		if ruleEffectBlockedSelectionMatches(g, effect, attacker) {
			return true
		}
	}
	return false
}

// ruleEffectBlockedSelectionMatches reports whether attacker satisfies a
// conditional can't-block restriction's protected-object Selection ("can't block
// creatures you control"). It builds a permanent selection subject viewed from
// the effect's controller so the selection's controller relation resolves "you"
// to the source controller.
func ruleEffectBlockedSelectionMatches(g *game.Game, effect *game.RuleEffect, attacker *game.Permanent) bool {
	if attacker == nil {
		return false
	}
	sel := &effect.BlockedSelection
	values := effectivePermanentValues(g, attacker)
	subject := selectionSubject{
		kind:           subjectPermanent,
		g:              g,
		permanent:      attacker,
		values:         &values,
		viewer:         effect.Controller,
		sourceObjectID: effect.SourceObjectID,
	}
	if sel.Controller != game.ControllerAny {
		subject.controller = effectiveController(g, attacker)
	}
	return matchSelection(&subject, sel)
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

// ruleEffectRequiresBeingBlockedByAllAble reports whether attacker carries a
// true-lure requirement (every creature able to block it must do so, CR 509.1c).
func ruleEffectRequiresBeingBlockedByAllAble(g *game.Game, attacker *game.Permanent) bool {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind == game.RuleEffectMustBeBlockedByAllAble && ruleEffectMatchesPermanent(g, effect, attacker) {
			return true
		}
	}
	return false
}

// ruleEffectAssignsCombatDamageAsThoughUnblocked reports whether attacker may
// assign its combat damage to its attack target as though it weren't blocked.
func ruleEffectAssignsCombatDamageAsThoughUnblocked(g *game.Game, attacker *game.Permanent) bool {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind == game.RuleEffectAssignCombatDamageAsThoughUnblocked && ruleEffectMatchesPermanent(g, effect, attacker) {
			return true
		}
	}
	return false
}

// ruleEffectAssignsCombatDamageByToughness reports whether a static combat-damage
// replacement ("<subject> assigns combat damage equal to its toughness rather
// than its power.") applies to permanent, making it assign combat damage equal to
// its toughness instead of its power (Doran, the Siege Tower; Assault Formation).
func ruleEffectAssignsCombatDamageByToughness(g *game.Game, permanent *game.Permanent) bool {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind == game.RuleEffectAssignCombatDamageUsingToughness && ruleEffectMatchesPermanent(g, effect, permanent) {
			return true
		}
	}
	return false
}

// combatDamageAssignedBy returns the amount of combat damage permanent assigns:
// its toughness when a static combat-damage replacement applies to it ("assigns
// combat damage equal to its toughness rather than its power.", Doran, the Siege
// Tower), and otherwise its power (CR 510.1a). Negative toughness contributes no
// damage, matching the power floor.
func combatDamageAssignedBy(g *game.Game, permanent *game.Permanent) int {
	if ruleEffectAssignsCombatDamageByToughness(g, permanent) {
		if toughness, ok := effectiveToughness(g, permanent); ok {
			return max(0, toughness)
		}
	}
	return effectivePower(g, permanent)
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

// blockerBlockLimit reports how many attackers the given blocker may block. A
// creature blocks at most one attacker by default (CR 509.1a); each active
// RuleEffectCanBlockAdditional matching the blocker raises that limit by its
// AdditionalBlockCount ("This creature can block an additional creature each
// combat.", Brave the Sands, and the group "Each creature you control can block
// an additional creature each combat." forms).
func blockerBlockLimit(g *game.Game, blocker *game.Permanent) int {
	limit := 1
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind == game.RuleEffectCanBlockAdditional && ruleEffectMatchesPermanent(g, effect, blocker) {
			limit += effect.AdditionalBlockCount
		}
	}
	return limit
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
	case game.BlockerRestrictionDefender:
		return hasKeyword(g, blocker, game.Defender)
	case game.BlockerRestrictionLegendary:
		return permanentHasSupertype(g, blocker, types.Legendary)
	case game.BlockerRestrictionControlledByMonarch:
		player, ok := playerByID(g, effectiveController(g, blocker))
		return ok && player.IsMonarch
	default:
		return false
	}
}

// ruleEffectRestrictsBlockerExcept reports whether a "can't be blocked except by
// ..." prohibition on attacker stops the given blocker because the blocker does
// not match the prohibition's BlockerRestriction. The restriction names the only
// blockers allowed to block the attacker; every other blocker is prohibited.
func ruleEffectRestrictsBlockerExcept(g *game.Game, attacker, blocker *game.Permanent) bool {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind != game.RuleEffectCantBeBlockedExceptBy {
			continue
		}
		if !ruleEffectMatchesPermanent(g, effect, attacker) {
			continue
		}
		if !blockerMatchesRestriction(g, blocker, effect.BlockerRestriction) {
			return true
		}
	}
	return false
}

// ruleEffectLimitsBlockerToCreaturesWith reports whether a blocker-side "can
// block only creatures with ..." permission restriction on blocker stops it from
// blocking attacker because attacker does not match the restriction's
// BlockerRestriction characteristic ("This creature can block only creatures with
// flying."). The restriction describes the attacker the blocker may block, so the
// match is tested against attacker rather than against another blocker.
func ruleEffectLimitsBlockerToCreaturesWith(g *game.Game, blocker, attacker *game.Permanent) bool {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind != game.RuleEffectCanBlockOnlyCreaturesWith {
			continue
		}
		if !ruleEffectMatchesPermanent(g, effect, blocker) {
			continue
		}
		if !blockerMatchesRestriction(g, attacker, effect.BlockerRestriction) {
			return true
		}
	}
	return false
}

func ruleEffectPreventsUntap(g *game.Game, permanent *game.Permanent) bool {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind != game.RuleEffectDoesntUntap || !ruleEffectMatchesPermanent(g, effect, permanent) {
			continue
		}
		if effect.UntapUnlessControllerIsMonarch {
			// "... doesn't untap during its controller's untap step unless that
			// player is the monarch": the prohibition lifts while the affected
			// permanent's controller currently holds the monarch designation, so the
			// permanent untaps normally (Fall from Favor).
			controller := effectiveController(g, permanent)
			if player, ok := playerByID(g, controller); ok && player.IsMonarch {
				continue
			}
		}
		return true
	}
	return false
}

// ruleEffectPreventsTransform reports whether an active rule effect forbids
// permanent from transforming ("Non-Human Werewolves you control can't
// transform.", Immerwolf). A matching prohibition stops the transform (CR
// 701.28), so transformPermanent does nothing.
func ruleEffectPreventsTransform(g *game.Game, permanent *game.Permanent) bool {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind == game.RuleEffectCantTransform && ruleEffectMatchesPermanent(g, effect, permanent) {
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
	if effect.AffectedSpecificPlayer.Exists {
		if effectiveController(g, permanent) != effect.AffectedSpecificPlayer.Val {
			return false
		}
	} else if !controllerRelationMatches(effect.Controller, effectiveController(g, permanent), effect.AffectedController) {
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
	if !effect.AffectedSelection.Empty() && !ruleEffectAffectedSelectionMatches(g, effect, permanent) {
		return false
	}
	return true
}

// ruleEffectAffectedSelectionMatches reports whether permanent satisfies a
// group-scoped rule effect's affected-permanent Selection filter ("Blue creatures
// you control ...", "Creatures you control with +1/+1 counters on them ..."). It
// builds a permanent selection subject viewed from the effect's controller and
// excluding the effect's source object, mirroring group-membership matching.
func ruleEffectAffectedSelectionMatches(g *game.Game, effect *game.RuleEffect, permanent *game.Permanent) bool {
	sel := &effect.AffectedSelection
	values := effectivePermanentValues(g, permanent)
	subject := selectionSubject{
		kind:           subjectPermanent,
		g:              g,
		permanent:      permanent,
		values:         &values,
		viewer:         effect.Controller,
		sourceObjectID: effect.SourceObjectID,
	}
	if sel.Controller != game.ControllerAny {
		subject.controller = effectiveController(g, permanent)
	}
	return matchSelection(&subject, sel)
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

// ruleEffectAffectsPlayer reports whether a play/cast-from-zone permission
// authorizes candidate. It resolves the owner-scoped permission Prowl, Stoic
// Strategist grants ("its owner may play it"): when AffectToOwner is set, only
// the owner of the exiled card AffectedCardID is authorized, which may be an
// opponent of the effect's controller. Every other permission stays
// controller-relative through playerRelationMatches, so this is a no-op for
// existing effects.
func ruleEffectAffectsPlayer(g *game.Game, effect *game.RuleEffect, candidate game.PlayerID) bool {
	if effect.AffectToOwner {
		card, ok := g.GetCardInstance(effect.AffectedCardID)
		return ok && card.Owner == candidate
	}
	return playerRelationMatches(effect.Controller, candidate, effect.AffectedPlayer)
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

// permanentCantBeSacrificed reports whether an active RuleEffectCantBeSacrificed
// protects permanent from being sacrificed ("Creatures you control but don't own
// ... can't be sacrificed."). It matches whether the sacrifice would be a cost
// or an effect, so a protected permanent is never a legal sacrifice.
func permanentCantBeSacrificed(g *game.Game, permanent *game.Permanent) bool {
	if permanent == nil {
		return false
	}
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind == game.RuleEffectCantBeSacrificed &&
			ruleEffectMatchesPermanent(g, effect, permanent) {
			return true
		}
	}
	return false
}

// playerDamageRedirectPermanent returns the permanent an active
// RuleEffectRedirectDamageToSource redirects the given player's damage to ("All
// damage that would be dealt to you is dealt to this creature instead." —
// Protector of the Crown). The redirect target is the rule effect's source
// permanent; it returns the first active redirect whose affected player is the
// damaged player.
func playerDamageRedirectPermanent(g *game.Game, playerID game.PlayerID) (*game.Permanent, bool) {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind != game.RuleEffectRedirectDamageToSource ||
			!playerRelationMatches(effect.Controller, playerID, effect.AffectedPlayer) {
			continue
		}
		if permanent, ok := permanentByObjectID(g, effect.SourceObjectID); ok && activeBattlefieldPermanent(permanent) {
			return permanent, true
		}
	}
	return nil, false
}

// payLifeForManaColorActive reports whether an active RuleEffectPayLifeForColoredMana
// effect lets playerID pay 2 life instead of a mana of color c when paying a cost
// ("For each {B} in a cost, you may pay 2 life rather than pay that mana.", K'rrik).
func payLifeForManaColorActive(g *game.Game, playerID game.PlayerID, c mana.Color) bool {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind != game.RuleEffectPayLifeForColoredMana {
			continue
		}
		if effect.ManaColor == c &&
			playerRelationMatches(effect.Controller, playerID, effect.AffectedPlayer) {
			return true
		}
	}
	return false
}

func staticCostModifiersForContext(g *game.Game, playerID game.PlayerID, card *game.CardDef, sourceZone zone.Type, targets []game.Target) []game.CostModifier {
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
		if !spellCostModifierMatchesZone(modifier, sourceZone) {
			continue
		}
		if !spellCostModifierMatchesTargets(modifier, effect.SourceObjectID, targets) {
			continue
		}
		if modifier.SharedExiledCardTypeReduction > 0 {
			reduction := sharedExiledCardTypeReduction(g, effect.SourceObjectID, modifier, card)
			if reduction <= 0 {
				continue
			}
			modifiers = append(modifiers, game.CostModifier{Kind: game.CostModifierSpell, GenericReduction: reduction})
			continue
		}
		if modifier.PerObjectReduction > 0 {
			if !actionRestrictionTurnActive(g, effect) {
				continue
			}
			reduction := perObjectGroupReduction(g, effect.SourceObjectID, effect.Controller, modifier)
			if reduction <= 0 {
				continue
			}
			modifiers = append(modifiers, game.CostModifier{Kind: game.CostModifierSpell, GenericReduction: reduction})
			continue
		}
		modifiers = append(modifiers, modifier)
	}
	return modifiers
}

// perObjectGroupReduction resolves the concrete generic reduction for a group
// spell cost modifier whose amount scales with a countable battlefield permanent
// the modifier controller controls ("[<filter>] spells you cast cost {N} less to
// cast for each <permanent> you control[ with power M or greater].", Temur
// Battlecrier; Hamza, Guardian of Arashin). It counts the permanents the source
// permanent's controller controls that match CountSelection and multiplies that
// count by the per-permanent amount.
func perObjectGroupReduction(g *game.Game, sourceObjectID id.ID, controller game.PlayerID, modifier game.CostModifier) int {
	if modifier.CountSelection == nil {
		return 0
	}
	owner := controller
	if source, ok := permanentByObjectID(g, sourceObjectID); ok {
		owner = effectiveController(g, source)
	}
	count := countPermanentsMatchingGroup(g, nil, owner, game.BattlefieldGroup(*modifier.CountSelection))
	return count * modifier.PerObjectReduction
}

// sharedExiledCardTypeReduction resolves the concrete generic reduction for a
// SharedExiledCardTypeReduction spell modifier ("Spells you cast cost {N} less
// to cast for each card type they share with cards exiled with this creature.",
// Cemetery Prowler). It reads the source permanent's linked-exile set named by
// the modifier's ExiledLinkKey, gathers the distinct card types among the cards
// still exiled with the source, counts how many of those types the casting card
// also has, and multiplies that shared count by the per-type amount.
func sharedExiledCardTypeReduction(g *game.Game, sourceObjectID id.ID, modifier game.CostModifier, card *game.CardDef) int {
	if card == nil {
		return 0
	}
	source, ok := permanentByObjectID(g, sourceObjectID)
	if !ok {
		return 0
	}
	key := game.LinkedObjectKey{SourceID: source.CardInstanceID, LinkID: string(modifier.ExiledLinkKey)}
	shared := 0
	for cardType := range exiledLinkedCardTypes(g, key) {
		if card.HasType(cardType) {
			shared++
		}
	}
	return shared * modifier.SharedExiledCardTypeReduction
}

// exiledLinkedCardTypes returns the set of distinct card types among the cards
// still exiled under key. A linked card whose instance is gone or that has left
// its owner's exile zone is skipped, so the count tracks only cards currently
// exiled with the source. The links are recorded by card identity (a graveyard
// exile publishes a card, not a permanent), so membership is checked against the
// owner's exile zone rather than last-known object information.
func exiledLinkedCardTypes(g *game.Game, key game.LinkedObjectKey) map[types.Card]bool {
	distinct := make(map[types.Card]bool)
	for _, ref := range linkedObjects(g, key) {
		card, ok := g.GetCardInstance(ref.CardID)
		if !ok {
			continue
		}
		owner, ok := playerByID(g, card.Owner)
		if !ok || !owner.Exile.Contains(card.ID) {
			continue
		}
		for _, cardType := range graveyardCardTypes(card) {
			distinct[cardType] = true
		}
	}
	return distinct
}

// spellCostModifierMatchesTargets reports whether a spell cost modifier's
// optional targets-source filter admits the spell being cast. A modifier
// without the filter applies regardless of the spell's targets; one with the
// filter applies only when one of the spell's chosen targets is exactly the
// permanent (sourceObjectID) whose static ability carries the modifier.
func spellCostModifierMatchesTargets(modifier game.CostModifier, sourceObjectID id.ID, targets []game.Target) bool {
	if !modifier.TargetsSource {
		return true
	}
	if sourceObjectID == 0 {
		return false
	}
	for _, target := range targets {
		if target.Kind == game.TargetPermanent && target.PermanentID == sourceObjectID {
			return true
		}
	}
	return false
}

// spellCostModifierMatchesZone reports whether a spell cost modifier's optional
// source-zone filter admits a spell being cast from sourceZone. A modifier with
// no zone filter applies regardless of the casting zone. A SourceZones set
// (generalizing "Spells you cast from anywhere other than your hand ...") admits
// only spells cast from one of its listed zones. Otherwise a single-zone
// SourceZone applies only when the spell is cast from exactly that zone.
func spellCostModifierMatchesZone(modifier game.CostModifier, sourceZone zone.Type) bool {
	if len(modifier.SourceZones) > 0 {
		return slices.Contains(modifier.SourceZones, sourceZone)
	}
	return !modifier.SourceZone.Exists || modifier.SourceZone.Val == sourceZone
}

// spellCostModifierMatchesCard reports whether a spell cost modifier's card-type
// and color filters admit the given spell card. A nil card fails any active
// filter. The colorless sentinel (MatchColor with an empty Color) matches spells
// that have no colors; otherwise the spell must carry the named color.
func spellCostModifierEffectMatchesCard(g *game.Game, effect *game.RuleEffect, card *game.CardDef) bool {
	modifier := effect.CostModifier
	if !spellCostModifierBaseMatchesCard(g, modifier, card) {
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

func spellCostModifierMatchesCard(g *game.Game, modifier game.CostModifier, card *game.CardDef) bool {
	return !modifier.ChosenSubtypeFromEntryChoice && spellCostModifierBaseMatchesCard(g, modifier, card)
}

// spellCostModifierBaseMatchesCard reports whether a spell cost modifier's
// card-subject filter admits the given spell card. It converts the modifier's
// card filters into the canonical card-subject Selection and matches through the
// shared matchSelection so cost modifiers describe their spells the same way
// triggers and additional costs do. A modifier with no card filter matches any
// card (including a nil card); any active filter fails a nil card.
func spellCostModifierBaseMatchesCard(g *game.Game, modifier game.CostModifier, card *game.CardDef) bool {
	sel := modifier.CardSelection
	if sel.Empty() {
		return true
	}
	return cardDefMatchesCostSelection(g, card, sel)
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
	if face == game.FaceFront && cardOK && cardHasJumpStartAlternative(card) {
		permissions = append(permissions, payment.SpellCastPermissionFlashback)
	}
	if face == game.FaceFront && cardOK && cardHasEscapeAlternative(card) {
		permissions = append(permissions, payment.SpellCastPermissionEscape)
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
		if !ruleEffectAffectsPlayer(g, effect, playerID) {
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

// castFromZoneAllowsAnyMana reports whether an active play/cast-from-zone
// permission lets playerID spend mana of any color to cast cardID from
// sourceZone ("mana of any type can be spent to cast it.", Court of Locthwain;
// "you may spend mana as though it were mana of any color to cast it.", Evelyn,
// the Covetous; "you may spend mana as though it were mana of any color to cast
// that spell.", Grenzo, Havoc Raiser). A per-card RuleEffectPlayFromZone or
// RuleEffectCastFromZone permission carries the flag on the self-scoped card
// grant; a RuleEffectCastSpellsFromZone permission carries it on the group cast
// grant, in which case the full CastSpellsFromZone match (including its
// exile-counter, provenance, and once-per-turn filters) must also hold.
func castFromZoneAllowsAnyMana(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex) bool {
	if effect, ok := matchingCastSpellsFromZoneEffect(g, playerID, cardID, sourceZone, face); ok && effect.SpendAnyMana {
		return true
	}
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if (effect.Kind != game.RuleEffectPlayFromZone && effect.Kind != game.RuleEffectCastFromZone) ||
			!effect.SpendAnyMana ||
			effect.CastFromZone != sourceZone {
			continue
		}
		if !playerRelationMatches(effect.Controller, playerID, effect.AffectedPlayer) {
			continue
		}
		if effect.AffectedCardID != 0 && effect.AffectedCardID != cardID {
			continue
		}
		if effect.CastFace.Exists && effect.CastFace.Val != face {
			continue
		}
		return true
	}
	return false
}

// castFromZoneWithoutPayingManaCost reports whether an active per-card
// RuleEffectPlayFromZone permission lets playerID cast cardID from sourceZone
// without paying its mana cost ("You may play it this turn without paying its
// mana cost.", Dauthi Voidwalker). The flag rides the self-scoped card grant that
// handlePlayChosenExiledCard emits; a played land has no mana cost, so this only
// affects the spell cast. It is false for every paying play-from-exile grant
// (ImpulseExile, Prowl), leaving their casts to pay normally.
func castFromZoneWithoutPayingManaCost(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex) bool {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind != game.RuleEffectPlayFromZone || !effect.WithoutPayingManaCost ||
			effect.CastFromZone != sourceZone {
			continue
		}
		if !ruleEffectAffectsPlayer(g, effect, playerID) {
			continue
		}
		if effect.AffectedCardID != 0 && effect.AffectedCardID != cardID {
			continue
		}
		if effect.CastFace.Exists && effect.CastFace.Val != face {
			continue
		}
		return true
	}
	return false
}

// RuleEffectCastLinkedExileForFree permission, if any, that lets playerID cast
// cardID from exile without paying its mana cost ("cast a spell from among cards
// exiled with this enchantment without paying its mana cost.", Court of
// Locthwain). It matches the affected player and requires cardID to belong to
// the source-keyed linked-exile pool the effect names.
func castLinkedExileForFreePermission(g *game.Game, playerID game.PlayerID, cardID id.ID) (game.RuleEffect, bool) {
	if cardID == 0 {
		return game.RuleEffect{}, false
	}
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind != game.RuleEffectCastLinkedExileForFree || effect.ExiledLinkKey == "" {
			continue
		}
		if !playerRelationMatches(effect.Controller, playerID, effect.AffectedPlayer) {
			continue
		}
		key := game.LinkedObjectKey{SourceID: effect.SourceCardID, LinkID: string(effect.ExiledLinkKey)}
		if cardInLinkedObjectPool(g, key, cardID) {
			return *effect, true
		}
	}
	return game.RuleEffect{}, false
}

// castLinkedExileForFree reports whether an active RuleEffectCastLinkedExileForFree
// permission lets playerID cast cardID from exile without paying its mana cost.
func castLinkedExileForFree(g *game.Game, playerID game.PlayerID, cardID id.ID) bool {
	_, ok := castLinkedExileForFreePermission(g, playerID, cardID)
	return ok
}

// cardInLinkedObjectPool reports whether cardID was remembered in the linked
// object set keyed by key.
func cardInLinkedObjectPool(g *game.Game, key game.LinkedObjectKey, cardID id.ID) bool {
	for _, ref := range linkedObjects(g, key) {
		if ref.CardID == cardID {
			return true
		}
	}
	return false
}

// consumeCastLinkedExileForFreePermission removes the one-shot
// RuleEffectCastLinkedExileForFree permission identified by effectID after its
// player casts a spell under it, matching the singular "cast a spell". It is a
// no-op when effectID is zero or the effect is no longer present.
func consumeCastLinkedExileForFreePermission(g *game.Game, effectID id.ID) {
	if effectID == 0 || len(g.RuleEffects) == 0 {
		return
	}
	kept := g.RuleEffects[:0]
	for i := range g.RuleEffects {
		if g.RuleEffects[i].ID == effectID && g.RuleEffects[i].Kind == game.RuleEffectCastLinkedExileForFree {
			continue
		}
		kept = append(kept, g.RuleEffects[i])
	}
	g.RuleEffects = kept
}

// exileCounterPermissionBlocked reports whether the provenance or once-per-turn
// riders on an exile-counter play/cast permission (Evelyn, the Covetous) forbid
// using it for cardID right now. The provenance rider
// (ExileCounterExiledByController) requires cardID to have been exiled by an
// ability the permission's controller controlled; the OncePerTurn rider blocks
// the permission once its source permanent has already played or cast a card
// this turn. Both riders are false for every ordinary play/cast-from-zone
// permission, so this is a no-op for them.
func exileCounterPermissionBlocked(g *game.Game, effect *game.RuleEffect, cardID id.ID) bool {
	if effect.ExileCounterExiledByController && !g.ExileCounterExiledByController(cardID, effect.Controller) {
		return true
	}
	if effect.OncePerTurn && g.ExilePlayPermissionUsedThisTurn[effect.SourceObjectID] {
		return true
	}
	return false
}

// recordExilePlayPermissionUse marks a once-per-turn play/cast-from-exile
// permission as used this turn, keyed by its source permanent so a source that
// grants both a land-play and a spell-cast permission spends the single shared
// use (Evelyn, the Covetous). It is a no-op for a permission with no per-turn cap
// or an unset source, so callers can invoke it unconditionally after a play or
// cast authorized by an exile-counter permission.
func recordExilePlayPermissionUse(g *game.Game, effect game.RuleEffect) {
	if !effect.OncePerTurn || effect.SourceObjectID == 0 {
		return
	}
	if g.ExilePlayPermissionUsedThisTurn == nil {
		g.ExilePlayPermissionUsedThisTurn = make(map[game.ObjectID]bool)
	}
	g.ExilePlayPermissionUsedThisTurn[effect.SourceObjectID] = true
}

func canPlayLandFromZoneByRuleEffect(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type) bool {
	_, ok := matchingPlayLandFromZoneEffect(g, playerID, cardID, sourceZone)
	return ok
}

// matchingPlayLandFromZoneEffect returns the first continuous play-from-zone
// permission that lets playerID play cardID as a land from sourceZone, applying
// the same TopCardOnly, exile-counter, provenance, and once-per-turn filters as
// canPlayLandFromZoneByRuleEffect. Callers record a once-per-turn use against the
// returned effect after the land is actually played.
func matchingPlayLandFromZoneEffect(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type) (game.RuleEffect, bool) {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.CastFromZone != sourceZone ||
			!ruleEffectAffectsPlayer(g, effect, playerID) {
			continue
		}
		switch effect.Kind {
		case game.RuleEffectPlayFromZone:
			if effect.AffectedCardID == cardID {
				return *effect, true
			}
		case game.RuleEffectPlayLandsFromZone:
			if effect.TopCardOnly && !cardIsTopOfLibrary(g, playerID, cardID) {
				continue
			}
			if effect.ExileCounterFilter.Exists && !g.HasExileCounter(cardID, effect.ExileCounterFilter.Val) {
				continue
			}
			if exileCounterPermissionBlocked(g, effect, cardID) {
				continue
			}
			return *effect, true
		default:
			continue
		}
	}
	return game.RuleEffect{}, false
}

// canCastSpellsFromZoneByRuleEffect reports whether a continuous
// RuleEffectCastSpellsFromZone permission lets playerID cast the face of cardID
// from sourceZone ("You may cast spells from the top of your library.", Future
// Sight). A SpellTypes and/or SpellColorless filter requires the cast face to
// have one of the listed card types or be colorless ("You may cast artifact
// spells and colorless spells from the top of your library.", Mystic Forge);
// TopCardOnly requires the card to be on top of the player's library.
func canCastSpellsFromZoneByRuleEffect(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex) bool {
	_, ok := matchingCastSpellsFromZoneEffect(g, playerID, cardID, sourceZone, face)
	return ok
}

// matchingCastSpellsFromZoneEffect returns the first continuous
// RuleEffectCastSpellsFromZone permission that lets playerID cast the face of
// cardID from sourceZone, applying the same TopCardOnly, chosen-subtype, and
// type/colorless filters as canCastSpellsFromZoneByRuleEffect.
func matchingCastSpellsFromZoneEffect(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex) (game.RuleEffect, bool) {
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return game.RuleEffect{}, false
	}
	faceDef := cardFaceOrDefault(card, face)
	faceTypes := faceDef.Types
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
		if effect.ExileCounterFilter.Exists && !g.HasExileCounter(cardID, effect.ExileCounterFilter.Val) {
			continue
		}
		if exileCounterPermissionBlocked(g, effect, cardID) {
			continue
		}
		if effect.SpellChosenSubtypeFrom != "" && !cardMatchesSourceEntryChosenSubtype(g, effect, faceDef) {
			continue
		}
		if len(effect.SpellTypes) > 0 || effect.SpellColorless {
			typeMatch := len(effect.SpellTypes) > 0 && slices.ContainsFunc(effect.SpellTypes, func(t types.Card) bool {
				return slices.Contains(faceTypes, t)
			})
			colorMatch := effect.SpellColorless && len(faceDef.Colors) == 0
			if !typeMatch && !colorMatch {
				continue
			}
		}
		return *effect, true
	}
	return game.RuleEffect{}, false
}

// castFromZoneRequiresPayLife reports whether the permission authorizing
// playerID to cast the face of cardID from sourceZone replaces the spell's mana
// cost with paying life equal to its mana value ("If you cast a spell this way,
// pay life equal to its mana value rather than pay its mana cost.", Bolas's
// Citadel, Gwenom, Remorseless).
func castFromZoneRequiresPayLife(g *game.Game, playerID game.PlayerID, cardID id.ID, sourceZone zone.Type, face game.FaceIndex) bool {
	effect, ok := matchingCastSpellsFromZoneEffect(g, playerID, cardID, sourceZone, face)
	return ok && effect.PayLifeEqualToManaValue
}

// cardMatchesSourceEntryChosenSubtype reports whether faceDef shares the creature
// subtype the source permanent of effect chose as it entered ("creature spells of
// the chosen type", Realmwalker), reading the choice stored under
// effect.SpellChosenSubtypeFrom in the source's EntryChoices.
func cardMatchesSourceEntryChosenSubtype(g *game.Game, effect *game.RuleEffect, faceDef *game.CardDef) bool {
	source, ok := permanentByObjectID(g, effect.SourceObjectID)
	if !ok {
		return false
	}
	choice, ok := source.EntryChoices[effect.SpellChosenSubtypeFrom]
	return ok &&
		choice.Kind == game.ResolutionChoiceSubtype &&
		types.KnownSubtypeForType(types.Creature, choice.Subtype) &&
		faceDef.HasSubtype(choice.Subtype)
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

// playerCanLookAtTopCardAnyTime reports whether playerID may privately look at
// the top card of their library at any time ("You may look at the top card of
// your library any time.", Bolas's Citadel). It is a private-visibility static,
// the look-at-your-own-card counterpart of playerPlaysWithTopCardRevealed.
func playerCanLookAtTopCardAnyTime(g *game.Game, playerID game.PlayerID) bool {
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind != game.RuleEffectLookAtTopCardAnyTime {
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
			if g.AdventureCards[cardID] || cardIsPlottedInExile(g, cardID) || cardIsForetoldInExile(g, cardID) || slices.ContainsFunc(card.Def.LegalCastFaces(), func(face game.FaceIndex) bool {
				return canCastFromZoneByRuleEffect(g, playerID, cardID, zone.Exile, face) ||
					canCastSpellsFromZoneByRuleEffect(g, playerID, cardID, zone.Exile, face)
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
	if !slices.Contains(zones, zone.Exile) && len(foreignExileCastableCards(g, playerID)) > 0 {
		zones = append(zones, zone.Exile)
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

func cardHasEscapeAlternative(card *game.CardInstance) bool {
	frontDef := cardFaceOrDefault(card, game.FaceFront)
	if !frontDef.HasKeyword(game.Escape) {
		return false
	}
	return slices.ContainsFunc(frontDef.AlternativeCosts, isEscapeAlternative)
}

// cardHasJumpStartAlternative reports whether the card's front face has the
// Jump-start keyword (CR 702.134), which grants a graveyard cast paying the
// card's other costs plus discarding a card, then exiling the card on
// resolution. The graveyard cast reuses the Flashback permission and
// exile-on-resolution; the discard additional cost is synthesized by the
// payment planner.
func cardHasJumpStartAlternative(card *game.CardInstance) bool {
	return cardFaceOrDefault(card, game.FaceFront).HasKeyword(game.JumpStart)
}
