package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func cardFaceDef(card *game.CardInstance, face game.FaceIndex) (*game.CardDef, bool) {
	if card == nil || card.Def == nil {
		return nil, false
	}
	return card.Def.FaceDefView(face)
}

func cardFaceOrDefault(card *game.CardInstance, face game.FaceIndex) *game.CardDef {
	def, ok := cardFaceDef(card, face)
	if ok {
		return def
	}
	return card.Def
}

func permanentFaceDef(g *game.Game, permanent *game.Permanent) (*game.CardDef, bool) {
	if permanent.Token {
		if permanent.TokenDef == nil {
			return nil, false
		}
		return permanent.TokenDef.FaceDefView(permanent.Face)
	}
	card, ok := g.GetCardInstance(permanent.CardInstanceID)
	if !ok {
		return nil, false
	}
	return cardFaceDef(card, permanent.Face)
}

func visibleFace(permanent *game.Permanent) game.FaceIndex {
	if permanent == nil {
		return game.FaceFront
	}
	return permanent.Face
}

func stackObjectFace(obj *game.StackObject) game.FaceIndex {
	if obj == nil {
		return game.FaceFront
	}
	return obj.Face
}

func transformPermanent(g *game.Game, permanent *game.Permanent) bool {
	def, ok := physicalPermanentDef(g, permanent)
	if !ok {
		return false
	}
	if !def.IsTransformingDoubleFaced() || !def.Back.Exists {
		return false
	}
	if ruleEffectPreventsTransform(g, permanent) {
		return false
	}
	if permanent.Face == game.FaceFront {
		permanent.Face = game.FaceBack
		permanent.Transformed = true
		return true
	}
	permanent.Face = game.FaceFront
	permanent.Transformed = false
	return true
}

func physicalPermanentDef(g *game.Game, permanent *game.Permanent) (*game.CardDef, bool) {
	if permanent.Token {
		return permanent.TokenDef, permanent.TokenDef != nil
	}
	card, ok := g.GetCardInstance(permanent.CardInstanceID)
	if !ok {
		return nil, false
	}
	return card.Def, true
}

func cardTypes(def *game.CardDef) []types.Card {
	return append([]types.Card(nil), def.Types...)
}

func cardSupertypes(def *game.CardDef) []types.Super {
	return append([]types.Super(nil), def.Supertypes...)
}

func cardSubtypes(def *game.CardDef) []types.Sub {
	return append([]types.Sub(nil), def.Subtypes...)
}

// stackObjectCardTypes returns the effective card types of a spell on the stack,
// applying the Bestow cast transform (CR 702.103b) when the spell was cast for
// its bestow cost: a bestowed spell is an Aura spell and not a creature spell.
// Callers pass the already-resolved face def to avoid re-resolving it. Abilities
// and non-bestowed spells keep their printed types.
func stackObjectCardTypes(obj *game.StackObject, def *game.CardDef) []types.Card {
	printed := cardTypes(def)
	if obj != nil && obj.Kind == game.StackSpell && obj.Bestowed {
		return game.BestowSpellTypes(printed)
	}
	return printed
}

// stackObjectCardSubtypes returns the effective subtypes of a spell on the
// stack: its printed subtypes as its own characteristic-defining abilities
// define them (CR 604.3; a Changeling spell is every creature type), plus the
// Aura subtype for a bestowed spell (CR 702.103b).
func stackObjectCardSubtypes(obj *game.StackObject, def *game.CardDef) []types.Sub {
	subtypes := cardSubtypes(characteristicDefinedFace(def))
	if obj != nil && obj.Kind == game.StackSpell && obj.Bestowed {
		return game.BestowSpellSubtypes(subtypes)
	}
	return subtypes
}

