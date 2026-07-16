package eval

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// ScorableAbilityOf reduces a sealed ability body to its scorable cost and
// effect terms. For a modal ability it unions every mode (an over-approximation);
// callers that know which modes were chosen should use ScorableAbilityOfModes.
func ScorableAbilityOf(body game.Ability) ScorableAbility {
	return ScorableAbilityOfModes(body, nil)
}

// ScorableAbilityOfModes reduces a sealed ability body to its scorable cost and
// effect terms, scoring only the chosen modes. An empty chosenModes scores the
// sole mode of a non-modal ability, or unions all modes of a modal ability whose
// choice is unknown.
func ScorableAbilityOfModes(body game.Ability, chosenModes []int) ScorableAbility {
	return ScorableAbility{
		Costs:  game.BodyAdditionalCosts(body),
		Effect: ScorableEffectModes(game.BodyContent(body), chosenModes),
	}
}

// ScorableEffect summarizes an ability's effect content as value-relevant atoms,
// unioning the atoms of every mode. It is the mode-unaware form; see
// ScorableEffectModes to score only the chosen modes of a modal ability.
func ScorableEffect(content game.AbilityContent) []EffectAtom {
	return ScorableEffectModes(content, nil)
}

// ScorableEffectModes summarizes an ability's effect content as value-relevant
// atoms, scoring only the modes the controller chose. When chosenModes is empty
// it scores the sole mode of non-modal content, and unions all modes of modal
// content whose choice is unknown (an over-approximation). Atoms produced by a
// conditional or optional instruction are marked dynamic, so a scorer treats
// their magnitude as uncertain rather than trusting an effect that may not
// happen. Primitives the translator does not model contribute no atom, which a
// scorer reads as value-neutral.
func ScorableEffectModes(content game.AbilityContent, chosenModes []int) []EffectAtom {
	var atoms []EffectAtom
	for _, m := range scorableModeIndices(content, chosenModes) {
		mode := content.Modes[m]
		for i := range mode.Sequence {
			instruction := mode.Sequence[i]
			before := len(atoms)
			atoms = appendPrimitiveAtoms(atoms, instruction.Primitive)
			if instructionUncertain(instruction) {
				for j := before; j < len(atoms); j++ {
					atoms[j].IsDynamic = true
				}
			}
		}
	}
	return atoms
}

// scorableModeIndices returns the mode indices to score. Chosen modes win;
// otherwise a single-mode (non-modal) ability scores its one mode, and a modal
// ability with no known choice falls back to every mode.
func scorableModeIndices(content game.AbilityContent, chosenModes []int) []int {
	if len(chosenModes) > 0 {
		valid := make([]int, 0, len(chosenModes))
		for _, m := range chosenModes {
			if m >= 0 && m < len(content.Modes) {
				valid = append(valid, m)
			}
		}
		return valid
	}
	indices := make([]int, len(content.Modes))
	for m := range content.Modes {
		indices[m] = m
	}
	return indices
}

// instructionUncertain reports whether an instruction's effect may not happen or
// happen in an amount the static body does not fix: a gating condition, a
// referenced-card condition, an "if you do/don't" result gate, or an optional
// instruction the controller may decline.
func instructionUncertain(instruction game.Instruction) bool {
	return instruction.Condition.Exists ||
		instruction.CardCondition.Exists ||
		instruction.ResultGate.Exists ||
		instruction.Optional
}

