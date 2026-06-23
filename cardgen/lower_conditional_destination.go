package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// conditionalDestinationLinkedKey links the searched-and-revealed card to the
// ConditionalDestinationPlace that routes it. It is local to one resolution and
// never observed across abilities.
const conditionalDestinationLinkedKey = game.LinkedKey("conditional-destination-card")

// conditionalDestinationParts holds the recognized pieces of a
// search/reveal-then-conditional-placement ability body.
type conditionalDestinationParts struct {
	search       *compiler.CompiledEffect
	gate         *compiler.CompiledCondition
	elseZone     zone.Type
	elseBottom   bool
	elseOptional bool
	entryTapped  bool
	shuffle      bool
}

// recognizeConditionalDestination matches the closed "search your library for a
// card and reveal it; if <gate>, you may put it onto the battlefield (tapped);
// if you don't, put it into your hand (or on the bottom of your library); then
// shuffle" shape (Scholar of New Horizons). It fails closed on any other body so
// an unrepresentable variant is never silently dropped. The gate must lower in
// the effect-gate context, which excludes the look-family "if it's a land card"
// characteristic gate the compiler does not yet structure.
func recognizeConditionalDestination(content compiler.AbilityContent) (conditionalDestinationParts, bool) {
	if len(content.Modes) != 0 ||
		len(content.Targets) != 0 ||
		len(content.Conditions) != 2 {
		return conditionalDestinationParts{}, false
	}
	effects := content.Effects
	if len(effects) != 5 && len(effects) != 6 {
		return conditionalDestinationParts{}, false
	}
	search := &effects[0]
	reveal := &effects[1]
	mayPut := &effects[2]
	negatedPut := &effects[3]
	elsePut := &effects[4]
	if search.Kind != compiler.EffectSearch ||
		search.Context != parser.EffectContextController ||
		len(search.Targets) != 0 {
		return conditionalDestinationParts{}, false
	}
	if reveal.Kind != compiler.EffectReveal ||
		reveal.Connection != parser.EffectConnectionAnd {
		return conditionalDestinationParts{}, false
	}
	if mayPut.Kind != compiler.EffectPut ||
		!mayPut.Optional ||
		mayPut.Negated ||
		mayPut.ToZone != zone.Battlefield {
		return conditionalDestinationParts{}, false
	}
	if negatedPut.Kind != compiler.EffectPut ||
		!negatedPut.Negated ||
		negatedPut.ToZone != zone.Battlefield {
		return conditionalDestinationParts{}, false
	}
	if elsePut.Kind != compiler.EffectPut ||
		elsePut.Negated {
		return conditionalDestinationParts{}, false
	}
	elseBottom := false
	switch elsePut.ToZone {
	case zone.Hand:
	case zone.Library:
		if elsePut.Destination != parser.EffectDestinationBottom {
			return conditionalDestinationParts{}, false
		}
		elseBottom = true
	default:
		return conditionalDestinationParts{}, false
	}
	shuffle := false
	if len(effects) == 6 {
		tail := &effects[5]
		if tail.Kind != compiler.EffectShuffle ||
			tail.Connection != parser.EffectConnectionThen {
			return conditionalDestinationParts{}, false
		}
		shuffle = true
	}
	return conditionalDestinationParts{
		search:       search,
		gate:         &content.Conditions[0],
		elseZone:     elsePut.ToZone,
		elseBottom:   elseBottom,
		elseOptional: elsePut.Optional,
		entryTapped:  mayPut.EntersTapped,
		shuffle:      shuffle,
	}, true
}

// lowerConditionalDestinationPlace lowers the recognized search/reveal then
// gated-optional placement body to a RevealOnly search that publishes the found
// card, a ConditionalDestinationPlace that routes it to the battlefield or the
// fallback zone under the lowered gate, and the closing library shuffle.
func lowerConditionalDestinationPlace(ctx contentCtx) (game.AbilityContent, bool) {
	parts, ok := recognizeConditionalDestination(ctx.content)
	if !ok {
		return game.AbilityContent{}, false
	}
	if ctx.optional {
		return game.AbilityContent{}, false
	}
	spec, ok := searchSpecForSelector(parts.search.Selector)
	if !ok {
		return game.AbilityContent{}, false
	}
	if spec.Filter.Empty() || spec.MaxManaValueFromX {
		return game.AbilityContent{}, false
	}
	spec.SourceZone = zone.Library
	spec.RevealOnly = true
	spec.Reveal = true
	gate, ok := lowerCondition(*parts.gate, conditionContextEffectGate)
	if !ok {
		return game.AbilityContent{}, false
	}
	linkedCard := game.CardReference{Kind: game.CardReferenceLinked, LinkID: string(conditionalDestinationLinkedKey)}
	sequence := []game.Instruction{
		{
			Primitive: game.Search{
				Player:        game.ControllerReference(),
				Spec:          spec,
				Amount:        game.Fixed(1),
				PublishLinked: conditionalDestinationLinkedKey,
			},
		},
		{
			Primitive: game.ConditionalDestinationPlace{
				Card:         linkedCard,
				FromZone:     zone.Library,
				Condition:    opt.Val(game.EffectCondition{Condition: opt.Val(gate)}),
				EntryTapped:  parts.entryTapped,
				Else:         parts.elseZone,
				ElseBottom:   parts.elseBottom,
				ElseOptional: parts.elseOptional,
			},
		},
	}
	if parts.shuffle {
		sequence = append(sequence, game.Instruction{
			Primitive: game.ShuffleLibrary{Player: game.ControllerReference()},
		})
	}
	return game.Mode{Text: ctx.text, Sequence: sequence}.Ability(), true
}
