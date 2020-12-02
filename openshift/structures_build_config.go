package openshift

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	api "github.com/openshift/api/build/v1"
	corev1 "k8s.io/api/core/v1"
	types "k8s.io/apimachinery/pkg/types"
)

func flattenBuildConfigSpec(in api.BuildConfigSpec, d *schema.ResourceData) ([]interface{}, error) {
	att := make(map[string]interface{})

	if in.CompletionDeadlineSeconds != nil {
		att["completion_deadline_seconds"] = *in.CompletionDeadlineSeconds
	}

	if in.FailedBuildsHistoryLimit != nil {
		att["failed_builds_history_limit"] = *in.FailedBuildsHistoryLimit
	}

	if in.SuccessfulBuildsHistoryLimit != nil {
		att["successful_builds_history_limit"] = *in.SuccessfulBuildsHistoryLimit
	}

	if len(in.NodeSelector) > 0 {
		att["node_selector"] = in.NodeSelector
	}

	if in.ServiceAccount != "" {
		att["service_account"] = in.ServiceAccount
	}

	res, err := flattenBuildConfigPostCommit(in.PostCommit)
	if err != nil {
		return nil, err
	}
	att["post_commit"] = res

	res, err = flattenBuildConfigOutput(in.Output)
	if err != nil {
		return nil, err
	}
	att["output"] = res

	res, err = flattenContainerResourceRequirements(in.Resources)
	if err != nil {
		return nil, err
	}
	att["resources"] = res

	if in.RunPolicy != "" {
		att["run_policy"] = in.RunPolicy
	}

	res, err = flattenBuildConfigSource(in.Source)
	if err != nil {
		return nil, err
	}
	att["source"] = res

	res, err = flattenBuildConfigStrategy(in.Strategy)
	if err != nil {
		return nil, err
	}
	att["strategy"] = res

	if len(in.Triggers) > 0 {
		v, err := flattenBuildConfigTriggers(in.Triggers)
		if err != nil {
			return []interface{}{att}, err
		}
		att["trigger"] = v
	}

	return []interface{}{att}, nil
}

func flattenBuildConfigTriggers(triggers []api.BuildTriggerPolicy) ([]interface{}, error) {
	att := make([]interface{}, len(triggers))
	for i, v := range triggers {
		obj := map[string]interface{}{}

		if v.Type != "" {
			obj["type"] = v.Type
		}
		if v.GitHubWebHook != nil {
			obj["github"] = flattenWebHookTrigger(v.GitHubWebHook)
		}
		if v.GenericWebHook != nil {
			obj["generic"] = flattenWebHookTrigger(v.GenericWebHook)
		}
		if v.GitLabWebHook != nil {
			obj["gitlab"] = flattenWebHookTrigger(v.GitLabWebHook)
		}
		if v.BitbucketWebHook != nil {
			obj["bitbucket"] = flattenWebHookTrigger(v.BitbucketWebHook)
		}
		att[i] = obj
	}
	return att, nil
}

func flattenWebHookTrigger(in *api.WebHookTrigger) []interface{} {
	att := make(map[string]interface{})

	if in.Secret != "" {
		att["secret"] = in.Secret
	}

	return []interface{}{att}
}

func flattenBuildConfigPostCommit(in api.BuildPostCommitSpec) ([]interface{}, error) {
	att := make(map[string]interface{})

	if len(in.Command) > 0 {
		att["command"] = in.Command
	}
	if len(in.Args) > 0 {
		att["args"] = in.Args
	}
	if in.Script != "" {
		att["script"] = in.Script
	}

	return []interface{}{att}, nil
}

func flattenBuildConfigOutput(in api.BuildOutput) ([]interface{}, error) {
	att := make(map[string]interface{})

	if in.To != nil {
		att["to"] = flattenObjectReference(*in.To)
	}

	if in.PushSecret != nil {
		att["push_secret"] = flattenLocalObjectReference(in.PushSecret)
	}

	return []interface{}{att}, nil
}

