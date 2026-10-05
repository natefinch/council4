package game

import "testing"

func TestLocalProductDeclarationsAreDeepCloned(t *testing.T) {
	content := Mode{Sequence: []Instruction{{
		Primitive:     LookAtLibraryTop{Player: ControllerReference(), PublishLinked: "link"},
		LocalProducts: LocalProducts{Links: []LinkedKey{"link"}, Results: []ResultKey{"result"}},
	}}}.Ability()
	cloned := cloneAbilityContent(content)
	cloned.Modes[0].Sequence[0].LocalProducts.Links[0] = "other"
	cloned.Modes[0].Sequence[0].LocalProducts.Results[0] = "other"
	if !content.Modes[0].Sequence[0].LocalProducts.HasLink("link") ||
		!content.Modes[0].Sequence[0].LocalProducts.HasResult("result") {
		t.Fatal("ability clone shared mutable product declarations")
	}
}
