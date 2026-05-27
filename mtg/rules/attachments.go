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
	if isAuraPermanent(g, attachment) {
		return auraCanAttachToPermanent(g, attachment, target)
	}
	if isEquipmentPermanent(g, attachment) {
		return permanentHasType(g, target, game.TypeCreature)
	}
	return false
}

func auraCanAttachToPermanent(g *game.Game, aura *game.Permanent, target *game.Permanent) bool {
	spec := enchantTargetSpecForPermanent(g, aura)
	if spec.Allow != game.TargetAllowUnspecified && spec.Allow&game.TargetAllowPermanent == 0 {
		return false
	}
	return permanentTargetMatchesSpec(g, effectiveController(g, aura), aura.ObjectID, spec, target.ObjectID)
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

func enchantTargetSpecForPermanent(g *game.Game, aura *game.Permanent) game.TargetSpec {
	def, ok := permanentCardDef(g, aura)
	if !ok {
		return defaultEnchantTargetSpec()
	}
	return enchantTargetSpecForCard(def)
}

func enchantTargetSpecForCard(card *game.CardDef) game.TargetSpec {
	for _, ability := range card.Abilities {
		if !abilityHasKeyword(&ability, game.Enchant) || !ability.EnchantTarget.Exists {
			continue
		}
		spec := ability.EnchantTarget.Val
		if spec.MinTargets == 0 {
			spec.MinTargets = 1
		}
		if spec.MaxTargets == 0 {
			spec.MaxTargets = 1
		}
		return spec
	}
	return defaultEnchantTargetSpec()
}

func defaultEnchantTargetSpec() game.TargetSpec {
	return game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowPermanent,
		Predicate: game.TargetPredicate{
			PermanentTypes: []game.CardType{game.TypeCreature},
		},
	}
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
