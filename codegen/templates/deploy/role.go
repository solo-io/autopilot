package deploy

import (
	"github.com/solo-io/autopilot/codegen/model"
	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sort"
)

func Role(data *model.TemplateData) runtime.Object {
	return role(data)
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
		verbs = append(verbs, "create", "update", "delete")
	}
	return verbs
}

func role(data *model.TemplateData) *v1.Role {
	requiredPermissions := make(map[model.Parameter]permission)

	setRead := func(param model.Parameter) {
		perm := requiredPermissions[param]
		perm.read = true
		requiredPermissions[param] = perm
	}
	setWrite := func(param model.Parameter) {
		perm := requiredPermissions[param]
		perm.write = true
		requiredPermissions[param] = perm
	}

	for _, phase := range data.Phases {
		for _, param := range phase.Inputs {
			if param == model.Metrics {
				continue
			}
			setRead(param)
		}
		for _, param := range phase.Outputs {
			setWrite(param)
		}
	}

	// always require read on pods and configmaps
	setRead(model.Pods)
	setRead(model.ConfigMaps)
	setWrite(model.ConfigMaps)

	var rules []v1.PolicyRule
	for param, perm := range requiredPermissions {
		verbs := perm.verbs()
		if len(verbs) == 0 {
			continue
		}
		rules = append(rules, v1.PolicyRule{
			Verbs:     verbs,
			APIGroups: []string{model.ParamApiVersion(param)},
			Resources: []string{param.String()},
		})
	}

	sort.SliceStable(rules, func(i, j int) bool {
		return rules[i].APIGroups[0] < rules[j].APIGroups[0] &&
			rules[i].Resources[0] < rules[j].Resources[0]
	})

	rules = append(rules, v1.PolicyRule{
		Verbs:     []string{"*"},
		APIGroups: []string{data.ApiVersion},
		Resources: []string{data.KindLowerPlural},
	})

	return &v1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name: data.OperatorName,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1.SchemeGroupVersion.String(),
			Kind:       "Role",
		},
		Rules: rules,
	}
}
