package payment

import (
	"slices"
	"strings"

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
	if c, ok := basicLandManaColor(s, permanent); ok {
		return []permanentManaOutputResult{{
			color:        c,
			amount:       1,
			snow:         s.PermanentHasSupertype(permanent, types.Snow),
			abilityIndex: -1,
		}}
	}
	controller := s.EffectiveController(permanent)
	abilities := paymentManaAbilities(s, controller, permanent)
	var outputs []permanentManaOutputResult
	for _, ability := range abilities {
		for _, color := range ability.colors {
			outputs = append(outputs, permanentManaOutputResult{
				color:        color,
				amount:       ability.amount,
				snow:         s.PermanentHasSupertype(permanent, types.Snow),
				untap:        ability.untap,
				sacrifice:    ability.sacrifice,
				abilityIndex: ability.index,
				timing:       ability.timing,
			})
		}
	}
	return outputs
}

func basicLandManaColor(s State, permanent *game.Permanent) (mana.Color, bool) {
	card, ok := s.PermanentCardDef(permanent)
	if !ok || !card.HasType(types.Land) {
		return "", false
	}
	for _, landType := range basicLandTypes {
		if card.HasSubtype(landType.subtype) || strings.EqualFold(card.Name, string(landType.subtype)) {
			return landType.color, true
		}
	}
	return "", false
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
			continue
		}
		if permanent.Tapped != output.untap {
			continue
		}
		if s.PermanentHasType(permanent, types.Creature) && permanent.SummoningSick {
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
	// A mana-spend rider attaches a one-shot delayed trigger to the produced
	// mana, which is a strategic consequence (it can later scry). Automatic
	// activation adds untagged pool mana and would silently drop the rider, so
	// rider-bearing abilities stay manual agent choices where activation tags the
	// mana with its rider (CR 106.12).
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
	if !unconditionalPaymentInstruction(sequence[0]) || !unconditionalPaymentInstruction(sequence[1]) {
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

func unconditionalPaymentInstruction(instruction game.Instruction) bool {
	return !instruction.Condition.Exists &&
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
			if sacrifice || sacrificeCost != want {
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

func delveCandidates(s State, playerID game.PlayerID, manaCost *cost.Mana, xValue int, sourceCardID id.ID, sourceZone zone.Type) ([]id.ID, int, bool) {
	_, generic, ok := costRequirements(manaCost, xValue)
	if !ok || generic <= 0 {
		return nil, 0, false
	}
	player, ok := s.Player(playerID)
	if !ok {
		return nil, 0, false
	}
	var exiles []id.ID
	for _, cardID := range player.Graveyard.All() {
		if len(exiles) == generic {
			break
		}
		if sourceZone == zone.Graveyard && cardID == sourceCardID {
			continue
		}
		exiles = append(exiles, cardID)
	}
	if len(exiles) == 0 {
		return nil, 0, false
	}
	return exiles, generic, true
}

func convokePayment(s State, playerID game.PlayerID, manaCost *cost.Mana, xValue int, exclude map[id.ID]bool) ([]*game.Permanent, *cost.Mana, bool) {
	_, generic, ok := costRequirements(manaCost, xValue)
	if !ok {
		return nil, manaCost, false
	}
	candidates := convokeCandidates(s, playerID, exclude)
	paidColored := make(map[int]bool)
	var taps []*game.Permanent
	used := make(map[id.ID]bool)
	if manaCost != nil {
		for symbolIndex, symbol := range *manaCost {
			for _, color := range symbol.Colors() {
				permanent, ok := chooseConvokeColoredCreature(s, candidates, used, color)
				if !ok {
					continue
				}
				taps = append(taps, permanent)
				used[permanent.ObjectID] = true
				paidColored[symbolIndex] = true
				break
			}
		}
	}
	genericReduction := 0
	for _, permanent := range candidates {
		if genericReduction == generic {
			break
		}
		if used[permanent.ObjectID] {
			continue
		}
		taps = append(taps, permanent)
		used[permanent.ObjectID] = true
		genericReduction++
	}
	if len(taps) == 0 {
		return nil, manaCost, false
	}
	return taps, costWithConvokePayments(manaCost, genericReduction, paidColored), true
}

func chooseConvokeColoredCreature(s State, candidates []*game.Permanent, used map[id.ID]bool, m mana.Color) (*game.Permanent, bool) {
	if m == mana.C {
		// can't pay for colorless pips via convoke.
		return nil, false
	}
	for _, permanent := range candidates {
		if used[permanent.ObjectID] {
			continue
		}
		for _, c := range s.PermanentEffectiveColors(permanent) {
			if cost.ManaForColor(c) == m {
				return permanent, true
			}
		}
	}
	return nil, false
}

func costWithConvokePayments(manaCost *cost.Mana, genericReduction int, paidColored map[int]bool) *cost.Mana {
	generic := max(genericCostAmount(manaCost)-genericReduction, 0)
	var modified cost.Mana
	if generic > 0 {
		modified = append(modified, cost.O(generic))
	}
	if manaCost != nil {
		for i, symbol := range *manaCost {
			if symbol.Kind == cost.GenericSymbol || paidColored[i] {
				continue
			}
			modified = append(modified, symbol)
		}
	}
	return &modified
}

func costWithGenericRequirement(manaCost *cost.Mana, generic int) *cost.Mana {
	if generic < 0 {
		generic = 0
	}
	var modified cost.Mana
	if generic > 0 {
		modified = append(modified, cost.O(generic))
	}
	if manaCost != nil {
		for _, symbol := range *manaCost {
			if symbol.Kind == cost.GenericSymbol || symbol.Kind == cost.VariableSymbol {
				continue
			}
			modified = append(modified, symbol)
		}
	}
	return &modified
}

// availableManaSources groups sources by color. Callers must consume it through
// paymentColors or explicit symbol colors, never by ranging over the map, so
// payment ordering remains deterministic.
func availableManaSources(s State, playerID game.PlayerID, exclude map[id.ID]bool) map[mana.Color][]manaSource {
	available := make(map[mana.Color][]manaSource)
	for _, permanent := range s.Battlefield() {
		if s.EffectiveController(permanent) != playerID || exclude[permanent.ObjectID] {
			continue
		}
		for _, output := range permanentManaOutputs(s, permanent) {
			if permanent.Tapped != output.untap {
				continue
			}
			available[output.color] = append(available[output.color], manaSource{
				permanent:    permanent,
				color:        output.color,
				amount:       output.amount,
				snow:         output.snow,
				untap:        output.untap,
				sacrifice:    output.sacrifice,
				abilityIndex: output.abilityIndex,
				timing:       output.timing,
			})
		}
	}
	return available
}
