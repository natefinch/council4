package game

import (
	"fmt"
	"strings"

	"github.com/natefinch/council4/mtg/game/color"
)

// HexproofFromKeyword parameterizes "hexproof from [quality]" for one or more
// colors (CR 702.11e). It is a source-color-qualified targeting restriction:
// the permanent or player with this ability can't be the target of spells an
// opponent controls or abilities an opponent controls from a source of any
// named color. Unlike full Hexproof it never blocks the controller's own
// spells and abilities, and unlike Protection it is a targeting restriction
// only — it never prevents damage, enchanting/equipping, or blocking.
type HexproofFromKeyword struct {
	FromColors []color.Color
}

func (HexproofFromKeyword) isKeywordAbility() {}

func (HexproofFromKeyword) keyword() Keyword { return HexproofFrom }

func (ability HexproofFromKeyword) cloneKeywordAbility() KeywordAbility {
	ability.FromColors = append([]color.Color(nil), ability.FromColors...)
	return ability
}

// HexproofFromColorsStaticAbility builds the complete static ability for
// "hexproof from" one or more colors.
func HexproofFromColorsStaticAbility(colors ...color.Color) StaticAbility {
	protectedColors := append([]color.Color(nil), colors...)
	validateHexproofFromColors(protectedColors)
	return StaticAbility{
		Text: hexproofFromColorsText(protectedColors),
		KeywordAbilities: []KeywordAbility{
			HexproofFromKeyword{FromColors: protectedColors},
		},
	}
}

// StaticBodyHexproofFromKeyword returns the HexproofFromKeyword from a
// StaticAbility body.
func StaticBodyHexproofFromKeyword(body *StaticAbility) (HexproofFromKeyword, bool) {
	ka, ok := BodyKeywordAbility(body, HexproofFrom)
	if !ok {
		return HexproofFromKeyword{}, false
	}
	hexproof, ok := ka.(HexproofFromKeyword)
	return hexproof, ok
}

func hexproofFromColorsText(colors []color.Color) string {
	phrases := make([]string, len(colors))
	for i, c := range colors {
		phrases[i] = "from " + strings.ToLower(string(c))
	}
	switch len(phrases) {
	case 0:
		return "Hexproof"
	case 1:
		return "Hexproof " + phrases[0]
	case 2:
		return "Hexproof " + phrases[0] + " and " + phrases[1]
	default:
		return "Hexproof " +
			strings.Join(phrases[:len(phrases)-1], ", ") +
			", and " +
			phrases[len(phrases)-1]
	}
}

// validateHexproofFromColors panics if colors is empty or names an unknown or
// duplicate color. It mirrors validateProtectionColors.
func validateHexproofFromColors(colors []color.Color) {
	if len(colors) == 0 {
		panic("game: hexproof from requires at least one color")
	}
	seen := make(map[color.Color]struct{}, len(colors))
	for _, c := range colors {
		switch c {
		case color.White, color.Blue, color.Black, color.Red, color.Green:
		default:
			panic(fmt.Sprintf("game: invalid hexproof-from color %q", c))
		}
		if _, ok := seen[c]; ok {
			panic(fmt.Sprintf("game: duplicate hexproof-from color %q", c))
		}
		seen[c] = struct{}{}
	}
}
