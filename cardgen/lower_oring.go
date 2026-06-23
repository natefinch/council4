package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// exileUntilLeavesKey is the constant linked key binding an O-Ring exile to the
// source permanent that exiled it. The runtime keys linked objects by source
// card-instance id plus this string, so a fixed key still keeps each prison
// permanent's exiled card distinct. The enters-the-battlefield exile publishes
// it and the synthesized leaves-the-battlefield trigger consumes it to return
// the card.
const exileUntilLeavesKey = game.LinkedKey("exile-until-leaves")

// lowerExileUntilLeavesContent lowers the O-Ring enters-the-battlefield clause
// "exile target <permanent> until <this permanent> leaves the battlefield."
// (Banisher Priest, Banishing Light, Fairgrounds Warden) into a linked exile of
// the single target. The trailing self-reference is the duration anchor naming
// the source permanent, not a second object, so it is consumed rather than
// bound to a target. A paired leaves-the-battlefield return trigger is
// synthesized at the face level (synthesizeExileUntilLeavesReturns) using the
// same linked key.
//
// It returns ok=false for any shape it does not fully consume: a missing or
// non-single target, a plural (MaxTargets above one) target, a non-exile or
// residual optional effect, a condition, mode, or keyword rider, or a reference
// that is not the source duration anchor. The "up to one target" cardinality
// and an optional "you may" offer are accepted.
func lowerExileUntilLeavesContent(ctx contentCtx) (game.AbilityContent, bool) {
	if !exileUntilLeavesEnclosingKind(ctx.enclosingKind) ||
		ctx.optional ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectExile ||
		!effect.ExileUntilSourceLeaves ||
		!effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.Context != parser.EffectContextController {
		return game.AbilityContent{}, false
	}
	if !referencesAreOnlySourceAnchors(ctx.content.References) {
		return game.AbilityContent{}, false
	}
	// permanentTargetSpecWithCardinality carries the target's own
	// MinTargets/MaxTargets range, admitting the "exile up to one target ..."
	// wording (MinTargets 0, MaxTargets 1) in addition to the single mandatory
	// target. MaxTargets must stay 1: the exile binds a single card under the
	// exile-until-leaves link key, so a plural ("any number of target") wording
	// stays closed. An optional "you may" offer is recognized at the parser
	// level (exactExileUntilSourceLeavesEffectSyntax) and surfaces as the
	// trigger's Optional flag after prepareTriggerBody strips the "you may"
	// prefix from the leading effect, so neither effect.Optional nor ctx.optional
	// is set on this clause by then.
	targetSpec, ok := permanentTargetSpecWithCardinality(ctx.content.Targets[0])
	if !ok || targetSpec.MaxTargets != 1 {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{{
			Primitive: game.Exile{
				Object:         game.TargetPermanentReference(0),
				ExileLinkedKey: exileUntilLeavesKey,
			},
		}},
	}.Ability(), true
}

// exileUntilLeavesEnclosingKind reports whether an ability kind may host the
// O-Ring exile-until-leaves clause. A triggered ability carries the
// enters-the-battlefield form (Banisher Priest, Banishing Light); a Saga chapter
// ability carries the chapter form ("I, II, III — Exile target ... until this
// Saga leaves the battlefield.", e.g. Summon: Ixion, Trial of a Time Lord). Both
// publish the linked exile under the source, and the face-level
// synthesizeExileUntilLeavesReturns adds the paired return-on-leave trigger.
func exileUntilLeavesEnclosingKind(kind compiler.AbilityKind) bool {
	return kind == compiler.AbilityTriggered || kind == compiler.AbilityChapter
}

// referencesAreOnlySourceAnchors reports whether every reference (if any) is the
// "this permanent" / source-name duration anchor bound to the source permanent,
// as in "... until this creature leaves the battlefield." Such references name
// the source as the exile's lifetime anchor and carry no resolving object, so
// the O-Ring lowering consumes them in place of a target binding.
func referencesAreOnlySourceAnchors(references []compiler.CompiledReference) bool {
	for _, reference := range references {
		if reference.Binding != compiler.ReferenceBindingSource {
			return false
		}
		if reference.Kind != compiler.ReferenceThisObject &&
			reference.Kind != compiler.ReferenceSelfName {
			return false
		}
	}
	return true
}

