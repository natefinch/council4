package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

// lowerSwitchPTContent lowers the one-shot continuous "switch power and
// toughness until end of turn" effect (CR 613.4e, layer 7e) into a single
// ApplyContinuous at LayerPowerToughnessSwitch over the source permanent for the
// turn ("Switch this creature's power and toughness until end of turn.",
// Aeromoeba). The parser only recognizes the source-affecting form, so any other
// shape — a condition, optional wrapper, keyword/mode rider, or a target — fails
// closed.
func lowerSwitchPTContent(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if !effect.SwitchPTSource ||
		effect.Negated ||
		effect.Optional ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported switch power/toughness effect",
			"the executable source backend supports only switching the source creature's power and toughness until end of turn",
		)
	}

	continuousEffects := []game.ContinuousEffect{{Layer: game.LayerPowerToughnessSwitch}}
	return sourceContinuousMode(continuousEffects), nil
}
