package ignorepatcher

func PatchMap(original, patch map[string]interface{}) {
	for key, patchValue := range patch {
		originalValue, ok := original[key]
		if ok {
			// If the key exists in both the original and patch maps,
			// recursively merge the nested maps.
			patchMap, patchMapOk := patchValue.(map[string]interface{})
			originalMap, originalMapOk := originalValue.(map[string]interface{})
			if patchMapOk && originalMapOk {
				PatchMap(originalMap, patchMap)
				// If the value is a map and it exists in the original map,
				// merge the two maps and update the original map.
				for k, v := range patchMap {
					if originalValue, ok := originalMap[k]; ok {
						if originalValueMap, ok := originalValue.(map[string]interface{}); ok {
							PatchMap(originalValueMap, v.(map[string]interface{}))
							continue
						}
					}
					originalMap[k] = v
				}
				continue
			}
			// If the key exists in the original map, do not overwrite its value.
			continue
		}
		// If the key doesn't exist in the original map, or the value is not a map,
		// set the value to the patch value.
		original[key] = patchValue
	}
}
