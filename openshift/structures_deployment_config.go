package openshift

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	api "github.com/openshift/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func flattenDeploymentConfigSpec(in api.DeploymentConfigSpec, d *schema.ResourceData) ([]interface{}, error) {
	att := make(map[string]interface{})
	att["min_ready_seconds"] = in.MinReadySeconds
	att["replicas"] = in.Replicas

	if in.RevisionHistoryLimit != nil {
		att["revision_history_limit"] = *in.RevisionHistoryLimit
	}

	if in.Selector != nil {
		att["selector"] = in.Selector
	}

	att["strategy"] = flattenDeploymentStrategy(in.Strategy)

	podSpec, err := flattenPodSpec(in.Template.Spec)
	if err != nil {
		return nil, err
	}
	template := make(map[string]interface{})
	template["spec"] = podSpec
	template["metadata"] = flattenMetadata(in.Template.ObjectMeta, d, "spec.0.template.0.")
	att["template"] = []interface{}{template}

	if len(in.Triggers) > 0 {
		v, err := flattenTriggers(in.Triggers)
		if err != nil {
			return []interface{}{att}, err
		}
		att["trigger"] = v
	}

	return []interface{}{att}, nil
}

func flattenTriggers(triggers []api.DeploymentTriggerPolicy) ([]interface{}, error) {
	att := make([]interface{}, len(triggers))
	for i, v := range triggers {
		obj := map[string]interface{}{}

		if v.Type != "" {
			obj["type"] = v.Type
		}
		if v.ImageChangeParams != nil {
			obj["image_change_params"] = flattenImageChangeParams(v.ImageChangeParams)
		}
		att[i] = obj
	}
	return att, nil
}

func flattenImageChangeParams(in *api.DeploymentTriggerImageChangeParams) []interface{} {
	att := make(map[string]interface{})

	if in.Automatic {
		att["automatic"] = in.Automatic
	}

	if len(in.ContainerNames) > 0 {
		att["container_names"] = in.ContainerNames
	}

	att["from"] = flattenImageChangeParamsFrom(in.From)

	return []interface{}{att}
}

func flattenImageChangeParamsFrom(in corev1.ObjectReference) []interface{} {
	att := make(map[string]interface{})

	if in.Kind != "" {
		att["kind"] = in.Kind
	}

	if in.Name != "" {
		att["name"] = in.Name
	}

	if in.Namespace != "" {
		att["namespace"] = in.Namespace
	}

	return []interface{}{att}
}

func flattenDeploymentStrategy(in api.DeploymentStrategy) []interface{} {
	att := make(map[string]interface{})
	att["active_deadline_seconds"] = in.ActiveDeadlineSeconds

	if in.Type != "" {
		att["type"] = in.Type
	}

	res, err := flattenContainerResourceRequirements(in.Resources)
	if err != nil {
		return nil
	}

	att["resources"] = res

	if in.RollingParams != nil {
		att["rolling_params"] = flattenDeploymentStrategyRollingParams(in.RollingParams)
	}
	return []interface{}{att}
}

func flattenDeploymentStrategyRollingParams(in *api.RollingDeploymentStrategyParams) []interface{} {
	att := make(map[string]interface{})

	att["interval_seconds"] = in.IntervalSeconds
	att["timeout_seconds"] = in.TimeoutSeconds
	att["update_period_seconds"] = in.UpdatePeriodSeconds

	if in.MaxUnavailable != nil {
		att["max_unavailable"] = in.MaxUnavailable.String()
	}
	if in.MaxSurge != nil {
		att["max_surge"] = in.MaxSurge.String()
	}
	return []interface{}{att}
}

func expandDeploymentConfigSpec(deployment []interface{}) (*api.DeploymentConfigSpec, error) {
	obj := &api.DeploymentConfigSpec{}

	if len(deployment) == 0 || deployment[0] == nil {
		return obj, nil
	}

	in := deployment[0].(map[string]interface{})

	obj.MinReadySeconds = int32(in["min_ready_seconds"].(int))
	obj.Paused = in["paused"].(bool)
	obj.Replicas = int32(in["replicas"].(int))
	obj.RevisionHistoryLimit = ptrToInt32(int32(in["revision_history_limit"].(int)))

	obj.Selector = expandStringMap(in["selector"].(map[string]interface{}))

	if v, ok := in["strategy"].([]interface{}); ok && len(v) > 0 {
		obj.Strategy = expandDeploymentStrategy(v)
	}

	template, err := expandPodTemplate(in["template"].([]interface{}))
	if err != nil {
		return obj, err
	}
	obj.Template = template

	if v, ok := in["trigger"].([]interface{}); ok && len(v) > 0 {
		cs, err := expandTriggers(v)
		if err != nil {
			return obj, err
		}
		obj.Triggers = cs
	}

	return obj, nil
}

