package v2

import (
	"log"

	"github.com/observatorium/observatorium/configuration_go/k8sutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const defaultTerminationGracePeriodSeconds = 120

type StatefulSetBuilder struct {
	MetaConfig                    MetaConfig
	Containers                    []ContainerProvider
	Replicas                      int32
	TerminationGracePeriodSeconds int64
	Affinity                      *corev1.Affinity
	SecurityContext               *corev1.PodSecurityContext
	ServiceAccountName            string
}

func (s *StatefulSetBuilder) makeAffinity() *corev1.Affinity {
	if s.Affinity != nil {
		return s.Affinity
	}

	if s.MetaConfig.Labels == nil {
		s.MetaConfig.Labels = make(map[string]string)
	}

	for _, k := range []string{k8sutil.NameLabel, k8sutil.InstanceLabel} {
		if _, ok := s.MetaConfig.Labels[k]; !ok {
			log.Printf("Warning: key %s not found in compactor meta labels", k)
		}
	}

	labelSelectors := map[string]string{
		k8sutil.NameLabel:     s.MetaConfig.Labels[k8sutil.NameLabel],
		k8sutil.InstanceLabel: s.MetaConfig.Labels[k8sutil.InstanceLabel],
	}

	namespaces := []string{s.MetaConfig.Namespace}

	return NewAntiAffinity(namespaces, labelSelectors)
}

func (s *StatefulSetBuilder) makeSecurityContext() corev1.PodSecurityContext {
	if s.SecurityContext != nil {
		return *s.SecurityContext
	}

	return GetDefaultSecurityContext()
}

func (s *StatefulSetBuilder) makeTermintationGracePeriodSeconds() int64 {
	if s.TerminationGracePeriodSeconds != 0 {
		return s.TerminationGracePeriodSeconds
	}

	return defaultTerminationGracePeriodSeconds
}

func (s *StatefulSetBuilder) MakePod() *Pod {
	pod := &Pod{
		ContainerProviders:            s.Containers,
		TerminationGracePeriodSeconds: s.makeTermintationGracePeriodSeconds(),
		Affinity:                      s.makeAffinity(),
		SecurityContext:               s.makeSecurityContext(),
		ServiceAccountName:            s.ServiceAccountName,
	}

	return pod
}

func (s *StatefulSetBuilder) MakeManifest() runtime.Object {
	statefulSet := &StatefulSet{
		MetaConfig: s.MetaConfig,
		Replicas:   s.Replicas,
		Pod:        s.MakePod(),
	}

	return statefulSet.MakeManifest()
}
