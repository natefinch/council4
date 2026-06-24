package eval

import "github.com/natefinch/council4/mtg/game"

// ScorableAbilityOf reduces a sealed ability body to its scorable cost and
// effect terms.
func ScorableAbilityOf(body game.Ability) ScorableAbility {
	return ScorableAbility{
		Costs:  game.BodyAdditionalCosts(body),
		Effect: ScorableEffect(game.BodyContent(body)),
	}
}

// ScorableEffect summarizes an ability's effect content as value-relevant
// atoms. It unions the atoms of every mode, so a modal effect is summarized by
// the combined consequences of its options; refining modal handling is tracked
// separately. Primitives the translator does not model contribute no atom,
// which a scorer reads as value-neutral.
func ScorableEffect(content game.AbilityContent) []EffectAtom {
	var atoms []EffectAtom
	for m := range content.Modes {
		for i := range content.Modes[m].Sequence {
			atoms = appendPrimitiveAtoms(atoms, content.Modes[m].Sequence[i].Primitive)
		}
	}
	return atoms
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
	case game.GainLife:
		return append(atoms, quantityAtom(EffectLifeGained, p.Amount, affectedFromPlayer(p.Player)))
	case game.LoseLife:
		return append(atoms, quantityAtom(EffectLifeLost, p.Amount, affectedFromPlayer(p.Player)))
	case game.Damage:
		return append(atoms, quantityAtom(EffectDamageDealt, p.Amount, AffectedTarget))
	case game.Destroy:
		return append(atoms, EffectAtom{Kind: EffectPermanentRemoved, Affected: AffectedTarget})
	case game.Exile:
		return append(atoms, EffectAtom{Kind: EffectPermanentRemoved, Affected: AffectedTarget})
	case game.Bounce:
		return append(atoms, EffectAtom{Kind: EffectPermanentRemoved, Affected: AffectedTarget})
	case game.Tap:
		return append(atoms, EffectAtom{Kind: EffectPermanentTapped, Affected: AffectedTarget})
	case game.AddMana:
		return append(atoms, quantityAtom(EffectManaAdded, p.Amount, AffectedYou))
	case game.CreateToken:
		return append(atoms, quantityAtom(EffectTokenCreated, p.Amount, AffectedYou))
	case game.AddCounter:
		return append(atoms, quantityAtom(EffectCounterAdded, p.Amount, AffectedUnknown))
	case game.Search:
		return append(atoms, quantityAtom(EffectCardTutored, p.Amount, AffectedYou))
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