func flattenBuildConfigSource(in api.BuildSource) ([]interface{}, error) {
	att := make(map[string]interface{})

	if in.Type != "" {
		att["type"] = in.Type
	}

	if in.Binary != nil {
		att["binary"] = flattenBuildConfigBinary(*in.Binary)
	}

	if in.Dockerfile != nil {
		att["dockerfile"] = *in.Dockerfile
	}

	return []interface{}{att}, nil
}

func flattenBuildConfigBinary(in api.BinaryBuildSource) []interface{} {
	att := make(map[string]interface{})

	if in.AsFile != "" {
		att["as_file"] = in.AsFile
	}

	return []interface{}{att}
}

func flattenBuildConfigStrategy(in api.BuildStrategy) ([]interface{}, error) {
	att := make(map[string]interface{})

	if in.Type != "" {
		att["type"] = in.Type
	}

	if in.DockerStrategy != nil {
		att["docker_strategy"] = flattenBuildConfigDockerStrategy(*in.DockerStrategy)
	}

	if in.JenkinsPipelineStrategy != nil {
		att["jenkins_pipeline_strategy"] = flattenBuildConfigJenkinsPipelineStrategy(*in.JenkinsPipelineStrategy)
	}

	return []interface{}{att}, nil
}

func flattenBuildConfigDockerStrategy(in api.DockerBuildStrategy) []interface{} {
	att := make(map[string]interface{})

	if in.From != nil {
		att["from"] = flattenObjectReference(*in.From)
	}

	if in.PullSecret != nil {
		att["pull_secret"] = flattenLocalObjectReference(in.PullSecret)
	}

	return []interface{}{att}
}

func flattenBuildConfigJenkinsPipelineStrategy(in api.JenkinsPipelineBuildStrategy) []interface{} {
	att := make(map[string]interface{})

	if in.JenkinsfilePath != "" {
		att["jenkinsfile_path"] = in.JenkinsfilePath
	}

	if in.Jenkinsfile != "" {
		att["jenkinsfile"] = in.Jenkinsfile
	}

	return []interface{}{att}
}

func expandBuildConfigSpec(buildConfig []interface{}) (api.BuildConfigSpec, error) {
	obj := api.BuildConfigSpec{}

	if len(buildConfig) == 0 || buildConfig[0] == nil {
		return obj, nil
	}

	in := buildConfig[0].(map[string]interface{})

	obj.CompletionDeadlineSeconds = ptrToInt64(int64(in["completion_deadline_seconds"].(int)))
	obj.FailedBuildsHistoryLimit = ptrToInt32(int32(in["failed_builds_history_limit"].(int)))

	if v, ok := in["node_selector"].(map[string]interface{}); ok {
		obj.NodeSelector = expandStringMap(v)
	}

	if v, ok := in["output"].([]interface{}); ok && len(v) > 0 {
		obj.Output = expandBuildConfigOutput(v)
	}

	if v, ok := in["post_commit"].([]interface{}); ok && len(v) > 0 {
		obj.PostCommit = expandBuildConfigPostCommit(v)
	}

	if v, ok := in["resources"].([]interface{}); ok && len(v) > 0 {
		var err error
		resources, err := expandContainerResourceRequirements(v)
		if err != nil {
			return obj, err
		}
		obj.Resources = *resources
	}

	if v, ok := in["run_policy"].(string); ok {
		obj.RunPolicy = api.BuildRunPolicy(v)
	}

	if v, ok := in["service_account"].(string); ok {
		obj.ServiceAccount = v
	}

	if v, ok := in["source"].([]interface{}); ok && len(v) > 0 {
		obj.Source = expandBuildConfigSource(v)
	}

	if v, ok := in["strategy"].([]interface{}); ok && len(v) > 0 {
		obj.Strategy = expandBuildConfigStrategy(v)
	}

	obj.SuccessfulBuildsHistoryLimit = ptrToInt32(int32(in["successful_builds_history_limit"].(int)))

	if v, ok := in["trigger"].([]interface{}); ok && len(v) > 0 {
		obj.Triggers = expandBuildConfigTriggers(v)
	}

	return obj, nil
}

