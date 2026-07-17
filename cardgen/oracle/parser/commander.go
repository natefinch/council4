package parser

import "strings"

func emitCanBeCommander(abilities []Ability, cardName string) {
	cardName = strings.TrimSpace(cardName)
	if cardName == "" {
		return
	}
	want := strings.ToLower(cardName + " can be your commander.")
	for i := range abilities {
		if strings.ToLower(strings.TrimSpace(abilities[i].Text)) == want {
			abilities[i].CanBeCommander = true
		}
	}
}
