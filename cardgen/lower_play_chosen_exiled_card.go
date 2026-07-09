package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// lowerPlayChosenExiledCard lowers Dauthi Voidwalker's activated-ability body
// "Choose an exiled card an opponent owns with a void counter on it. You may
// play it this turn without paying its mana cost." into a single
// PlayChosenExiledCard primitive.
//
// The compiler models this as two effects: effect[0] is the mandatory
// EffectChooseExiledCard choice (source zone Exile, opponent owner scope, a
// named marker-counter filter), and effect[1] is the optional EffectPlay
// permission that back-references the chosen card ("it") and carries the play
// window plus the free-cast rider. Because choosing the card and granting its
// play permission share one card identity, they lower together into the
// combined primitive that resolves the choice and grants the per-card
// play-from-exile permission atomically, mirroring how lowerExileForPlay pairs
// an exile with its play grant.
//
// The choice itself is mandatory (the instruction is not optional); the granted
// "you may play it" permission is inherently optional to use. Only the exact
// two-effect shape with a known marker counter, an opponent owner scope, and a
// bounded play window is accepted; any other shape fails closed.
func lowerPlayChosenExiledCard(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Conditions) != 0 {
		return game.AbilityContent{}, false
	}
	choose := ctx.content.Effects[0]
	if choose.Kind != compiler.EffectChooseExiledCard ||
		!choose.Exact ||
		choose.Optional ||
		choose.Negated ||
		choose.Context != parser.EffectContextController ||
		choose.FromZone != zone.Exile ||
		!choose.ChooseExiledCardOwnerOpponent ||
		!choose.CounterKindKnown {
		return game.AbilityContent{}, false
	}
	play := ctx.content.Effects[1]
	if play.Kind != compiler.EffectPlay ||
		!play.Exact ||
		!play.Optional ||
		play.Negated ||
		play.Context != parser.EffectContextController ||
		!play.CastWithoutPayingManaCost ||
		len(play.References) != 1 {
		return game.AbilityContent{}, false
	}
	duration, ok := lowerImpulseExileDuration(play.Duration)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.PlayChosenExiledCard{
			Player:                game.ControllerReference(),
			Zone:                  zone.Exile,
			OwnerScope:            game.PlayerOpponent,
			Counter:               opt.Val(choose.CounterKind),
			Duration:              duration,
			WithoutPayingManaCost: true,
		},
	}}}.Ability(), true
}
