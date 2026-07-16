package compiler

import (
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/color"
)

// recognizeStaticDevotionNotCreatureDeclaration maps the parser-owned
// devotion-gated "isn't a creature" syntax (Purphoros, God of the Forge; the
// full Theros God family) onto its closed semantic payload. It consumes the
// typed devotion colors and numeric threshold the parser captured and never
// inspects Oracle text. The whole sentence is a self-referential static, so the
// ability carries no cost, trigger, modes, targets, keywords, or ability word.
func recognizeStaticDevotionNotCreatureDeclaration(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) (StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationDevotionNotCreature) {
		return StaticDeclaration{}, false
	}
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		ability.AbilityWord != "" {
		return StaticDeclaration{}, false
	}
	node := statics[0]
	colors, ok := devotionNotCreatureColors(node.DevotionColors)
	if !ok || node.DevotionThreshold < 1 {
		return StaticDeclaration{}, false
	}
	return StaticDeclaration{
		Kind:          StaticDeclarationDevotionNotCreature,
		Span:          node.Span,
		OperationSpan: node.OperationSpan,
		DevotionNotCreature: &StaticDevotionNotCreatureDeclaration{
			Colors:    colors,
			Threshold: node.DevotionThreshold,
		},
	}, true
}

// devotionNotCreatureColors maps the parser devotion colors onto runtime colors,
// failing closed if any color is unrecognized so a malformed color list yields
// no declaration rather than a wrong-color devotion count.
func devotionNotCreatureColors(parserColors []parser.Color) ([]color.Color, bool) {
	if len(parserColors) == 0 {
		return nil, false
	}
	colors := make([]color.Color, 0, len(parserColors))
	for _, parserColor := range parserColors {
		runtimeColor, ok := runtimeColorFromParser(parserColor)
		if !ok {
			return nil, false
		}
		colors = append(colors, runtimeColor)
	}
	return colors, true
}
