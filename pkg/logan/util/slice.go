package util

import (
	corev1 "k8s.io/api/core/v1"
)

// Returns
// diff1: in slice1, not in slice2
// diff2: not in slice1, in slice2
// https://stackoverflow.com/questions/19374219/how-to-find-the-difference-between-two-slices-of-strings-in-golang?answertab=votes#tab-top
func Difference(slice1 []string, slice2 []string) (diff1 []string, diff2 []string) {
	// Loop two times, first to find slice1 strings not in slice2,
	// second loop to find slice2 strings not in slice1
	for i := 0; i < 2; i++ {
		for _, s1 := range slice1 {
			found := false
			for _, s2 := range slice2 {
				if s1 == s2 {
					found = true
					break
				}
			}
			// String not found. We add it to return slice
			if !found {
				if i == 0 {
					diff1 = append(diff1, s1)
				} else {
					diff2 = append(diff2, s1)
				}
			}
		}
		// Swap the slices, only if it was the first loop
		if i == 0 {
			slice1, slice2 = slice2, slice1
		}
	}

	return
}

// Returns
// diff1: in slice1(origin), not in slice2(now) delete
// diff2: not in slice1(origin), in slice2(now) add
// modified: the slice2 value
// https://stackoverflow.com/questions/19374219/how-to-find-the-difference-between-two-slices-of-strings-in-golang?answertab=votes#tab-top
func Difference2(origin []corev1.EnvVar, now []corev1.EnvVar) (diff1 []corev1.EnvVar,
	diff2 []corev1.EnvVar, modified []corev1.EnvVar) {
	// Avoid the keys duplicate
	cMap := make(map[string]string)
	// Loop two times, first to find slice1 strings not in slice2,
	// second loop to find slice2 strings not in slice1
	for i := 0; i < 2; i++ {
		for _, s1 := range origin {
			found := false
			for _, s2 := range now {
				if s1.Name == s2.Name {
					if s1.Value != s2.Value {
						if i == 0 {
							modified = append(modified, s2)
						}
						cMap[s1.Name] = ""
					}
					found = true
					break
				}
			}
			// String not found. We add it to return slice
			if !found {
				if i == 0 {
					diff1 = append(diff1, s1)
				} else {
					diff2 = append(diff2, s1)
				}
			}
		}
		// Swap the slices, only if it was the first loop
		if i == 0 {
			origin, now = now, origin
		}
	}

	//for key, _ := range cMap {
	//	conflict = append(conflict, key)
	//}

	return
}
