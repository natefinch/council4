package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/opt"
)

// lowerLevelUpAbility lowers a leveler card's "Level up {cost}" activated
// ability (CR 711.2) into a sorcery-speed ActivatedAbility that puts one level
// counter on the source. The cost must be a pure mana cost; an empty cost leaves
// the ability unsupported so the card fails closed.
func lowerLevelUpAbility(
	cardName string,
	ability compiler.CompiledAbility,
) (game.ActivatedAbility, *shared.Diagnostic) {
	if len(ability.LevelUpCost) == 0 {
		return game.ActivatedAbility{}, executableDiagnostic(
			ability,
			"unsupported Level up ability",
			"the executable source backend requires a mana cost on a leveler card's Level up ability",
		)
	}
	manaCost := slices.Clone(ability.LevelUpCost)
	activated := game.ActivatedAbility{
		Text:     ability.Text,
		ManaCost: opt.Val(manaCost),
		Timing:   game.SorceryOnly,
		Content: game.Mode{
			Sequence: []game.Instruction{
				{
					Primitive: game.AddCounter{
						Object:      game.SourcePermanentReference(),
						Amount:      game.Fixed(1),
						CounterKind: counter.Level,
					},
				},
			},
		}.Ability(),
	}
	return activated, nil
}

// lowerLevelBandPowerToughness emits a static ability that sets the source's
// base power and toughness to a leveler band's printed values while the source
// has level counters within the band (CR 711.4). It returns emit=false for a
// band with no printed P/T (a non-creature leveler band), which carries no base
// P/T to set. The static is gated by the band's level-counter condition.
func lowerLevelBandPowerToughness(
	ability compiler.CompiledAbility,
) (loweredStaticAbility, *shared.Diagnostic, bool) {
	band := ability.LevelBand
	if band == nil {
		return loweredStaticAbility{}, executableDiagnostic(
			ability,
			"unsupported level band",
			"the executable source backend could not read a leveler card's LEVEL band",
		), false
	}
	if !band.HasPowerToughness {
		return loweredStaticAbility{}, nil, false
	}
	condition := levelBandCondition(band)
	static := loweredStaticAbility{
		Body: game.StaticAbility{
			Text:      ability.Text,
			Condition: opt.Val(condition),
			ContinuousEffects: []game.ContinuousEffect{
				{
					Layer:          game.LayerPowerToughnessSet,
					AffectedSource: true,
					SetPower:       opt.Val(game.PT{Value: band.Power}),
					SetToughness:   opt.Val(game.PT{Value: band.Toughness}),
				},
			},
		},
	}
	return static, nil, true
}

// levelBandCondition builds the level-counter condition gating a leveler band
// (CR 711.4). AtLeast applies the band's lower bound; a closed band also applies
// an exclusive upper bound (High+1), while the open-ended final band leaves it
// unset.
func levelBandCondition(band *compiler.CompiledLevelBand) game.Condition {
	condition := game.Condition{SourceLevelCountersAtLeast: band.Low}
	if band.High > 0 {
		condition.SourceLevelCountersLessThan = band.High + 1
	}
	return condition
}

// gateLoweredAbilityByLevelBand restricts a lowered leveler-band ability to the
// levels within its band by merging the band's level-counter condition into the
// ability's own condition (CR 711.4). It covers the activated, triggered, and
// static ability shapes leveler bands use and fails closed for any other lowered
// output so an unexpected shape stays unsupported.
func gateLoweredAbilityByLevelBand(
	lowered *abilityLowering,
	ability compiler.CompiledAbility,
	band *compiler.CompiledLevelBand,
) *shared.Diagnostic {
	const unsupportedDetail = "the executable source backend cannot gate this leveler-band ability shape by level"
	condition := levelBandCondition(band)
	gated := false
	if lowered.activatedAbility.Exists {
		activated := lowered.activatedAbility.Val
		activated.ActivationCondition = mergeLevelBandGate(activated.ActivationCondition, condition)
		lowered.activatedAbility = opt.Val(activated)
		gated = true
	}
	if lowered.triggeredAbility.Exists {
		triggered := lowered.triggeredAbility.Val
		triggered.Trigger.InterveningCondition = mergeLevelBandGate(triggered.Trigger.InterveningCondition, condition)
		lowered.triggeredAbility = opt.Val(triggered)
		gated = true
	}
	for i := range lowered.staticAbilities {
		if lowered.staticAbilities[i].VarName != "" {
			return executableDiagnostic(ability, "unsupported Level up ability", unsupportedDetail)
		}
		lowered.staticAbilities[i].Body.Condition = mergeLevelBandGate(lowered.staticAbilities[i].Body.Condition, condition)
		gated = true
	}
	if lowered.manaAbility.Exists ||
		lowered.loyaltyAbility.Exists ||
		lowered.chapterAbility.Exists ||
		lowered.replacementAbility.Exists ||
		lowered.spellAbility.Exists {
		return executableDiagnostic(ability, "unsupported Level up ability", unsupportedDetail)
	}
	if !gated {
		return executableDiagnostic(ability, "unsupported Level up ability", unsupportedDetail)
	}
	return nil
}

// mergeLevelBandGate merges a leveler band's level-counter bounds into a
// condition, preserving any existing condition fields so a banded ability that
// carries its own condition is gated by both its condition and the band.
func mergeLevelBandGate(existing opt.V[game.Condition], band game.Condition) opt.V[game.Condition] {
	merged := existing.Val
	merged.SourceLevelCountersAtLeast = band.SourceLevelCountersAtLeast
	merged.SourceLevelCountersLessThan = band.SourceLevelCountersLessThan
	return opt.Val(merged)
}
