package overwritepatcher

func PatchMap(original, patch map[string]interface{}) (map[string]interface{}, error) {
	for key, patchValue := range patch {
		originalValue, ok := original[key]
		if ok {
			// If the key exists in both the original and patch maps,
			// recursively merge the nested maps.
			patchMap, patchMapOk := patchValue.(map[string]interface{})
			originalMap, originalMapOk := originalValue.(map[string]interface{})
			if patchMapOk && originalMapOk {
				PatchMap(originalMap, patchMap)
				continue
			}
		}
		// If the key doesn't exist in the original map, or the value is not a map,
		// set the value to the patch value.
		original[key] = patchValue
	}
	return original, nil
}
