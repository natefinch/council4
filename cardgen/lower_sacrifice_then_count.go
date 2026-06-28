package cardgen

import (
	"reflect"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

// sacrificedThisWayResultKey is the published count of permanents sacrificed by a
// count-scaled "sacrifice {all|any number of} X, then <reward> that many/much"
// sequence. The reward's dynamic amount reads it through
// DynamicAmountPreviousEffectResult so the controller's reward scales to exactly
// the number of permanents they sacrificed.
const sacrificedThisWayResultKey = game.ResultKey("sacrificed-this-way")

// lowerSacrificeThenCountSequence lowers a count-scaled sacrifice sequence
// "Sacrifice all creatures you control, then create that many 4/4 red Hellion
// creature tokens." (Hellion Eruption) and "Sacrifice any number of lands, then
// add that much {C}." (Mana Seism) into a SacrificePermanents that publishes the
// number sacrificed followed by a reward (CreateToken, AddMana, or Draw) whose
// amount reads that count. The parser marks the sacrifice effect's
// SacrificeThenCount and SacrificeAnyNumber fields, so this lowering reads no
// Oracle words. It accepts only a faithful single-card-type sacrifice selection
// so a partially captured filter can never produce a wrong card; any other shape
// leaves the sequence unrecognized for the caller to fail closed.
func lowerSacrificeThenCountSequence(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		ctx.optional {
		return game.AbilityContent{}, false
	}
	sacrifice := ctx.content.Effects[0]
	reward := ctx.content.Effects[1]
	if sacrifice.Kind != compiler.EffectSacrifice ||
		!sacrifice.SacrificeThenCount ||
		sacrifice.Negated ||
		sacrifice.Context != parser.EffectContextController {
		return game.AbilityContent{}, false
	}
	selection, ok := faithfulSingleTypeSacrificeSelection(sacrifice.Selector)
	if !ok {
		return game.AbilityContent{}, false
	}
	sacrificePrim := game.SacrificePermanents{
		Player:    game.ControllerReference(),
		Selection: selection,
	}
	switch {
	case sacrifice.Selector.All:
		sacrificePrim.All = true
	case sacrifice.SacrificeAnyNumber:
		sacrificePrim.AnyNumber = true
	default:
		return game.AbilityContent{}, false
	}
	rewardPrim, ok := countScaledRewardPrimitive(reward)
	if !ok {
		return game.AbilityContent{}, false
	}
	sequence := []game.Instruction{
		{
			Primitive:     sacrificePrim,
			PublishResult: sacrificedThisWayResultKey,
		},
		{Primitive: rewardPrim},
	}
	return game.Mode{Sequence: sequence}.Ability(), true
}

// faithfulSingleTypeSacrificeSelection returns the runtime selection for a
// sacrifice whose filter is a single card type with no further constraints
// ("creatures you control", "lands"), the only forms whose canonical capture is
// guaranteed faithful. It rejects type unions, subtypes, exclusions, token
// qualifiers, and every other shape so a count-scaled sacrifice never sacrifices
// from a partially captured filter.
func faithfulSingleTypeSacrificeSelection(selector compiler.CompiledSelector) (game.Selection, bool) {
	selection, ok := sacrificeChoiceSelection(selector)
	if !ok {
		return game.Selection{}, false
	}
	if len(selection.RequiredTypes) != 1 {
		return game.Selection{}, false
	}
	if !reflect.DeepEqual(selection, game.Selection{RequiredTypes: selection.RequiredTypes}) {
		return game.Selection{}, false
	}
	return selection, true
}

// countScaledRewardPrimitive builds the reward primitive for a count-scaled
// sacrifice sequence, scaling its amount to the published sacrificed count. It
// supports a fixed creature-token creation ("create that many <token>"), a
// single fixed mana color ("add that much {C}"), and a controller draw ("draw
// that many cards"); any other reward shape is rejected.
func countScaledRewardPrimitive(reward compiler.CompiledEffect) (game.Primitive, bool) {
	if reward.Negated ||
		reward.Context != parser.EffectContextController ||
		reward.DelayedTiming != 0 ||
		reward.Duration != compiler.DurationNone {
		return nil, false
	}
	amount := game.Dynamic(game.DynamicAmount{
		Kind:      game.DynamicAmountPreviousEffectResult,
		ResultKey: sacrificedThisWayResultKey,
	})
	switch reward.Kind {
	case compiler.EffectCreate:
		return countScaledCreateToken(reward, amount)
	case compiler.EffectAddMana:
		return countScaledAddMana(reward, amount)
	case compiler.EffectDraw:
		return game.Draw{Player: game.ControllerReference(), Amount: amount}, true
	default:
		return nil, false
	}
}

// countScaledCreateToken builds a "create that many <token>" reward, creating the
// published count of a fixed creature token. It rejects copy tokens, choice
// tokens, multi-token clauses, variable-size tokens, and non-controller
// recipients so only the plain synthesized creature-token form is produced.
func countScaledCreateToken(reward compiler.CompiledEffect, amount game.Quantity) (game.Primitive, bool) {
	if reward.TokenCopyOfTarget ||
		reward.TokenCopyOfReference ||
		reward.TokenCopyOfAttached ||
		reward.TokenCopyOfTriggeringSet ||
		reward.TokenCopyOfForEach ||
		reward.TokenChoice ||
		reward.TokenPTVariableX ||
		reward.TokenGrantedAbility != nil ||
		len(reward.AdditionalTokens) != 0 {
		return nil, false
	}
	def, ok := synthesizeCreatureTokenDef(&reward, nil)
	if !ok {
		return nil, false
	}
	return game.CreateToken{
		Amount:         amount,
		Source:         game.TokenDef(def),
		EntryTapped:    reward.Selector.Tapped,
		EntryAttacking: reward.Selector.Attacking,
	}, true
}

// countScaledAddMana builds an "add that much <symbol>" reward, adding the
// published count of a single fixed-color mana. The "that much" wording leaves
// the produced symbol in Details.Symbol rather than the Mana colors, so the color
// is read from there. It rejects color choices, any-color, and every dynamic or
// filtered mana form so only a fixed single color is produced.
func countScaledAddMana(reward compiler.CompiledEffect, amount game.Quantity) (game.Primitive, bool) {
	m := reward.Mana
	if m.Choice ||
		m.AnyColor ||
		m.ChosenColor ||
		m.ChosenColorDevotion ||
		m.ChosenColorDynamic ||
		m.CommanderIdentity ||
		m.FilterPair ||
		m.LandsProduce ||
		m.LinkedExileColors ||
		m.ColorsAmongControlled ||
		len(m.Colors) != 0 ||
		len(m.Symbols) != 0 {
		return nil, false
	}
	color, ok := manaColorValue(strings.Trim(reward.Details.Symbol, "{}"))
	if !ok {
		return nil, false
	}
	return game.AddMana{Amount: amount, ManaColor: color}, true
}
