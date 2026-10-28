package rules

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/opt"
)

// attackingBandsLegal checks the bands announced with an attack declaration
// (CR 508.1e, 702.22c-d). Every member is declared in this declaration, each
// creature is in at most one band, every member attacks the same player,
// planeswalker or battle, and each band is a legal band (bandEligibility.legalBand):
// at least two members, either at least one with banding and at most one
// without, or all of one [quality] with a member that has "bands with other
// [quality]".
func attackingBandsLegal(g *game.Game, declarations []game.AttackDeclaration, bands []game.AttackingBand, eligibleByID map[id.ID]*game.Permanent) bool {
	if len(bands) == 0 {
		return true
	}
	targets := make(map[id.ID]game.AttackTarget, len(declarations))
	for _, declaration := range declarations {
		targets[declaration.Attacker] = declaration.Target
	}
	inBand := make(map[id.ID]bool)
	for _, band := range bands {
		creatures := make([]*game.Permanent, 0, len(band.Members))
		members := make([]int, 0, len(band.Members))
		for index, member := range band.Members {
			target, declared := targets[member]
			if !declared || inBand[member] || target != targets[band.Members[0]] {
				return false
			}
			inBand[member] = true
			attacker := eligibleByID[member]
			if attacker == nil {
				return false
			}
			creatures = append(creatures, attacker)
			members = append(members, index)
		}
		if !newBandEligibility(g, creatures).legalBand(members) {
			return false
		}
	}
	return true
}

// chooseAttackingBands asks the active player to form attacking bands among a
// completed attack declaration (CR 508.1e). Each declared creature not yet in a
// band that has banding, or a "bands with other [quality]" ability while having
// that quality, may start one; the player then adds members one at a time until
// finishing it. Only creatures attacking the same player, planeswalker or
// battle, not in another band, whose addition keeps the band legal
// (bandEligibility.legalBand) are offered, so every answer stays legal (CR
// 702.22c-d). Every legal band is reachable this way: its other members can
// join its anchor in any order once a creature with banding or the "bands with
// other" holder has joined. Choosing to finish immediately forms no band, so the
// deterministic default declares no bands. Requests describe creatures only by
// their public characteristics.
func (ce combatEngine) chooseAttackingBands(g *game.Game, player game.PlayerID, declarations []game.AttackDeclaration, agents [game.NumPlayers]PlayerAgent, log *TurnLog) []game.AttackingBand {
	attackers := make([]*game.Permanent, len(declarations))
	g.BeginStaticSourceFrame()
	for index, declaration := range declarations {
		if attacker, ok := permanentByObjectID(g, declaration.Attacker); ok {
			attackers[index] = attacker
		}
	}
	eligibility := newBandEligibility(g, attackers)
	g.EndStaticSourceFrame()

	banded := make([]bool, len(declarations))
	var bands []game.AttackingBand
	for anchor, declaration := range declarations {
		if attackers[anchor] == nil || !eligibility.canStartBand(anchor) || banded[anchor] {
			continue
		}
		members := []int{anchor}
		for {
			options := []game.ChoiceOption{{Index: 0, Label: "Finish this band"}}
			if len(members) == 1 {
				options[0].Label = "Do not form a band"
			}
			var candidates []int
			for index, candidate := range declarations {
				if banded[index] || slices.Contains(members, index) || attackers[index] == nil ||
					candidate.Target != declaration.Target ||
					!eligibility.legalBand(append(slices.Clip(members), index)) {
					continue
				}
				options = append(options, game.ChoiceOption{
					Index:  len(options),
					Label:  fmt.Sprintf("Band %s with %s", permanentChoiceLabel(g, attackers[index]), permanentChoiceLabel(g, attackers[anchor])),
					Card:   permanentChoiceInfo(g, attackers[index]),
					Attack: opt.Val(candidate),
				})
				candidates = append(candidates, index)
			}
			if len(candidates) == 0 {
				break
			}
			selected := ce.e.chooseChoice(g, agents, game.ChoiceRequest{
				Kind: game.ChoiceAttackingBand, Player: player,
				Prompt:         fmt.Sprintf("Choose a creature to add to %s's attacking band", permanentChoiceLabel(g, attackers[anchor])),
				Subject:        permanentChoiceInfo(g, attackers[anchor]),
				CombatAttacker: declaration.Attacker,
				Options:        options, MinChoices: 1, MaxChoices: 1, DefaultSelection: []int{0},
			}, log)
			if selected[0] == 0 {
				break
			}
			members = append(members, candidates[selected[0]-1])
		}
		if len(members) < 2 {
			continue
		}
		slices.Sort(members)
		band := game.AttackingBand{Members: make([]id.ID, 0, len(members))}
		for _, member := range members {
			banded[member] = true
			band.Members = append(band.Members, declarations[member].Attacker)
		}
		bands = append(bands, band)
	}
	return bands
}

// removeFromAttackingBands removes a creature leaving combat from its band (CR
// 702.22f). A band left with one member is no longer a band.
func removeFromAttackingBands(combat *game.CombatState, permanentID id.ID) {
	bands := combat.Bands[:0]
	for _, band := range combat.Bands {
		band.Members = slices.DeleteFunc(band.Members, func(member id.ID) bool {
			return member == permanentID
		})
		if len(band.Members) >= 2 {
			bands = append(bands, band)
		}
	}
	if len(bands) == 0 {
		bands = nil
	}
	combat.Bands = bands
}

// attackingBandMembers returns every creature in attackerID's attacking band,
// including attackerID, or attackerID alone when it is not in a band. Blocking
// any member blocks them all (CR 702.22h).
func attackingBandMembers(combat *game.CombatState, attackerID id.ID) []id.ID {
	for _, band := range combat.Bands {
		if slices.Contains(band.Members, attackerID) {
			return band.Members
		}
	}
	return []id.ID{attackerID}
}

// attackingBandIndex maps each banded attacker to its band's index in
// combat.Bands. Attackers not in a band are absent.
func attackingBandIndex(combat *game.CombatState) map[id.ID]int {
	if len(combat.Bands) == 0 {
		return nil
	}
	index := make(map[id.ID]int)
	for i, band := range combat.Bands {
		for _, member := range band.Members {
			index[member] = i
		}
	}
	return index
}