func expandTriggers(triggers []interface{}) ([]api.DeploymentTriggerPolicy, error) {
	if len(triggers) == 0 {
		return []api.DeploymentTriggerPolicy{}, nil
	}

	tg := make([]api.DeploymentTriggerPolicy, len(triggers))
	for i, c := range triggers {
		m := c.(map[string]interface{})

		if value, ok := m["type"].(string); ok {
			tg[i].Type = api.DeploymentTriggerType(value)
		}

		if value, ok := m["image_change_params"].([]interface{}); ok && len(value) > 0 {
			sc, err := expandImageChangeParams(value)
			if err != nil {
				return tg, err
			}
			tg[i].ImageChangeParams = sc
		}
	}

	return tg, nil
}

func expandPodTemplate(l []interface{}) (*corev1.PodTemplateSpec, error) {
	obj := &corev1.PodTemplateSpec{}
	if len(l) == 0 || l[0] == nil {
		return obj, nil
	}
	in := l[0].(map[string]interface{})

	obj.ObjectMeta = expandMetadata(in["metadata"].([]interface{}))

	if v, ok := in["spec"].([]interface{}); ok && len(v) > 0 {
		podSpec, err := expandPodSpec(in["spec"].([]interface{}))
		if err != nil {
			return obj, err
		}
		obj.Spec = *podSpec
	}
	return obj, nil
}

func expandImageChangeParams(l []interface{}) (*api.DeploymentTriggerImageChangeParams, error) {
	obj := &api.DeploymentTriggerImageChangeParams{}
	if len(l) == 0 || l[0] == nil {
		return obj, nil
	}

	in := l[0].(map[string]interface{})

	if v, ok := in["automatic"].(bool); ok {
		obj.Automatic = v
	}

	obj.ContainerNames = expandStringSlice(in["container_names"].(*schema.Set).List())

	if v, ok := in["from"].([]interface{}); ok && len(v) > 0 {
		obj.From = expandDeploymentTriggerImageChangeParamsFrom(v)
	}

	if v, ok := in["last_triggered_image "].(string); ok {
		obj.LastTriggeredImage = v
	}

	return obj, nil
}

func expandDeploymentStrategy(l []interface{}) api.DeploymentStrategy {
	obj := api.DeploymentStrategy{}

	if len(l) == 0 || l[0] == nil {
		obj.Type = api.DeploymentStrategyTypeRolling
		return obj
	}
	in := l[0].(map[string]interface{})

	obj.ActiveDeadlineSeconds = ptrToInt64(int64(in["active_deadline_seconds"].(int)))
	if v, ok := in["type"].(string); ok {
		obj.Type = api.DeploymentStrategyType(v)
	}
	if v, ok := in["rolling_params"].([]interface{}); ok && len(v) > 0 {
		obj.RollingParams = expandRollingParamsDeployment(v)
	}
	if v, ok := in["resources"].([]interface{}); ok && len(v) > 0 {
		var err error
		resources, err := expandContainerResourceRequirements(v)
		if err != nil {
			return obj
		}
		obj.Resources = *resources
	}
	return obj
}

func expandDeploymentTriggerImageChangeParamsFrom(l []interface{}) corev1.ObjectReference {
	obj := corev1.ObjectReference{}
	if len(l) == 0 || l[0] == nil {
		return obj
	}
	in := l[0].(map[string]interface{})

	if v, ok := in["kind"].(string); ok {
		obj.Kind = v
	}

	if v, ok := in["namespace"].(string); ok {
		obj.Namespace = v
	}

	if v, ok := in["name"].(string); ok {
		obj.Name = v
	}

	return obj
}

func expandRollingParamsDeployment(l []interface{}) *api.RollingDeploymentStrategyParams {
	obj := api.RollingDeploymentStrategyParams{}
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	in := l[0].(map[string]interface{})

	obj.IntervalSeconds = ptrToInt64(int64(in["interval_seconds"].(int)))
	obj.TimeoutSeconds = ptrToInt64(int64(in["timeout_seconds"].(int)))
	obj.UpdatePeriodSeconds = ptrToInt64(int64(in["update_period_seconds"].(int)))

	if v, ok := in["max_surge"].(string); ok {
		val := intstr.Parse(v)
		obj.MaxSurge = &val
	}
	if v, ok := in["max_unavailable"].(string); ok {
		val := intstr.Parse(v)
		obj.MaxUnavailable = &val
	}

	return &obj
}
