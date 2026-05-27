package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

func cardFaceDef(card *game.CardInstance, face game.FaceIndex) (*game.CardDef, bool) {
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
		return permanent.TokenDef.FaceDef(permanent.Face)
	}
	card, ok := g.GetCardInstance(permanent.CardInstanceID)
	if !ok {
		return nil, false
	}
	return card.Def.FaceDef(permanent.Face)
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
	if !def.IsTransformingDoubleFaced() || len(def.Faces) < 2 {
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

func cardTypes(def *game.CardDef) []game.CardType {
	return append([]game.CardType(nil), def.Types...)
}

func permanentCardID(permanent *game.Permanent) id.ID {
	return permanent.CardInstanceID
}