// lowerReturnExiledCardContent lowers the explicit O-Ring leaves-the-battlefield
// clause "return the exiled card to the battlefield under its owner's control."
// (Oblivion Ring, Journey to Nowhere, Fiend Hunter) into a linked battlefield
// return reading the exile-until-leaves key. The exiled card is identified by
// the source link rather than a target, so the clause carries no target; the
// paired exile (an enters-the-battlefield trigger, or an earlier Saga chapter as
// in The Princess Takes Flight) is rewritten to publish the same key by
// linkExplicitExileReturns at the face level. A triggered ability hosts the
// O-Ring form and a Saga chapter hosts the split "exile in an early chapter,
// return in a later chapter" form.
//
// It returns ok=false for any shape it does not fully consume: a target, a
// condition, mode, or keyword rider, an optional or negated effect, or a
// non-controller context.
func lowerReturnExiledCardContent(ctx contentCtx) (game.AbilityContent, bool) {
	if !exileUntilLeavesEnclosingKind(ctx.enclosingKind) ||
		ctx.optional ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectReturn ||
		!effect.ReturnExiledCard ||
		!effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.Context != parser.EffectContextController {
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.PutOnBattlefield{
			Source: game.LinkedBattlefieldSource(exileUntilLeavesKey),
		},
	}}}.Ability(), true
}

// linkExplicitExileReturns binds an O-Ring face's two explicit clauses (Oblivion
// Ring, Journey to Nowhere, Fiend Hunter): an enters-the-battlefield exile of a
// target permanent and a separate leaves-the-battlefield "return the exiled card"
// trigger. The same pairing also spans Saga chapters (The Princess Takes Flight:
// "I — Exile up to one target creature." … "III — Return the exiled card to the
// battlefield under its owner's control."). The return already lowered to a
// linked battlefield put reading exileUntilLeavesKey; this pass publishes that
// key on the paired exile so the runtime link binds them, mirroring the
// single-ability Shape B form where one clause carries both halves.
//
// It acts only when the face has the linked return and exactly one unlinked
// single-target exile to pair with it — an enters-the-battlefield self trigger
// first, then a Saga chapter — leaving any other shape unlinked so the
// unpublished-key validation fails the card closed.
func linkExplicitExileReturns(result *loweredFaceAbilities) {
	if !faceReturnsLinkedBattlefield(result, exileUntilLeavesKey) || faceExilesUntilLeaves(result) {
		return
	}
	if sequence, exile, ok := soleEntersTargetExile(result); ok {
		exile.ExileLinkedKey = exileUntilLeavesKey
		sequence[0].Primitive = exile
		return
	}
	if sequence, exile, ok := soleChapterTargetExile(result); ok {
		exile.ExileLinkedKey = exileUntilLeavesKey
		sequence[0].Primitive = exile
	}
}

// soleEntersTargetExile returns the single-instruction sequence and its exile
// for the one enters-the-battlefield self trigger whose sole instruction exiles
// a single target permanent with no existing linked key. It returns ok=false
// when there is not exactly one such exile, so the caller links nothing
// ambiguous.
func soleEntersTargetExile(result *loweredFaceAbilities) ([]game.Instruction, game.Exile, bool) {
	var foundSequence []game.Instruction
	var foundExile game.Exile
	count := 0
	for abilityIndex := range result.TriggeredAbilities {
		ability := &result.TriggeredAbilities[abilityIndex]
		if ability.Trigger.Pattern.Event != game.EventPermanentEnteredBattlefield ||
			ability.Trigger.Pattern.Source != game.TriggerSourceSelf {
			continue
		}
		for modeIndex := range ability.Content.Modes {
			sequence := ability.Content.Modes[modeIndex].Sequence
			if len(sequence) != 1 {
				continue
			}
			exile, ok := sequence[0].Primitive.(game.Exile)
			if !ok || exile.ExileLinkedKey != "" ||
				exile.Object.Kind() != game.ObjectReferenceTargetPermanent {
				continue
			}
			foundSequence = sequence
			foundExile = exile
			count++
		}
	}
	if count != 1 {
		return nil, game.Exile{}, false
	}
	return foundSequence, foundExile, true
}

