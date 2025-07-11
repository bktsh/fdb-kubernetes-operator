/*
 * exec_tests.go
 *
 * This source file is part of the FoundationDB open source project
 *
 * Copyright 2021 Apple Inc. and the FoundationDB project authors
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

package cmd

import (
	"bytes"
	"context"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	fdbv1beta2 "github.com/FoundationDB/fdb-kubernetes-operator/v2/api/v1beta2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

var _ = Describe("[plugin] exec command", func() {
	When("running exec command", func() {
		type testCase struct {
			ClusterName   string
			Context       string
			Command       []string
			ExpectedArgs  []string
			ExpectedError string
		}

		BeforeEach(func() {
			Expect(k8sClient.Create(context.TODO(), &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "storage-1",
					Namespace: namespace,
					Labels: map[string]string{
						fdbv1beta2.FDBProcessClassLabel: string(fdbv1beta2.ProcessClassStorage),
						fdbv1beta2.FDBClusterLabel:      clusterName,
					},
				},
			})).NotTo(HaveOccurred())
		})

		DescribeTable("should execute the provided command",
			func(input testCase) {
				outBuffer := bytes.Buffer{}
				errBuffer := bytes.Buffer{}
				inBuffer := bytes.Buffer{}

				rootCmd := NewRootCmd(
					genericclioptions.IOStreams{In: &inBuffer, Out: &outBuffer, ErrOut: &errBuffer},
					&MockVersionChecker{},
				)
				Expect(
					runExec(rootCmd, k8sClient, cluster, &rest.Config{}, input.Command),
				).NotTo(HaveOccurred())
			},
			Entry("Exec into instance with valid pod",
				testCase{
					ExpectedArgs: []string{
						"--namespace",
						"test",
						"exec",
						"-it",
						"storage-1",
						"--",
						"bash",
					},
				}),
			Entry("Exec into instance with explicit context",
				testCase{
					Context: "remote-kc",
					ExpectedArgs: []string{
						"--context",
						"remote-kc",
						"--namespace",
						"test",
						"exec",
						"-it",
						"storage-1",
						"--",
						"bash",
					},
				}),
		)
	})
})
