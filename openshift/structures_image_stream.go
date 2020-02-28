package openshift

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	api "github.com/openshift/api/image/v1"
)

func flattenImageStreamSpec(in api.ImageStreamSpec, d *schema.ResourceData) ([]interface{}, error) {
	att := make(map[string]interface{})

	res, err := flattenImageStreamLookupPolicy(in.LookupPolicy)
	if err != nil {
		return nil, err
	}
	att["lookup_policy"] = res

	if in.DockerImageRepository != "" {
		att["docker_image_repository"] = in.DockerImageRepository
	}

	if len(in.Tags) > 0 {
		v, err := flattenImageStreamTags(in.Tags)
		if err != nil {
			return []interface{}{att}, err
		}
		att["tag"] = v
	}

	return []interface{}{att}, nil
}

func flattenImageStreamLookupPolicy(in api.ImageLookupPolicy) ([]interface{}, error) {
	att := make(map[string]interface{})

	if in.Local {
		att["local"] = in.Local
	}

	return []interface{}{att}, nil
}

func flattenImageStreamTags(tags []api.TagReference) ([]interface{}, error) {
	att := make([]interface{}, len(tags))
	for i, v := range tags {
		obj := map[string]interface{}{}

		if v.Name != "" {
			obj["name"] = v.Name
		}
		obj["annotations"] = v.Annotations
		if v.From != nil {
			obj["from"] = flattenObjectReference(*v.From)
		}
		if v.Reference {
			obj["reference"] = v.Reference
		}
		if v.Generation != nil {
			obj["generation"] = *v.Generation
		}

		res, err := flattenImageStreamTagImportPolicy(v.ImportPolicy)
		if err != nil {
			return nil, err
		}
		obj["import_policy"] = res

		res, err = flattenImageStreamTagReferencePolicy(v.ReferencePolicy)
		if err != nil {
			return nil, err
		}
		obj["reference_policy"] = res

		att[i] = obj
	}
	return att, nil
}

func flattenImageStreamTagImportPolicy(in api.TagImportPolicy) ([]interface{}, error) {
	att := make(map[string]interface{})

	if in.Insecure {
		att["insecure"] = in.Insecure
	}
	if in.Scheduled {
		att["scheduled"] = in.Scheduled
	}

	return []interface{}{att}, nil
}

func flattenImageStreamTagReferencePolicy(in api.TagReferencePolicy) ([]interface{}, error) {
	att := make(map[string]interface{})

	if in.Type != "" {
		att["type"] = in.Type
	}

	return []interface{}{att}, nil
}

func expandImageStreamSpec(imageStream []interface{}) (api.ImageStreamSpec, error) {
	obj := api.ImageStreamSpec{}

	if len(imageStream) == 0 || imageStream[0] == nil {
		return obj, nil
	}

	in := imageStream[0].(map[string]interface{})

	if v, ok := in["lookup_policy"].([]interface{}); ok && len(v) > 0 {
		obj.LookupPolicy = expandImageStreamLookupPolicy(v)
	}

	if v, ok := in["docker_image_repository"].(string); ok {
		obj.DockerImageRepository = v
	}

	if v, ok := in["tag"].([]interface{}); ok && len(v) > 0 {
		obj.Tags = expandImageStreamTags(v)
	}

	return obj, nil
}

func expandImageStreamLookupPolicy(l []interface{}) api.ImageLookupPolicy {
	obj := api.ImageLookupPolicy{}
	if len(l) == 0 || l[0] == nil {
		return obj
	}

	in := l[0].(map[string]interface{})

	if v, ok := in["local"].(bool); ok {
		obj.Local = v
	}

	return obj
}

func expandImageStreamTags(tags []interface{}) []api.TagReference {
	if len(tags) == 0 {
		return []api.TagReference{}
	}

	tg := make([]api.TagReference, len(tags))
	for i, c := range tags {
		m := c.(map[string]interface{})

		if value, ok := m["name"].(string); ok {
			tg[i].Name = value
		}
		if value, ok := m["annotations"].(map[string]interface{}); ok && len(value) > 0 {
			tg[i].Annotations = expandStringMap(m["annotations"].(map[string]interface{}))
		}
		if value, ok := m["from"].([]interface{}); ok && len(value) > 0 {
			tg[i].From = expandBuildConfigImageDefinitionPtr(value)
		}
		if value, ok := m["reference"].(bool); ok {
			tg[i].Reference = value
		}
		tg[i].Generation = ptrToInt64(int64(m["generation"].(int)))

		if value, ok := m["import_policy"].([]interface{}); ok && len(value) > 0 {
			tg[i].ImportPolicy = expandImageStreamTagImportPolicy(value)
		}

		if value, ok := m["reference_policy"].([]interface{}); ok && len(value) > 0 {
			tg[i].ReferencePolicy = expandImageStreamTagReferencePolicy(value)
		}
	}

	return tg
}

func expandImageStreamTagImportPolicy(l []interface{}) api.TagImportPolicy {
	obj := api.TagImportPolicy{}
	if len(l) == 0 || l[0] == nil {
		return obj
	}

	in := l[0].(map[string]interface{})

	if v, ok := in["insecure"].(bool); ok {
		obj.Insecure = v
	}

	if v, ok := in["scheduled"].(bool); ok {
		obj.Scheduled = v
	}

	return obj
}

func expandImageStreamTagReferencePolicy(l []interface{}) api.TagReferencePolicy {
	obj := api.TagReferencePolicy{}
	if len(l) == 0 || l[0] == nil {
		return obj
	}

	in := l[0].(map[string]interface{})

	if v, ok := in["type"].(string); ok {
		obj.Type = api.TagReferencePolicyType(v)
	}

	return obj
}
