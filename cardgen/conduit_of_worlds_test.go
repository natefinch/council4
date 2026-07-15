package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// conduitOfWorldsOracleText is Conduit of Worlds's exact Oracle text: the shared
// "play lands from your graveyard" static plus the targeted, conditional
// cast-and-lock activated ability this card support adds.
const conduitOfWorldsOracleText = "You may play lands from your graveyard.\n" +
	"{T}: Choose target nonland permanent card in your graveyard. If you haven't cast a spell this turn, you may cast that card. If you do, you can't cast additional spells this turn. Activate only as a sorcery."

// TestGenerateConduitOfWorlds proves the whole card renders end to end (parse →
// compile → lower → render) to the exact typed nodes the runtime consumes: the
// shared play-lands-from-graveyard static, and an activated ability that taps,
// activates only as a sorcery, targets a nonland permanent card in the
// controller's graveyard, optionally pays for and casts that card only when no
// spell was cast this turn, and then — only when that cast actually happened —
// forbids the controller from casting further spells this turn.
func TestGenerateConduitOfWorlds(t *testing.T) {
	t.Parallel()
	generatedSourceContains(t, &ScryfallCard{
		Name:       "Conduit of Worlds",
		Layout:     "normal",
		ManaCost:   "{2}{G}{G}",
		TypeLine:   "Artifact",
		OracleText: conduitOfWorldsOracleText,
	}, []string{
		"game.PlayLandsFromGraveyardStaticBody",
		"AdditionalCosts: cost.Tap",
		"Timing:          game.SorceryOnly",
		"Allow:      game.TargetAllowCard",
		"TargetZone: zone.Graveyard",
		"ExcludedTypes: []types.Card{types.Land}",
		"Controller: game.ControllerYou",
		"game.CastForFree{",
		"Card:        game.CardReference{Kind: game.CardReferenceTarget}",
		"PayManaCost: true",
		"Negate: true",
		"Event:      game.EventSpellCast",
		"Controller: game.TriggerControllerYou",
		"game.EventHistoryCurrentTurn",
		"Optional:      true",
		`PublishResult: game.ResultKey("if-you-do")`,
		"Kind:           game.RuleEffectCantCastSpells",
		"AffectedPlayer: game.PlayerYou",
		"Duration: game.DurationThisTurn",
		`Key:       "if-you-do"`,
		"Succeeded: game.TriTrue",
	})
}

