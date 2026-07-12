package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
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
	if ability.UnlicensedHearseExile {
		content, ok := lowerTargetedGraveyardExile(contentCtx{
			text:          bodyText,
			span:          ability.Span,
			content:       bodyContent,
			enclosingKind: compiler.AbilityActivated,
		})
		if !ok || len(content.Modes) != 1 {
			panic("lowerActivatedBodyContent: exact Unlicensed Hearse exile did not lower")
		}
		for i := range content.Modes[0].Sequence {
			move, ok := content.Modes[0].Sequence[i].Primitive.(game.MoveCard)
			if !ok {
				panic("lowerActivatedBodyContent: exact Unlicensed Hearse exile emitted non-MoveCard")
			}
			move.PublishLinked = exiledWithSourceKey
			move.PublishLinkedObjectScoped = true
			content.Modes[0].Sequence[i].Primitive = move
		}
		return content, nil
	}
	if exchange := ability.LifeCharacteristicExchange; exchange != nil {
		var characteristic game.SourcePowerToughness
		switch exchange.Kind {
		case compiler.LifeCharacteristicExchangeSourcePower:
			characteristic = game.SourcePower
		case compiler.LifeCharacteristicExchangeSourceToughness:
			characteristic = game.SourceToughness
		default:
			panic("lowerActivatedBodyContent: life-characteristic exchange has no characteristic")
		}
		player := game.ControllerReference()
		var targets []game.TargetSpec
		if exchange.TargetOpponent {
			player = game.TargetPlayerReference(0)
			targets = []game.TargetSpec{controllerPlayerTargetSpec(true)}
		}
		return game.Mode{
			Targets: targets,
			Sequence: []game.Instruction{{Primitive: game.ExchangeLifeTotalWithSourceCharacteristic{
				Player:         player,
				Characteristic: characteristic,
			}}},
		}.Ability(), nil
	}
	if ability.SelesnyaEulogistPopulate {
		return game.AbilityContent{
			MinModes: 1,
			MaxModes: 1,
			Modes: []game.Mode{{
				Text: bodyText,
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target creature card from a graveyard",
					Allow:      game.TargetAllowCard,
					TargetZone: zone.Graveyard,
					Selection: opt.Val(game.Selection{
						RequiredTypes: []types.Card{types.Creature},
					}),
				}},
				Sequence: []game.Instruction{
					{Primitive: game.MoveCard{
						Card:        game.CardReference{Kind: game.CardReferenceTarget},
						FromZone:    zone.Graveyard,
						Destination: zone.Exile,
					}},
					{Primitive: game.CreateToken{
						Amount: game.Fixed(1),
						Source: game.TokenCopyOf(game.TokenCopySpec{
							Source: game.TokenCopySourceChosenControlledCreatureToken,
						}),
					}},
				},
			}},
		}, nil
	}
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
