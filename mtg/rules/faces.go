package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

func cardFaceDef(card *game.CardInstance, face game.FaceIndex) (*game.CardDef, bool) {
	if face == game.FaceFront {
		return card.Def, true
	}
	return card.Def.FaceDef(face)
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
		if permanent.Face == game.FaceFront {
			return permanent.TokenDef, true
		}
		return permanent.TokenDef.FaceDef(permanent.Face)
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

// spellColors returns the colors of a spell's effective face for use in
// EventSpellCast, paralleling cardTypes for type-based filters.
func spellColors(def *game.CardDef) []color.Color {
	return append([]color.Color(nil), def.Colors...)
}

// stackObjectColors returns the effective-face colors of a stack object. A spell
// keeps its card instance in SourceID, while an activated or triggered ability
// records its physical source in SourceCardID/SourceTokenDef; both are resolved
// so a resolving color condition ("if it's blue") can test the targeted object.
func stackObjectColors(g *game.Game, obj *game.StackObject) ([]color.Color, bool) {
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