// TestLowerConduitOfWorldsEnvelope proves the activated ability lowers to the
// exact two-instruction resolution envelope: an optional paid cast of the
// targeted graveyard card gated on the negated "cast a spell this turn" event
// history and publishing a result, followed by a controller-scoped
// "can't cast spells this turn" rule effect gated on that cast having succeeded.
func TestLowerConduitOfWorldsEnvelope(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Conduit of Worlds",
		Layout:     "normal",
		ManaCost:   "{2}{G}{G}",
		TypeLine:   "Artifact",
		OracleText: conduitOfWorldsOracleText,
	})

	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want 1", len(face.StaticAbilities))
	}
	if got := face.StaticAbilities[0].VarName; got != "game.PlayLandsFromGraveyardStaticBody" {
		t.Fatalf("static ability = %q, want game.PlayLandsFromGraveyardStaticBody", got)
	}

	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	ability := face.ActivatedAbilities[0]
	if len(ability.AdditionalCosts) != 1 || ability.AdditionalCosts[0].Kind != cost.AdditionalTap {
		t.Fatalf("activation cost = %+v, want a single tap", ability.AdditionalCosts)
	}
	if ability.Timing != game.SorceryOnly {
		t.Fatalf("timing = %v, want SorceryOnly", ability.Timing)
	}

	mode := ability.Content.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %d, want 1", len(mode.Targets))
	}
	target := mode.Targets[0]
	if target.MinTargets != 1 || target.MaxTargets != 1 {
		t.Fatalf("target cardinality = [%d,%d], want exactly one", target.MinTargets, target.MaxTargets)
	}
	if target.Allow != game.TargetAllowCard || target.TargetZone != zone.Graveyard {
		t.Fatalf("target allow=%v zone=%v, want a card in the graveyard", target.Allow, target.TargetZone)
	}
	sel := target.Selection.Val
	if sel.Controller != game.ControllerYou {
		t.Fatalf("target controller = %v, want ControllerYou (own graveyard)", sel.Controller)
	}
	if len(sel.ExcludedTypes) != 1 || sel.ExcludedTypes[0] != types.Land {
		t.Fatalf("target excluded types = %v, want [Land] (nonland)", sel.ExcludedTypes)
	}

	seq := mode.Sequence
	if len(seq) != 2 {
		t.Fatalf("resolution sequence = %d instructions, want 2 (cast, lock)", len(seq))
	}

	castInstr := seq[0]
	if !castInstr.Optional {
		t.Fatal("the cast instruction must be optional (\"you may cast that card\")")
	}
	if castInstr.PublishResult != game.ResultKey("if-you-do") {
		t.Fatalf("cast publishes %q, want if-you-do", castInstr.PublishResult)
	}
	castPrim, ok := castInstr.Primitive.(game.CastForFree)
	if !ok {
		t.Fatalf("cast instruction primitive = %T, want game.CastForFree", castInstr.Primitive)
	}
	if !castPrim.PayManaCost {
		t.Fatal("the cast must pay the card's mana cost, not cast it for free")
	}
	if castPrim.Card.Kind != game.CardReferenceTarget {
		t.Fatalf("cast card reference = %v, want CardReferenceTarget", castPrim.Card.Kind)
	}
	if castPrim.Zone != zone.Graveyard {
		t.Fatalf("cast source zone = %v, want Graveyard", castPrim.Zone)
	}
	cond := castInstr.Condition.Val.Condition.Val
	if !cond.Negate {
		t.Fatal("cast condition must be negated (\"haven't cast a spell\")")
	}
	hist := cond.EventHistory.Val
	if hist.Pattern.Event != game.EventSpellCast || hist.Pattern.Controller != game.TriggerControllerYou {
		t.Fatalf("cast condition history = %+v, want your spell-cast events", hist.Pattern)
	}
	if hist.Window != game.EventHistoryCurrentTurn {
		t.Fatalf("cast condition window = %v, want EventHistoryCurrentTurn", hist.Window)
	}

	lockInstr := seq[1]
	if lockInstr.Optional {
		t.Fatal("the lock instruction must be mandatory")
	}
	gate := lockInstr.ResultGate.Val
	if gate.Key != "if-you-do" || gate.Succeeded != game.TriTrue {
		t.Fatalf("lock result gate = %+v, want it gated on the cast having succeeded", gate)
	}
	apply, ok := lockInstr.Primitive.(game.ApplyRule)
	if !ok {
		t.Fatalf("lock instruction primitive = %T, want game.ApplyRule", lockInstr.Primitive)
	}
	if apply.Duration != game.DurationThisTurn {
		t.Fatalf("lock duration = %v, want DurationThisTurn", apply.Duration)
	}
	if len(apply.RuleEffects) != 1 {
		t.Fatalf("lock rule effects = %d, want 1", len(apply.RuleEffects))
	}
	rule := apply.RuleEffects[0]
	if rule.Kind != game.RuleEffectCantCastSpells {
		t.Fatalf("lock rule kind = %v, want RuleEffectCantCastSpells", rule.Kind)
	}
	if rule.AffectedPlayer != game.PlayerYou {
		t.Fatalf("lock affects %v, want PlayerYou (only the controller)", rule.AffectedPlayer)
	}
}

// TestLowerCantCastAdditionalOpponentScopeUnchanged is a fail-closed regression
// guard for the controller-scoped lowering: an opponents-scoped Silence-style
// prohibition must still lower to a PlayerOpponent rule effect, proving the
// PlayerYou path is gated on the controller wording alone and does not blanket
// every "can't cast spells this turn" prohibition.
func TestLowerCantCastAdditionalOpponentScopeUnchanged(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Opponent Silence Tester",
		Layout:     "normal",
		ManaCost:   "{W}",
		TypeLine:   "Instant",
		OracleText: "Your opponents can't cast spells this turn.",
	})
	seq := face.SpellAbility.Val.Modes[0].Sequence
	if len(seq) != 1 {
		t.Fatalf("spell sequence = %d instructions, want 1", len(seq))
	}
	apply, ok := seq[0].Primitive.(game.ApplyRule)
	if !ok {
		t.Fatalf("spell primitive = %T, want game.ApplyRule", seq[0].Primitive)
	}
	if got := apply.RuleEffects[0].AffectedPlayer; got != game.PlayerOpponent {
		t.Fatalf("opponents prohibition affects %v, want PlayerOpponent", got)
	}
}

// TestConduitOpponentGraveyardVariantUnsupported is a fail-closed near-miss: the
// same ability targeting an opponent's graveyard instead of the controller's own
// must not lower, because the choose-verb coverage recognizer accepts only the
// controller-scoped "target nonland permanent card in your graveyard" phrasing.
func TestConduitOpponentGraveyardVariantUnsupported(t *testing.T) {
	t.Parallel()
	face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:     "Conduit Opponent Tester",
		Layout:   "normal",
		ManaCost: "{2}{G}{G}",
		TypeLine: "Artifact",
		OracleText: "{T}: Choose target nonland permanent card in target opponent's graveyard. " +
			"If you haven't cast a spell this turn, you may cast that card. " +
			"If you do, you can't cast additional spells this turn. Activate only as a sorcery.",
	})
	if len(face.ActivatedAbilities) != 0 {
		t.Fatalf("opponent-graveyard variant produced %d activated abilities, want 0 (fail closed)", len(face.ActivatedAbilities))
	}
}
