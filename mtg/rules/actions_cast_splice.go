package rules

import (
	"fmt"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
)

// spellIsArcane reports whether a spell has the Arcane subtype (CR 205.3k), the
// only spells onto which "Splice onto Arcane" cards may be spliced (CR 702.47a).
func spellIsArcane(spellDef *game.CardDef) bool {
	return spellDef != nil && spellDef.HasSubtype(types.Arcane)
}

// spliceCandidate is one card in hand eligible to be spliced onto the Arcane
// spell being cast: it has a mana splice cost the caster can afford (on top of
// the host spell and any already-chosen splices), a spell ability, and at least
// one legal target combination for that ability.
type spliceCandidate struct {
	cardID     id.ID
	def        *game.CardDef
	content    game.AbilityContent
	spliceCost cost.Mana
}

// spliceSelection is one resolved splice choice: the spliced card's spell effects,
// the targets chosen for them, the per-spec target counts, and the mana splice
// cost paid as an additional cost of the host spell.
type spliceSelection struct {
	content      game.AbilityContent
	targets      []game.Target
	targetCounts []int
	manaCost     cost.Mana
}

// spliceCastResult is the outcome of the splice offer: the spliced spell
// contents, their announced targets, their per-spec target counts, and the mana
// splice costs paid as additional costs, all indexed in parallel in the order the
// caster chose. A zero value (all-nil) means nothing was spliced.
type spliceCastResult struct {
	contents     []game.AbilityContent
	targets      [][]game.Target
	targetCounts [][]int
	manaCosts    []cost.Mana
}

// chooseSpliceOntoArcane offers the caster each affordable, legally targetable
// "Splice onto Arcane" card in hand while an Arcane spell is being cast (CR
// 702.47). The caster may splice any number of them, in any order; each chosen
// card is revealed and stays in hand, its splice cost is paid as an additional
// cost of the host spell, and its spell effects (and targets) are appended to the
// host spell. When no card is spliced it returns the zero result so the cast is
// byte-for-byte identical to one with no splice opportunity.
func (e *Engine) chooseSpliceOntoArcane(
	g *game.Game,
	playerID game.PlayerID,
	hostCardID id.ID,
	hostDef *game.CardDef,
	cast action.CastSpellAction,
	permissions []payment.SpellCastPermission,
	agents [game.NumPlayers]PlayerAgent,
	log *TurnLog,
) spliceCastResult {
	chosen := make(map[id.ID]bool)
	var selections []spliceSelection
	for {
		accumulated := spliceManaCostsOf(selections)
		candidates := e.eligibleSpliceCandidates(g, playerID, hostCardID, hostDef, cast, permissions, accumulated, chosen)
		if len(candidates) == 0 {
			break
		}
		selectedID, ok := e.chooseSpliceCard(g, playerID, candidates, agents, log)
		if !ok {
			break
		}
		cand, ok := spliceCandidateByID(candidates, selectedID)
		if !ok {
			break
		}
		targets, targetCounts, ok := e.chooseSpliceTargets(g, playerID, cand.def, agents, log)
		if !ok {
			break
		}
		selections = append(selections, spliceSelection{
			content:      cand.content,
			targets:      targets,
			targetCounts: targetCounts,
			manaCost:     cand.spliceCost,
		})
		chosen[selectedID] = true
	}
	if len(selections) == 0 {
		return spliceCastResult{}
	}
	result := spliceCastResult{
		contents:     make([]game.AbilityContent, len(selections)),
		targets:      make([][]game.Target, len(selections)),
		targetCounts: make([][]int, len(selections)),
		manaCosts:    make([]cost.Mana, len(selections)),
	}
	for i, sel := range selections {
		result.contents[i] = sel.content
		result.targets[i] = sel.targets
		result.targetCounts[i] = sel.targetCounts
		result.manaCosts[i] = sel.manaCost
	}
	return result
}

// spliceManaCostsOf returns the mana splice costs of the already-chosen splices,
// used to check whether the caster can also afford the next candidate.
func spliceManaCostsOf(selections []spliceSelection) []cost.Mana {
	if len(selections) == 0 {
		return nil
	}
	costs := make([]cost.Mana, len(selections))
	for i, sel := range selections {
		costs[i] = sel.manaCost
	}
	return costs
}

