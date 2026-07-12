package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// lowerActivatedBodyContent lowers an activated ability's resolving body. It
// first routes the parser-recognized conditional look-at-top battlefield exact
// sequence — "{cost}: Look at the top card of your library. If it's a <type>
// card, you may put it onto the battlefield[ tapped]. If you don't put the card
// onto the battlefield, you may put it on the bottom of your library." (Lantern
// of Revealing, Parcelbeast) — to the same fixed instruction template as the
// triggered form, then defers to the generic ability-content lowerer for every
// other body. It fails closed unless the compiler recorded at least one card
// type, so a partial recognition never lowers to a silently-wrong body.
func lowerActivatedBodyContent(
	cardName string,
	ability compiler.CompiledAbility,
	bodyContent compiler.AbilityContent,
	bodySyntax *parser.Ability,
	bodyText string,
	variableCounterRemovalCost bool,
) (game.AbilityContent, *shared.Diagnostic) {
	if ability.StarCompassMana {
		return game.TapManaLandsProduceMatchingAbility(
			game.PlayerYou,
			false,
			game.Selection{Supertypes: []types.Super{types.Basic}},
		).Content, nil
	}
	if ability.ProgenitorIconNextFlash {
		return game.Mode{
			Text: bodyText,
			Sequence: []game.Instruction{{Primitive: game.ApplyRule{
				RuleEffects: []game.RuleEffect{{
					Kind:                   game.RuleEffectCastSpellsAsThoughFlash,
					AffectedPlayer:         game.PlayerYou,
					AppliesToNextSpellOnly: true,
					SpellChosenSubtypeFrom: game.EntryTypeChoiceKey,
				}},
				Duration: game.DurationThisTurn,
			}}},
		}.Ability(), nil
	}
	if ability.EvolutionaryLeapRevealUntil {
		return game.Mode{
			Text: bodyText,
			Sequence: []game.Instruction{{Primitive: game.RevealUntil{
				Player:                             game.ControllerReference(),
				Until:                              game.Selection{RequiredTypes: []types.Card{types.Creature}},
				Destination:                        zone.Hand,
				MatchToDestinationRestRandomBottom: true,
			}}},
		}.Ability(), nil
	}
	if ability.ExactSequence == compiler.ExactSequenceConditionalLookAtTopBattlefield &&
		len(ability.ExactSequenceLookAtTopTypes) > 0 {
		return game.Mode{
			Text:     bodyText,
			Sequence: conditionalLookAtTopBattlefieldSequence(ability),
		}.Ability(), nil
	}
	ctx := contentCtx{
		text:                       bodySyntax.Text,
		span:                       bodySyntax.Span,
		content:                    bodyContent,
		enclosingKind:              ability.Kind,
		variableCounterRemovalCost: variableCounterRemovalCost,
	}
	return lowerContent(cardName, ctx, bodySyntax)
}
