package util

// Converts a map of empty structs with string key to structure to an array of keys
func ArrayifyMap(in map[string]struct{}) []string {
	o := make([]string, len(in))
	i := 0
	for k, _ := range in {
		o[i] = k
		i++
	}
	return o
}

// Removes duplicates from a string array. Returning a new array with
// duplicates removed
func DeDupeStringArray(in []string) []string {
	m := make(map[string]struct{})
	for _, v := range in {
		m[v] = struct{}{}
	}

	return ArrayifyMap(m)
}
