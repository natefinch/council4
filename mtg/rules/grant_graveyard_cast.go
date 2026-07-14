package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// escapeGrantLabel and retraceGrantLabel are the cost-option labels for
// dynamically granted graveyard casts. They mirror the native keyword names so
// the payment UI reads the same whether the permission is printed or granted.
const (
	escapeGrantLabel  = "Escape"
	retraceGrantLabel = "Retrace"
)

// grantedGraveyardCastAlternatives returns the alternative costs that active
// rule effects grant for casting def from playerID's graveyard. It is the
// text-blind runtime behind RuleEffectGrantGraveyardCardKeyword: any card whose
// printed characteristics match the grant's selection may be cast from the
// graveyard for a computed alternative cost (Underworld Breach's escape, Six's
// retrace). Both mechanics reuse the repeatable Escape alternative (the spell is
// not exiled on resolution, so it returns to the graveyard and can be recast),
// which the existing planner and escape-resolution paths already model.
//
// The grant applies only when the effect reaches playerID (CR 800, "your
// graveyard" resolves to the effect's controller) and its during-your-turn
// restriction, if any, currently holds. Selection matching is delegated to the
// same printed-characteristic matcher the choice layer uses, so no card-name or
// Oracle-text inspection happens at runtime.
func grantedGraveyardCastAlternatives(g *game.Game, playerID game.PlayerID, def *game.CardDef) []cost.Alternative {
	if def == nil {
		return nil
	}
	var alternatives []cost.Alternative
	effects := activeRuleEffects(g)
	for i := range effects {
		effect := &effects[i]
		if effect.Kind != game.RuleEffectGrantGraveyardCardKeyword {
			continue
		}
		if !ruleEffectAffectsPlayer(g, effect, playerID) {
			continue
		}
		if !actionRestrictionTurnActive(g, effect) {
			continue
		}
		if !cardDefMatchesCostSelection(g, def, effect.CardSelection) {
			continue
		}
		alternative, ok := grantedGraveyardCastAlternative(def, effect)
		if !ok {
			continue
		}
		alternatives = append(alternatives, alternative)
	}
	return alternatives
}

// grantedGraveyardCastAlternative builds the alternative cost for a single
// graveyard-cast keyword grant, or reports false when the grant is not a runtime
// castable form.
//
// Escape uses the computed GraveyardCastCost the compiler lowered from the
// card's Oracle wording: the card's own mana cost plus the grant's additional
// costs (Underworld Breach's "exile three other cards from your graveyard").
// Retrace uses its intrinsic cost defined by the keyword itself (CR 702.83): the
// card's mana cost plus discarding a land card. Retrace carries no computed cost
// on the rule effect, so its cost is synthesized here rather than read from the
// payload, keeping generated card sources for retrace grants byte-identical.
//
// Any other granted keyword, an escape grant missing its computed cost, or a
// retrace grant that unexpectedly carries a computed cost fails closed.
func grantedGraveyardCastAlternative(def *game.CardDef, effect *game.RuleEffect) (cost.Alternative, bool) {
	switch effect.GrantedKeyword {
	case game.Escape:
		castCost := effect.GraveyardCastCost
		if castCost.IsZero() || !castCost.UseCardManaCost || len(castCost.AdditionalCosts) == 0 {
			return cost.Alternative{}, false
		}
		alternative := cost.Alternative{
			Label:           escapeGrantLabel,
			Mechanic:        cost.AlternativeMechanicEscape,
			AdditionalCosts: slices.Clone(castCost.AdditionalCosts),
		}
		if def.ManaCost.Exists {
			alternative.ManaCost = opt.Val(slices.Clone(def.ManaCost.Val))
		}
		return alternative, true
	case game.Retrace:
		if !effect.GraveyardCastCost.IsZero() {
			return cost.Alternative{}, false
		}
		alternative := cost.Alternative{
			Label:    retraceGrantLabel,
			Mechanic: cost.AlternativeMechanicEscape,
			AdditionalCosts: []cost.Additional{{
				Kind:          cost.AdditionalDiscard,
				Amount:        1,
				MatchCardType: true,
				CardType:      types.Land,
				Text:          "Discard a land card",
			}},
		}
		if def.ManaCost.Exists {
			alternative.ManaCost = opt.Val(slices.Clone(def.ManaCost.Val))
		}
		return alternative, true
	default:
		return cost.Alternative{}, false
	}
}
