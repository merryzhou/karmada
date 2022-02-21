// Code generated by informer-gen. DO NOT EDIT.

package externalversions

import (
	"fmt"

	v1alpha1 "github.com/karmada-io/karmada/pkg/apis/cluster/v1alpha1"
	configv1alpha1 "github.com/karmada-io/karmada/pkg/apis/config/v1alpha1"
	networkingv1alpha1 "github.com/karmada-io/karmada/pkg/apis/networking/v1alpha1"
	policyv1alpha1 "github.com/karmada-io/karmada/pkg/apis/policy/v1alpha1"
	workv1alpha1 "github.com/karmada-io/karmada/pkg/apis/work/v1alpha1"
	v1alpha2 "github.com/karmada-io/karmada/pkg/apis/work/v1alpha2"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	cache "k8s.io/client-go/tools/cache"
)

// GenericInformer is type of SharedIndexInformer which will locate and delegate to other
// sharedInformers based on type
type GenericInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() cache.GenericLister
}

type genericInformer struct {
	informer cache.SharedIndexInformer
	resource schema.GroupResource
}

// Informer returns the SharedIndexInformer.
func (f *genericInformer) Informer() cache.SharedIndexInformer {
	return f.informer
}

// Lister returns the GenericLister.
func (f *genericInformer) Lister() cache.GenericLister {
	return cache.NewGenericLister(f.Informer().GetIndexer(), f.resource)
}

// ForResource gives generic access to a shared informer of the matching type
// TODO extend this to unknown resources with a client pool
func (f *sharedInformerFactory) ForResource(resource schema.GroupVersionResource) (GenericInformer, error) {
	switch resource {
	// Group=cluster.karmada.io, Version=v1alpha1
	case v1alpha1.SchemeGroupVersion.WithResource("clusters"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Cluster().V1alpha1().Clusters().Informer()}, nil

		// Group=config.karmada.io, Version=v1alpha1
	case configv1alpha1.SchemeGroupVersion.WithResource("resourceinterpreterwebhookconfigurations"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Config().V1alpha1().ResourceInterpreterWebhookConfigurations().Informer()}, nil

		// Group=networking.karmada.io, Version=v1alpha1
	case networkingv1alpha1.SchemeGroupVersion.WithResource("multiclusteringresses"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Networking().V1alpha1().MultiClusterIngresses().Informer()}, nil

		// Group=policy.karmada.io, Version=v1alpha1
	case policyv1alpha1.SchemeGroupVersion.WithResource("clusteroverridepolicies"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Policy().V1alpha1().ClusterOverridePolicies().Informer()}, nil
	case policyv1alpha1.SchemeGroupVersion.WithResource("clusterpropagationpolicies"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Policy().V1alpha1().ClusterPropagationPolicies().Informer()}, nil
	case policyv1alpha1.SchemeGroupVersion.WithResource("overridepolicies"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Policy().V1alpha1().OverridePolicies().Informer()}, nil
	case policyv1alpha1.SchemeGroupVersion.WithResource("propagationpolicies"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Policy().V1alpha1().PropagationPolicies().Informer()}, nil

		// Group=work.karmada.io, Version=v1alpha1
	case workv1alpha1.SchemeGroupVersion.WithResource("clusterresourcebindings"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Work().V1alpha1().ClusterResourceBindings().Informer()}, nil
	case workv1alpha1.SchemeGroupVersion.WithResource("resourcebindings"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Work().V1alpha1().ResourceBindings().Informer()}, nil
	case workv1alpha1.SchemeGroupVersion.WithResource("works"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Work().V1alpha1().Works().Informer()}, nil

		// Group=work.karmada.io, Version=v1alpha2
	case v1alpha2.SchemeGroupVersion.WithResource("clusterresourcebindings"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Work().V1alpha2().ClusterResourceBindings().Informer()}, nil
	case v1alpha2.SchemeGroupVersion.WithResource("resourcebindings"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Work().V1alpha2().ResourceBindings().Informer()}, nil

	}

	return nil, fmt.Errorf("no informer found for %v", resource)
}
