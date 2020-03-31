/*
Copyright 2020 The Kruise Authors.

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

package migration

import (
	"github.com/openkruise/kruise-tools/pkg/api"
	"k8s.io/apimachinery/pkg/types"
)

type Control interface {
	Submit(src api.ResourceRef, dst api.ResourceRef, opts Options) (Result, error)
	Query(ID types.UID) (Result, error)
}

type Options struct {
	// Specify Replicas that should be migrated.
	// Default to migrate all replicas
	Replicas *int32
	// The maximum number of pods that can be scheduled above the desired number of pods.
	// This can not be 0 if MaxUnavailable is 0.
	// Defaults to 1.
	MaxSurge *int32
	// TimeoutSeconds indicates the timeout seconds that migration exceeded.
	// Defaults to no limited.
	TimeoutSeconds *int32
}

type Result struct {
	ID      types.UID
	State   MigrateState
	Message string

	SrcMigratedReplicas int32
	DstMigratedReplicas int32
}

type MigrateState string

const (
	MigrateExecuting MigrateState = "Executing"
	MigrateSucceeded MigrateState = "Succeeded"
	MigrateFailed    MigrateState = "Failed"
)