func expandBuildConfigOutput(l []interface{}) api.BuildOutput {
	obj := api.BuildOutput{}
	if len(l) == 0 || l[0] == nil {
		return obj
	}
	in := l[0].(map[string]interface{})

	if v, ok := in["to"].([]interface{}); ok && len(v) > 0 {
		obj.To = expandBuildConfigImageDefinitionPtr(v)
	}

	if v, ok := in["push_secret"].([]interface{}); ok && len(v) > 0 {
		obj.PushSecret = expandBuildConfigSecret(v)
	}

	if v, ok := in["image_label"].([]interface{}); ok && len(v) > 0 {
		obj.ImageLabels = expandBuildConfigOutputImageLabels(v)
	}

	return obj
}

func expandBuildConfigPostCommit(l []interface{}) api.BuildPostCommitSpec {
	obj := api.BuildPostCommitSpec{}
	if len(l) == 0 || l[0] == nil {
		return obj
	}

	in := l[0].(map[string]interface{})

	obj.Command = expandStringSlice(in["command"].(*schema.Set).List())
	obj.Args = expandStringSlice(in["args"].(*schema.Set).List())
	if v, ok := in["script"].(string); ok {
		obj.Script = v
	}

	return obj
}

func expandBuildConfigSource(l []interface{}) api.BuildSource {
	obj := api.BuildSource{}
	if len(l) == 0 || l[0] == nil {
		obj.Type = api.BuildSourceNone
		return obj
	}

	in := l[0].(map[string]interface{})

	if v, ok := in["type"].(string); ok {
		obj.Type = api.BuildSourceType(v)
	}

	if v, ok := in["binary"].([]interface{}); ok && len(v) > 0 {
		obj.Binary = expandBuildConfigSourceBinary(v)
	}

	if v, ok := in["dockerfile"].(string); ok && v != "" {
		obj.Dockerfile = ptrToString(v)
	}

	if v, ok := in["git"].([]interface{}); ok && len(v) > 0 {
		obj.Git = expandBuildConfigSourceGit(v)
	}

	if v, ok := in["image"].([]interface{}); ok && len(v) > 0 {
		obj.Images = expandBuildConfigSourceImages(v)
	}

	if v, ok := in["context_dir"].(string); ok {
		obj.ContextDir = v
	}

	if v, ok := in["source_secret"].([]interface{}); ok && len(v) > 0 {
		obj.SourceSecret = expandBuildConfigSecret(v)
	}

	if v, ok := in["secret"].([]interface{}); ok && len(v) > 0 {
		obj.Secrets = expandBuildConfigSecrets(v)
	}

	return obj
}

func expandBuildConfigSourceBinary(l []interface{}) *api.BinaryBuildSource {
	obj := &api.BinaryBuildSource{}
	if len(l) == 0 || l[0] == nil {
		return obj
	}

	in := l[0].(map[string]interface{})

	if v, ok := in["as_file"].(string); ok {
		obj.AsFile = v
	}

	return obj
}

func expandBuildConfigSourceGit(l []interface{}) *api.GitBuildSource {
	obj := &api.GitBuildSource{}
	if len(l) == 0 || l[0] == nil {
		return obj
	}

	in := l[0].(map[string]interface{})

	if v, ok := in["ref"].(string); ok {
		obj.Ref = v
	}

	if v, ok := in["uri"].(string); ok {
		obj.URI = v
	}

	//	if v, ok := m["proxy_config"].(string); ok {
	//		obj.ProxyConfig = expandBuildConfigProxyConfig(v)
	//		if v, ok := in["http_proxy"].(string); ok {
	//			obj.HttpProxy = v
	//		}
	//
	//		if v, ok := in["https_proxy"].(string); ok {
	//			obj.HttpsProxy = v
	//		}
	//
	//		if v, ok := in["no_proxy"].(string); ok {
	//			obj.NoProxy = v
	//		}
	//	}

	return obj
}

func expandBuildConfigSourceImages(images []interface{}) []api.ImageSource {
	if len(images) == 0 {
		return []api.ImageSource{}
	}

	img := make([]api.ImageSource, len(images))
	for i, c := range images {
		m := c.(map[string]interface{})

		if v, ok := m["from"].([]interface{}); ok {
			img[i].From = expandBuildConfigImageDefinition(v)
		}

		if v, ok := m["path"].([]interface{}); ok && len(v) > 0 {
			img[i].Paths = expandBuildConfigSourceImagesPaths(v)
		}

		if v, ok := m["pull_secret"].([]interface{}); ok && len(v) > 0 {
			img[i].PullSecret = expandBuildConfigSecret(v)
		}
	}

	return img
}