// eligibleSpliceCandidates gathers the "Splice onto Arcane" cards in the caster's
// hand that can still be spliced onto the host spell: excluding the host card and
// any already-chosen splice, requiring a spell ability, requiring the caster to
// be able to pay the host cost plus every accumulated splice cost plus this one,
// and requiring at least one legal target combination for the spliced effects.
func (*Engine) eligibleSpliceCandidates(
	g *game.Game,
	playerID game.PlayerID,
	hostCardID id.ID,
	hostDef *game.CardDef,
	cast action.CastSpellAction,
	permissions []payment.SpellCastPermission,
	accumulated []cost.Mana,
	chosen map[id.ID]bool,
) []spliceCandidate {
	player := g.Players[playerID]
	var candidates []spliceCandidate
	for _, handID := range player.Hand.All() {
		if handID == hostCardID || chosen[handID] {
			continue
		}
		card, ok := g.GetCardInstance(handID)
		if !ok {
			continue
		}
		def := cardFaceOrDefault(card, game.FaceFront)
		spliceCost, ok := def.SpliceCost()
		if !ok {
			continue
		}
		ability, ok := firstSpellAbility(def)
		if !ok {
			continue
		}
		combined := append(append([]cost.Mana(nil), accumulated...), append(cost.Mana(nil), spliceCost...))
		if !paymentOrch.canPaySpellCosts(g, payment.SpellRequest{
			PlayerID:        playerID,
			CardID:          hostCardID,
			SourceZone:      zone.Hand,
			Card:            hostDef,
			XValue:          cast.XValue,
			KickerPaid:      cast.KickerPaid,
			KickerCount:     cast.KickerCount,
			ChosenModes:     cast.ChosenModes,
			CastPermissions: permissions,
			Targets:         cast.Targets,
			SpliceManaCosts: combined,
		}) {
			continue
		}
		result := targetChoicesForSpell(g, playerID, def, nil, game.CastBranch{})
		if result.kind == targetNoLegalChoices || result.kind == targetInvalidSpec {
			continue
		}
		candidates = append(candidates, spliceCandidate{
			cardID:     handID,
			def:        def,
			content:    *ability,
			spliceCost: append(cost.Mana(nil), spliceCost...),
		})
	}
	return candidates
}

// spliceCandidateByID returns the candidate with the given card ID.
func spliceCandidateByID(candidates []spliceCandidate, cardID id.ID) (spliceCandidate, bool) {
	for _, cand := range candidates {
		if cand.cardID == cardID {
			return cand, true
		}
	}
	return spliceCandidate{}, false
}

// chooseSpliceCard asks the caster which eligible card to splice next, or to
// decline. The request offers no default and permits an empty selection
// (MinChoices 0), so an agent that does not answer declines to splice — keeping
// existing simulations unchanged whenever no scripted decision is supplied.
func (e *Engine) chooseSpliceCard(
	g *game.Game,
	playerID game.PlayerID,
	candidates []spliceCandidate,
	agents [game.NumPlayers]PlayerAgent,
	log *TurnLog,
) (id.ID, bool) {
	options := make([]game.ChoiceOption, len(candidates))
	for i, cand := range candidates {
		options[i] = game.ChoiceOption{
			Index: i,
			Label: fmt.Sprintf("Splice %s onto this Arcane spell", cand.def.Name),
			Card:  cardChoiceInfo(g, cand.cardID),
		}
	}
	selected := e.chooseChoice(g, agents, game.ChoiceRequest{
		Kind:       game.ChoiceZoneSelection,
		Player:     playerID,
		Prompt:     "Choose a card to splice onto this Arcane spell, or decline.",
		Options:    options,
		MinChoices: 0,
		MaxChoices: 1,
	}, log)
	if len(selected) != 1 {
		return 0, false
	}
	index := selected[0]
	if index < 0 || index >= len(candidates) {
		return 0, false
	}
	return candidates[index].cardID, true
}

// chooseSpliceTargets chooses the targets for one spliced card's spell effects,
// mirroring how a triggered ability announces its targets (CR 601.2c via the
// splice card's own target specs). The returned targets are indexed from zero to
// match the spliced content's own target references. It reports ok=false only
// when the effects require targets but none are legal, in which case the splice
// is skipped.
func (e *Engine) chooseSpliceTargets(
	g *game.Game,
	playerID game.PlayerID,
	def *game.CardDef,
	agents [game.NumPlayers]PlayerAgent,
	log *TurnLog,
) ([]game.Target, []int, bool) {
	result := targetChoicesForSpell(g, playerID, def, nil, game.CastBranch{})
	switch result.kind {
	case targetNoLegalChoices, targetInvalidSpec:
		return nil, nil, false
	default:
	}
	choices := result.choices
	counts := result.targetCounts
	index := 0
	if len(choices) > 1 {
		selected := e.chooseChoice(g, agents, targetChoiceRequest(playerID, "Choose targets for the spliced spell.", choices), log)
		if len(selected) == 1 && selected[0] >= 0 && selected[0] < len(choices) {
			index = selected[0]
		}
	}
	bound, ok := bindCardTargetZoneVersions(g, choices[index])
	if !ok {
		return nil, nil, false
	}
	var chosenCounts []int
	if index < len(counts) {
		chosenCounts = append([]int(nil), counts[index]...)
	}
	return bound, chosenCounts, true
}
