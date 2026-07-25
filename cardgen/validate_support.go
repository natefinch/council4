package cardgen

import (
	"errors"

	"github.com/natefinch/council4/mtg/game"
)

// This file holds the support gates that the generated struct renderers run
// before building a literal.
//
// Rendering and support-gating used to be fused: a hand-written renderer
// decided both "can the executable backend run this?" and "what Go source
// spells it?" in one pass. Generating the renderer separates them, and the
// gates have to move here rather than stay at a call site, because the
// generated renderer reaches a nested value directly instead of going through
// whatever hand-written wrapper used to guard it.
//
// Only runtime-capability gates live here — rules the engine genuinely cannot
// execute. Gates that merely recorded "no one wrote a render arm for this"
// are intentionally gone; removing them is why the renderer is generated.
//
// The generated renderer calls these; the wiring is preRenderValidators in
// cardgen/cmd/genrender/validators.go.

// validateTriggerPattern rejects trigger patterns whose field combinations the
// executable backend cannot match.
func validateTriggerPattern(p game.TriggerPattern) error {
	// renderEventKind is a support gate rather than a spelling: it refuses the
	// unrenderedEventKinds, which runtime machinery produces and no card's text
	// lowers to, and it refuses the unset zero, which a pattern must never
	// carry. The generated renderer spells every declared constant, so this has
	// to stay explicit. UnionEvent is optional, so only its kind is checked.
	if _, err := renderEventKind(p.Event); err != nil {
		return err
	}
	if p.UnionEvent != game.EventUnknown {
		if _, err := renderEventKind(p.UnionEvent); err != nil {
			return err
		}
	}
	if (p.Event == game.EventBeginningOfStep) != (p.Step != game.StepNone) {
		return errors.New("render: beginning-of-step trigger pattern must set exactly one supported step")
	}
	allowZoneChangeZones := p.Event == game.EventZoneChanged
	allowFromZone := p.MatchFromZone &&
		(p.Event == game.EventSpellCast || p.Event == game.EventPermanentEnteredBattlefield || allowZoneChangeZones) &&
		!p.MatchToZone
	allowExcludeFromZone := p.ExcludeFromZone &&
		(p.Event == game.EventSpellCast || allowZoneChangeZones)
	if len(p.RequireCardTypes) != 0 ||
		len(p.ExcludeCardTypes) != 0 ||
		(p.MatchFromZone && !allowFromZone && !allowZoneChangeZones) ||
		(p.MatchToZone && !allowZoneChangeZones) ||
		(p.ExcludeToZone && !allowZoneChangeZones) ||
		(p.MatchToZone && p.ExcludeToZone) ||
		(p.ExcludeFromZone && !allowExcludeFromZone) ||
		(p.MatchFromZone && p.ExcludeFromZone) ||
		(len(p.FromZones) > 0 &&
			(!allowZoneChangeZones || p.MatchFromZone || p.ExcludeFromZone || len(p.FromZones) < 2)) ||
		p.DamageRecipientCombatState != game.CombatStateAny ||
		(p.SpellTargetsSource && p.Event != game.EventSpellCast) ||
		((p.SpellTargetAllow != game.TargetAllowUnspecified || p.SpellTargetPattern.Exists) && p.Event != game.EventSpellCast) ||
		(p.RequireKickerPaid && p.Event != game.EventSpellCast) ||
		(p.RequireHistoric && p.Event != game.EventSpellCast) ||
		(p.MatchSpellCopy && p.Event != game.EventSpellCast) ||
		(p.SelfWasCast && p.Event != game.EventSpellCast) ||
		(p.RequireTappedForMana && p.Event != game.EventPermanentTapped && p.Event != game.EventManaProduced) ||
		(p.RequireProducedManaColorFromEntryChoice && p.Event != game.EventPermanentTapped && p.Event != game.EventManaProduced) ||
		(p.RequireManaProducedByLand && p.Event != game.EventManaProduced) ||
		(p.ExcludeManaAbility && p.Event != game.EventAbilityActivated) ||
		(p.Event == game.EventAbilityActivated && !p.ExcludeManaAbility) ||
		(p.PlayerEventOrdinalThisTurn > 0 &&
			p.Event != game.EventCardDrawn &&
			p.Event != game.EventCardDiscarded &&
			p.Event != game.EventCycled &&
			p.Event != game.EventLifeGained &&
			p.Event != game.EventLifeLost &&
			p.Event != game.EventScry &&
			p.Event != game.EventSurveil &&
			p.Event != game.EventSpellCast) ||
		(p.RequireCombatDamage && p.RequireNonCombatDamage) ||
		(p.AttackAlone && p.Event != game.EventAttackerDeclared) ||
		(p.AttackerCaptured && p.Event != game.EventAttackerDeclared) ||
		(p.DyingObjectCaptured && p.Event != game.EventPermanentDied) ||
		(p.AttackWhileSaddled && p.Event != game.EventAttackerDeclared) ||
		(p.AttacksDifferentPlayerThanAnother && p.Event != game.EventAttackerDeclared) ||
		(p.AttackedPlayerIsSourceEnchantedPlayer && p.Event != game.EventAttackerDeclared) ||
		(p.StepPlayerIsSourceEnchantedPlayer && p.Event != game.EventBeginningOfStep) ||
		(p.FirstUpkeepStepEachTurn && (p.Event != game.EventBeginningOfStep || p.Step != game.StepUpkeep)) ||
		(p.AttacksAlongsideCount != 0 &&
			(p.Event != game.EventAttackerDeclared || p.AttacksAlongsideSelection.Empty())) ||
		(p.DyingDamagedBySource && p.Event != game.EventPermanentDied) ||
		(p.ExcludeFirstDrawInDrawStep && p.Event != game.EventCardDrawn) ||
		(p.ClassBecameLevel > 0 && p.Event != game.EventClassLevelGained) ||
		(p.PlaysLinkedExileCard != "" && p.Event != game.EventCardPlayedFromExile) ||
		(p.AttackerCountAtLeast != 0 &&
			(p.Event != game.EventAttackerDeclared || p.AttackAlone || p.AttackerCountAtLeast < 2 ||
				(!p.OneOrMore && p.Source != game.TriggerSourceSelf))) {
		return errors.New("render: unsupported trigger pattern fields")
	}
	if err := validateTriggerPatternCardSelection(&p); err != nil {
		return err
	}
	return validateTriggerPatternZones(p)
}

