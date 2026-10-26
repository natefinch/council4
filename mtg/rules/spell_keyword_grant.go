package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// spellGrantedKeyword reports whether an active RuleEffectGrantSpellKeyword
// confers keyword on the spell playerID is casting. It backs the payment
// planner's cost-affecting keyword check, so a spell a static or one-shot grant
// touches ("Nonartifact spells you cast have improvise.", Inspiring Statuary;
// "The next spell you cast this turn has improvise.", Archway of Innovation) can
// pay with the granted keyword's machinery in addition to any keyword the card
// carries natively.
//
// A grant matches when its granted keyword equals keyword and
// spellKeywordGrantMatches accepts the spell. A one-shot next-spell grant stays
// active here; it is consumed separately, and only when a matching spell is
// actually cast.
func spellGrantedKeyword(g *game.Game, playerID game.PlayerID, card *game.CardDef, cardID id.ID, fromZone zone.Type, keyword game.Keyword) bool {
	if card == nil || keyword == game.KeywordNone {
		return false
	}
	for _, effect := range activeRuleEffects(g) {
		if effect.Kind == game.RuleEffectGrantSpellKeyword &&
			effect.GrantedKeyword == keyword &&
			spellKeywordGrantMatches(g, &effect, playerID, card, cardID, fromZone) {
			return true
		}
	}
	return false
}

// spellKeywordGrantMatches reports whether a RuleEffectGrantSpellKeyword applies
// to a spell caster casts from fromZone whose characteristics are card and whose
// card instance is cardID (zero for a cast copy, which is never a commander).
// The controller relation must resolve to the caster, a set CastFromZone must
// equal fromZone, CardSelection.MatchCommander requires the card to carry the
// commander designation (CR 903.3, an attribute of the card rather than a
// characteristic), and the rest of the selection reads the spell's
// characteristics as spellGrantSelectionFace computes them.
func spellKeywordGrantMatches(g *game.Game, effect *game.RuleEffect, caster game.PlayerID, card *game.CardDef, cardID id.ID, fromZone zone.Type) bool {
	if card == nil || !controllerRelationMatches(effect.Controller, caster, effect.AffectedController) {
		return false
	}
	if effect.CastFromZone != zone.None && effect.CastFromZone != fromZone {
		return false
	}
	selection := effect.CardSelection
	if selection.MatchCommander {
		if cardID == 0 || !g.CommanderIDs[cardID] {
			return false
		}
		selection.MatchCommander = false
	}
	return selection.Empty() || cardDefMatchesCostSelection(g, spellGrantSelectionFace(card), selection)
}

// spellGrantSelectionFace returns the characteristics a spell-keyword grant's
// selection reads for a spell whose face is card: its printed characteristics as
// its own characteristic-defining abilities modify them. CDAs function in every
// zone (CR 604.3), and a CDA defines only colors, subtypes or power/toughness
// (CR 604.3a), so this applies:
//   - Changeling, which makes the spell every creature type (CR 702.73a) while
//     keeping its other subtypes;
//   - Devoid, which makes the spell colorless (CR 702.114a);
//   - every unconditional static ability whose continuous effects change only
//     the card's own subtypes or colors ("Mistform Ultimus is every creature
//     type (even if this card isn't on the battlefield).", "Transguild Courier
//     is all colors."), as characteristicDefiningStatic recognizes them.
//
// A self type effect keyed to a battlefield entry choice (Adaptive Automaton's
// chosen type) is not a CDA and has no value off the battlefield. card itself is
// never mutated; a face with no CDA is returned unchanged.
func spellGrantSelectionFace(card *game.CardDef) *game.CardDef {
	if card == nil {
		return nil
	}
	subtypes := slices.Clone(card.Subtypes)
	colors := slices.Clone(card.Colors)
	changed := false
	if card.HasKeyword(game.Changeling) {
		subtypes = appendUniqueSubtypes(slices.Clone(types.SubtypesForType(types.Creature)), subtypes...)
		changed = true
	}
	for i := range card.StaticAbilities {
		static := &card.StaticAbilities[i]
		if !characteristicDefiningStatic(static) {
			continue
		}
		changed = true
		for j := range static.ContinuousEffects {
			effect := &static.ContinuousEffects[j]
			switch effect.Layer {
			case game.LayerType:
				if effect.SetSubtypes != nil {
					subtypes = slices.Clone(effect.SetSubtypes)
				}
				subtypes = removeSubtypes(subtypes, effect.RemoveSubtypes)
				subtypes = appendUniqueSubtypes(subtypes, effect.AddSubtypes...)
				if effect.AddEveryCreatureType {
					subtypes = appendUniqueSubtypes(subtypes, types.SubtypesForType(types.Creature)...)
				}
			case game.LayerColor:
				if effect.SetColorless {
					colors = nil
				} else if effect.SetColors != nil {
					colors = slices.Clone(effect.SetColors)
				}
				colors = removeColors(colors, effect.RemoveColors)
				colors = appendUniqueColors(colors, effect.AddColors...)
			}
		}
	}
	if card.HasKeyword(game.Devoid) && len(colors) > 0 {
		colors = nil
		changed = true
	}
	if !changed {
		return card
	}
	expanded := *card
	expanded.Subtypes = subtypes
	expanded.Colors = colors
	return &expanded
}

