package helpers

const ZendeskConfigs string = "ZendeskConfigs"
const Users string = "Users"
const OfflineTickets string = "OfflineTickets"
const Keys string = "APIKeys"
const AssignedTickets string = "AssignedTickets"
const Appointments string = "Appointments"
const Agents string = "Agents"

func substr(input string, start int, length int) string {
	asRunes := []rune(input)
	if start >= len(asRunes) {
		return ""
	}
	if start+length > len(asRunes) {
		length = len(asRunes) - start
	}
	return string(asRunes[start : start+length])
}

func indexOf(el string, data []string) (int) {
	for k, v := range data {
		if el == v {
			return k
		}
	}
	return -1    //not found.
 }