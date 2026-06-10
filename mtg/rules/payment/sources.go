package payment

import (
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
	abilityIndex int
	timing       game.TimingRestriction
}

type simpleManaAbilityResult struct {
	index int
	body  *game.ManaAbility
	untap bool
}

// permanentManaOutput derives the mana output of a permanent by checking
// basic land types and simple tap mana abilities.
func permanentManaOutput(s State, permanent *game.Permanent) (permanentManaOutputResult, bool) {
	if c, ok := basicLandManaColor(s, permanent); ok {
		return permanentManaOutputResult{
			color:        c,
			amount:       1,
			snow:         s.PermanentHasSupertype(permanent, types.Snow),
			abilityIndex: -1,
		}, true
	}
	controller := s.EffectiveController(permanent)
	ability, ok := simpleManaAbility(s, controller, permanent)
	if !ok {
		return permanentManaOutputResult{}, false
	}
	addMana, ok := simpleAddMana(ability.body)
	if !ok {
		return permanentManaOutputResult{}, false
	}
	amount := addMana.Amount.Value()
	if amount <= 0 {
		amount = 1
	}
	return permanentManaOutputResult{
		color:        addMana.ManaColor,
		amount:       amount,
		snow:         s.PermanentHasSupertype(permanent, types.Snow),
		untap:        ability.untap,
		abilityIndex: ability.index,
		timing:       ability.body.Timing,
	}, true
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

func simpleManaAbility(s State, playerID game.PlayerID, permanent *game.Permanent) (simpleManaAbilityResult, bool) {
	card, ok := s.PermanentCardDef(permanent)
	if !ok {
		return simpleManaAbilityResult{}, false
	}
	face := &card.CardFace
	offset := 0
	if face.SpellAbility.Exists {
		offset++
	}
	offset += len(face.ActivatedAbilities)
	for i := range face.ManaAbilities {
		body := &face.ManaAbilities[i]
		untap, ok := simpleManaAbilityTapState(body.AdditionalCosts)
		if !ok || body.ManaCost.Exists || !isSimpleAddMana(body) {
			continue
		}
		if permanent.Tapped != untap {
			continue
		}
		if s.PermanentHasType(permanent, types.Creature) && permanent.SummoningSick {
			continue
		}
		if !s.ActivationConditionSatisfied(playerID, permanent, body.ActivationCondition) {
			continue
		}
		abilityIndex := offset + i
		if !s.ManaAbilityTimingAllowed(playerID, permanent, abilityIndex, body.Timing) {
			continue
		}
		return simpleManaAbilityResult{
			index: abilityIndex,
			body:  body,
			untap: untap,
		}, true
	}
	return simpleManaAbilityResult{}, false
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

func isSimpleAddMana(body *game.ManaAbility) bool {
	_, ok := simpleAddMana(body)
	return ok
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
		output, ok := permanentManaOutput(s, permanent)
		if !ok || permanent.Tapped != output.untap {
			continue
		}
		available[output.color] = append(available[output.color], manaSource{
			permanent:    permanent,
			color:        output.color,
			amount:       output.amount,
			snow:         output.snow,
			untap:        output.untap,
			abilityIndex: output.abilityIndex,
			timing:       output.timing,
		})
	}
	return available
}