func expandBuildConfigSecrets(secrets []interface{}) []api.SecretBuildSource {
	if len(secrets) == 0 {
		return []api.SecretBuildSource{}
	}

	s := make([]api.SecretBuildSource, len(secrets))
	for i, c := range secrets {
		m := c.(map[string]interface{})

		if v, ok := m["secret"].([]interface{}); ok && len(v) > 0 {
			s[i].Secret = *expandBuildConfigSecret(v)
		}

		if v, ok := m["destination_dir"].(string); ok {
			s[i].DestinationDir = v
		}
	}

	return s
}

func expandBuildConfigSourceImagesPaths(sourcePaths []interface{}) []api.ImageSourcePath {
	if len(sourcePaths) == 0 {
		return []api.ImageSourcePath{}
	}

	sp := make([]api.ImageSourcePath, len(sourcePaths))
	for i, c := range sourcePaths {
		m := c.(map[string]interface{})

		if v, ok := m["source_path"].(string); ok {
			sp[i].SourcePath = v
		}

		if v, ok := m["destination_dir"].(string); ok {
			sp[i].DestinationDir = v
		}
	}

	return sp
}

func expandBuildConfigStrategy(l []interface{}) api.BuildStrategy {
	obj := api.BuildStrategy{}
	if len(l) == 0 || l[0] == nil {
		return obj
	}

	in := l[0].(map[string]interface{})

	if v, ok := in["type"].(string); ok {
		obj.Type = api.BuildStrategyType(v)
	}

	if v, ok := in["docker_strategy"].([]interface{}); ok && len(v) > 0 {
		obj.DockerStrategy = expandBuildConfigStrategyDockerBuildStrategy(v)
	}

	//	if v, ok := in["source_strategy"].([]interface{}); ok && len(v) > 0 {
	//		obj.SourceStrategy = expandBuildConfigStrategySourceStrategy(v)
	//	}

	if v, ok := in["jenkins_pipeline_strategy"].([]interface{}); ok && len(v) > 0 {
		obj.JenkinsPipelineStrategy = expandBuildConfigStrategyJenkinsPipelineStrategy(v)
	}

	return obj
}

func expandBuildConfigStrategyDockerBuildStrategy(l []interface{}) *api.DockerBuildStrategy {
	obj := &api.DockerBuildStrategy{}
	if len(l) == 0 || l[0] == nil {
		return obj
	}

	in := l[0].(map[string]interface{})

	if v, ok := in["from"].([]interface{}); ok && len(v) > 0 {
		obj.From = expandBuildConfigImageDefinitionPtr(v)
	}

	if v, ok := in["pull_secret"].([]interface{}); ok && len(v) > 0 {
		obj.PullSecret = expandBuildConfigSecret(v)
	}

	if v, ok := in["dockerfile_path"].(string); ok {
		obj.DockerfilePath = v
	}

	return obj
}

//func expandBuildConfigStrategySourceStrategy() {
//
//}

func expandBuildConfigStrategyJenkinsPipelineStrategy(l []interface{}) *api.JenkinsPipelineBuildStrategy {
	obj := &api.JenkinsPipelineBuildStrategy{}
	if len(l) == 0 || l[0] == nil {
		return obj
	}

	in := l[0].(map[string]interface{})

	if v, ok := in["jenkinsfile_path"].(string); ok {
		obj.JenkinsfilePath = v
	}

	if v, ok := in["jenkinsfile"].(string); ok {
		obj.Jenkinsfile = v
	}

	return obj
}

