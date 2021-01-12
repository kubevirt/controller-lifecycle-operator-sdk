/*
Copyright 2020 The CDI Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sdk

import (
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const lastsAppliedConfigurationAnnotation = "lastAppliedConfiguration"

var _ = Describe("MergeLabelsAndAnnotations", func() {
	It("Should properly merge labels and annotations, if no dest labels/anns", func() {
		source := createPod("source", map[string]string{"l1": "test"}, map[string]string{"a1": "ann"})
		dest := createPod("dest", nil, nil)
		MergeLabelsAndAnnotations(&source.ObjectMeta, &dest.ObjectMeta)
		Expect(dest.GetObjectMeta).ToNot(BeNil())
		Expect(dest.GetLabels()["l1"]).To(Equal("test"))
		Expect(dest.GetAnnotations()["a1"]).To(Equal("ann"))
	})

	It("Should properly merge labels and annotations, if no dest labels", func() {
		source := createPod("source", map[string]string{"l1": "test"}, map[string]string{"a1": "ann"})
		dest := createPod("dest", nil, map[string]string{"a1": "ann2"})
		MergeLabelsAndAnnotations(&source.ObjectMeta, &dest.ObjectMeta)
		Expect(dest.GetObjectMeta).ToNot(BeNil())
		Expect(dest.GetLabels()["l1"]).To(Equal("test"))
		// Check that dest is now equal to source
		Expect(dest.GetAnnotations()["a1"]).To(Equal("ann"))
	})

	It("Should properly merge labels and annotations, if no dest labels, and different ann", func() {
		source := createPod("source", map[string]string{"l1": "test"}, map[string]string{"a1": "ann"})
		dest := createPod("dest", nil, map[string]string{"a2": "ann2"})
		MergeLabelsAndAnnotations(&source.ObjectMeta, &dest.ObjectMeta)
		Expect(dest.GetObjectMeta).ToNot(BeNil())
		Expect(dest.GetLabels()["l1"]).To(Equal("test"))
		Expect(dest.GetAnnotations()["a1"]).To(Equal("ann"))
		Expect(dest.GetAnnotations()["a2"]).To(Equal("ann2"))
	})

	It("Should properly merge labels and annotations, if no dest ann", func() {
		source := createPod("source", map[string]string{"l1": "test"}, map[string]string{"a1": "ann"})
		dest := createPod("dest", map[string]string{"l1": "test2"}, nil)
		MergeLabelsAndAnnotations(&source.ObjectMeta, &dest.ObjectMeta)
		Expect(dest.GetObjectMeta).ToNot(BeNil())
		// Check that dest is now equal to source
		Expect(dest.GetLabels()["l1"]).To(Equal("test"))
		Expect(dest.GetAnnotations()["a1"]).To(Equal("ann"))
	})

	It("Should properly merge labels and annotations, if no dest ann, and different label", func() {
		source := createPod("source", map[string]string{"l1": "test"}, map[string]string{"a1": "ann"})
		dest := createPod("dest", map[string]string{"l2": "test2"}, nil)
		MergeLabelsAndAnnotations(&source.ObjectMeta, &dest.ObjectMeta)
		Expect(dest.GetObjectMeta).ToNot(BeNil())
		Expect(dest.GetLabels()["l1"]).To(Equal("test"))
		Expect(dest.GetLabels()["l2"]).To(Equal("test2"))
		Expect(dest.GetAnnotations()["a1"]).To(Equal("ann"))
	})

	// TODO: fix the problem with preserving unknown fields
	PIt("will not merge CRD correctly", func() {
		obj1 := &extv1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: "obj",
			},
			Spec: extv1.CustomResourceDefinitionSpec{
				Group: "foo",
			},
		}

		err := SetLastAppliedConfiguration(obj1, lastsAppliedConfigurationAnnotation)
		Expect(err).ToNot(HaveOccurred())

		obj2 := obj1.DeepCopy()
		obj2.Spec.PreserveUnknownFields = true

		obj3 := obj1.DeepCopy()
		// not necessary but let's be explicit
		obj3.Spec.PreserveUnknownFields = false
		err = SetLastAppliedConfiguration(obj3, lastsAppliedConfigurationAnnotation)
		Expect(err).ToNot(HaveOccurred())

		obj4, err := MergeObject(obj3, obj2, lastsAppliedConfigurationAnnotation)
		Expect(err).ToNot(HaveOccurred())

		crd := obj4.(*extv1.CustomResourceDefinition)
		Expect(crd.Spec.PreserveUnknownFields).To(BeFalse())
	})

	It("will merge CRD correctly", func() {
		obj1 := &extv1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: "obj",
			},
			Spec: extv1.CustomResourceDefinitionSpec{
				Group:                 "foo",
				PreserveUnknownFields: true,
			},
		}

		err := SetLastAppliedConfiguration(obj1, lastsAppliedConfigurationAnnotation)
		Expect(err).ToNot(HaveOccurred())

		obj2 := obj1.DeepCopy()

		obj3 := obj1.DeepCopy()
		// not necessary but let's be explicit
		obj3.Spec.PreserveUnknownFields = false
		err = SetLastAppliedConfiguration(obj3, lastsAppliedConfigurationAnnotation)
		Expect(err).ToNot(HaveOccurred())

		obj4, err := MergeObject(obj3, obj2, lastsAppliedConfigurationAnnotation)
		Expect(err).ToNot(HaveOccurred())

		crd := obj4.(*extv1.CustomResourceDefinition)
		Expect(crd.Spec.PreserveUnknownFields).To(BeFalse())
	})
})

var _ = Describe("StripStatusFromObject", func() {
	It("Should not alter object without status", func() {
		in := &corev1.PodList{}
		out, err := StripStatusFromObject(in.DeepCopyObject())
		Expect(err).ToNot(HaveOccurred())
		Expect(reflect.DeepEqual(out, in)).To(BeTrue())
	})

	DescribeTable("Should strip object status", func(in, expected controllerutil.Object) {

		out, err := StripStatusFromObject(in)
		Expect(err).ToNot(HaveOccurred())
		Expect(reflect.DeepEqual(out, in)).To(BeFalse())
		Expect(reflect.DeepEqual(out, expected)).To(BeTrue())
	},
		Entry("status@Deployment",
			&appsv1.Deployment{
				Status: appsv1.DeploymentStatus{
					Replicas: 128,
				},
			},
			&appsv1.Deployment{Status: appsv1.DeploymentStatus{}},
		),
		Entry("Status@Pod",
			&corev1.Pod{
				Status: corev1.PodStatus{
					PodIP: "pod-ip",
				},
			},
			&corev1.Pod{Status: corev1.PodStatus{}},
		),
	)

	It("Should strip object status", func() {
		in := &appsv1.Deployment{
			Status: appsv1.DeploymentStatus{
				Replicas: 128,
			},
		}
		expected := &appsv1.Deployment{
			Status: appsv1.DeploymentStatus{},
		}
		out, err := StripStatusFromObject(in.DeepCopyObject())
		Expect(err).ToNot(HaveOccurred())
		Expect(reflect.DeepEqual(out, in)).To(BeFalse())
		Expect(reflect.DeepEqual(out, expected)).To(BeTrue())
	})

})

func createPod(name string, labels, annotations map[string]string) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	if len(labels) > 0 {
		pod.ObjectMeta.Labels = labels
	}
	if len(annotations) > 0 {
		pod.ObjectMeta.Annotations = annotations
	}
	return pod
}