// validateTriggerPatternZones rejects a pattern that turns on a zone match
// without naming the zone it matches. renderZone refuses zone.None, so the
// hand-written renderer enforced this while spelling the field; the generated
// renderer spells zone.None happily, which would produce a pattern the engine
// can never satisfy.
func validateTriggerPatternZones(p game.TriggerPattern) error {
	if p.MatchFromZone || p.ExcludeFromZone {
		if _, err := renderZone(p.FromZone); err != nil {
			return err
		}
	}
	if p.MatchToZone || p.ExcludeToZone {
		if _, err := renderZone(p.ToZone); err != nil {
			return err
		}
	}
	for _, from := range p.FromZones {
		if _, err := renderZone(from); err != nil {
			return err
		}
	}
	return nil
}

// validateTriggerCondition rejects trigger types the executable backend cannot
// run. renderTriggerType is the gate: it names the supported constants, and the
// generated renderer spells every declared one.
func validateTriggerCondition(c game.TriggerCondition) error {
	_, err := renderTriggerType(c.Type)
	return err
}

// validateContinuousEffect rejects continuous effects whose recipient and layer
// fields the executable backend cannot apply together.
func validateContinuousEffect(e game.ContinuousEffect) error {
	if e.AffectedSource && !e.Group.Empty() {
		return errors.New("render: continuous effect cannot set both AffectedSource and Group")
	}
	if err := validateContinuousEffectLayerFields(&e); err != nil {
		return err
	}
	_, err := renderContinuousLayer(e.Layer)
	return err
}

// validateRuleEffect rejects rule-effect kinds the executable backend cannot run.
func validateRuleEffect(e game.RuleEffect) error {
	_, err := renderRuleEffectKind(e.Kind)
	return err
}

// validateCostModifier rejects cost-modifier kinds the executable backend
// cannot apply.
func validateCostModifier(m game.CostModifier) error {
	_, err := renderCostModifierKind(m.Kind)
	return err
}
