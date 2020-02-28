package openshift

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	v1 "github.com/openshift/api/route/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func flattenRouteSpec(in v1.RouteSpec) ([]interface{}, error) {
	att := make(map[string]interface{})
	if in.Host != "" {
		att["host"] = in.Host
	}
	if in.Path != "" {
		att["path"] = in.Path
	}
	if in.Port != nil {
		att["port"] = flattenRoutePort(in.Port)
	}
	if in.TLS != nil {
		att["tls"] = flattenTLSConfig(in.TLS)
	}
	if in.To.Size() > 0 {
		att["to"] = flattenRouteTargetReference(in.To)
	}
	if in.WildcardPolicy != "" {
		att["wildcard_policy"] = in.WildcardPolicy
	}
	return []interface{}{att}, nil
}

func flattenRoutePort(in *v1.RoutePort) []interface{} {
	att := make(map[string]interface{})

	if (in.TargetPort).Type == intstr.String || (in.TargetPort).Type == intstr.Int {
		att["target_port"] = in.TargetPort.String()
	}

	return []interface{}{att}
}

func flattenTLSConfig(in *v1.TLSConfig) []interface{} {
	att := make(map[string]interface{})

	if in.Termination != "" {
		att["termination"] = in.Termination
	}

	return []interface{}{att}
}

func flattenRouteTargetReference(in v1.RouteTargetReference) []interface{} {
	att := make(map[string]interface{})
	if in.Kind != "" {
		att["kind"] = in.Kind
	}
	if in.Name != "" {
		att["name"] = in.Name
	}
	if in.Weight != nil {
		att["weight"] = in.Weight
	}

	return []interface{}{att}
}

func expandRouteSpec(l []interface{}) v1.RouteSpec {
	if len(l) == 0 || l[0] == nil {
		return v1.RouteSpec{}
	}
	in := l[0].(map[string]interface{})
	obj := v1.RouteSpec{}

	if v, ok := in["host"].(string); ok {
		obj.Host = v
	}
	if v, ok := in["path"].(string); ok {
		obj.Path = v
	}
	if v, ok := in["port"].([]interface{}); ok && len(v) > 0 {
		obj.Port = expandRoutePort(v)
	}
	if v, ok := in["tls"].([]interface{}); ok && len(v) > 0 {
		obj.TLS = expandTLSConfig(v)
	}
	if v, ok := in["to"].([]interface{}); ok && len(v) > 0 {
		obj.To = expandRouteTargetReference(v)
	}

	return obj
}

func expandRoutePort(l []interface{}) *v1.RoutePort {
	if len(l) == 0 || l[0] == nil {
		return &v1.RoutePort{}
	}
	in := l[0].(map[string]interface{})
	obj := v1.RoutePort{}

	if v, ok := in["target_port"].(string); ok {
		obj.TargetPort = intstr.Parse(v)
	}

	return &obj
}

func expandTLSConfig(l []interface{}) *v1.TLSConfig {
	if len(l) == 0 || l[0] == nil {
		return &v1.TLSConfig{}
	}
	in := l[0].(map[string]interface{})
	obj := v1.TLSConfig{}

	if v, ok := in["termination"].(string); ok {
		obj.Termination = v1.TLSTerminationType(v)
	}

	return &obj
}

func expandRouteTargetReference(l []interface{}) v1.RouteTargetReference {
	if len(l) == 0 || l[0] == nil {
		return v1.RouteTargetReference{}
	}
	in := l[0].(map[string]interface{})
	obj := v1.RouteTargetReference{}
	if v, ok := in["kind"].(string); ok {
		obj.Kind = v
	}
	if v, ok := in["name"].(string); ok {
		obj.Name = v
	}

	return obj
}

func patchRouteSpec(keyPrefix, pathPrefix string, data *schema.ResourceData) (PatchOperations, error) {
	ops := make([]PatchOperation, 0)
	if data.HasChange(keyPrefix + "host") {
		ops = append(ops, &ReplaceOperation{
			Path:  pathPrefix + "host",
			Value: data.Get(keyPrefix + "host").(string),
		})
	}
	if data.HasChange(keyPrefix + "path") {
		ops = append(ops, &ReplaceOperation{
			Path:  pathPrefix + "path",
			Value: data.Get(keyPrefix + "path").(string),
		})
	}
	if data.HasChange(keyPrefix + "port") {
		ops = append(ops, &ReplaceOperation{
			Path:  pathPrefix + "port",
			Value: expandRoutePort(data.Get(keyPrefix + "port").([]interface{})),
		})
	}
	if data.HasChange(keyPrefix + "tls") {
		ops = append(ops, &ReplaceOperation{
			Path:  pathPrefix + "tls",
			Value: expandTLSConfig(data.Get(keyPrefix + "tls").([]interface{})),
		})
	}
	if data.HasChange(keyPrefix + "to") {
		ops = append(ops, &ReplaceOperation{
			Path:  pathPrefix + "to",
			Value: expandRouteTargetReference(data.Get(keyPrefix + "to").([]interface{})),
		})
	}
	if data.HasChange(keyPrefix + "wildcard_policy") {
		ops = append(ops, &ReplaceOperation{
			Path:  pathPrefix + "wildcardPolicy",
			Value: data.Get(keyPrefix + "wildcard_policy").(string),
		})
	}
	return ops, nil
}