// appendPrimitiveAtoms is the single place that maps the engine's resolution
// primitives to value atoms. It is intentionally the only switch over the
// primitive surface, so the ~100 primitives never leak into strategy or report
// code. Unmodeled primitives append nothing (value-neutral).
func appendPrimitiveAtoms(atoms []EffectAtom, primitive game.Primitive) []EffectAtom {
	switch p := primitive.(type) {
	case game.Draw:
		return append(atoms, quantityAtom(EffectCardsDrawn, p.Amount, affectedFromPlayer(p.Player)))
	case game.Discard:
		atom := quantityAtom(EffectCardsLost, p.Amount, affectedFromPlayer(p.Player))
		if p.EntireHand {
			atom.IsDynamic = true
		}
		return append(atoms, atom)
	case game.Mill:
		return append(atoms, quantityAtom(EffectCardsLost, p.Amount, affectedFromPlayer(p.Player)))
	case game.DiscardThenDraw:
		if p.DrawOffset == 0 {
			return atoms
		}
		return append(atoms, EffectAtom{Kind: EffectCardsDrawn, Amount: p.DrawOffset, Affected: affectedFromPlayer(p.Player)})
	case game.DiscardUnlessType:
		return append(atoms, EffectAtom{Kind: EffectCardsLost, Amount: p.Amount, Affected: affectedFromPlayer(p.Player)})
	case game.GainLife:
		return append(atoms, quantityAtom(EffectLifeGained, p.Amount, affectedFromPlayer(p.Player)))
	case game.LoseLife:
		return append(atoms, quantityAtom(EffectLifeLost, p.Amount, affectedFromPlayer(p.Player)))
	case game.ExchangeLifeTotalWithSourceCharacteristic:
		// The exchange can gain or lose life and raises or lowers a source
		// characteristic depending on runtime values, so it has no stable atom.
		return atoms
	case game.Damage:
		return append(atoms, quantityAtom(EffectDamageDealt, p.Amount, AffectedTarget))
	case game.Destroy:
		return append(atoms, EffectAtom{Kind: EffectPermanentRemoved, Affected: AffectedTarget})
	case game.Exile:
		return append(atoms, EffectAtom{Kind: EffectPermanentRemoved, Affected: AffectedTarget})
	case game.ExilePermanentForPlay:
		return append(atoms, EffectAtom{Kind: EffectPermanentRemoved, Affected: AffectedTarget})
	case game.ChampionExile:
		return append(atoms, EffectAtom{Kind: EffectPermanentRemoved, Affected: AffectedYou})
	case game.Bounce:
		return append(atoms, EffectAtom{Kind: EffectPermanentRemoved, Affected: AffectedTarget})
	case game.RemoveTargetsForToken:
		return append(atoms, EffectAtom{Kind: EffectPermanentRemoved, Affected: AffectedTarget})
	case game.Tap:
		return append(atoms, EffectAtom{Kind: EffectPermanentTapped, Affected: AffectedTarget})
	case game.Fight:
		// Two creatures fight, each dealing its power to the other (CR 701.12).
		// The value is damaging the opposing creature, so treat it as targeted
		// damage; the target-value scoring resolves it against the fight's targets.
		return append(atoms, EffectAtom{Kind: EffectDamageDealt, Affected: AffectedTarget})
	case game.CorrelatedFight:
		// A correlated fight pairs a created-token group with a counted-permanent
		// group and fights each pair (Ezuri's Predation), damaging an
		// indeterminate number of opposing creatures; value it as dynamic
		// targeted damage, matching the single Fight's targeted-damage atom.
		return append(atoms, EffectAtom{Kind: EffectDamageDealt, IsDynamic: true, Affected: AffectedTarget})
	case game.Monstrosity:
		// Monstrosity puts +1/+1 counters on the controller's own creature and
		// makes it monstrous (CR 701.33) — board-boosting counters.
		return append(atoms, quantityAtom(EffectCounterAdded, p.Amount, AffectedUnknown))
	case game.Bolster:
		// Bolster puts +1/+1 counters on the controller's creature with the least
		// toughness (CR 701.37) — board-boosting counters, matching Monstrosity.
		return append(atoms, quantityAtom(EffectCounterAdded, p.Amount, AffectedUnknown))
	case game.AddMana:
		return append(atoms, quantityAtom(EffectManaAdded, p.Amount, AffectedYou))
	case game.CreateToken:
		return append(atoms, quantityAtom(EffectTokenCreated, p.Amount, AffectedYou))
	case game.AddCounter:
		if p.DoubleKind || p.AllKinds {
			return append(atoms, EffectAtom{Kind: EffectCounterAdded, IsDynamic: true, Affected: AffectedUnknown})
		}
		return append(atoms, quantityAtom(EffectCounterAdded, p.Amount, AffectedUnknown))
	case game.OptionalCounterForEachPlayer:
		return append(atoms, EffectAtom{Kind: EffectCounterAdded, IsDynamic: true, Affected: AffectedUnknown})
	case game.Search:
		if isLandRampSearch(p) {
			return append(atoms, quantityAtom(EffectLandRamp, p.Amount, AffectedYou))
		}
		return append(atoms, quantityAtom(EffectCardTutored, p.Amount, AffectedYou))
	case game.RevealTopPartition:
		// The player puts an indeterminate number of the revealed cards into
		// their hand (every revealed card matching the filter), so the gain is
		// dynamic card advantage for that player.
		return append(atoms, EffectAtom{Kind: EffectCardsDrawn, IsDynamic: true, Affected: affectedFromPlayer(p.Player)})
	case game.ReturnExiledCardsWithCounter:
		// The player returns every exiled card they own bearing the named marker
		// counter to their hand — an indeterminate, accumulating set — so the
		// gain is dynamic card advantage for that player, matching how the other
		// "unknown number of cards into hand" primitives are valued.
		return append(atoms, EffectAtom{Kind: EffectCardsDrawn, IsDynamic: true, Affected: affectedFromPlayer(p.Player)})
	case game.IterativeLibraryProcess:
		// The controller digs through the top of their library and puts a single
		// found card into their hand (Tainted Pact's selection, Demonic
		// Consultation's named tutor), exiling the rest — card advantage of one
		// card for that player, matching how the other "card into hand" library
		// primitives are valued.
		return append(atoms, EffectAtom{Kind: EffectCardsDrawn, Amount: 1, Affected: affectedFromPlayer(p.Player)})
	default:
		return atoms
	}
}

// quantityAtom builds an amount-bearing atom, marking dynamic ({X}, "for each")
// amounts so a scorer estimates rather than trusting the fixed value.
func quantityAtom(kind EffectKind, amount game.Quantity, affected Affected) EffectAtom {
	if amount.IsDynamic() {
		return EffectAtom{Kind: kind, IsDynamic: true, Affected: affected}
	}
	return EffectAtom{Kind: kind, Amount: amount.Value(), Affected: affected}
}

// affectedFromPlayer resolves a player reference to an audience only when it is
// unambiguous (the controller), leaving the rest AffectedUnknown so a scorer
// never infers a wrong value sign.
func affectedFromPlayer(ref game.PlayerReference) Affected {
	if ref.Kind() == game.PlayerReferenceController {
		return AffectedYou
	}
	return AffectedUnknown
}

// isLandRampSearch reports whether a search puts a land onto the battlefield —
// land ramp (Rampant Growth, Cultivate, Farseek) — as opposed to a tutor that
// moves a card to hand or another zone. It matches a battlefield destination
// with a land-typed filter.
func isLandRampSearch(p game.Search) bool {
	return p.Spec.Destination == zone.Battlefield &&
		(slices.Contains(p.Spec.Filter.RequiredTypes, types.Land) ||
			slices.Contains(p.Spec.Filter.RequiredTypesAny, types.Land))
}
