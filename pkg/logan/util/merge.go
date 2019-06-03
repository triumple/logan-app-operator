package util

import (
	"github.com/imdario/mergo"
	appv1 "github.com/logancloud/logan-app-operator/pkg/apis/app/v1"
	corev1 "k8s.io/api/core/v1"
	"reflect"
)

type overrideEnvTransformer struct {
}

// Use envTransfomer to handle the merge of corev1.EnvVar

// slice Merge default
// 1. No Options: slice will no overwrite, src keep not changed.
// 2. option=mergo.WithOverride：slice will overwrite to the src whole slice.
// 3. option=mergo.WithAppendSlice：slice will append the src's，but the keys will be duplicated.
func (t overrideEnvTransformer) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
	if typ == reflect.TypeOf([]corev1.EnvVar{}) {
		return func(dst, src reflect.Value) error {
			// Replace the old slice index with src slice value
			mergeSlice(dst, src, true)
			return nil
		}
	}
	return nil
}

type unOverrideEnvTransformer struct {
}

func (t unOverrideEnvTransformer) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
	if typ == reflect.TypeOf([]corev1.EnvVar{}) {
		return func(dst, src reflect.Value) error {
			// Keep the old slice value, not replaced with src slice value
			mergeSlice(dst, src, false)
			return nil
		}
	}
	return nil
}

func mergeSlice(dst, src reflect.Value, override bool) error {
	if dst.CanSet() {
		for i := 0; i < src.Len(); i++ {
			// 每个要merge的值
			foundIndex := -1
			srcValue := src.Index(i)
			srcVar := srcValue.Interface().(corev1.EnvVar)
			for j := 0; j < dst.Len(); j++ {
				dstValue := dst.Index(j).Interface().(corev1.EnvVar)
				if srcVar.Name == dstValue.Name {
					foundIndex = j
					break
				}
			}

			if foundIndex >= 0 {
				if override {
					dst.Index(foundIndex).Set(srcValue)
				}
			} else {
				dst.Set(reflect.Append(dst, srcValue))
			}
		}
	}
	return nil
}

// MergeOverride merge the src and dst, handle the logic of same slice keys.
// same slice keys: src will overwrite dst.
func MergeOverride(dst, src interface{}) error {
	return mergo.Merge(dst, src, mergo.WithOverride, mergo.WithTransformers(overrideEnvTransformer{}))
}

// MergeOverride merge the src and dst, handle the logic of same slice keys.
// same slice keys: dst will keep not changed.
func MergeUnOverride(dst, src interface{}) error {
	return mergo.Merge(dst, src, mergo.WithOverride, mergo.WithTransformers(unOverrideEnvTransformer{}))
}

func MergeAppEnvs(dstSpec *appv1.BootSpec, src []corev1.EnvVar) ([]corev1.EnvVar, error) {
	keys := make([]corev1.EnvVar, 0)
	dst := dstSpec.Env
	for i := 0; i < len(src); i++ {
		// 每个要merge的值
		foundIndex := -1
		srcValue := src[i]
		for j := 0; j < len(dst); j++ {
			if srcValue.Name == dst[j].Name {
				foundIndex = j
				break
			}
		}

		if foundIndex >= 0 {
			keys = append(keys, dst[foundIndex])
		} else {
			dst = append(dst, srcValue)
		}
	}

	return keys, nil
}
