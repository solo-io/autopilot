package render

import (
	"github.com/solo-io/autopilot/codegen/templates/deploy"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
	"strings"
)

// creates a k8s resource for a group
// this gets turned into a k8s manifest file
type MakeResourceFunc func(group Group) []metav1.Object

// renders kubernetes from templates
type ManifestsRenderer struct {
	AppName       string // used for labeling
	ResourceFuncs map[OutFile]MakeResourceFunc
	ManifestDir   string
}

func RenderManifests(appName, manifestDir string, grp Group) ([]OutFile, error) {
	defaultManifestsRenderer := ManifestsRenderer{
		AppName:     appName,
		ManifestDir: manifestDir,
		ResourceFuncs: map[OutFile]MakeResourceFunc{
			{
				Path: "crds.yaml",
			}: deploy.CustomResourceDefinitions,
		},
	}
	return defaultManifestsRenderer.RenderManifests(grp)
}

func (r ManifestsRenderer) RenderManifests(grp Group) ([]OutFile, error) {
	var renderedFiles []OutFile
	for out, mkFunc := range r.ResourceFuncs {
		content, err := r.renderManifest(mkFunc, grp)
		if err != nil {
			return nil, err
		}
		out.Content = content
		out.Path = r.ManifestDir + "/" + grp.Group + "_" + grp.Version + "_" + out.Path
		renderedFiles = append(renderedFiles, out)
	}
	return renderedFiles, nil
}

func (r ManifestsRenderer) renderManifest(mk MakeResourceFunc, group Group) (string, error) {
	objs := mk(group)

	var objManifests []string
	for _, obj := range objs {
		manifest, err := r.marshalObjToYaml(obj)
		if err != nil {
			return "", err
		}
		objManifests = append(objManifests, manifest)
	}

	return strings.Join(objManifests, "\n---\n"), nil
}

func (r ManifestsRenderer) marshalObjToYaml(obj metav1.Object) (string, error) {
	labels := obj.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}

	labels["app"] = r.AppName
	labels["app.kubernetes.io/name"] = r.AppName

	obj.SetLabels(labels)

	yam, err := yaml.Marshal(obj)
	if err != nil {
		return "", err
	}
	var v map[string]interface{}

	if err := yaml.Unmarshal(yam, &v); err != nil {
		return "", err
	}

	delete(v, "status")
	// why do we have to do this? Go problem???
	meta := v["metadata"].(map[string]interface{})

	delete(meta, "creationTimestamp")
	v["metadata"] = meta

	if spec, ok := v["spec"].(map[string]interface{}); ok {
		if template, ok := spec["template"].(map[string]interface{}); ok {
			if meta, ok := template["metadata"].(map[string]interface{}); ok {
				delete(meta, "creationTimestamp")
				template["metadata"] = meta
				spec["template"] = template
				v["spec"] = spec
			}
		}
	}

	yam, err = yaml.Marshal(v)

	return string(yam), err
}
