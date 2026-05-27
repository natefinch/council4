package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/opt"
)

func attachPermanent(g *game.Game, attachment *game.Permanent, target *game.Permanent) bool {
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

func canAttachPermanent(g *game.Game, attachment *game.Permanent, target *game.Permanent) bool {
	if attachment.ObjectID == target.ObjectID {
		return false
	}
	if isAuraPermanent(g, attachment) || isEquipmentPermanent(g, attachment) {
		return permanentHasType(g, target, game.TypeCreature)
	}
	return false
}

func isAuraPermanent(g *game.Game, permanent *game.Permanent) bool {
	return permanentHasType(g, permanent, game.TypeEnchantment) && (permanentHasSubtype(g, permanent, "Aura") || hasKeyword(g, permanent, game.Enchant))
}

func isEquipmentPermanent(g *game.Game, permanent *game.Permanent) bool {
	return permanentHasType(g, permanent, game.TypeArtifact) && (permanentHasSubtype(g, permanent, "Equipment") || hasKeyword(g, permanent, game.Equip))
}

func isAttachmentPermanent(g *game.Game, permanent *game.Permanent) bool {
	return isAuraPermanent(g, permanent) || isEquipmentPermanent(g, permanent)
}

func isAuraCard(card *game.CardDef) bool {
	return card != nil && card.HasType(game.TypeEnchantment) && (card.HasSubtype("Aura") || card.HasKeyword(game.Enchant))
}

func isEquipmentCard(card *game.CardDef) bool {
	return card != nil && card.HasType(game.TypeArtifact) && (card.HasSubtype("Equipment") || card.HasKeyword(game.Equip))
}

func permanentIDsContain(ids []id.ID, want id.ID) bool {
	for _, got := range ids {
		if got == want {
			return true
		}
	}
	return false
}

func removePermanentID(ids []id.ID, remove id.ID) []id.ID {
	for i, got := range ids {
		if got == remove {
			return append(ids[:i], ids[i+1:]...)
		}
	}
	return ids
}
