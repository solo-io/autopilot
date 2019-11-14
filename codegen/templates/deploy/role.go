package deploy

import (
	"sort"

	"github.com/solo-io/autopilot/codegen/model"
	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Role(data *model.ProjectData) runtime.Object {
	return role(data)
}

func role(data *model.ProjectData) *v1.Role {
	return &v1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name: data.OperatorName,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1.SchemeGroupVersion.String(),
			Kind:       "Role",
		},
		Rules: rules(data),
	}
}

func ClusterRole(data *model.ProjectData) runtime.Object {
	return clusterRole(data)
}

func clusterRole(data *model.ProjectData) *v1.ClusterRole {
	return &v1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: data.OperatorName,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1.SchemeGroupVersion.String(),
			Kind:       "ClusterRole",
		},
		Rules: rules(data),
	}
}

type permission struct {
	read  bool
	write bool
}

func (p permission) verbs() []string {
	var verbs []string
	if p.read {
		verbs = append(verbs, "get", "list", "watch")
	}
	if p.write {
		// writing currently requires reading as we use the EzKube.Ensure method
		// delete is added here as well as controllers may need to remove resources they write
		verbs = []string{"*"}
	}
	return verbs
}

func rules(data *model.ProjectData) []v1.PolicyRule {
	type paramPermission struct {
		model.Parameter
		permission
	}
	requiredPermissions := make(map[string]paramPermission)

	setRead := func(param model.Parameter) {
		perm := requiredPermissions[param.String()]
		perm.Parameter = param
		perm.read = true
		requiredPermissions[param.String()] = perm
	}
	setWrite := func(param model.Parameter) {
		perm := requiredPermissions[param.String()]
		perm.Parameter = param
		perm.write = true
		requiredPermissions[param.String()] = perm
	}

	for _, phase := range data.Phases {
		for _, param := range phase.Inputs {
			if param.Equals(model.Metrics) {
				continue
			}
			setRead(param)
		}
		for _, param := range phase.Outputs {
			setWrite(param)
		}
	}

	// these permissions required by controller-runtime
	setRead(model.ReplicaSets)
	setRead(model.Pods)

	// required by leader election
	setWrite(model.ConfigMaps)
	setWrite(model.Events)

	var rules []v1.PolicyRule
	for _, param := range requiredPermissions {
		verbs := param.verbs()
		if len(verbs) == 0 {
			continue
		}
		rules = append(rules, v1.PolicyRule{
			Verbs:     verbs,
			APIGroups: []string{param.ApiGroup},
			Resources: []string{param.String()},
		})
	}

	sort.SliceStable(rules, func(i, j int) bool {
		if rules[i].APIGroups[0] < rules[j].APIGroups[0] {
			return true
		}
		if rules[i].Resources[0] < rules[j].Resources[0] {
			return true
		}
		return rules[i].Verbs[0] < rules[i].Verbs[0]
	})

	rules = append(rules, v1.PolicyRule{
		Verbs:     []string{"get", "list", "watch"},
		APIGroups: []string{data.Group},
		Resources: []string{
			data.KindLowerPlural,
		},
	}, v1.PolicyRule{
		Verbs:     []string{"update"},
		APIGroups: []string{data.Group},
		Resources: []string{
			data.KindLowerPlural + "/status",
		},
	})

	return rules
}
