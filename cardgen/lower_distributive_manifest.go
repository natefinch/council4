package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

// exiledForEachPlayerManifestKey links a distributive per-player exile to a
// per-controller manifest or cloak payoff.
const exiledForEachPlayerManifestKey = game.LinkedKey("exiled-for-each-player-manifest")

// lowerExileForEachPlayerManifestChainContent lowers "For each player, exile up
// to one target <permanent> that player controls. For each permanent exiled this
// way, its controller <manifests/cloaks> ..." into a distributive exile followed
// by one face-down library action per linked permanent's last-known controller.
func lowerExileForEachPlayerManifestChainContent(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) != 2 ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	exileEffect := ctx.content.Effects[0]
	if exileEffect.Kind != compiler.EffectExile ||
		!exileEffect.ExileForEachPlayer ||
		!exileEffect.Exact ||
		exileEffect.Negated ||
		exileEffect.Optional ||
		exileEffect.Context != parser.EffectContextController {
		return game.AbilityContent{}, false
	}
	manifestEffect := ctx.content.Effects[1]
	if manifestEffect.Kind != compiler.EffectCloak ||
		!manifestEffect.CloakForEachExiledThisWay ||
		!manifestEffect.Exact ||
		manifestEffect.Negated ||
		manifestEffect.Optional ||
		manifestEffect.Context != parser.EffectContextReferencedObjectController {
		return game.AbilityContent{}, false
	}
	if !referencesAreThatPlayerOrItsController(ctx.content.References) {
		return game.AbilityContent{}, false
	}
	selection, ok := SelectionForSelector(exileEffect.Selector)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: []game.Instruction{
		{Primitive: game.ExileForEachPlayer{
			Chooser:   game.ControllerReference(),
			Selection: selection,
			LinkedKey: exiledForEachPlayerManifestKey,
		}},
		{Primitive: game.ManifestForEachLinked{
			Cloak:     true,
			LinkedKey: exiledForEachPlayerManifestKey,
		}},
	}}.Ability(), true
}
