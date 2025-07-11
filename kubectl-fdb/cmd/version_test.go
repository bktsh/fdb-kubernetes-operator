/*
 * version.go
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
	"fmt"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
)

var _ = Describe("[plugin] version command", func() {
	When("running the version command with client only", func() {
		var outBuffer bytes.Buffer
		var errBuffer bytes.Buffer
		var inBuffer bytes.Buffer

		BeforeEach(func() {
			// We use these buffers to check the input/output
			outBuffer = bytes.Buffer{}
			errBuffer = bytes.Buffer{}
			inBuffer = bytes.Buffer{}

			rootCmd := NewRootCmd(
				genericclioptions.IOStreams{In: &inBuffer, Out: &outBuffer, ErrOut: &errBuffer},
				&MockVersionChecker{},
			)

			args := []string{"version", "--client-only"}
			rootCmd.SetArgs(args)

			err := rootCmd.Execute()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should print out the client version", func() {
			Expect(outBuffer.String()).To(HavePrefix("kubectl-fdb build information"))
		})
	})

	When("running the version command", func() {
		operatorName := "fdb-operator"

		type testCase struct {
			deployment    *appsv1.Deployment
			expected      string
			expectedError error
			hasError      bool
		}

		DescribeTable("should return the correct version",
			func(input testCase) {
				Expect(k8sClient.Create(context.TODO(), input.deployment))

				operatorVersion, err := version(k8sClient, operatorName, "default", "manager")
				if input.hasError {
					Expect(err).To(Equal(input.expectedError))
				} else {
					Expect(err).To(BeNil())
				}

				Expect(operatorVersion).To(Equal(input.expected))
			},
			Entry("Single container",
				testCase{
					expected: "0.27.0",
					deployment: &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "default",
							Name:      operatorName,
						},
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name:  "manager",
											Image: "foundationdb/fdb-kubernetes-operator:0.27.0",
										},
									},
								},
							},
						},
					},
					expectedError: nil,
					hasError:      false,
				}),
			Entry("Multi container",
				testCase{
					expected: "0.27.0",
					deployment: &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "default",
							Name:      operatorName,
						},
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name:  "test",
											Image: "test:1337",
										},
										{
											Name:  "test2",
											Image: "test:1337-2",
										},
										{
											Name:  "manager",
											Image: "foundationdb/fdb-kubernetes-operator:0.27.0",
										},
									},
								},
							},
						},
					},
					expectedError: nil,
					hasError:      false,
				}),
			Entry("No container",
				testCase{
					expected: "",
					deployment: &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "default",
							Name:      operatorName,
						},
						Spec: appsv1.DeploymentSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{},
									},
								},
							},
						},
					},
					expectedError: fmt.Errorf(
						"could not find container: manager in default/fdb-operator",
					),
					hasError: true,
				}),
		)
	})

	When("running the version command with old version", func() {
		var outBuffer bytes.Buffer
		var errBuffer bytes.Buffer
		var inBuffer bytes.Buffer

		AfterEach(func() {
			pluginVersion = "latest"
		})
		BeforeEach(func() {
			pluginVersion = "1.0.0"
			// We use these buffers to check the input/output
			outBuffer = bytes.Buffer{}
			errBuffer = bytes.Buffer{}
			inBuffer = bytes.Buffer{}

			rootCmd := NewRootCmd(
				genericclioptions.IOStreams{In: &inBuffer, Out: &outBuffer, ErrOut: &errBuffer},
				&MockVersionChecker{MockedVersion: "2.0.0"},
			)

			args := []string{"version", "--client-only"}
			rootCmd.SetArgs(args)

			err := rootCmd.Execute()
			Expect(err).To(HaveOccurred())
		})

		It("should print out the client version", func() {
			Expect(outBuffer.String()).To(ContainSubstring(
				"kubectl-fdb plugin is not up-to-date, please install the latest version and try again!",
			))
		})
	})
})
