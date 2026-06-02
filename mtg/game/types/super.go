package types

// Super represents a card's supertype (CR 205.4).
type Super string

// Super values identify supported card supertypes.
const (
	Legendary Super = "Legendary"
	Basic     Super = "Basic"
	Snow      Super = "Snow"
	World     Super = "World"
	Ongoing   Super = "Ongoing"
)
