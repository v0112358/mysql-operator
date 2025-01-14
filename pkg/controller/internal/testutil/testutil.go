/*
Copyright 2018 Pressinfra SRL

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

//nolint: golint
package testutil

import (
	"context"
	"io"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	gomegatypes "github.com/onsi/gomega/types"

	// loggging
	"github.com/go-logr/logr"

	core "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	drainTimeout = 200 * time.Millisecond
)

// DrainChan drains the request chan time for drainTimeout
func DrainChan(requests <-chan reconcile.Request) {
	for {
		select {
		case <-requests:
			continue
		case <-time.After(drainTimeout):
			return
		}
	}
}

// NewTestLogger returns a logger good for tests
func NewTestLogger(w io.Writer) logr.Logger {
	klog.SetOutput(w)
	return klogr.New()
}

// SetupTestReconcile returns a reconcile.Reconcile implementation that delegates to inner and
// writes the request to requests after Reconcile is finished.
func SetupTestReconcile(inner reconcile.Reconciler) (reconcile.Reconciler, chan reconcile.Request) {
	requests := make(chan reconcile.Request)
	fn := reconcile.Func(func(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
		result, err := inner.Reconcile(ctx, req)
		requests <- req
		return result, err
	})
	return fn, requests
}

// StartTestManager adds recFn
func StartTestManager(mgr manager.Manager) (context.Context, func()) {
	ctx, cancel := context.WithCancel(context.TODO())

	go func() {
		defer GinkgoRecover()
		Expect(mgr.Start(ctx)).NotTo(HaveOccurred())
	}()
	return ctx, cancel
}

// PodHaveCondition is a helper func that returns a matcher to check for an
// existing condition in condition list list
func PodHaveCondition(condType core.PodConditionType, status core.ConditionStatus) gomegatypes.GomegaMatcher {
	return PointTo(MatchFields(IgnoreExtras, Fields{
		"Status": MatchFields(IgnoreExtras, Fields{
			"Conditions": ContainElement(MatchFields(IgnoreExtras, Fields{
				"Type":   Equal(condType),
				"Status": Equal(status),
			})),
		}),
	}))
}
