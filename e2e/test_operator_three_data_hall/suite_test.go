/*
 * suite_test.go
 *
 * This source file is part of the FoundationDB open source project
 *
 * Copyright 2018-2024 Apple Inc. and the FoundationDB project authors
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

package operator

import (
	"testing"
	"time"

	"github.com/FoundationDB/fdb-kubernetes-operator/e2e/fixtures"
	"github.com/onsi/gomega"
)

func TestOperator(t *testing.T) {
	gomega.SetDefaultEventuallyTimeout(10 * time.Second)
	fixtures.SetTestSuiteName("operator-test-3dh")
	fixtures.RunGinkgoTests(t, "Operator three data hall test suite")
}