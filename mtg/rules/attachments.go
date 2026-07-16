package rules

import (
	"maps"
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func attachPermanent(g *game.Game, attachment, target *game.Permanent) bool {
	return attachPermanentWithChoices(g, attachment, target, nil)
}

func attachPermanentWithChoices(g *game.Game, attachment, target *game.Permanent, choiceCtx *replacementChoiceContext) bool {
	if !canAttachPermanent(g, attachment, target) {
		return false
	}
	if attachment.AttachedTo.Exists && attachment.AttachedTo.Val == target.ObjectID {
		return true
	}
	choices, ok := attachmentChoices(g, attachment, choiceCtx)
	if !ok {
		return false
	}
	detachPermanent(g, attachment)
	if len(choices) > 0 {
		if attachment.EntryChoices == nil {
			attachment.EntryChoices = make(map[game.ChoiceKey]game.ResolutionChoiceResult)
		}
		maps.Copy(attachment.EntryChoices, choices)
	}
	attachment.AttachedTo = opt.Val(target.ObjectID)
	if !permanentIDsContain(target.Attachments, attachment.ObjectID) {
		target.Attachments = append(target.Attachments, attachment.ObjectID)
	}
	return true
}

func attachmentChoices(g *game.Game, attachment *game.Permanent, choiceCtx *replacementChoiceContext) (map[game.ChoiceKey]game.ResolutionChoiceResult, bool) {
	if attachment == nil || attachment.FaceDown {
		return nil, true
	}
	def, ok := permanentCardDef(g, attachment)
	if !ok || def == nil {
		return nil, true
	}
	if choiceCtx == nil {
		if current, present := replacementChoiceContextFor(g); present {
			choiceCtx = current
		} else {
			choiceCtx = &replacementChoiceContext{engine: NewEngine(nil)}
		}
	}
	controller := effectiveController(g, attachment)
	results := make(map[game.ChoiceKey]game.ResolutionChoiceResult)
	for i := range def.ReplacementAbilities {
		replacement := &def.ReplacementAbilities[i].Replacement
		if replacement.AttachCardNameChoiceType != "" {
			result, chosen := choiceCtx.engine.choosePersistentValue(g, choiceCtx.agents, controller, &game.ResolutionChoice{
				Kind:         game.ResolutionChoiceCardName,
				Prompt:       "Choose a card name.",
				CardNameType: replacement.AttachCardNameChoiceType,
			}, choiceCtx.log)
			if !chosen {
				return nil, false
			}
			results[game.AttachmentCardNameChoiceKey] = result
		}
		if replacement.AttachSubtypeChoiceType != "" {
			result, chosen := choiceCtx.engine.choosePersistentValue(g, choiceCtx.agents, controller, &game.ResolutionChoice{
				Kind:          game.ResolutionChoiceSubtype,
				Prompt:        "Choose a subtype.",
				SubtypeOfType: replacement.AttachSubtypeChoiceType,
			}, choiceCtx.log)
			if !chosen {
				return nil, false
			}
			results[game.AttachmentSubtypeChoiceKey] = result
		}
	}
	return results, true
}

// attachAuraToPlayer attaches a player-enchanting Aura to a player (CR 303.4h,
// CR 701.3). It clears any prior permanent or player attachment first, then
// records the player pointer. It returns false without mutating anything when
// the Aura may not legally enchant the player, mirroring attachPermanent.
func attachAuraToPlayer(g *game.Game, aura *game.Permanent, playerID game.PlayerID) bool {
	if !auraCanAttachToPlayer(g, aura, playerID) {
		return false
	}
	detachPermanent(g, aura)
	aura.AttachedToPlayer = opt.Val(playerID)
	return true
}

// attachResolvingAura attaches a just-resolved Aura permanent to the target its
// spell chose, which is a player for an Enchant-player Aura or a permanent
// otherwise. It returns false when no legal target remains, so the caller
// applies the unattached-Aura outcome (owner's graveyard, or, for a bestowed
// Aura, ceasing to be bestowed). Reanimation Auras have their own resolution
// path and never reach here.
func attachResolvingAura(g *game.Game, obj *game.StackObject, aura *game.Permanent) bool {
	if playerID, ok := effectPlayerTarget(g, obj, 0); ok {
		return attachAuraToPlayer(g, aura, playerID)
	}
	target, ok := effectPermanentTarget(g, obj, 0)
	if !ok {
		return false
	}
	return attachPermanent(g, aura, target)
}

func detachPermanent(g *game.Game, attachment *game.Permanent) {
	if attachment.AttachedTo.Exists {
		if target, ok := permanentByObjectID(g, attachment.AttachedTo.Val); ok {
			target.Attachments = removePermanentID(target.Attachments, attachment.ObjectID)
		}
		attachment.AttachedTo = opt.V[id.ID]{}
	}
	// An Aura attached to a player has no permanent host tracking a reverse
	// Attachments entry, so clearing the player pointer fully detaches it.
	attachment.AttachedToPlayer = opt.V[game.PlayerID]{}
}

func detachAttachmentsFromPermanent(g *game.Game, target *game.Permanent) {
	for _, attachmentID := range target.Attachments {
		attachment, ok := permanentByObjectID(g, attachmentID)
		if ok && attachment.AttachedTo.Exists && attachment.AttachedTo.Val == target.ObjectID {
			attachment.AttachedTo = opt.V[id.ID]{}
		}
	}
	target.Attachments = nil
}

func canAttachPermanent(g *game.Game, attachment, target *game.Permanent) bool {
	if attachment.ObjectID == target.ObjectID {
		return false
	}
	if isAuraPermanent(g, attachment) {
		return auraCanAttachToPermanent(g, attachment, target)
	}
	if isEquipmentPermanent(g, attachment) {
		return permanentHasType(g, target, types.Creature)
	}
	return false
}

func auraCanAttachToPermanent(g *game.Game, aura, target *game.Permanent) bool {
	if aura.ReanimationLinkedObject != 0 {
		// A resolved graveyard-reanimation Aura (Animate Dead, Dance of the Dead)
		// has lost "enchant creature card in a graveyard" and gained "enchant
		// creature put onto the battlefield with this Aura". Its only legal host
		// is the specific permanent it returned to the battlefield, tracked by
		// ReanimationLinkedObject; every other object is an illegal attachment.
		return target.ObjectID == aura.ReanimationLinkedObject
	}
	spec, ok := enchantTargetSpecForPermanent(g, aura)
	if !ok {
		return false
	}
	if spec.Allow != game.TargetAllowUnspecified && spec.Allow&game.TargetAllowPermanent == 0 {
		return false
	}
	return permanentTargetMatchesSpec(g, effectiveController(g, aura), aura.ObjectID, game.Event{}, &spec, target.ObjectID)
}

// auraCanAttachToPlayer reports whether aura may legally enchant playerID. The
// Aura's Enchant restriction must permit a player, and the player must satisfy
// the Aura's player relation (for example "Enchant opponent") and still be in
// the game (CR 704.5m). Reanimation Auras never enchant players, so their
// linked-object restriction never applies here.
func auraCanAttachToPlayer(g *game.Game, aura *game.Permanent, playerID game.PlayerID) bool {
	spec, ok := enchantTargetSpecForPermanent(g, aura)
	if !ok {
		return false
	}
	if spec.Allow != game.TargetAllowUnspecified && spec.Allow&game.TargetAllowPlayer == 0 {
		return false
	}
	return playerTargetMatchesSpec(g, effectiveController(g, aura), &spec, playerID)
}

func isAuraPermanent(g *game.Game, permanent *game.Permanent) bool {
	return permanentHasType(g, permanent, types.Enchantment) && (permanentHasSubtype(g, permanent, types.Aura) || hasKeyword(g, permanent, game.Enchant))
}

func permanentIsEnchanted(g *game.Game, permanent *game.Permanent) bool {
	for _, attachmentID := range permanent.Attachments {
		attachment, ok := permanentByObjectID(g, attachmentID)
		if ok && activeBattlefieldPermanent(attachment) && isAuraPermanent(g, attachment) {
			return true
		}
	}
	return false
}

func permanentIsEquipped(g *game.Game, permanent *game.Permanent) bool {
	for _, attachmentID := range permanent.Attachments {
		attachment, ok := permanentByObjectID(g, attachmentID)
		if ok && activeBattlefieldPermanent(attachment) && isEquipmentPermanent(g, attachment) {
			return true
		}
	}
	return false
}

func isEquipmentPermanent(g *game.Game, permanent *game.Permanent) bool {
	return permanentHasType(g, permanent, types.Artifact) && (permanentHasSubtype(g, permanent, types.Equipment) || hasKeyword(g, permanent, game.Equip))
}

// bodyAttachesLikeEquip reports whether an activated ability is an Equip-style
// attachment activation. Reconfigure (CR 702.151) shares Equip's sorcery-speed
// attach-to-target-creature-you-control activation and resolution, so the rules
// layer dispatches both the same way.
func bodyAttachesLikeEquip(body *game.ActivatedAbility) bool {
	return game.BodyHasKeyword(body, game.Equip) || game.BodyHasKeyword(body, game.Reconfigure)
}

func isAttachmentPermanent(g *game.Game, permanent *game.Permanent) bool {
	return isAuraPermanent(g, permanent) || isEquipmentPermanent(g, permanent)
}

func isAuraCard(card *game.CardDef) bool {
	return card != nil && card.HasType(types.Enchantment) && (card.HasSubtype(types.Aura) || card.HasKeyword(game.Enchant))
}

func enchantTargetSpecForPermanent(g *game.Game, aura *game.Permanent) (game.TargetSpec, bool) {
	def, ok := permanentCardDef(g, aura)
	if !ok {
		return game.TargetSpec{}, false
	}
	return enchantTargetSpecForCard(def)
}

func enchantTargetSpecForCard(card *game.CardDef) (game.TargetSpec, bool) {
	for i := range card.StaticAbilities {
		spec, ok := game.StaticBodyEnchantTarget(&card.StaticAbilities[i])
		if !ok {
			continue
		}
		if spec.MinTargets == 0 {
			spec.MinTargets = 1
		}
		if spec.MaxTargets == 0 {
			spec.MaxTargets = 1
		}
		return spec, true
	}
	// A card with Bestow (CR 702.103) has no Enchant static ability while it is a
	// creature, but its Bestow keyword carries the enchant-creature target the
	// bestowed Aura permanent must keep as its attachment legality.
	if bestow, ok := game.CardDefBestow(card); ok {
		spec := bestow.Target
		if spec.MinTargets == 0 {
			spec.MinTargets = 1
		}
		if spec.MaxTargets == 0 {
			spec.MaxTargets = 1
		}
		return spec, true
	}
	return game.TargetSpec{}, false
}

func isEquipmentCard(card *game.CardDef) bool {
	return card != nil && card.HasType(types.Artifact) && (card.HasSubtype(types.Equipment) || card.HasKeyword(game.Equip))
}

func permanentIDsContain(ids []id.ID, want id.ID) bool {
	return slices.Contains(ids, want)
}

func removePermanentID(ids []id.ID, remove id.ID) []id.ID {
	for i, got := range ids {
		if got == remove {
			return append(ids[:i], ids[i+1:]...)
		}
	}
	return ids
}
