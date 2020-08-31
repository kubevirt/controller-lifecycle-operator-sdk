package sdk

import (
	"github.com/kubevirt/controller-lifecycle-operator-sdk/pkg/sdk/api"
	v1 "github.com/openshift/custom-resource-status/conditions/v1"
	v12 "k8s.io/api/core/v1"
)

// IsUpgrading checks whether cr status represents upgrade in progress
func IsUpgrading(crStatus *api.Status) bool {
	deploying := crStatus.Phase == api.PhaseDeploying
	return (crStatus.ObservedVersion != "" || !deploying) && crStatus.ObservedVersion != crStatus.TargetVersion
}

// GetConditionValues gets the conditions and put them into a map for easy comparison
func GetConditionValues(conditionList []v1.Condition) map[v1.ConditionType]v12.ConditionStatus {
	result := make(map[v1.ConditionType]v12.ConditionStatus)
	for _, cond := range conditionList {
		result[cond.Type] = cond.Status
	}
	return result
}

// ConditionsChanged compares condition maps and return true if any of the conditions changed, false otherwise.
func ConditionsChanged(originalValues, newValues map[v1.ConditionType]v12.ConditionStatus) bool {
	if len(originalValues) != len(newValues) {
		return true
	}
	for k, v := range newValues {
		oldV, ok := originalValues[k]
		if !ok || oldV != v {
			return true
		}
	}
	return false
}

// MarkCrHealthyMessage marks the passed in CR as healthy. The CR object needs to be updated by the caller afterwards.
// Healthy means the following status conditions are set:
// ApplicationAvailable: true
// Progressing: false
// Degraded: false
func MarkCrHealthyMessage(crStatus *api.Status, reason, message string) {
	v1.SetStatusCondition(&crStatus.Conditions, v1.Condition{
		Type:    v1.ConditionAvailable,
		Status:  v12.ConditionTrue,
		Reason:  reason,
		Message: message,
	})
	v1.SetStatusCondition(&crStatus.Conditions, v1.Condition{
		Type:   v1.ConditionProgressing,
		Status: v12.ConditionFalse,
	})
	v1.SetStatusCondition(&crStatus.Conditions, v1.Condition{
		Type:   v1.ConditionDegraded,
		Status: v12.ConditionFalse,
	})
}

// MarkCrUpgradeHealingDegraded marks the passed CR as upgrading and degraded. The CR object needs to be updated by the caller afterwards.
// Failed means the following status conditions are set:
// ApplicationAvailable: true
// Progressing: true
// Degraded: true
func MarkCrUpgradeHealingDegraded(crStatus *api.Status, reason, message string) {
	v1.SetStatusCondition(&crStatus.Conditions, v1.Condition{
		Type:   v1.ConditionAvailable,
		Status: v12.ConditionTrue,
	})
	v1.SetStatusCondition(&crStatus.Conditions, v1.Condition{
		Type:   v1.ConditionProgressing,
		Status: v12.ConditionTrue,
	})
	v1.SetStatusCondition(&crStatus.Conditions, v1.Condition{
		Type:    v1.ConditionDegraded,
		Status:  v12.ConditionTrue,
		Reason:  reason,
		Message: message,
	})
}

// MarkCrFailed marks the passed CR as failed and requiring human intervention. The CR object needs to be updated by the caller afterwards.
// Failed means the following status conditions are set:
// ApplicationAvailable: false
// Progressing: false
// Degraded: true
func MarkCrFailed(crStatus *api.Status, reason, message string) {
	v1.SetStatusCondition(&crStatus.Conditions, v1.Condition{
		Type:   v1.ConditionAvailable,
		Status: v12.ConditionFalse,
	})
	v1.SetStatusCondition(&crStatus.Conditions, v1.Condition{
		Type:   v1.ConditionProgressing,
		Status: v12.ConditionFalse,
	})
	v1.SetStatusCondition(&crStatus.Conditions, v1.Condition{
		Type:    v1.ConditionDegraded,
		Status:  v12.ConditionTrue,
		Reason:  reason,
		Message: message,
	})
}

// MarkCrFailedHealing marks the passed CR as failed and healing. The CR object needs to be updated by the caller afterwards.
// FailedAndHealing means the following status conditions are set:
// ApplicationAvailable: false
// Progressing: true
// Degraded: true
func MarkCrFailedHealing(crStatus *api.Status, reason, message string) {
	v1.SetStatusCondition(&crStatus.Conditions, v1.Condition{
		Type:   v1.ConditionAvailable,
		Status: v12.ConditionFalse,
	})
	v1.SetStatusCondition(&crStatus.Conditions, v1.Condition{
		Type:   v1.ConditionProgressing,
		Status: v12.ConditionTrue,
	})
	v1.SetStatusCondition(&crStatus.Conditions, v1.Condition{
		Type:    v1.ConditionDegraded,
		Status:  v12.ConditionTrue,
		Reason:  reason,
		Message: message,
	})
}

// MarkCrDeploying marks the passed CR as currently deploying. The CR object needs to be updated by the caller afterwards.
// Deploying means the following status conditions are set:
// ApplicationAvailable: false
// Progressing: true
// Degraded: false
func MarkCrDeploying(crStatus *api.Status, reason, message string) {
	v1.SetStatusCondition(&crStatus.Conditions, v1.Condition{
		Type:   v1.ConditionAvailable,
		Status: v12.ConditionFalse,
	})
	v1.SetStatusCondition(&crStatus.Conditions, v1.Condition{
		Type:    v1.ConditionProgressing,
		Status:  v12.ConditionTrue,
		Reason:  reason,
		Message: message,
	})
	v1.SetStatusCondition(&crStatus.Conditions, v1.Condition{
		Type:   v1.ConditionDegraded,
		Status: v12.ConditionFalse,
	})
}
