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

// cardMode is one resolution mode paired with whether the mode is targeted —
// either by its own target slice or by the shared targets of the enclosing
// ability content (modal abilities such as charms and commands place their
// targets in SharedTargets, not on each mode; see game.BodyTargets).
type cardMode struct {
	sequence []game.Instruction
	targeted bool
}

// cardModes returns every resolution mode across all of a card's faces and
// ability kinds, so its effect primitives can be inspected for tagging.
func cardModes(def *game.CardDef) []cardMode {
	var modes []cardMode
	for _, face := range cardFaces(def) {
		if face.SpellAbility.Exists {
			modes = appendContentModes(modes, &face.SpellAbility.Val)
		}
		for i := range face.ActivatedAbilities {
			modes = appendContentModes(modes, &face.ActivatedAbilities[i].Content)
		}
		for i := range face.ManaAbilities {
			modes = appendContentModes(modes, &face.ManaAbilities[i].Content)
		}
		for i := range face.TriggeredAbilities {
			modes = appendContentModes(modes, &face.TriggeredAbilities[i].Content)
		}
		for i := range face.ChapterAbilities {
			modes = appendContentModes(modes, &face.ChapterAbilities[i].Content)
		}
		for i := range face.LoyaltyAbilities {
			modes = appendContentModes(modes, &face.LoyaltyAbilities[i].Content)
		}
	}
	return modes
}

func appendContentModes(modes []cardMode, content *game.AbilityContent) []cardMode {
	shared := len(content.SharedTargets) > 0
	for i := range content.Modes {
		modes = append(modes, cardMode{
			sequence: content.Modes[i].Sequence,
			targeted: shared || len(content.Modes[i].Targets) > 0,
		})
	}
	return modes
}

func sequencePrimitiveKinds(sequence []game.Instruction) map[game.PrimitiveKind]bool {
	kinds := make(map[game.PrimitiveKind]bool, len(sequence))
	for i := range sequence {
		if sequence[i].Primitive == nil {
			continue
		}
		kinds[sequence[i].Primitive.Kind()] = true
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
