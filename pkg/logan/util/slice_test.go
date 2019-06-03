package util

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Diff", func() {
	BeforeEach(func() {

	})

	Context("With String Diff", func() {
		It("test string", func() {
			testStringS := []struct {
				SLICE1        []string
				SLICE2        []string
				ExpectedDiff1 []string
				ExpectedDiff2 []string
			}{
				{
					[]string{"foo", "bar", "hello"},
					[]string{"foo", "world", "bar", "foo"},
					[]string{"hello"},
					[]string{"world"},
				},
			}

			for _, data := range testStringS {
				diff1, diff2 := Difference(data.SLICE1, data.SLICE2)

				Expect(diff1).Should(HaveLen(len(data.ExpectedDiff1)))
				for i, val := range diff1 {
					Expect(val).To(Equal(data.ExpectedDiff1[i]))
				}

				Expect(diff2).Should(HaveLen(len(data.ExpectedDiff2)))
				for i, val := range diff2 {
					Expect(val).To(Equal(data.ExpectedDiff2[i]))
				}
			}
		})
	})

	Context("With Env Diff", func() {
		It("test envs", func() {
			testDataS := []struct {
				ORIGIN           []corev1.EnvVar
				NOW              []corev1.EnvVar
				ExpectedDiff1    []corev1.EnvVar
				ExpectedDiff2    []corev1.EnvVar
				ExpectedModified []corev1.EnvVar
			}{
				{
					[]corev1.EnvVar{
						{Name: "key1", Value: "value1"},
						{Name: "key2", Value: "value2"},
					},
					[]corev1.EnvVar{
						{Name: "key2", Value: "valueNew"},
						{Name: "key3", Value: "value3"},
					},
					[]corev1.EnvVar{
						{Name: "key1", Value: "value1"},
					},
					[]corev1.EnvVar{
						{Name: "key3", Value: "value3"},
					},
					[]corev1.EnvVar{
						{Name: "key2", Value: "valueNew"},
					},
				},
			}

			for _, data := range testDataS {
				diff1, diff2, modified := Difference2(data.ORIGIN, data.NOW)

				Expect(diff1).Should(HaveLen(len(data.ExpectedDiff1)))
				for i, val := range diff1 {
					Expect(val).To(Equal(data.ExpectedDiff1[i]))
				}

				Expect(diff2).Should(HaveLen(len(data.ExpectedDiff2)))
				for i, val := range diff2 {
					Expect(val).To(Equal(data.ExpectedDiff2[i]))
				}

				Expect(modified).Should(HaveLen(len(data.ExpectedModified)))
				for i, val := range modified {
					Expect(val).To(Equal(data.ExpectedModified[i]))
				}
			}
		})
	})
})