// characteristicDefiningStatic reports whether static is a characteristic-
// defining ability of the subtype or color kind (CR 604.3a): it is
// unconditional, and each of its continuous effects affects only its own object
// and changes only subtypes (set, remove, add or every creature type) at the
// type layer or colors at the color layer. Abilities that also grant rules
// effects, keywords or other abilities, change card types or supertypes, depend
// on a battlefield entry choice or work only from a particular zone are not.
func characteristicDefiningStatic(static *game.StaticAbility) bool {
	if static == nil || static.Condition.Exists || static.ZoneOfFunction != zone.None ||
		len(static.ContinuousEffects) == 0 || len(static.RuleEffects) != 0 || len(static.KeywordAbilities) != 0 {
		return false
	}
	for i := range static.ContinuousEffects {
		effect := &static.ContinuousEffects[i]
		if !effect.AffectedSource || !effect.Group.Empty() || effect.CardZone != zone.None {
			return false
		}
		switch effect.Layer {
		case game.LayerType:
			if effect.SetSupertypes != nil || len(effect.AddSupertypes) != 0 || len(effect.RemoveSupertypes) != 0 ||
				effect.SetTypes != nil || len(effect.AddTypes) != 0 || len(effect.RemoveTypes) != 0 ||
				len(effect.SetBasicLandTypes) != 0 || effect.AddEveryBasicLandType ||
				effect.AddSubtypeFromEntryChoice != "" || effect.SetSubtypeFromSourceChoice != "" {
				return false
			}
			if effect.SetSubtypes == nil && len(effect.RemoveSubtypes) == 0 && len(effect.AddSubtypes) == 0 && !effect.AddEveryCreatureType {
				return false
			}
		case game.LayerColor:
			if !effect.SetColorless && effect.SetColors == nil && len(effect.RemoveColors) == 0 && len(effect.AddColors) == 0 {
				return false
			}
		default:
			return false
		}
	}
	return true
}

// consumeNextSpellKeywordGrantEffects applies and consumes any one-shot "The
// next spell you cast this turn has <keyword>." rule effects (Archway of
// Innovation) that match the spell just cast from fromZone. Each matching effect
// is removed from the game's rule effects so later spells are unaffected; the
// grant already fed the payment planner while the spell's cost was being paid,
// so nothing needs to persist onto the spell. It is a no-op when no such
// one-shot grant is active.
//
// Only a matching spell that is actually cast reaches this point: it runs from
// emitSpellCastEvents after the spell is on the stack and its costs are paid, so
// activating abilities, land plays, failed cast attempts, and nonmatching spells
// never consume the grant.
func consumeNextSpellKeywordGrantEffects(g *game.Game, obj *game.StackObject, fromZone zone.Type) {
	if len(g.RuleEffects) == 0 {
		return
	}
	spellDef, ok := spellDefForStackObject(g, obj)
	if !ok {
		return
	}
	cardID := obj.SourceID
	if obj.SourceTokenDef != nil {
		cardID = 0
	}
	kept := g.RuleEffects[:0]
	for i := range g.RuleEffects {
		effect := &g.RuleEffects[i]
		if effect.Kind != game.RuleEffectGrantSpellKeyword ||
			!effect.AppliesToNextSpellOnly ||
			!spellKeywordGrantMatches(g, effect, obj.Controller, spellDef, cardID, fromZone) {
			kept = append(kept, *effect)
			continue
		}
	}
	g.RuleEffects = kept
}
