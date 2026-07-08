package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// gateStaticOnMaxSpeed gates every static ability produced by lowering a "Max
// speed" static (CR 702.179, the Start your engines! speed subsystem) on
// game.Condition.ControllerHasMaxSpeed, so the static is active only while its
// controller has maximum speed (speed 4). The runtime already evaluates that
// condition. The caller lowers the ability with its "Max speed" ability word
// stripped (so the ordinary static paths accept the body); this then wraps the
// result. A body that already carries its own condition, or a lowering that
// produced anything other than static abilities, fails closed rather than
// silently dropping the max-speed gate.
func gateStaticOnMaxSpeed(lowered abilityLowering, ability compiler.CompiledAbility) (abilityLowering, *shared.Diagnostic) {
	unsupported := func() (abilityLowering, *shared.Diagnostic) {
		return abilityLowering{}, executableDiagnostic(
			ability,
			"unsupported Max speed ability",
			"the executable source backend supports only a Max speed static ability with no other condition",
		)
	}
	if len(lowered.staticAbilities) == 0 ||
		lowered.activatedAbility.Exists ||
		lowered.manaAbility.Exists ||
		lowered.loyaltyAbility.Exists ||
		lowered.triggeredAbility.Exists ||
		lowered.chapterAbility.Exists ||
		lowered.replacementAbility.Exists ||
		lowered.spellAbility.Exists {
		return unsupported()
	}
	for i := range lowered.staticAbilities {
		body := &lowered.staticAbilities[i].Body
		// A static ability's own KeywordAbilities field is applied through a path
		// that ignores the ability's Condition (only level-band conditions gate
		// it), so a ControllerHasMaxSpeed condition on such a body would be
		// silently dropped and the keyword granted regardless of speed. Continuous
		// and rule effects are correctly gated by Condition, so only fail closed on
		// the keyword-ability shape.
		if body.Condition.Exists || len(body.KeywordAbilities) != 0 {
			return unsupported()
		}
		body.Condition = opt.Val(game.Condition{ControllerHasMaxSpeed: true})
	}
	return lowered, nil
}
