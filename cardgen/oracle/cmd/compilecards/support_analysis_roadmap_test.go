package main

import (
	"slices"
	"testing"
)

func cardNamed(name string, summaries ...string) unsupported {
	card := cardWithSummaries(summaries...)
	card.Name = name
	return card
}

func TestAnalyzeUnblockRoadmapGreedyOrder(t *testing.T) {
	t.Parallel()
	// Blocker sets: c1{a}, c2{a,b}, c3{b}, c4{a,b,cc}.
	// Step 1: fixing "a" completes only c1 (its sole blocker); "b" completes only
	//   c3. Tie at 1 -> alphabetical "a". c2 -> {b}, c4 -> {b,cc}.
	// Step 2: "b" now completes c2 and c3 (2). c4 -> {cc}.
	// Step 3: "cc" completes c4 (1).
	output := report{Unsupported: []unsupported{
		cardNamed("c1", "a"),
		cardNamed("c2", "a", "b"),
		cardNamed("c3", "b"),
		cardNamed("c4", "a", "b", "cc"),
	}}

	got := analyzeUnblockRoadmap(output)
	want := []roadmapStep{
		{summary: "a", capability: capabilityOther, newlyUnblocked: 1, cumulativeUnblocked: 1, sampleCards: []string{"c1"}},
		{summary: "b", capability: capabilityOther, newlyUnblocked: 2, cumulativeUnblocked: 3, sampleCards: []string{"c2", "c3"}},
		{summary: "cc", capability: capabilityOther, newlyUnblocked: 1, cumulativeUnblocked: 4, sampleCards: []string{"c4"}},
	}
	if len(got) != len(want) {
		t.Fatalf("steps = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i].summary != want[i].summary ||
			got[i].capability != want[i].capability ||
			got[i].newlyUnblocked != want[i].newlyUnblocked ||
			got[i].cumulativeUnblocked != want[i].cumulativeUnblocked ||
			!slices.Equal(got[i].sampleCards, want[i].sampleCards) {
			t.Fatalf("step %d = %#v, want %#v", i, got[i], want[i])
		}
	}
}

func TestAnalyzeUnblockRoadmapAssignsCapability(t *testing.T) {
	t.Parallel()
	output := report{Unsupported: []unsupported{
		cardNamed("Ankh", "unsupported enters-tapped replacement"),
	}}
	got := analyzeUnblockRoadmap(output)
	if len(got) != 1 {
		t.Fatalf("steps = %d, want 1", len(got))
	}
	if got[0].capability != capabilityReplacement {
		t.Fatalf("capability = %q, want %q", got[0].capability, capabilityReplacement)
	}
}

func TestAnalyzeUnblockRoadmapEmpty(t *testing.T) {
	t.Parallel()
	if steps := analyzeUnblockRoadmap(report{}); len(steps) != 0 {
		t.Fatalf("steps = %#v, want none", steps)
	}
}
