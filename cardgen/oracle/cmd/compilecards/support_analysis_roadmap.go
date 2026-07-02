package main

import "slices"

// unblockRoadmapStepLimit caps the rendered roadmap so the planning document
// stays focused on the highest-leverage fixes.
const unblockRoadmapStepLimit = 30

// unblockRoadmapSampleLimit caps the sample card names shown per step.
const unblockRoadmapSampleLimit = 5

// roadmapStep is one fix in the greedy unblock roadmap: resolving reason
// `summary` — after every earlier step's reason is already resolved — newly
// completes `newlyUnblocked` cards, i.e. cards whose only remaining distinct
// diagnostic summary is this one. cumulativeUnblocked is the running total of
// cards fully unblocked through this step.
type roadmapStep struct {
	summary             string
	capability          supportCapabilityID
	newlyUnblocked      int
	cumulativeUnblocked int
	sampleCards         []string
}

// analyzeUnblockRoadmap computes a greedy set-cover ordering of unsupported
// reasons. A card is generated only when every one of its distinct diagnostic
// summaries is resolved, so the highest-leverage plan repeatedly fixes the reason
// that — given the reasons already fixed — newly completes the most still-blocked
// cards. It returns those fixes in priority order with marginal and cumulative
// fully-unblocked counts, turning "how do we unblock the most cards at once?" into
// a ranked to-do list. Each ability's fan-out lowerers (ordered sequence, modal,
// optional) now report every independent blocker they carry, so a card's distinct
// summary set reflects all of its blockers rather than only the first encountered;
// the marginal counts below therefore no longer overstate how many cards a single
// fix unblocks.
func analyzeUnblockRoadmap(output report) []roadmapStep {
	type cardBlockers struct {
		name      string
		remaining int
	}
	// bySummary maps each still-unresolved reason to the cards that still carry
	// it. A card leaves a reason's set only when that reason is fixed (the whole
	// set is dropped) or the card is completed.
	bySummary := make(map[string]map[*cardBlockers]bool)
	for _, card := range output.Unsupported {
		summaries := distinctDiagnosticSummaries(card.Diagnostics)
		if len(summaries) == 0 {
			continue
		}
		blockers := &cardBlockers{name: card.Name, remaining: len(summaries)}
		for _, summary := range summaries {
			cards := bySummary[summary]
			if cards == nil {
				cards = make(map[*cardBlockers]bool)
				bySummary[summary] = cards
			}
			cards[blockers] = true
		}
	}

	var steps []roadmapStep
	cumulative := 0
	for len(steps) < unblockRoadmapStepLimit {
		// Choose the reason that would complete the most cards now: a card in a
		// reason's set is completed by fixing that reason exactly when it is the
		// card's last remaining blocker (remaining == 1).
		candidates := make([]string, 0, len(bySummary))
		for summary := range bySummary {
			candidates = append(candidates, summary)
		}
		slices.Sort(candidates)
		bestSummary := ""
		bestComplete := 0
		for _, summary := range candidates {
			complete := 0
			for blockers := range bySummary[summary] {
				if blockers.remaining == 1 {
					complete++
				}
			}
			if complete > bestComplete {
				bestComplete = complete
				bestSummary = summary
			}
		}
		if bestComplete == 0 {
			break
		}

		completed := make([]string, 0, bestComplete)
		for blockers := range bySummary[bestSummary] {
			blockers.remaining--
			if blockers.remaining == 0 {
				completed = append(completed, blockers.name)
			}
		}
		delete(bySummary, bestSummary)
		cumulative += len(completed)
		slices.Sort(completed)
		samples := completed
		if len(samples) > unblockRoadmapSampleLimit {
			samples = samples[:unblockRoadmapSampleLimit]
		}
		steps = append(steps, roadmapStep{
			summary:             bestSummary,
			capability:          capabilityForDiagnostic(bestSummary),
			newlyUnblocked:      len(completed),
			cumulativeUnblocked: cumulative,
			sampleCards:         slices.Clone(samples),
		})
	}
	return steps
}
