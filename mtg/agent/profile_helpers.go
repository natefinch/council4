package agent

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

// evasionKeywords are the keywords that make a creature hard to block, used to
// flag a creature as a threat even when its power is modest.
var evasionKeywords = []game.Keyword{game.Flying, game.Menace, game.Trample}

func defHasType(def *game.CardDef, cardType types.Card) bool {
	return slices.Contains(def.Types, cardType)
}

// curveBucket maps a mana value to its mana-curve bucket index: the value itself
// for 0..6, and the final bucket for 7 or more.
func curveBucket(manaValue int) int {
	if manaValue >= curveBuckets-1 {
		return curveBuckets - 1
	}
	if manaValue < 0 {
		return 0
	}
	return manaValue
}

func defPower(def *game.CardDef) int {
	if def.Power.Exists {
		return def.Power.Val.Value
	}
	return 0
}

func defHasEvasion(def *game.CardDef) bool {
	return slices.ContainsFunc(evasionKeywords, def.HasKeyword)
}

func hasIntrinsicManaAbility(def *game.CardDef) bool {
	for _, face := range cardFaces(def) {
		if len(face.ManaAbilities) > 0 {
			return true
		}
	}
	return false
}

// cardFaces returns every printed face whose abilities should be analysed: the
// front face plus any back or alternate face (double-faced cards, adventures,
// split cards). Faces are returned by pointer to avoid copying the large
// CardFace struct.
func cardFaces(def *game.CardDef) []*game.CardFace {
	faces := []*game.CardFace{&def.CardFace}
	if def.Back.Exists {
		faces = append(faces, &def.Back.Val)
	}
	if def.Alternate.Exists {
		faces = append(faces, &def.Alternate.Val)
	}
	return faces
}

// cardModes returns every resolution mode across all of a card's faces and
// ability kinds, so its effect primitives can be inspected for tagging.
func cardModes(def *game.CardDef) []game.Mode {
	var modes []game.Mode
	for _, face := range cardFaces(def) {
		if face.SpellAbility.Exists {
			modes = append(modes, face.SpellAbility.Val.Modes...)
		}
		for i := range face.ActivatedAbilities {
			modes = append(modes, face.ActivatedAbilities[i].Content.Modes...)
		}
		for i := range face.ManaAbilities {
			modes = append(modes, face.ManaAbilities[i].Content.Modes...)
		}
		for i := range face.TriggeredAbilities {
			modes = append(modes, face.TriggeredAbilities[i].Content.Modes...)
		}
		for i := range face.ChapterAbilities {
			modes = append(modes, face.ChapterAbilities[i].Content.Modes...)
		}
		for i := range face.LoyaltyAbilities {
			modes = append(modes, face.LoyaltyAbilities[i].Content.Modes...)
		}
	}
	return modes
}

func modePrimitiveKinds(mode game.Mode) map[game.PrimitiveKind]bool {
	kinds := make(map[game.PrimitiveKind]bool, len(mode.Sequence))
	for i := range mode.Sequence {
		if mode.Sequence[i].Primitive == nil {
			continue
		}
		kinds[mode.Sequence[i].Primitive.Kind()] = true
	}
	return kinds
}

func orderedColors(present map[color.Color]bool) []color.Color {
	if len(present) == 0 {
		return nil
	}
	ordered := make([]color.Color, 0, len(present))
	for _, c := range color.AllColors() {
		if present[c] {
			ordered = append(ordered, c)
		}
	}
	return ordered
}
