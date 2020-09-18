package sdk_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	v1 "github.com/openshift/custom-resource-status/conditions/v1"
	v12 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"kubevirt.io/controller-lifecycle-operator-sdk/pkg/sdk"
	sdkapi "kubevirt.io/controller-lifecycle-operator-sdk/pkg/sdk/api"
	testcr "kubevirt.io/controller-lifecycle-operator-sdk/tests/cr"
)

var _ = Describe("CR status", func() {
	var recorder = &record.FakeRecorder{}
	DescribeTable("should be upgrading for", func(phase sdkapi.Phase, observedVersion, targetVersion string) {
		crStatus := sdkapi.Status{TargetVersion: targetVersion, ObservedVersion: observedVersion, Phase: phase}

		upgrading := sdk.IsUpgrading(&crStatus)

		Expect(upgrading).To(BeTrue())
	},
		Entry("Upgrading and versions observed version empty", sdkapi.PhaseUpgrading, "", "v0.0.2"),
		Entry("Deploying and non-empty versions differ", sdkapi.PhaseDeploying, "v0.0.1", "v0.0.2"),
	)

	DescribeTable("should not be upgrading for", func(phase sdkapi.Phase, observedVersion, targetVersion string) {
		crStatus := sdkapi.Status{TargetVersion: targetVersion, ObservedVersion: observedVersion, Phase: phase}

		upgrading := sdk.IsUpgrading(&crStatus)

		Expect(upgrading).To(BeFalse())
	},
		Entry("Upgrading and versions are equal", sdkapi.PhaseUpgrading, "v0.0.1", "v0.0.1"),
		Entry("Deploying and non-empty versions don't differ", sdkapi.PhaseDeploying, "v0.0.1", "v0.0.1"),
	)

	It("should be marked healthy", func() {
		cr := testcr.Config{}
		crStatus := sdkapi.Status{
			Conditions: []v1.Condition{
				{
					Type:   v1.ConditionAvailable,
					Status: v12.ConditionFalse,
				},
				{
					Type:   v1.ConditionProgressing,
					Status: v12.ConditionTrue,
				},
				{
					Type:   v1.ConditionDegraded,
					Status: v12.ConditionTrue,
				},
			},
		}
		reason := "TheReason"
		message := "the message"

		sdk.MarkCrHealthyMessage(&cr, &crStatus, reason, message, recorder)

		Expect(crStatus.Conditions).To(HaveLen(3))

		availableCondition := v1.FindStatusCondition(crStatus.Conditions, v1.ConditionAvailable)
		Expect(availableCondition.Status).To(Equal(v12.ConditionTrue))
		Expect(availableCondition.Message).To(Equal(message))
		Expect(availableCondition.Reason).To(Equal(reason))

		correctProgressing := v1.IsStatusConditionPresentAndEqual(crStatus.Conditions, v1.ConditionProgressing, v12.ConditionFalse)
		Expect(correctProgressing).To(BeTrue())

		correctDegraded := v1.IsStatusConditionPresentAndEqual(crStatus.Conditions, v1.ConditionDegraded, v12.ConditionFalse)
		Expect(correctDegraded).To(BeTrue())
	})

	It("should be marked upgrade healing degraded", func() {
		cr := testcr.Config{}
		crStatus := sdkapi.Status{
			Conditions: []v1.Condition{
				{
					Type:   v1.ConditionAvailable,
					Status: v12.ConditionFalse,
				},
				{
					Type:   v1.ConditionProgressing,
					Status: v12.ConditionFalse,
				},
				{
					Type:   v1.ConditionDegraded,
					Status: v12.ConditionFalse,
				},
			},
		}
		reason := "TheReason"
		message := "the message"

		sdk.MarkCrUpgradeHealingDegraded(&cr, &crStatus, reason, message, recorder)

		Expect(crStatus.Conditions).To(HaveLen(3))

		degradedCondition := v1.FindStatusCondition(crStatus.Conditions, v1.ConditionDegraded)
		Expect(degradedCondition.Status).To(Equal(v12.ConditionTrue))
		Expect(degradedCondition.Message).To(Equal(message))
		Expect(degradedCondition.Reason).To(Equal(reason))

		correctProgressing := v1.IsStatusConditionPresentAndEqual(crStatus.Conditions, v1.ConditionProgressing, v12.ConditionTrue)
		Expect(correctProgressing).To(BeTrue())

		correctAvailable := v1.IsStatusConditionPresentAndEqual(crStatus.Conditions, v1.ConditionAvailable, v12.ConditionTrue)
		Expect(correctAvailable).To(BeTrue())
	})

	It("should be marked failed", func() {
		cr := testcr.Config{}
		crStatus := sdkapi.Status{
			Conditions: []v1.Condition{
				{
					Type:   v1.ConditionAvailable,
					Status: v12.ConditionTrue,
				},
				{
					Type:   v1.ConditionProgressing,
					Status: v12.ConditionTrue,
				},
				{
					Type:   v1.ConditionDegraded,
					Status: v12.ConditionFalse,
				},
			},
		}
		reason := "TheReason"
		message := "the message"

		sdk.MarkCrFailed(&cr, &crStatus, reason, message, recorder)

		Expect(crStatus.Conditions).To(HaveLen(3))

		degradedCondition := v1.FindStatusCondition(crStatus.Conditions, v1.ConditionDegraded)
		Expect(degradedCondition.Status).To(Equal(v12.ConditionTrue))
		Expect(degradedCondition.Message).To(Equal(message))
		Expect(degradedCondition.Reason).To(Equal(reason))

		correctProgressing := v1.IsStatusConditionPresentAndEqual(crStatus.Conditions, v1.ConditionProgressing, v12.ConditionFalse)
		Expect(correctProgressing).To(BeTrue())

		correctAvailable := v1.IsStatusConditionPresentAndEqual(crStatus.Conditions, v1.ConditionAvailable, v12.ConditionFalse)
		Expect(correctAvailable).To(BeTrue())
	})

	It("should be marked failed and healing", func() {
		cr := testcr.Config{}
		crStatus := sdkapi.Status{
			Conditions: []v1.Condition{
				{
					Type:   v1.ConditionAvailable,
					Status: v12.ConditionTrue,
				},
				{
					Type:   v1.ConditionProgressing,
					Status: v12.ConditionFalse,
				},
				{
					Type:   v1.ConditionDegraded,
					Status: v12.ConditionFalse,
				},
			},
		}
		reason := "TheReason"
		message := "the message"

		sdk.MarkCrFailedHealing(&cr, &crStatus, reason, message, recorder)

		Expect(crStatus.Conditions).To(HaveLen(3))

		degradedCondition := v1.FindStatusCondition(crStatus.Conditions, v1.ConditionDegraded)
		Expect(degradedCondition.Status).To(Equal(v12.ConditionTrue))
		Expect(degradedCondition.Message).To(Equal(message))
		Expect(degradedCondition.Reason).To(Equal(reason))

		correctProgressing := v1.IsStatusConditionPresentAndEqual(crStatus.Conditions, v1.ConditionProgressing, v12.ConditionTrue)
		Expect(correctProgressing).To(BeTrue())

		correctAvailable := v1.IsStatusConditionPresentAndEqual(crStatus.Conditions, v1.ConditionAvailable, v12.ConditionFalse)
		Expect(correctAvailable).To(BeTrue())
	})

	It("should be marked deploying", func() {
		cr := testcr.Config{}
		crStatus := sdkapi.Status{}
		reason := "TheReason"
		message := "the message"

		sdk.MarkCrDeploying(&cr, &crStatus, reason, message, recorder)

		Expect(crStatus.Conditions).To(HaveLen(3))

		progressingCondition := v1.FindStatusCondition(crStatus.Conditions, v1.ConditionProgressing)
		Expect(progressingCondition.Status).To(Equal(v12.ConditionTrue))
		Expect(progressingCondition.Message).To(Equal(message))
		Expect(progressingCondition.Reason).To(Equal(reason))

		correctDegraded := v1.IsStatusConditionPresentAndEqual(crStatus.Conditions, v1.ConditionDegraded, v12.ConditionFalse)
		Expect(correctDegraded).To(BeTrue())

		correctAvailable := v1.IsStatusConditionPresentAndEqual(crStatus.Conditions, v1.ConditionAvailable, v12.ConditionFalse)
		Expect(correctAvailable).To(BeTrue())
	})
})

