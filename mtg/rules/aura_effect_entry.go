package rules

import (
	"maps"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// CR 303.4f-g: choose an enchant-legal attachment, without targeting, before
// committing an effect-driven Aura entry. No legal choice means no zone move.
func (r *effectResolver) prepareAuraEffectEntry(card *game.CardInstance, controller game.PlayerID, options permanentCreationOptions) (permanentCreationOptions, bool) {
	face := game.FaceFront
	if options.EntersTransformed && card.Def.IsTransformingDoubleFaced() && card.Def.Back.Exists {
		face = game.FaceBack
	}
	def, ok := cardFaceDef(card, face)
	if !ok {
		return options, false
	}
	if !isAuraCard(def) || options.Bestowed {
		return options, true
	}
	prototype := &game.Permanent{
		ObjectID: card.ID, CardInstanceID: card.ID, Owner: card.Owner, Controller: controller, Face: face,
	}
	controller = effectiveController(r.game, prototype)
	var choices [][]game.Target
	for _, host := range r.game.Battlefield {
		if auraEffectAttachmentLegal(r.game, prototype, game.PermanentTarget(host.ObjectID)) {
			choices = append(choices, []game.Target{game.PermanentTarget(host.ObjectID)})
		}
	}
	for player := range game.PlayerID(game.NumPlayers) {
		if auraEffectAttachmentLegal(r.game, prototype, game.PlayerTarget(player)) {
			choices = append(choices, []game.Target{game.PlayerTarget(player)})
		}
	}
	if len(choices) == 0 {
		return options, false
	}
	index := 0
	if len(choices) > 1 {
		selected := r.engine.chooseChoice(r.game, r.agents,
			targetChoiceRequest(controller, "Choose what the Aura enchants.", choices), r.log)
		if len(selected) != 1 || selected[0] < 0 || selected[0] >= len(choices) {
			return options, false
		}
		index = selected[0]
	}
	options.auraAttachment = opt.Val(choices[index][0])
	entryChoices, ok := attachmentChoices(r.game, prototype, &replacementChoiceContext{
		engine: r.engine, agents: r.agents, log: r.log,
	})
	options.auraEntryChoices = entryChoices
	return options, ok && auraEffectAttachmentLegal(r.game, prototype, options.auraAttachment.Val)
}

// resolvingAuraSpellTarget is the object or player a resolving Aura spell
// becomes attached to (CR 608.3c). Graveyard-reanimation Auras attach through
// their own resolution and report none.
func resolvingAuraSpellTarget(g *game.Game, obj *game.StackObject, spellDef *game.CardDef) opt.V[game.Target] {
	if obj == nil || len(obj.Targets) == 0 {
		return opt.V[game.Target]{}
	}
	if _, reanimates := game.CardDefReanimationEnchant(spellDef); reanimates {
		return opt.V[game.Target]{}
	}
	if player, ok := effectPlayerTarget(g, obj, 0); ok {
		return opt.Val(game.PlayerTarget(player))
	}
	if host, ok := effectPermanentTarget(g, obj, 0); ok {
		return opt.Val(game.PermanentTarget(host.ObjectID))
	}
	return opt.V[game.Target]{}
}

// attachEnteringAuraSpell attaches a resolving Aura spell's permanent to its
// target before the permanent enters, so it enters already attached (CR
// 608.3c) and its enters abilities see what it enchants (CR 303.4m, 603.6a).
// An attachment that is not legal leaves the entering Aura unattached for the
// resolving spell's CR 303.4i or 702.103f outcome.
func attachEnteringAuraSpell(e *Engine, g *game.Game, aura *game.Permanent, target opt.V[game.Target], agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	if !target.Exists || !isAuraPermanent(g, aura) || !auraEffectAttachmentLegal(g, aura, target.Val) {
		return
	}
	choices, ok := attachmentChoices(g, aura, &replacementChoiceContext{engine: e, agents: agents, log: log})
	if !ok {
		return
	}
	prepareAuraEffectAttachment(g, aura, permanentCreationOptions{auraAttachment: target, auraEntryChoices: choices})
}

func auraEffectAttachmentLegal(g *game.Game, aura *game.Permanent, target game.Target) bool {
	if target.Kind == game.TargetPermanent {
		host, found := permanentByObjectID(g, target.PermanentID)
		return found && canAttachPermanent(g, aura, host) && !permanentProtectedFromPermanentEffective(g, host, aura)
	}
	values := effectivePermanentValues(g, aura)
	def := &game.CardDef{CardFace: game.CardFace{Colors: values.colors, Types: values.types, Subtypes: values.subtypes}}
	return target.Kind == game.TargetPlayer && auraCanAttachToPlayer(g, aura, target.PlayerID) &&
		!playerProtectedFromSource(g, target.PlayerID, 0, 0, def)
}

func commitAuraEffectAttachment(g *game.Game, aura *game.Permanent, options permanentCreationOptions) bool {
	if !prepareAuraEffectAttachment(g, aura, options) {
		return false
	}
	if aura.AttachedTo.Exists {
		host, _ := permanentByObjectID(g, aura.AttachedTo.Val)
		host.Attachments = append(host.Attachments, aura.ObjectID)
	}
	return true
}

func prepareAuraEffectAttachment(g *game.Game, aura *game.Permanent, options permanentCreationOptions) bool {
	if !options.auraAttachment.Exists {
		return true
	}
	target := options.auraAttachment.Val
	if !auraEffectAttachmentLegal(g, aura, target) {
		return false
	}
	if target.Kind == game.TargetPermanent {
		aura.AttachedTo = opt.Val(target.PermanentID)
	} else {
		aura.AttachedToPlayer = opt.Val(target.PlayerID)
	}
	if len(options.auraEntryChoices) != 0 {
		if aura.EntryChoices == nil {
			aura.EntryChoices = make(map[game.ChoiceKey]game.ResolutionChoiceResult)
		}
		maps.Copy(aura.EntryChoices, options.auraEntryChoices)
	}
	return true
}
