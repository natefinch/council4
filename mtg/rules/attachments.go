package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func attachPermanent(g *game.Game, attachment, target *game.Permanent) bool {
	if !canAttachPermanent(g, attachment, target) {
		return false
	}
	detachPermanent(g, attachment)
	attachment.AttachedTo = opt.Val(target.ObjectID)
	if !permanentIDsContain(target.Attachments, attachment.ObjectID) {
		target.Attachments = append(target.Attachments, attachment.ObjectID)
	}
	return true
}

func detachPermanent(g *game.Game, attachment *game.Permanent) {
	if !attachment.AttachedTo.Exists {
		return
	}
	if target, ok := permanentByObjectID(g, attachment.AttachedTo.Val); ok {
		target.Attachments = removePermanentID(target.Attachments, attachment.ObjectID)
	}
	attachment.AttachedTo = opt.V[id.ID]{}
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
	spec, ok := enchantTargetSpecForPermanent(g, aura)
	if !ok {
		return false
	}
	if spec.Allow != game.TargetAllowUnspecified && spec.Allow&game.TargetAllowPermanent == 0 {
		return false
	}
	return permanentTargetMatchesSpec(g, effectiveController(g, aura), aura.ObjectID, &spec, target.ObjectID)
}

func isAuraPermanent(g *game.Game, permanent *game.Permanent) bool {
	return permanentHasType(g, permanent, types.Enchantment) && (permanentHasSubtype(g, permanent, types.Aura) || hasKeyword(g, permanent, game.Enchant))
}

func isEquipmentPermanent(g *game.Game, permanent *game.Permanent) bool {
	return permanentHasType(g, permanent, types.Artifact) && (permanentHasSubtype(g, permanent, types.Equipment) || hasKeyword(g, permanent, game.Equip))
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
	for i := range card.Abilities {
		ability := &card.Abilities[i]
		if !abilityHasKeyword(ability, game.Enchant) || !ability.EnchantTarget.Exists {
			continue
		}
		spec := ability.EnchantTarget.Val
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
