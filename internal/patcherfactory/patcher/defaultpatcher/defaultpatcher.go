package defaultpatcher

func PatchMap(original map[string]interface{}, patch map[string]interface{}, overwrite bool) {
	for key, patchValue := range patch {
		originalValue, exists := original[key]
		if !exists {
			// Rule 2: key not exists in original map, add it
			original[key] = patchValue
		} else if patchMap, isMap := patchValue.(map[string]interface{}); isMap {
			if originalMap, isMap := originalValue.(map[string]interface{}); isMap {
				// Rule 1: value is map[string]interface{} and key exists in original map, patch it
				PatchMap(originalMap, patchMap, overwrite)
			}
		} else {
			// Rule 3: value is not a map[string]interface{} and key exists in original map
			if overwrite {
				original[key] = patchValue
			}
		}
	}
}
