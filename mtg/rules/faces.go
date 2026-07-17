package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

func cardFaceDef(card *game.CardInstance, face game.FaceIndex) (*game.CardDef, bool) {
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

// stackObjectCardSubtypes mirrors stackObjectCardTypes for subtypes, adding the
// Aura subtype to a bestowed spell (CR 702.103b).
func stackObjectCardSubtypes(obj *game.StackObject, def *game.CardDef) []types.Sub {
	printed := cardSubtypes(def)
	if obj != nil && obj.Kind == game.StackSpell && obj.Bestowed {
		return game.BestowSpellSubtypes(printed)
	}
	return printed
}

// castSelectionFace returns the CardDef characteristics a spell presents to
// cost-modifier CardSelection matching during cost determination (CR 601.2f).
// For a bestowed cast (CR 702.103b) it returns a shallow copy whose front-face
// types drop Creature and subtypes add Aura, so "creature spells cost less"
// modifiers no longer match and "Aura spells cost less" modifiers do. Power and
// every other characteristic — which the rules read as printed — are preserved by
// the copy, and the original CardDef is never mutated. A non-bestowed cast
// returns the card unchanged so native/printed paths are unaffected; future cast
// transformations can compose here.
func castSelectionFace(card *game.CardDef, bestowed bool) *game.CardDef {
	if card == nil || !bestowed {
		return card
	}
	transformed := *card
	transformed.Types = game.BestowSpellTypes(card.Types)
	transformed.Subtypes = game.BestowSpellSubtypes(card.Subtypes)
	return &transformed
}

// spellColors returns the colors of a spell's effective face for use in
// EventSpellCast, paralleling cardTypes for type-based filters.
func spellColors(def *game.CardDef) []color.Color {
	if def == nil || def.HasKeyword(game.Devoid) {
		return nil
	}
	return append([]color.Color(nil), def.Colors...)
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
		if card, ok := g.GetCardInstance(obj.SourceID); ok {
			if def, ok := card.Def.FaceDef(obj.Face); ok {
				return spellColors(def), true
			}
		}
	}
	return nil, false
}

func permanentCardID(permanent *game.Permanent) id.ID {
	return permanent.CardInstanceID
}
