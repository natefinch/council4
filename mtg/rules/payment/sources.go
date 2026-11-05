package payment

import (
	"reflect"
	"slices"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

type permanentManaOutputResult struct {
	color        mana.Color
	amount       int
	snow         bool
	fromCreature bool
	untap        bool
	sacrifice    bool
	abilityIndex int
	timing       game.TimingRestriction
}

type paymentManaAbilityResult struct {
	index     int
	colors    []mana.Color
	amount    int
	untap     bool
	sacrifice bool
	timing    game.TimingRestriction
}

type paymentManaOutput struct {
	colors    []mana.Color
	amount    int
	untap     bool
	sacrifice bool
}

// permanentManaOutput returns the first payment-safe output of a permanent.
func permanentManaOutput(s State, permanent *game.Permanent) (permanentManaOutputResult, bool) {
	outputs := permanentManaOutputs(s, permanent)
	if len(outputs) == 0 {
		return permanentManaOutputResult{}, false
	}
	return outputs[0], true
}

func permanentManaOutputForActivation(s State, permanent *game.Permanent, activation manaTap) (permanentManaOutputResult, bool) {
	for _, output := range permanentManaOutputs(s, permanent) {
		if output.color == activation.color &&
			output.amount == activation.amount &&
			output.snow == activation.snow &&
			output.fromCreature == activation.fromCreature &&
			output.untap == activation.untap &&
			output.sacrifice == activation.sacrifice &&
			output.abilityIndex == activation.abilityIndex &&
			output.timing == activation.timing {
			return output, true
		}
	}
	return permanentManaOutputResult{}, false
}

func permanentManaOutputs(s State, permanent *game.Permanent) []permanentManaOutputResult {
	var outputs []permanentManaOutputResult
	controller := s.EffectiveController(permanent)
	multiplier := s.ManaProductionMultiplier(controller)
	fromCreature := s.PermanentHasType(permanent, types.Creature)
	abilities := paymentManaAbilities(s, controller, permanent)
	for _, ability := range abilities {
		for _, color := range ability.colors {
			outputs = appendUniquePermanentManaOutput(outputs, permanentManaOutputResult{
				color:        color,
				amount:       multiplyTappedManaAmount(ability.amount, ability.untap, multiplier),
				snow:         s.PermanentHasSupertype(permanent, types.Snow),
				fromCreature: fromCreature,
				untap:        ability.untap,
				sacrifice:    ability.sacrifice,
				abilityIndex: ability.index,
				timing:       ability.timing,
			})
		}
	}
	return outputs
}

// multiplyTappedManaAmount scales a mana ability's produced amount by an active
// RuleEffectManaProductionMultiplier (Mana Reflection, Nyxbloom Ancient). The
// replacement applies only when a permanent is tapped to produce the mana, so an
// untap-cost ability ({Q}) is left unscaled. With no multiplier (factor 1) the
// amount is unchanged.
func multiplyTappedManaAmount(amount int, untap bool, multiplier int) int {
	if untap || multiplier <= 1 {
		return amount
	}
	return amount * multiplier
}

func appendUniquePermanentManaOutput(outputs []permanentManaOutputResult, candidate permanentManaOutputResult) []permanentManaOutputResult {
	for _, output := range outputs {
		if output.color == candidate.color &&
			output.amount == candidate.amount &&
			output.snow == candidate.snow &&
			output.fromCreature == candidate.fromCreature &&
			output.untap == candidate.untap &&
			output.sacrifice == candidate.sacrifice &&
			output.timing == candidate.timing {
			return outputs
		}
	}
	return append(outputs, candidate)
}

// effectiveBasicLandManaColors returns the color of each basic land type the
// permanent currently has (CR 305.6), in WUBRG order.
func effectiveBasicLandManaColors(s State, permanent *game.Permanent) []mana.Color {
	var colors []mana.Color
	for _, landType := range basicLandTypes {
		if s.PermanentHasSubtype(permanent, landType.subtype) {
			colors = append(colors, landType.color)
		}
	}
	return colors
}

// intrinsicLandManaChoiceOutput recognizes an effective "{T}: Add {U} or {B}."
// ability that states the intrinsic mana abilities of a land's basic land types
// (CR 305.6), such as a dual land's reminder text. It is recognized only while
// the permanent still has every basic land type the choice names, so an ability
// removed by type-setting (CR 305.7) or ability-removing effects stays removed.
func intrinsicLandManaChoiceOutput(s State, permanent *game.Permanent, body *game.ManaAbility) (paymentManaOutput, bool) {
	colors, amount, ok := game.ManaAbilityChoiceOutput(body)
	if !ok || !s.PermanentHasType(permanent, types.Land) {
		return paymentManaOutput{}, false
	}
	landColors := effectiveBasicLandManaColors(s, permanent)
	for _, c := range colors {
		if !slices.Contains(landColors, c) {
			return paymentManaOutput{}, false
		}
	}
	if !reflect.DeepEqual(*body, game.TapManaChoiceAbility(colors...)) {
		return paymentManaOutput{}, false
	}
	return paymentManaOutput{colors: slices.Clone(colors), amount: amount}, true
}

var basicLandTypes = []struct {
	subtype types.Sub
	color   mana.Color
}{
	{subtype: types.Plains, color: mana.W},
	{subtype: types.Island, color: mana.U},
	{subtype: types.Swamp, color: mana.B},
	{subtype: types.Mountain, color: mana.R},
	{subtype: types.Forest, color: mana.G},
}

func paymentManaAbilities(s State, playerID game.PlayerID, permanent *game.Permanent) []paymentManaAbilityResult {
	var results []paymentManaAbilityResult
	for abilityIndex, ability := range s.PermanentEffectiveAbilities(permanent) {
		body, ok := ability.(*game.ManaAbility)
		if !ok {
			continue
		}
		output, ok := paymentManaAbilityOutput(body)
		if !ok {
			output, ok = intrinsicLandManaChoiceOutput(s, permanent, body)
		}
		if !ok {
			continue
		}
		if permanent.Tapped != output.untap {
			continue
		}
		if !summoningSicknessAllowsAbilityCost(s, permanent) {
			continue
		}
		if !s.ActivationConditionSatisfied(playerID, permanent, body.ActivationCondition) {
			continue
		}
		if !s.ManaAbilityTimingAllowed(playerID, permanent, abilityIndex, body.Timing) {
			continue
		}
		results = append(results, paymentManaAbilityResult{
			index:     abilityIndex,
			colors:    output.colors,
			amount:    output.amount,
			untap:     output.untap,
			sacrifice: output.sacrifice,
			timing:    body.Timing,
		})
	}
	return results
}

// IsAutomaticManaAbility reports whether body should be hidden as a standalone
// strategic action because the planner can activate its fixed output on demand.
// Choice-bearing Treasure abilities remain standalone actions even though their
// exact safe shape is also available to the planner during payment.
func IsAutomaticManaAbility(body *game.ManaAbility) bool {
	_, ok := automaticManaAbilityTapState(body)
	return ok
}

func automaticManaAbilityTapState(body *game.ManaAbility) (untap, ok bool) {
	if body == nil || body.ManaCost.Exists {
		return false, false
	}
	untap, ok = simpleManaAbilityTapState(body.AdditionalCosts)
	if !ok {
		return false, false
	}
	addMana, ok := simpleAddMana(body)
	if !ok || addMana.EntryChoiceFrom != "" || !slices.Contains(paymentColors, addMana.ManaColor) {
		return false, false
	}
	// Spend-linked mana carries strategic provenance (a later trigger,
	// restriction, or spell rule effect). Automatic activation would add
	// untagged pool mana, so rider-bearing abilities stay manual agent choices.
	if addMana.SpendRider.Exists {
		return false, false
	}
	return untap, true
}

func paymentManaAbilityOutput(body *game.ManaAbility) (paymentManaOutput, bool) {
	if untap, ok := automaticManaAbilityTapState(body); ok {
		addMana, _ := simpleAddMana(body)
		return paymentManaOutput{
			colors: []mana.Color{addMana.ManaColor},
			amount: max(addMana.Amount.Value(), 1),
			untap:  untap,
		}, true
	}
	if colors, amount, ok := game.ManaAbilityChoiceOutput(body); ok &&
		game.IsTapAnyColorManaAbility(body) {
		return paymentManaOutput{
			colors: colors,
			amount: amount,
		}, true
	}
	colors, amount, ok := sacrificeManaChoiceOutput(body)
	return paymentManaOutput{
		colors:    colors,
		amount:    amount,
		sacrifice: ok,
	}, ok
}

func sacrificeManaChoiceOutput(body *game.ManaAbility) ([]mana.Color, int, bool) {
	if body == nil || body.ManaCost.Exists || !tapAndSacrificeSourceCosts(body.AdditionalCosts) ||
		len(game.BodyTargets(body)) != 0 || len(body.Content.Modes) == 0 || body.Content.IsModal() {
		return nil, 0, false
	}
	sequence := body.Content.Modes[0].Sequence
	if len(sequence) != 2 || sequence[0].Primitive == nil || sequence[1].Primitive == nil {
		return nil, 0, false
	}
	if !unconditionalPaymentInstruction(&sequence[0]) || !unconditionalPaymentInstruction(&sequence[1]) {
		return nil, 0, false
	}
	choose, ok := sequence[0].Primitive.(game.Choose)
	if !ok || choose.Choice.Kind != game.ResolutionChoiceMana ||
		choose.PublishChoice == "" || choose.Choice.UsePlayer ||
		choose.Choice.Player != 0 ||
		choose.Choice.ColorSource != 0 || len(choose.Choice.Colors) == 0 ||
		len(choose.Choice.CardTypes) != 0 || choose.Choice.PlayerRelation != 0 ||
		choose.Choice.Zone != zone.None || choose.Choice.SubtypeOfType != "" ||
		choose.Choice.EntryChoiceKey != "" || choose.Choice.IncludeColorless {
		return nil, 0, false
	}
	addMana, ok := sequence[1].Primitive.(game.AddMana)
	if !ok || addMana.Amount.IsDynamic() || addMana.ChoiceFrom != choose.PublishChoice ||
		addMana.ManaColor != "" || addMana.EntryChoiceFrom != "" || addMana.SpendRider.Exists {
		return nil, 0, false
	}
	colors := slices.Clone(choose.Choice.Colors)
	seen := make(map[mana.Color]bool, len(colors))
	if slices.ContainsFunc(colors, func(color mana.Color) bool {
		if !slices.Contains(paymentColors, color) || seen[color] {
			return true
		}
		seen[color] = true
		return false
	}) {
		return nil, 0, false
	}
	amount := max(addMana.Amount.Value(), 1)
	return colors, amount, true
}

func unconditionalPaymentInstruction(instruction *game.Instruction) bool {
	return !instruction.Condition.Exists &&
		instruction.ConditionGate == "" &&
		instruction.PublishCondition == "" &&
		!instruction.CardCondition.Exists &&
		!instruction.ResultGate.Exists &&
		!instruction.Optional &&
		!instruction.OptionalActor.Exists &&
		instruction.PublishResult == ""
}

func tapAndSacrificeSourceCosts(costs []cost.Additional) bool {
	if len(costs) != 2 {
		return false
	}
	tap := false
	sacrifice := false
	for _, additional := range costs {
		switch additional.Kind {
		case cost.AdditionalTap:
			tapCost := additional
			tapCost.Text = ""
			if tap || tapCost != (cost.Additional{Kind: cost.AdditionalTap}) {
				return false
			}
			tap = true
		case cost.AdditionalSacrificeSource:
			sacrificeCost := additional
			sacrificeCost.Text = ""
			if sacrificeCost.Amount == 0 {
				sacrificeCost.Amount = 1
			}
			want := cost.Additional{
				Kind:               cost.AdditionalSacrificeSource,
				Amount:             1,
				MatchPermanentType: true,
				PermanentType:      types.Artifact,
			}
			self := cost.Additional{Kind: cost.AdditionalSacrificeSource, Amount: 1}
			if sacrifice || sacrificeCost != want && sacrificeCost != self {
				return false
			}
			sacrifice = true
		default:
			return false
		}
	}
	return tap && sacrifice
}

func simpleManaAbilityTapState(costs []cost.Additional) (untap, ok bool) {
	if len(costs) != 1 {
		return false, false
	}
	switch costs[0].Kind {
	case cost.AdditionalTap:
		return false, true
	case cost.AdditionalUntap:
		return true, true
	default:
		return false, false
	}
}

func simpleAddMana(body *game.ManaAbility) (game.AddMana, bool) {
	if len(body.Content.Modes) == 0 || body.Content.IsModal() {
		return game.AddMana{}, false
	}
	sequence := body.Content.Modes[0].Sequence
	if len(sequence) != 1 || sequence[0].Primitive == nil ||
		sequence[0].Primitive.Kind() != game.PrimitiveAddMana {
		return game.AddMana{}, false
	}
	addMana, ok := sequence[0].Primitive.(game.AddMana)
	return addMana, ok && !addMana.Amount.IsDynamic() && addMana.ChoiceFrom == ""
}

func convokeCandidates(s State, playerID game.PlayerID, exclude map[id.ID]bool) []*game.Permanent {
	var nonMana []*game.Permanent
	var manaCreatures []*game.Permanent
	for _, permanent := range s.Battlefield() {
		if !canConvokeWith(s, playerID, permanent, exclude) {
			continue
		}
		if _, ok := permanentManaOutput(s, permanent); ok {
			manaCreatures = append(manaCreatures, permanent)
			continue
		}
		nonMana = append(nonMana, permanent)
	}
	return append(nonMana, manaCreatures...)
}

func improviseCandidates(s State, playerID game.PlayerID, exclude map[id.ID]bool) []*game.Permanent {
	var nonMana []*game.Permanent
	var manaArtifacts []*game.Permanent
	for _, permanent := range s.Battlefield() {
		if !canImproviseWith(s, playerID, permanent, exclude) {
			continue
		}
		if _, ok := permanentManaOutput(s, permanent); ok {
			manaArtifacts = append(manaArtifacts, permanent)
			continue
		}
		nonMana = append(nonMana, permanent)
	}
	return append(nonMana, manaArtifacts...)
}

// availableManaSources groups sources by color. Callers must consume it through
// paymentColors or explicit symbol colors, never by ranging over the map, so
// payment ordering remains deterministic.
func availableManaSources(s State, playerID game.PlayerID, restrictions manaSourceRestrictions) map[mana.Color][]manaSource {
	available := make(map[mana.Color][]manaSource)
	for _, permanent := range s.Battlefield() {
		if s.EffectiveController(permanent) != playerID || restrictions.excluded[permanent.ObjectID] {
			continue
		}
		outputs := permanentManaOutputs(s, permanent)
		for _, output := range outputs {
			if permanent.Tapped != output.untap ||
				restrictions.nonSacrificingOnly[permanent.ObjectID] && output.sacrifice {
				continue
			}
			available[output.color] = append(available[output.color], manaSource{
				permanent:    permanent,
				color:        output.color,
				amount:       output.amount,
				snow:         output.snow,
				fromCreature: output.fromCreature,
				untap:        output.untap,
				sacrifice:    output.sacrifice,
				abilityIndex: output.abilityIndex,
				timing:       output.timing,
				flexibility:  len(outputs),
			})
		}
	}
	return available
}
