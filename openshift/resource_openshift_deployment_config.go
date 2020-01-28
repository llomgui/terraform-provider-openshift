package openshift

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	api "github.com/openshift/api/apps/v1"
	client_v1 "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	pkgApi "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

const (
	// https://github.com/kubernetes/kubernetes/blob/master/pkg/controller/deployment/util/deployment_util.go#L84
	TimedOutReason = "ProgressDeadlineExceeded"
)

func resourceOpenshiftDeploymentConfig() *schema.Resource {
	return &schema.Resource{
		Create: resourceOpenshiftDeploymentConfigCreate,
		Read:   resourceOpenshiftDeploymentConfigRead,
		Update: resourceOpenshiftDeploymentConfigUpdate,
		Delete: resourceOpenshiftDeploymentConfigDelete,
		Exists: resourceOpenshiftDeploymentConfigExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"metadata": namespacedMetadataSchemaIsTemplate("pod", true, true),
			"spec": {
				Type:        schema.TypeList,
				Description: "(v1.DeploymentConfigSpec) Spec represents a desired deployment state and how to deploy to it.",
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"min_ready_seconds": {
							Type:        schema.TypeInt,
							Description: "(integer) MinReadySeconds is the minimum number of seconds for which a newly created pod should be ready without any of its container crashing, for it to be considered available. Defaults to 0 (pod will be considered available as soon as it is ready)",
							Optional:    true,
							Default:     0,
						},
						"paused": {
							Type:        schema.TypeBool,
							Description: "(boolean) Paused indicates that the deployment config is paused resulting in no new deployments on template changes or changes in the template caused by other triggers.",
							Optional:    true,
							Default:     false,
						},
						"replicas": {
							Type:        schema.TypeInt,
							Description: "(integer) Replicas is the number of desired replicas.",
							Optional:    true,
							Default:     1,
						},
						"revision_history_limit": {
							Type:        schema.TypeInt,
							Description: "(integer) RevisionHistoryLimit is the number of old ReplicationControllers to retain to allow for rollbacks. This field is a pointer to allow for differentiation between an explicit zero and not specified. Defaults to 10. (This only applies to DeploymentConfigs created via the new group API resource, not the legacy resource.)",
							Optional:    true,
							Default:     10,
						},
						"selector": {
							Type:        schema.TypeMap,
							Description: "(object) Selector is a label query over pods that should match the Replicas count.",
							Optional:    true,
							ForceNew:    true,
						},
						"strategy": {
							Type:        schema.TypeList,
							Description: "(v1.DeploymentStrategy) Strategy describes how a deployment is executed.",
							Optional:    true,
							Computed:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"active_deadline_seconds": {
										Type:        schema.TypeInt,
										Description: "(integer) ActiveDeadlineSeconds is the duration in seconds that the deployer pods for this deployment config may be active on a node before the system actively tries to terminate them.",
										Optional:    true,
										Default:     21600,
									},
									"resources": {
										Type:        schema.TypeList,
										Optional:    true,
										MaxItems:    1,
										Computed:    true,
										Description: "(v1.ResourceRequirements) Resources contains resource requirements to execute the deployment and any hooks.",
										Elem: &schema.Resource{
											Schema: resourcesField(),
										},
									},
									"rolling_params": {
										Type:        schema.TypeList,
										Description: "(v1.RollingDeploymentStrategyParams) RollingParams are the input to the Rolling deployment strategy.",
										Optional:    true,
										Computed:    true,
										MaxItems:    1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"interval_seconds": {
													Type:        schema.TypeInt,
													Description: "(integer) IntervalSeconds is the time to wait between polling deployment status after update. If the value is nil, a default will be used.",
													Optional:    true,
													Default:     1,
												},
												"max_surge": {
													Type:         schema.TypeString,
													Description:  "(intstr.IntOrString) MaxSurge is the maximum number of pods that can be scheduled above the original number of pods. Value can be an absolute number (ex: 5) or a percentage of total pods at the start of the update (ex: 10%). Absolute number is calculated from percentage by rounding up. This cannot be 0 if MaxUnavailable is 0. By default, 25% is used. Example: when this is set to 30%, the new RC can be scaled up by 30% immediately when the rolling update starts. Once old pods have been killed, new RC can be scaled up further, ensuring that total number of pods running at any time during the update is atmost 130% of original pods.",
													Optional:     true,
													Default:      "25%",
													ValidateFunc: validation.StringMatch(regexp.MustCompile(`^([0-9]+|[1-9][0-9]%|[1-9]%|100%)$`), ""),
												},
												"max_unavailable": {
													Type:         schema.TypeString,
													Description:  "(intstr.IntOrString) MaxUnavailable is the maximum number of pods that can be unavailable during the update. Value can be an absolute number (ex: 5) or a percentage of total pods at the start of update (ex: 10%). Absolute number is calculated from percentage by rounding down. This cannot be 0 if MaxSurge is 0. By default, 25% is used. Example: when this is set to 30%, the old RC can be scaled down by 30% immediately when the rolling update starts. Once new pods are ready, old RC can be scaled down further, followed by scaling up the new RC, ensuring that at least 70% of original number of pods are available at all times during the update.",
													Optional:     true,
													Default:      "25%",
													ValidateFunc: validation.StringMatch(regexp.MustCompile(`^([0-9]+|[1-9][0-9]%|[1-9]%|100%)$`), ""),
												},
												"timeout_seconds": {
													Type:        schema.TypeInt,
													Description: "(integer) TimeoutSeconds is the time to wait for updates before giving up. If the value is nil, a default will be used.",
													Optional:    true,
													Default:     600,
												},
												"update_period_seconds": {
													Type:        schema.TypeInt,
													Description: "(integer) UpdatePeriodSeconds is the time to wait between individual pod updates. If the value is nil, a default will be used.",
													Optional:    true,
													Default:     1,
												},
											},
										},
									},
									"type": {
										Type:         schema.TypeString,
										Description:  "(string) Type is the name of a deployment strategy.",
										Optional:     true,
										Default:      "Rolling",
										ValidateFunc: validation.StringInSlice([]string{"Rolling", "Recreate", "Custom"}, false),
									},
								},
							},
						},
						"template": {
							Type:        schema.TypeList,
							Description: "(v1.PodTemplateSpec) Template is the object that describes the pod that will be created if insufficient replicas are detected.",
							Required:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"metadata": namespacedMetadataSchema("pod", true),
									"spec": {
										Type:        schema.TypeList,
										Description: "(v1.PodSpec) Specification of the desired behavior of the pod. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status",
										Required:    true,
										MaxItems:    1,
										Elem: &schema.Resource{
											Schema: podSpecFields(true, false, false),
										},
									},
								},
							},
						},
						"trigger": {
							Type:        schema.TypeList,
							Description: "(array) Triggers determine how updates to a DeploymentConfig result in new deployments. If no triggers are defined, a new deployment can only occur as a result of an explicit client update to the DeploymentConfig with a new LatestVersion. If null, defaults to having a config change trigger.",
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:         schema.TypeString,
										Description:  "(string) Type of the trigger",
										Optional:     true,
										Default:      "ConfigChange",
										ValidateFunc: validation.StringInSlice([]string{"ImageChange", "ConfigChange"}, false),
									},
									"image_change_params": {
										Type:        schema.TypeList,
										Description: "(v1.DeploymentTriggerImageChangeParams) ImageChangeParams represents the parameters for the ImageChange trigger.",
										Optional:    true,
										MaxItems:    1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"automatic": {
													Type:        schema.TypeBool,
													Description: "(boolean) Automatic means that the detection of a new tag value should result in an image update inside the pod template.",
													Optional:    true,
													Computed:    true,
												},
												"container_names": {
													Type:        schema.TypeSet,
													Description: "(array) ContainerNames is used to restrict tag updates to the specified set of container names in a pod. If multiple triggers point to the same containers, the resulting behavior is undefined. Future API versions will make this a validation error. If ContainerNames does not point to a valid container, the trigger will be ignored. Future API versions will make this a validation error.",
													Required:    true,
													Elem:        &schema.Schema{Type: schema.TypeString},
													Set:         schema.HashString,
												},
												"from": {
													Type:        schema.TypeList,
													Description: "(v1.ObjectReference) From is a reference to an image stream tag to watch for changes. From.Name is the only required subfield - if From.Namespace is blank, the namespace of the current deployment trigger will be used.",
													Optional:    true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"kind": {
																Type:        schema.TypeString,
																Description: "(string) Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
																Optional:    true,
															},
															"name": {
																Type:        schema.TypeString,
																Description: "(string) Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names",
																Optional:    true,
															},
															"namespace": {
																Type:        schema.TypeString,
																Description: "(string) Namespace of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/",
																Optional:    true,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceOpenshiftDeploymentConfigCreate(d *schema.ResourceData, meta interface{}) error {
	client, err := client_v1.NewForConfig(meta.(*rest.Config))
	if err != nil {
		return err
	}

	metadata := expandMetadata(d.Get("metadata").([]interface{}))
	spec, err := expandDeploymentConfigSpec(d.Get("spec").([]interface{}))
	if err != nil {
		return err
	}

	deploymentConfig := api.DeploymentConfig{
		ObjectMeta: metadata,
		Spec:       *spec,
	}

	log.Printf("[INFO] Creating new deploymentconfig: %#v", deploymentConfig)
	out, err := client.DeploymentConfigs(metadata.Namespace).Create(&deploymentConfig)
	if err != nil {
		return fmt.Errorf("Failed to create deploymentconfig: %s", err)
	}

	d.SetId(buildId(out.ObjectMeta))

	log.Printf("[DEBUG] Waiting for deploymentconfig %s to schedule %d replicas", d.Id(), out.Spec.Replicas)

	err = resource.Retry(d.Timeout(schema.TimeoutCreate),
		waitForDeploymentReplicasFunc(client, out.GetNamespace(), out.GetName()))
	if err != nil {
		return err
	}

	log.Printf("[INFO] Submitted new deploymentconfig: %#v", out)

	return resourceOpenshiftDeploymentConfigRead(d, meta)
}

func resourceOpenshiftDeploymentConfigRead(d *schema.ResourceData, meta interface{}) error {
	client, err := client_v1.NewForConfig(meta.(*rest.Config))
	if err != nil {
		return err
	}

	namespace, name, err := idParts(d.Id())
	if err != nil {
		return err
	}

	log.Printf("[INFO] Reading deploymentconfig %s", name)
	deploymentConfig, err := client.DeploymentConfigs(namespace).Get(name, meta_v1.GetOptions{})
	if err != nil {
		log.Printf("[DEBUG] Received error: %#v", err)
		return err
	}
	log.Printf("[INFO] Received deploymentconfig: %#v", deploymentConfig)

	err = d.Set("metadata", flattenMetadata(deploymentConfig.ObjectMeta, d))
	if err != nil {
		return err
	}

	spec, err := flattenDeploymentConfigSpec(deploymentConfig.Spec, d)
	if err != nil {
		return err
	}

	err = d.Set("spec", spec)
	if err != nil {
		return err
	}

	return nil
}

func resourceOpenshiftDeploymentConfigUpdate(d *schema.ResourceData, meta interface{}) error {
	client, err := client_v1.NewForConfig(meta.(*rest.Config))
	if err != nil {
		return err
	}

	namespace, name, err := idParts(d.Id())
	if err != nil {
		return err
	}

	ops := patchMetadata("metadata.0.", "/metadata/", d)

	if d.HasChange("spec") {
		spec, err := expandDeploymentConfigSpec(d.Get("spec").([]interface{}))
		if err != nil {
			return err
		}

		ops = append(ops, &ReplaceOperation{
			Path:  "/spec",
			Value: spec,
		})
	}
	data, err := ops.MarshalJSON()
	if err != nil {
		return fmt.Errorf("Failed to marshal update operations: %s", err)
	}
	log.Printf("[INFO] Updating deploymentconfig %q: %v", name, string(data))
	out, err := client.DeploymentConfigs(namespace).Patch(name, pkgApi.JSONPatchType, data)
	if err != nil {
		return fmt.Errorf("Failed to update deployment: %s", err)
	}
	log.Printf("[INFO] Submitted updated deploymentconfig: %#v", out)

	err = resource.Retry(d.Timeout(schema.TimeoutUpdate),
		waitForDeploymentReplicasFunc(client, namespace, name))
	if err != nil {
		return err
	}

	return resourceOpenshiftDeploymentConfigRead(d, meta)
}

func resourceOpenshiftDeploymentConfigDelete(d *schema.ResourceData, meta interface{}) error {
	client, err := client_v1.NewForConfig(meta.(*rest.Config))
	if err != nil {
		return err
	}

	namespace, name, err := idParts(d.Id())
	if err != nil {
		return err
	}

	log.Printf("[INFO] Deleting deploymentconfig: %#v", name)

	err = client.DeploymentConfigs(namespace).Delete(name, &meta_v1.DeleteOptions{})
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}

func resourceOpenshiftDeploymentConfigExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client, err := client_v1.NewForConfig(meta.(*rest.Config))
	if err != nil {
		return true, err
	}

	namespace, name, err := idParts(d.Id())
	if err != nil {
		return false, err
	}

	log.Printf("[INFO] Checking deploymentconfig %s", name)
	_, err = client.DeploymentConfigs(namespace).Get(name, meta_v1.GetOptions{})
	if err != nil {
		if statusErr, ok := err.(*errors.StatusError); ok && statusErr.ErrStatus.Code == 404 {
			return false, nil
		}
		log.Printf("[DEBUG] Received error: %#v", err)
	}
	return true, err
}

func GetDeploymentConfigCondition(status api.DeploymentConfigStatus, condType api.DeploymentConditionType) *api.DeploymentCondition {
	for i := range status.Conditions {
		c := status.Conditions[i]
		if c.Type == condType {
			return &c
		}
	}
	return nil
}

func waitForDeploymentReplicasFunc(client *client_v1.AppsV1Client, ns, name string) resource.RetryFunc {
	return func() *resource.RetryError {
		// Query the deployment to get a status update.
		dply, err := client.DeploymentConfigs(ns).Get(name, meta_v1.GetOptions{})
		if err != nil {
			return resource.NonRetryableError(err)
		}

		if dply.Generation <= dply.Status.ObservedGeneration {
			cond := GetDeploymentConfigCondition(dply.Status, api.DeploymentProgressing)
			if cond != nil && cond.Reason == TimedOutReason {
				err := fmt.Errorf("Deployment exceeded its progress deadline")
				return resource.NonRetryableError(err)
			}

			if dply.Status.UpdatedReplicas < dply.Spec.Replicas {
				return resource.RetryableError(fmt.Errorf("Waiting for rollout to finish: %d out of %d new replicas have been updated...", dply.Status.UpdatedReplicas, dply.Spec.Replicas))
			}

			if dply.Status.Replicas > dply.Status.UpdatedReplicas {
				return resource.RetryableError(fmt.Errorf("Waiting for rollout to finish: %d old replicas are pending termination...", dply.Status.Replicas-dply.Status.UpdatedReplicas))
			}

			if dply.Status.AvailableReplicas < dply.Status.UpdatedReplicas {
				return resource.RetryableError(fmt.Errorf("Waiting for rollout to finish: %d of %d updated replicas are available...", dply.Status.AvailableReplicas, dply.Status.UpdatedReplicas))
			}
		} else if dply.Status.ObservedGeneration == 0 {
			return resource.RetryableError(fmt.Errorf("Waiting for rollout to start"))
		}
		return nil
	}
}
