package payment

import (
	"reflect"
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ManaAbilityClass distinguishes the CR 605.1a boundary from the planner's
// narrower modeling boundary. Opaque abilities remain real mana abilities.
type ManaAbilityClass uint8

const (
	NotManaAbility ManaAbilityClass = iota
	PlannableManaAbility
	OpaqueManaAbility
)

// ActivationSchema is an inspectable deterministic mana ability. Definition
// references are immutable; per-game sources and choices live in its plan node.
type ActivationSchema struct {
	body       *game.ManaAbility
	production []game.AddMana
	choices    []game.Choose
	effects    []game.Primitive
	tap        bool
	untap      bool
	sacrifice  bool
	exile      bool
	automatic  bool
}

// Automatic reports whether activation has no strategic choices or side effects
// beyond tapping/untapping its source. Other schemas remain explicit choices.
func (s ActivationSchema) Automatic() bool { return s.automatic }

// ResourceManaCondition is the condition boundary for resource-only forecasts.
// Battlefield selections and resource predicates use the canonical evaluators;
// zone-, resolution- and other event-dependent shapes remain opaque.
func ResourceManaCondition(condition opt.V[game.Condition]) bool {
	if !condition.Exists {
		return true
	}
	remaining := condition.Val
	remaining.ControlsMatching = opt.V[game.SelectionCount]{}
	remaining.AnyPlayerLifeAtMost = 0
	remaining.SourceCounterKind = 0
	remaining.SourceCounterKindKnown = false
	remaining.SourceCountersAtLeast = 0
	remaining.SourceLevelCountersAtLeast = 0
	remaining.SourceLevelCountersLessThan = 0
	remaining.SourceControllerTurn = false
	return reflect.DeepEqual(remaining, game.Condition{})
}

// ClassifyManaAbility fails closed on unmodeled costs, effects and choices.
// Library-moving abilities are not opaque: they are not mana abilities at all.
func ClassifyManaAbility(body *game.ManaAbility) (ActivationSchema, ManaAbilityClass) {
	if body == nil || len(game.BodyTargets(body)) != 0 || game.ManaAbilityLibraryMovement(body) != "" {
		return ActivationSchema{}, NotManaAbility
	}
	if len(body.Content.Modes) != 1 || body.Content.IsModal() ||
		body.Content.Modes[0].Program.ID != "" || body.Content.Modes[0].Program.Version != "" ||
		costHasVariableMana(manaCostPtr(body.ManaCost)) || !ResourceManaCondition(body.ActivationCondition) {
		return ActivationSchema{}, OpaqueManaAbility
	}
	schema := ActivationSchema{body: body}
	for _, additional := range body.AdditionalCosts {
		if additional.Random || additional.ChoiceGroup != 0 || additional.AmountFromX || additional.AmountDynamic != cost.AdditionalDynamicAmountNone {
			return ActivationSchema{}, OpaqueManaAbility
		}
		switch additional.Kind {
		case cost.AdditionalTap:
			if schema.tap {
				return ActivationSchema{}, OpaqueManaAbility
			}
			schema.tap = true
		case cost.AdditionalUntap:
			if schema.untap {
				return ActivationSchema{}, OpaqueManaAbility
			}
			schema.untap = true
		case cost.AdditionalSacrificeSource:
			schema.sacrifice = true
		case cost.AdditionalExileSource:
			schema.exile = true
		case cost.AdditionalPayLife, cost.AdditionalEnergy, cost.AdditionalRemoveCounter, cost.AdditionalPutCounter, cost.AdditionalSacrifice:
		case cost.AdditionalDiscard:
			if additional.Amount != 1 {
				return ActivationSchema{}, OpaqueManaAbility
			}
		case cost.AdditionalExile:
			if additional.Amount != 1 || additional.Source != zone.Hand && additional.Source != zone.Graveyard {
				return ActivationSchema{}, OpaqueManaAbility
			}
		default:
			return ActivationSchema{}, OpaqueManaAbility
		}
	}
	if schema.tap && schema.untap {
		return ActivationSchema{}, OpaqueManaAbility
	}
	keys := make(map[game.ChoiceKey]bool)
	for _, instruction := range body.Content.Modes[0].Sequence {
		if !unconditionalPaymentInstruction(&instruction) || instruction.Primitive == nil ||
			len(instruction.TemptingOfferBody) != 0 {
			return ActivationSchema{}, OpaqueManaAbility
		}
		if choice, ok := instruction.Primitive.(game.Choose); ok {
			if choice.Choice.Kind != game.ResolutionChoiceMana || choice.PublishChoice == "" ||
				keys[choice.PublishChoice] || choice.Choice.UsePlayer || choice.Choice.PlayerReference != nil || choice.Choice.Zone != zone.None ||
				choice.Choice.ColorSource != game.ResolutionChoiceColorSourceStatic && choice.Choice.ColorSource != game.ResolutionChoiceColorSourceCommanderIdentity && choice.Choice.ColorSource != game.ResolutionChoiceColorSourceFixedOrEntryChosen || choice.Choice.AtRandom {
				return ActivationSchema{}, OpaqueManaAbility
			}
			keys[choice.PublishChoice] = true
			if choice.Choice.ColorSource == game.ResolutionChoiceColorSourceStatic {
				if len(choice.Choice.Colors) == 0 || len(choice.Choice.Colors) > 6 {
					return ActivationSchema{}, OpaqueManaAbility
				}
				for index, color := range choice.Choice.Colors {
					if !slices.Contains(paymentColors, color) || slices.Contains(choice.Choice.Colors[:index], color) {
						return ActivationSchema{}, OpaqueManaAbility
					}
				}
			}
			schema.choices = append(schema.choices, choice)
			continue
		}
		if add, ok := instruction.Primitive.(game.AddMana); ok {
			if !ResourceManaAmount(add.Amount) || len(add.CombinationColors) != 0 ||
				add.Player.Exists && add.Player.Val.Kind() != game.PlayerReferenceController || add.EachControlledColor != nil ||
				add.ChoiceFrom != "" && !keys[add.ChoiceFrom] {
				return ActivationSchema{}, OpaqueManaAbility
			}
			if add.ManaColor == "" && add.ChoiceFrom == "" && add.EntryChoiceFrom == "" {
				return ActivationSchema{}, OpaqueManaAbility
			}
			if add.ManaColor != "" && !slices.Contains(paymentColors, add.ManaColor) {
				return ActivationSchema{}, OpaqueManaAbility
			}
			schema.production = append(schema.production, add)
			continue
		}
		if damage, ok := instruction.Primitive.(game.Damage); ok {
			player, self := damage.Recipient.PlayerReference()
			if !self || player.Kind() != game.PlayerReferenceController ||
				!damage.DamageSource.Exists || damage.DamageSource.Val.Kind() != game.ObjectReferenceSourcePermanent ||
				damage.Amount.IsDynamic() || damage.Amount.Value() < 0 || damage.Divided || damage.EachTarget || damage.CantBePrevented || damage.CreatureDamageCantBePreventedOrRedirected {
				return ActivationSchema{}, OpaqueManaAbility
			}
			schema.effects = append(schema.effects, damage)
			continue
		}
		if gain, ok := instruction.Primitive.(game.GainLife); ok {
			if gain.Player.Kind() != game.PlayerReferenceController || gain.PlayerGroup.Kind != game.PlayerGroupReferenceNone || gain.Amount.IsDynamic() || gain.Amount.Value() < 0 {
				return ActivationSchema{}, OpaqueManaAbility
			}
			schema.effects = append(schema.effects, gain)
			continue
		}
		if loss, ok := instruction.Primitive.(game.LoseLife); ok {
			if loss.Player.Kind() != game.PlayerReferenceController || loss.PlayerGroup.Kind != game.PlayerGroupReferenceNone || loss.Amount.IsDynamic() || loss.Amount.Value() < 0 {
				return ActivationSchema{}, OpaqueManaAbility
			}
			schema.effects = append(schema.effects, loss)
			continue
		}
		if add, ok := instruction.Primitive.(game.AddCounter); ok {
			if add.Object.Kind() != game.ObjectReferenceSourcePermanent || add.Amount.IsDynamic() || add.Amount.Value() < 0 ||
				add.Group.Valid() || add.PublishLinked != "" || add.AllKinds || add.ChooseOne || add.Distribute || add.DoubleKind || len(add.KindChoices) != 0 {
				return ActivationSchema{}, OpaqueManaAbility
			}
			schema.effects = append(schema.effects, add)
			continue
		}
		if remove, ok := instruction.Primitive.(game.RemoveCounter); ok {
			if remove.Object.Kind() != game.ObjectReferenceSourcePermanent || remove.Amount.IsDynamic() || remove.Amount.Value() < 0 ||
				remove.Group.Valid() || remove.ChooseKind || remove.AllKinds || remove.AllOfKind {
				return ActivationSchema{}, OpaqueManaAbility
			}
			schema.effects = append(schema.effects, remove)
			continue
		}
		return ActivationSchema{}, OpaqueManaAbility
	}
	if len(schema.production) == 0 {
		return ActivationSchema{}, NotManaAbility
	}
	if !schema.tap && !schema.untap && !schema.sacrifice && !schema.exile && !body.ManaCost.Exists &&
		!slices.ContainsFunc(body.AdditionalCosts, func(c cost.Additional) bool {
			return c.Kind == cost.AdditionalPayLife || c.Kind == cost.AdditionalRemoveCounter || c.Kind == cost.AdditionalEnergy ||
				c.Kind == cost.AdditionalSacrifice || c.Kind == cost.AdditionalDiscard || c.Kind == cost.AdditionalExile
		}) {
		return ActivationSchema{}, OpaqueManaAbility
	}
	schema.automatic = (schema.tap || schema.untap) && !schema.sacrifice && !schema.exile &&
		!body.ManaCost.Exists && len(body.AdditionalCosts) == 1 && len(schema.choices) == 0 &&
		len(schema.effects) == 0 && len(schema.production) == 1 && !schema.production[0].SpendRider.Exists &&
		!schema.production[0].Amount.IsDynamic() && schema.production[0].EntryChoiceFrom == "" &&
		len(schema.production[0].CombinationColors) == 0
	return schema, PlannableManaAbility
}

// ResourceManaAmount recognizes constant mana and public graveyard counts.
// A public graveyard count is deterministic, including after the cast source
// leaves that zone. The adapter must reject a projection that could change
// its cards through an unmodeled zone replacement.
func ResourceManaAmount(amount game.Quantity) bool {
	if !amount.IsDynamic() {
		return amount.Value() >= 0
	}
	dynamic := amount.DynamicAmount().Val
	if dynamic.Kind != game.DynamicAmountCountCardsInZone || dynamic.CardZone != zone.Graveyard ||
		dynamic.Player == nil || dynamic.Player.Kind() != game.PlayerReferenceController {
		return false
	}
	dynamic.Kind, dynamic.CardZone, dynamic.Player = 0, zone.None, nil
	dynamic.Multiplier, dynamic.Addend = 0, 0
	dynamic.Selection = nil
	return reflect.DeepEqual(dynamic, game.DynamicAmount{})
}
