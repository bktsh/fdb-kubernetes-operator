/*
 * operator_migration_test.go
 *
 * This source file is part of the FoundationDB open source project
 *
 * Copyright 2023 Apple Inc. and the FoundationDB project authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package operatormigration

/*
This test suite includes test that make sure that the migrations and exclusion strategy of the operator is working as
expected under different scenarios.
*/

import (
	"log"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/FoundationDB/fdb-kubernetes-operator/e2e/fixtures"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	factory     *fixtures.Factory
	fdbCluster  *fixtures.FdbCluster
	testOptions *fixtures.FactoryOptions
)

func init() {
	testOptions = fixtures.InitFlags()
}

var _ = BeforeSuite(func() {
	factory = fixtures.CreateFactory(testOptions)
	fdbCluster = factory.CreateFdbCluster(
		fixtures.DefaultClusterConfig(false),
		factory.GetClusterOptions()...,
	)
	// Load some data into the cluster.
	factory.CreateDataLoaderIfAbsent(fdbCluster)
})

var _ = AfterSuite(func() {
	if CurrentSpecReport().Failed() {
		log.Printf("failed due to %s", CurrentSpecReport().FailureMessage())
	}
	factory.Shutdown()
})

var _ = Describe("Operator Migrations", Label("e2e", "pr"), func() {
	AfterEach(func() {
		if CurrentSpecReport().Failed() {
			factory.DumpState(fdbCluster)
		}
		Expect(fdbCluster.WaitForReconciliation()).ToNot(HaveOccurred())
	})

	When("a migration is triggered and the namespace quota is limited", func() {
		prefix := "banana"

		BeforeEach(func() {
			processCounts, err := fdbCluster.GetCluster().GetProcessCountsWithDefaults()
			Expect(err).NotTo(HaveOccurred())
			// Create Quota to limit the additional Pods that can be created to 5, the actual value here is 7 ,because we run
			// 2 Operator Pods.
			Expect(factory.CreateIfAbsent(&corev1.ResourceQuota{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testing-quota",
					Namespace: fdbCluster.Namespace(),
				},
				Spec: corev1.ResourceQuotaSpec{
					Hard: corev1.ResourceList{
						"count/pods": resource.MustParse(strconv.Itoa(processCounts.Total() + 7)),
					},
				},
			})).NotTo(HaveOccurred())

			Expect(fdbCluster.SetProcessGroupPrefix(prefix)).NotTo(HaveOccurred())
		})

		It("should add the prefix to all instances", func() {
			lastForcedReconciliationTime := time.Now()
			forceReconcileDuration := 4 * time.Minute

			Eventually(func(g Gomega) bool {
				// Force a reconcile if needed to make sure we speed up the reconciliation if needed.
				if time.Since(lastForcedReconciliationTime) >= forceReconcileDuration {
					fdbCluster.ForceReconcile()
					lastForcedReconciliationTime = time.Now()
				}

				// Check if all process groups are migrated
				for _, processGroup := range fdbCluster.GetCluster().Status.ProcessGroups {
					if processGroup.IsMarkedForRemoval() && processGroup.IsExcluded() {
						continue
					}
					g.Expect(string(processGroup.ProcessGroupID)).To(HavePrefix(prefix))
				}

				return true
			}).WithTimeout(40 * time.Minute).WithPolling(5 * time.Second).Should(BeTrue())
			Expect(fdbCluster.WaitForReconciliation()).NotTo(HaveOccurred())
		})
	})
})
