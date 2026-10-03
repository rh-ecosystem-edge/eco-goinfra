package sriov

import (
	"strings"

	srIovV1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func conditionMatches(conditions []metav1.Condition, generation int64, expected metav1.Condition) bool {
	for _, condition := range conditions {
		if !isConditionGenerationCurrent(generation, &condition) {
			continue
		}

		if expected.Type != "" && condition.Type != expected.Type {
			continue
		}

		if expected.Status != "" && condition.Status != expected.Status {
			continue
		}

		if expected.Reason != "" && condition.Reason != expected.Reason {
			continue
		}

		if expected.Message != "" && !strings.Contains(condition.Message, expected.Message) {
			continue
		}

		if expected.ObservedGeneration != 0 &&
			condition.ObservedGeneration != expected.ObservedGeneration {
			continue
		}

		return true
	}

	return false
}

func isConditionGenerationCurrent(generation int64, condition *metav1.Condition) bool {
	if condition == nil {
		return false
	}

	return condition.ObservedGeneration == generation
}

const conditionDegraded = "Degraded"

func isReadyConditionTrue(conditions []metav1.Condition, generation int64) bool {
	ready := meta.FindStatusCondition(conditions, srIovV1.ConditionReady)

	return isConditionGenerationCurrent(generation, ready) && ready.Status == metav1.ConditionTrue
}

// isPoolConfigStable matches TC-02: Ready=True, Progressing != True, Degraded != True.
// Missing Progressing/Degraded is allowed (nil is not True), matching merged #1161.
func isPoolConfigStable(pool *srIovV1.SriovNetworkPoolConfig) bool {
	if pool == nil {
		return false
	}

	generation := pool.GetGeneration()

	if !isReadyConditionTrue(pool.Status.Conditions, generation) {
		return false
	}

	progressing := meta.FindStatusCondition(pool.Status.Conditions, srIovV1.ConditionProgressing)
	if progressing != nil && isConditionGenerationCurrent(generation, progressing) &&
		progressing.Status == metav1.ConditionTrue {
		return false
	}

	degraded := meta.FindStatusCondition(pool.Status.Conditions, conditionDegraded)
	if degraded != nil && isConditionGenerationCurrent(generation, degraded) &&
		degraded.Status == metav1.ConditionTrue {
		return false
	}

	return true
}