// castSelectionFace returns the CardDef characteristics a spell presents to
// cost-modifier, cast-permission and cast-prohibition CardSelection matching
// while it is being cast (CR 601.2f). It first applies the card's own
// characteristic-defining abilities, which function in every zone (CR 604.3), so
// a Changeling spell is every creature type and a Devoid spell is colorless. For
// a bestowed cast (CR 702.103b) it then drops Creature from the types and adds
// the Aura subtype, so "creature spells cost less" modifiers no longer match and
// "Aura spells cost less" modifiers do. Power and every other characteristic are
// preserved, and the original CardDef is never mutated. A card with neither a
// CDA nor a bestowed cast is returned unchanged.
func castSelectionFace(card *game.CardDef, bestowed bool) *game.CardDef {
	card = characteristicDefinedFace(card)
	if card == nil || !bestowed {
		return card
	}
	transformed := *card
	transformed.Types = game.BestowSpellTypes(card.Types)
	transformed.Subtypes = game.BestowSpellSubtypes(card.Subtypes)
	return &transformed
}

// spellColors returns the colors of a spell's effective face for use in
// EventSpellCast, paralleling cardTypes for type-based filters. The face's own
// color characteristic-defining abilities apply (CR 604.3): Devoid makes it
// colorless (CR 702.114a) and "is all colors" sets its colors.
func spellColors(def *game.CardDef) []color.Color {
	def = characteristicDefinedFace(def)
	if def == nil || len(def.Colors) == 0 {
		return nil
	}
	return append([]color.Color(nil), def.Colors...)
}

// characteristicDefinedFace returns card's characteristics as its own
// characteristic-defining abilities define them. CDAs function in every zone
// (CR 604.3), and a CDA defines only colors, subtypes or power/toughness
// (CR 604.3a), so this applies:
//   - Changeling, which makes the card every creature type (CR 702.73a) while
//     keeping its other subtypes;
//   - Devoid, which makes the card colorless (CR 702.114a);
//   - every unconditional static ability whose continuous effects change only
//     the card's own subtypes or colors ("Mistform Ultimus is every creature
//     type (even if this card isn't on the battlefield).", "Transguild Courier
//     is all colors."), as characteristicDefiningStatic recognizes them.
//
// Spell routes (cast events, cost modifiers, cast permissions and spell-keyword
// grants) read a spell's subtypes and colors through it. A self type effect
// keyed to a battlefield entry choice (Adaptive Automaton's chosen type) is not
// a CDA and has no value off the battlefield. card itself is never mutated; a
// face with no such CDA is returned unchanged.
func characteristicDefinedFace(card *game.CardDef) *game.CardDef {
	if card == nil {
		return nil
	}
	changeling := card.HasKeyword(game.Changeling)
	devoid := card.HasKeyword(game.Devoid)
	hasStatic := false
	for i := range card.StaticAbilities {
		if characteristicDefiningStatic(&card.StaticAbilities[i]) {
			hasStatic = true
			break
		}
	}
	if !changeling && !hasStatic && (!devoid || len(card.Colors) == 0) {
		return card
	}
	subtypes := slices.Clone(card.Subtypes)
	colors := slices.Clone(card.Colors)
	if changeling {
		subtypes = appendUniqueSubtypes(slices.Clone(types.SubtypesForType(types.Creature)), subtypes...)
	}
	for i := range card.StaticAbilities {
		static := &card.StaticAbilities[i]
		if !characteristicDefiningStatic(static) {
			continue
		}
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
	if devoid {
		colors = nil
	}
	expanded := *card
	expanded.Subtypes = subtypes
	expanded.Colors = colors
	return &expanded
}

// stackObjectColors returns the effective-face colors of a stack object. A spell
// keeps its card instance in SourceID, while an activated or triggered ability
// records its physical source in SourceCardID/SourceTokenDef; both are resolved
// so a resolving color condition ("if it's blue") can test the targeted object.
func stackObjectColors(g *game.Game, obj *game.StackObject) ([]color.Color, bool) {
	if obj != nil && obj.Kind == game.StackSpell {
		if def, ok := stackObjectSpellDef(g, obj); ok {
			return spellColors(def), true
		}
	}
	if def, ok := stackObjectSourceDef(g, obj); ok {
		return spellColors(def), true
	}
	if obj.SourceID != 0 {
		if card, ok := g.GetCardInstance(obj.SourceID); ok && card.Def != nil {
			if def, ok := card.Def.FaceDef(obj.Face); ok {
				return spellColors(def), true
			}
		}
	}
	return nil, false
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

func permanentCardID(permanent *game.Permanent) id.ID {
	return permanent.CardInstanceID
}
