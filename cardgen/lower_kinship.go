package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// Internal sequencing keys for the Kinship look-then-reveal flow. They are local
// to a single resolution and never observed across abilities.
const (
	kinshipLookedKey   = game.LinkedKey("kinship-top")
	kinshipRevealedKey = game.ResultKey("kinship-revealed")
)

// lowerKinshipReveal lowers the Kinship ability word's resolving body, "you may
// look at the top card of your library. If it shares a creature type with this
// creature, you may reveal it. If you do, <EFFECT>." into its fixed
// look-then-reveal template followed by the lowered payoff. The look and reveal
// are a fixed prefix; the trailing payoff varies per card, so it is lowered
// through the shared per-effect path and gated on whether the card was revealed.
// It is text-blind: it switches on the compiler-recognized look effect, optional
// reveal, shares-creature-type gate, and prior-instruction-accepted gate. It
// fails closed on any other shape.
func lowerKinshipReveal(
	cardName string,
	ctx contentCtx,
	syntax *parser.Ability,
) (game.AbilityContent, bool) {
	content := ctx.content
	if len(content.Effects) < 3 ||
		len(content.Modes) != 0 ||
		len(content.Targets) != 0 ||
		len(content.Keywords) != 0 {
		return game.AbilityContent{}, false
	}
	look := content.Effects[0]
	reveal := content.Effects[1]
	if look.Kind != compiler.EffectLookAtLibraryTop || !look.Optional {
		return game.AbilityContent{}, false
	}
	if reveal.Kind != compiler.EffectReveal || !reveal.Optional {
		return game.AbilityContent{}, false
	}
	if len(content.Conditions) != 2 {
		return game.AbilityContent{}, false
	}
	sawShares, sawPrior := false, false
	for i := range content.Conditions {
		switch content.Conditions[i].Predicate {
		case compiler.ConditionPredicateSubjectSharesCreatureTypeWithSource:
			sawShares = true
		case compiler.ConditionPredicatePriorInstructionAccepted:
			sawPrior = true
		default:
			return game.AbilityContent{}, false
		}
	}
	if !sawShares || !sawPrior {
		return game.AbilityContent{}, false
	}

	payoff, ok := lowerKinshipPayoff(cardName, ctx, syntax)
	if !ok {
		return game.AbilityContent{}, false
	}

	lookedCard := game.CardReference{Kind: game.CardReferenceLinked, LinkID: string(kinshipLookedKey)}
	sequence := []game.Instruction{
		{
			Primitive: game.LookAtLibraryTop{
				Player:        game.ControllerReference(),
				PublishLinked: kinshipLookedKey,
			},
			Optional: true,
		},
		{
			Primitive: game.Reveal{Card: lookedCard},
			CardCondition: opt.Val(game.CardSelection{
				Card: lookedCard,
				Selection: game.Selection{
					SharesCreatureTypeWithSource: true,
				},
			}),
			Optional:      true,
			PublishResult: kinshipRevealedKey,
		},
	}
	sequence = append(sequence, payoff...)
	return game.Mode{Text: ctx.text, Sequence: sequence}.Ability(), true
}

// lowerKinshipPayoff lowers each effect after the look-and-reveal prefix through
// the shared per-effect path and gates every produced instruction on the card
// having been revealed ("If you do, ..."). It fails closed if any payoff effect
// does not lower to a single non-modal mode, or already carries a result gate.
func lowerKinshipPayoff(
	cardName string,
	ctx contentCtx,
	syntax *parser.Ability,
) ([]game.Instruction, bool) {
	gate := game.InstructionResultGate{
		Key:       kinshipRevealedKey,
		Accepted:  game.TriTrue,
		Succeeded: game.TriTrue,
	}
	clauseSyntaxes := splitEffectSyntaxes(syntax, ctx.content.Effects)
	var gated []game.Instruction
	for i := 2; i < len(ctx.content.Effects); i++ {
		effect := &ctx.content.Effects[i]
		effectCtx := contextForEffect(ctx, effect)
		effectCtx.content.Conditions = nil
		clauseSyntax := clauseSyntaxes[i]
		content, diagnostic := lowerSequenceClauseContent(
			cardName,
			ctx,
			effectCtx.content,
			effect.Optional,
			&clauseSyntax,
			false,
		)
		if diagnostic != nil ||
			content.IsModal() ||
			len(content.Modes) != 1 ||
			len(content.SharedTargets) != 0 ||
			len(content.Modes[0].Targets) != 0 {
			return nil, false
		}
		for _, instruction := range content.Modes[0].Sequence {
			if instruction.ResultGate.Exists {
				return nil, false
			}
			instruction.ResultGate = opt.Val(gate)
			gated = append(gated, instruction)
		}
	}
	if len(gated) == 0 {
		return nil, false
	}
	return gated, true
}