func expandBuildConfigTriggers(triggers []interface{}) []api.BuildTriggerPolicy {
	if len(triggers) == 0 {
		return []api.BuildTriggerPolicy{}
	}

	tg := make([]api.BuildTriggerPolicy, len(triggers))
	for i, c := range triggers {
		m := c.(map[string]interface{})

		if value, ok := m["type"].(string); ok {
			tg[i].Type = api.BuildTriggerType(value)
		}

		if value, ok := m["generic"].([]interface{}); ok && len(value) > 0 {
			tg[i].GenericWebHook = expandBuildConfigTriggerGit(value)
		}

		if value, ok := m["github"].([]interface{}); ok && len(value) > 0 {
			tg[i].GitHubWebHook = expandBuildConfigTriggerGit(value)
		}

		if value, ok := m["gitlab"].([]interface{}); ok && len(value) > 0 {
			tg[i].GitLabWebHook = expandBuildConfigTriggerGit(value)
		}

		if value, ok := m["bitbucket"].([]interface{}); ok && len(value) > 0 {
			tg[i].BitbucketWebHook = expandBuildConfigTriggerGit(value)
		}

		if value, ok := m["image_change"].([]interface{}); ok && len(value) > 0 {
			tg[i].ImageChange = expandBuildConfigTriggerImageChange(value)
		}
	}

	return tg
}

func expandBuildConfigTriggerImageChange(l []interface{}) *api.ImageChangeTrigger {
	obj := &api.ImageChangeTrigger{}
	if len(l) == 0 || l[0] == nil {
		return obj
	}

	in := l[0].(map[string]interface{})

	if v, ok := in["last_triggered_image_id"].(string); ok {
		obj.LastTriggeredImageID = v
	}

	if v, ok := in["from"].([]interface{}); ok {
		obj.From = expandBuildConfigImageDefinitionPtr(v)
	}

	return obj
}

func expandBuildConfigTriggerGit(l []interface{}) *api.WebHookTrigger {
	obj := &api.WebHookTrigger{}
	if len(l) == 0 || l[0] == nil {
		return obj
	}

	in := l[0].(map[string]interface{})

	if v, ok := in["secret"].(string); ok {
		obj.Secret = v
	}

	if v, ok := in["allow_env"].(bool); ok {
		obj.AllowEnv = v
	}

	if v, ok := in["secret_reference"].([]interface{}); ok {
		obj.SecretReference = expandBuildConfigSecretReference(v)
	}

	return obj
}

func expandBuildConfigImageDefinition(l []interface{}) corev1.ObjectReference {
	obj := corev1.ObjectReference{}
	if len(l) == 0 || l[0] == nil {
		return obj
	}

	in := l[0].(map[string]interface{})

	if v, ok := in["api_version"].(string); ok {
		obj.APIVersion = v
	}

	if v, ok := in["field_path"].(string); ok {
		obj.FieldPath = v
	}

	if v, ok := in["kind"].(string); ok {
		obj.Kind = v
	}

	if v, ok := in["name"].(string); ok {
		obj.Name = v
	}

	if v, ok := in["namespace"].(string); ok {
		obj.Namespace = v
	}

	if v, ok := in["resource_version"].(string); ok {
		obj.ResourceVersion = v
	}

	if v, ok := in["uid"].(string); ok {
		obj.UID = types.UID(v)
	}

	return obj
}

func expandBuildConfigSecretReference(l []interface{}) *api.SecretLocalReference {
	obj := &api.SecretLocalReference{}
	if len(l) == 0 || l[0] == nil {
		return obj
	}

	in := l[0].(map[string]interface{})

	if v, ok := in["name"].(string); ok {
		obj.Name = v
	}

	return obj
}

func expandBuildConfigSecret(l []interface{}) *corev1.LocalObjectReference {
	obj := &corev1.LocalObjectReference{}
	if len(l) == 0 || l[0] == nil {
		return obj
	}

	in := l[0].(map[string]interface{})

	if v, ok := in["name"].(string); ok {
		obj.Name = v
	}

	return obj
}

func expandBuildConfigOutputImageLabels(imageLabels []interface{}) []api.ImageLabel {
	if len(imageLabels) == 0 {
		return []api.ImageLabel{}
	}

	il := make([]api.ImageLabel, len(imageLabels))
	for i, c := range imageLabels {
		m := c.(map[string]interface{})

		if v, ok := m["name"].(string); ok {
			il[i].Name = v
		}

		if v, ok := m["value"].(string); ok {
			il[i].Value = v
		}
	}

	return il
}
