package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// lowerStaticDevotionNotCreature lowers the Theros Gods' devotion-gated
// type-changing static "As long as your devotion to <color(s)> is less than N,
// <source> isn't a creature." (Purphoros, God of the Forge; the full God
// family) into a battlefield continuous static ability. The ability removes the
// creature card type from its source while the controller's devotion to the
// declared colors is below the threshold, reusing the type layer and the
// devotion aggregate condition.
//
// Per the Gods' rulings the type-changing ability functions only on the
// battlefield (a God is always a creature card in other zones and a creature
// spell on the stack), which a plain battlefield LayerType effect models. The
// gate reads devotion — the count of matching mana symbols among controlled
// permanents' mana costs (CR 700.5) — which never depends on the creature type
// it removes, so the effect turns on and off cleanly as devotion crosses the
// threshold. Like Bestow's self type-change, only the creature card type is
// removed; the God keeps its printed subtype and stays a legendary enchantment.
func lowerStaticDevotionNotCreature(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (abilityLowering, bool, *shared.Diagnostic) {
	declarations := ability.Static.Declarations
	if len(declarations) != 1 || declarations[0].Kind != compiler.StaticDeclarationDevotionNotCreature {
		return abilityLowering{}, false, nil
	}
	declaration := declarations[0]
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		!rulesFreeAbilityWordLabel(ability.AbilityWord) {
		return abilityLowering{}, true, staticDeclarationDiagnostic(
			ability,
			"unsupported static declaration shell",
			"the recognized static declarations require an otherwise empty static ability shell",
		)
	}
	if !staticDeclarationPayloadValid(declaration) ||
		declaration.Condition != nil ||
		declaration.DevotionNotCreature == nil ||
		len(declaration.DevotionNotCreature.Colors) == 0 ||
		declaration.DevotionNotCreature.Threshold < 1 {
		return abilityLowering{}, true, staticDeclarationDiagnostic(
			ability,
			"unsupported static declaration operation",
			"the recognized static declaration operation is not representable by the runtime static-value vocabulary",
		)
	}
	body := game.StaticAbility{
		Text: ability.Text,
		Condition: opt.Val(game.Condition{
			Aggregates: []game.AggregateComparison{{
				Aggregate: game.AggregateControllerDevotion,
				Colors:    append([]color.Color(nil), declaration.DevotionNotCreature.Colors...),
				Op:        compare.LessThan,
				Value:     declaration.DevotionNotCreature.Threshold,
			}},
		}),
		ContinuousEffects: []game.ContinuousEffect{{
			Layer:          game.LayerType,
			AffectedSource: true,
			RemoveTypes:    []types.Card{types.Creature},
		}},
	}
	spans := make([]shared.Span, 0, 1+len(syntax.Reminders))
	spans = append(spans, declaration.Span)
	for _, reminder := range syntax.Reminders {
		spans = append(spans, reminder.Span)
	}
	return abilityLowering{
		staticAbilities: []loweredStaticAbility{{
			Body:    body,
			VarName: canonicalStaticDeclarationVarName(declaration),
		}},
		consumed: semanticConsumption{
			conditions:   len(ability.Content.Conditions),
			effects:      len(ability.Content.Effects),
			keywords:     len(ability.Content.Keywords),
			references:   len(ability.Content.References),
			declarations: len(declarations),
		},
		sourceSpans: spans,
	}, true, nil
}