var _ = Describe("For Conditions", func() {
	It("should return values", func() {
		conditions := []v1.Condition{
			{
				Type:   v1.ConditionAvailable,
				Status: v12.ConditionTrue,
			},
			{
				Type:   v1.ConditionProgressing,
				Status: v12.ConditionFalse,
			},
			{
				Type:   v1.ConditionDegraded,
				Status: v12.ConditionTrue,
			},
		}

		conditionValues := sdk.GetConditionValues(conditions)

		Expect(conditionValues).To(HaveKeyWithValue(v1.ConditionAvailable, v12.ConditionTrue))
		Expect(conditionValues).To(HaveKeyWithValue(v1.ConditionProgressing, v12.ConditionFalse))
		Expect(conditionValues).To(HaveKeyWithValue(v1.ConditionDegraded, v12.ConditionTrue))

	})

	It("should detect no changes", func() {
		conditions := []v1.Condition{
			{
				Type:   v1.ConditionAvailable,
				Status: v12.ConditionTrue,
			},
			{
				Type:   v1.ConditionProgressing,
				Status: v12.ConditionFalse,
			},
		}
		notChanged := sdk.GetConditionValues(conditions)

		changed := sdk.ConditionsChanged(notChanged, notChanged)

		Expect(changed).To(BeFalse())
	})

	DescribeTable("should detect changes", func(originalValues, newValues map[v1.ConditionType]v12.ConditionStatus) {
		changed := sdk.ConditionsChanged(originalValues, newValues)

		Expect(changed).To(BeTrue())
	},
		Entry("changed value with only one existing",
			map[v1.ConditionType]v12.ConditionStatus{v1.ConditionAvailable: v12.ConditionTrue},
			map[v1.ConditionType]v12.ConditionStatus{v1.ConditionAvailable: v12.ConditionFalse},
		),
		Entry("changed value with more existing",
			map[v1.ConditionType]v12.ConditionStatus{v1.ConditionAvailable: v12.ConditionTrue, v1.ConditionDegraded: v12.ConditionFalse},
			map[v1.ConditionType]v12.ConditionStatus{v1.ConditionAvailable: v12.ConditionFalse, v1.ConditionDegraded: v12.ConditionFalse},
		),
		Entry("added value to no values",
			make(map[v1.ConditionType]v12.ConditionStatus),
			map[v1.ConditionType]v12.ConditionStatus{v1.ConditionAvailable: v12.ConditionTrue, v1.ConditionDegraded: v12.ConditionFalse},
		),
		Entry("added value",
			map[v1.ConditionType]v12.ConditionStatus{v1.ConditionAvailable: v12.ConditionTrue},
			map[v1.ConditionType]v12.ConditionStatus{v1.ConditionAvailable: v12.ConditionTrue, v1.ConditionDegraded: v12.ConditionFalse},
		),
		Entry("removed value",
			map[v1.ConditionType]v12.ConditionStatus{v1.ConditionAvailable: v12.ConditionTrue, v1.ConditionDegraded: v12.ConditionFalse},
			map[v1.ConditionType]v12.ConditionStatus{v1.ConditionAvailable: v12.ConditionTrue},
		),
	)
})
