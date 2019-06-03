package util

import (
	appv1 "github.com/logancloud/logan-app-operator/pkg/apis/app/v1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

type MyStruct struct {
	Name  string
	Envs  []corev1.EnvVar
	Other []string
}

var _ = Describe("Merge", func() {
	BeforeEach(func() {

	})

	Context("With Env merged", func() {
		It("test", func() {
			testDataS := []struct {
				DST      MyStruct
				SRC      MyStruct
				Expected MyStruct
			}{
				{
					MyStruct{
						Name: "DST",
						Envs: []corev1.EnvVar{
							{Name: "key1", Value: "value1"},
							{Name: "key2", Value: "value2"},
						},
						Other: []string{"a", "b"},
					},
					MyStruct{
						Name: "SRC",
						Envs: []corev1.EnvVar{
							{Name: "key2", Value: "valueNew"},
							{Name: "key3", Value: "value3"},
						},
						Other: []string{"b", "c"},
					},
					MyStruct{
						Name: "SRC",
						Envs: []corev1.EnvVar{
							{Name: "key1", Value: "value1"},
							{Name: "key2", Value: "valueNew"},
							{Name: "key3", Value: "value3"},
						},
						Other: []string{"b", "c"},
					},
				},
			}

			for _, data := range testDataS {
				err := MergeOverride(&data.DST, data.SRC)
				Expect(err).NotTo(HaveOccurred())
				Expect(data.DST.Name).To(Equal(data.Expected.Name))

				//corev1.EnvVar[]: merge
				Expect(data.DST.Envs).Should(HaveLen(len(data.Expected.Envs)))
				for i, val := range data.DST.Envs {
					Expect(val).To(Equal(data.Expected.Envs[i]))
				}

				//string[]
				Expect(data.DST.Other).Should(HaveLen(len(data.Expected.Other)))
				for i, val := range data.DST.Other {
					Expect(val).To(Equal(data.Expected.Other[i]))
				}
			}
		})
	})

	Context("With Env merged", func() {
		It("test", func() {
			testDataS := []struct {
				DST      appv1.BootSpec
				SRC      appv1.BootSpec
				Expected appv1.BootSpec
			}{
				{
					appv1.BootSpec{
						Env: []corev1.EnvVar{
							{Name: "key1", Value: "value1"},
							{Name: "key2", Value: "value2"},
						},
					},
					appv1.BootSpec{
						Env: []corev1.EnvVar{
							{Name: "key2", Value: "valueNew"},
							{Name: "key3", Value: "value3"},
						},
					},
					appv1.BootSpec{
						Env: []corev1.EnvVar{
							{Name: "key1", Value: "value1"},
							{Name: "key2", Value: "value2"},
							{Name: "key3", Value: "value3"},
						},
					},
				},
			}

			for _, data := range testDataS {
				keys, err := MergeAppEnvs(&data.DST, data.SRC.Env)
				Expect(len(keys)).To(Equal(1))
				Expect(keys).To(Equal([]corev1.EnvVar{
					{Name: "key2", Value: "value2"},
				}))

				Expect(err).NotTo(HaveOccurred())

				for i, val := range data.DST.Env {
					Expect(val).To(Equal(data.Expected.Env[i]))
				}
			}
		})
	})
})