// soleChapterTargetExile returns the single-instruction sequence and its exile
// for the one Saga chapter whose sole instruction exiles a single target
// permanent with no existing linked key, mirroring soleEntersTargetExile for the
// chapter form. It returns ok=false when there is not exactly one such chapter
// exile, so the caller links nothing ambiguous (The Princess Takes Flight has
// exactly one: chapter I).
func soleChapterTargetExile(result *loweredFaceAbilities) ([]game.Instruction, game.Exile, bool) {
	var foundSequence []game.Instruction
	var foundExile game.Exile
	count := 0
	for abilityIndex := range result.ChapterAbilities {
		content := &result.ChapterAbilities[abilityIndex].Content
		for modeIndex := range content.Modes {
			sequence := content.Modes[modeIndex].Sequence
			if len(sequence) != 1 {
				continue
			}
			exile, ok := sequence[0].Primitive.(game.Exile)
			if !ok || exile.ExileLinkedKey != "" ||
				exile.Object.Kind() != game.ObjectReferenceTargetPermanent {
				continue
			}
			foundSequence = sequence
			foundExile = exile
			count++
		}
	}
	if count != 1 {
		return nil, game.Exile{}, false
	}
	return foundSequence, foundExile, true
}

// for an O-Ring face. When a lowered trigger exiles a permanent under the
// exile-until-leaves linked key and the face declares no return for that key,
// it adds "When this permanent leaves the battlefield, return the exiled card to
// the battlefield under its owner's control." so the prison releases its captive
// when it leaves play.
func synthesizeExileUntilLeavesReturns(result *loweredFaceAbilities) {
	if !faceExilesUntilLeaves(result) || faceReturnsLinkedBattlefield(result, exileUntilLeavesKey) {
		return
	}
	result.TriggeredAbilities = append(result.TriggeredAbilities, exileUntilLeavesReturnAbility())
}

func faceExilesUntilLeaves(result *loweredFaceAbilities) bool {
	for abilityIndex := range result.TriggeredAbilities {
		if contentExilesUntilLeaves(&result.TriggeredAbilities[abilityIndex].Content) {
			return true
		}
	}
	for abilityIndex := range result.ChapterAbilities {
		if contentExilesUntilLeaves(&result.ChapterAbilities[abilityIndex].Content) {
			return true
		}
	}
	return false
}

func contentExilesUntilLeaves(content *game.AbilityContent) bool {
	for modeIndex := range content.Modes {
		for instructionIndex := range content.Modes[modeIndex].Sequence {
			exile, ok := content.Modes[modeIndex].Sequence[instructionIndex].Primitive.(game.Exile)
			if ok && exile.ExileLinkedKey == exileUntilLeavesKey {
				return true
			}
		}
	}
	return false
}

func faceReturnsLinkedBattlefield(result *loweredFaceAbilities, key game.LinkedKey) bool {
	for abilityIndex := range result.TriggeredAbilities {
		if contentReturnsLinkedBattlefield(&result.TriggeredAbilities[abilityIndex].Content, key) {
			return true
		}
	}
	for abilityIndex := range result.ChapterAbilities {
		if contentReturnsLinkedBattlefield(&result.ChapterAbilities[abilityIndex].Content, key) {
			return true
		}
	}
	return false
}

func contentReturnsLinkedBattlefield(content *game.AbilityContent, key game.LinkedKey) bool {
	for modeIndex := range content.Modes {
		for instructionIndex := range content.Modes[modeIndex].Sequence {
			put, ok := content.Modes[modeIndex].Sequence[instructionIndex].Primitive.(game.PutOnBattlefield)
			if !ok {
				continue
			}
			if linked, ok := put.Source.LinkedKey(); ok && linked == key {
				return true
			}
		}
	}
	return false
}

func exileUntilLeavesReturnAbility() game.TriggeredAbility {
	return game.TriggeredAbility{
		Text: "When this permanent leaves the battlefield, return the exiled card to the battlefield under its owner's control.",
		Trigger: game.TriggerCondition{
			Type: game.TriggerWhen,
			Pattern: game.TriggerPattern{
				Event:         game.EventZoneChanged,
				Source:        game.TriggerSourceSelf,
				MatchFromZone: true,
				FromZone:      zone.Battlefield,
			},
		},
		Content: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.PutOnBattlefield{
				Source: game.LinkedBattlefieldSource(exileUntilLeavesKey),
			},
		}}}.Ability(),
	}
}
