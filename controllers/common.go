/*
Copyright 2021.

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

package controllers

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	SyncAnnotation = "config-sync-operator/sync"
)

func ListNamespaces(cltn client.Client, ctx context.Context, selector string) (corev1.NamespaceList, error) {
	parsedSelector, _ := labels.Parse(selector)

	opts := []client.ListOption{
		client.MatchingLabelsSelector{Selector: parsedSelector},
	}

	var namespaces corev1.NamespaceList

	cltn.List(ctx, &namespaces, opts...)

	return namespaces, nil
}
