/*
Copyright 2025 The kcp Authors.

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

package envtest

import (
	"time"

	conditionsv1alpha1 "github.com/kcp-dev/sdk/apis/third_party/conditions/apis/conditions/v1alpha1"
	"github.com/kcp-dev/sdk/apis/third_party/conditions/util/conditions"
	kcptestinghelpers "github.com/kcp-dev/sdk/testing/helpers"
)

// ConditionEvaluator is a helper for evaluating conditions.
type ConditionEvaluator = kcptestinghelpers.ConditionEvaluator

// Eventually asserts that given condition will be met in waitFor time, periodically checking target function
// each tick. In addition to require.Eventually, this function t.Logs the reason string value returned by the condition
// function (eventually after 20% of the wait time) to aid in debugging.
func Eventually(t TestingT, condition func() (success bool, reason string), waitFor time.Duration, tick time.Duration, msgAndArgs ...interface{}) {
	t.Helper()
	kcptestinghelpers.Eventually(t, condition, waitFor, tick, msgAndArgs...)
}

// EventuallyReady asserts that the object returned by getter() eventually has a ready condition.
func EventuallyReady(t TestingT, getter func() (conditions.Getter, error), msgAndArgs ...interface{}) {
	t.Helper()
	kcptestinghelpers.EventuallyReady(t, getter, msgAndArgs...)
}

// EventuallyCondition asserts that the object returned by getter() eventually has a condition that matches the evaluator.
func EventuallyCondition(t TestingT, getter func() (conditions.Getter, error), evaluator *ConditionEvaluator, msgAndArgs ...interface{}) {
	t.Helper()
	kcptestinghelpers.EventuallyCondition(t, getter, evaluator, msgAndArgs...)
}

// Is matches if the given condition type is True.
func Is(conditionType conditionsv1alpha1.ConditionType) *ConditionEvaluator {
	return kcptestinghelpers.Is(conditionType)
}

// IsNot matches if the given condition type is False.
func IsNot(conditionType conditionsv1alpha1.ConditionType) *ConditionEvaluator {
	return kcptestinghelpers.IsNot(conditionType)
}
