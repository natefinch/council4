package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// applyCopyTokenOverride maps a recognized copy-token characteristic-overriding
// "except" exception onto the token-copy spec. The created token first copies
// its source and then applies these power/toughness, color, card-type, subtype,
// and keyword overrides. Colors and subtypes replace the copied values unless
// the parser marked them additive ("in addition to its other colors and
// types"), in which case they append; card types always append. It reports
// false (fail closed) when an override keyword has no runtime form.
func applyCopyTokenOverride(spec *game.TokenCopySpec, effect *compiler.CompiledEffect) bool {
	if !effect.TokenCopyOverride {
		return true
	}
	if effect.TokenCopyOverridePTKnown {
		spec.SetPower = opt.Val(game.PT{Value: effect.TokenCopyOverridePower})
		spec.SetToughness = opt.Val(game.PT{Value: effect.TokenCopyOverrideToughness})
	}
	if len(effect.TokenCopyOverrideColors) > 0 {
		if effect.TokenCopyOverrideAdditiveColors {
			spec.AddColors = append(spec.AddColors, effect.TokenCopyOverrideColors...)
		} else {
			spec.SetColors = append(spec.SetColors, effect.TokenCopyOverrideColors...)
		}
	}
	if len(effect.TokenCopyOverrideSubtypes) > 0 {
		if effect.TokenCopyOverrideAdditiveTypes {
			spec.AddSubtypes = append(spec.AddSubtypes, effect.TokenCopyOverrideSubtypes...)
		} else {
			spec.SetSubtypes = append(spec.SetSubtypes, effect.TokenCopyOverrideSubtypes...)
		}
	}
	spec.AddTypes = append(spec.AddTypes, effect.TokenCopyOverrideTypes...)
	for _, kind := range effect.TokenCopyOverrideKeywords {
		keyword, ok := runtimeKeyword(kind)
		if !ok {
			return false
		}
		spec.AddKeywords = append(spec.AddKeywords, keyword)
	}
	return true
}
