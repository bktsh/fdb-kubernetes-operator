/*
 * service_helper.go
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

package internal

import (
	"github.com/FoundationDB/fdb-kubernetes-operator/api/v1beta1"
	v1 "k8s.io/api/core/v1"
)

// GetHeadlessService builds a headless service for a FoundationDB cluster.
func GetHeadlessService(cluster *v1beta1.FoundationDBCluster) *v1.Service {
	headless := cluster.Spec.Routing.HeadlessService
	if headless == nil || !*headless {
		return nil
	}

	service := &v1.Service{
		ObjectMeta: GetObjectMetadata(cluster, nil, "", ""),
	}
	service.ObjectMeta.Name = cluster.ObjectMeta.Name
	service.Spec.ClusterIP = "None"
	service.Spec.Selector = cluster.Spec.LabelConfig.MatchLabels

	return service
}