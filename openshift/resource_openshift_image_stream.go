package openshift

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	api "github.com/openshift/api/image/v1"
	client_v1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	pkgApi "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

func resourceOpenshiftImageStream() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpenshiftImageStreamCreate,
		ReadContext:   resourceOpenshiftImageStreamRead,
		UpdateContext: resourceOpenshiftImageStreamUpdate,
		DeleteContext: resourceOpenshiftImageStreamDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"metadata": namespacedMetadataSchema("imagestream", true),
			"spec": {
				Type:        schema.TypeList,
				Description: "(v1.ImageStreamSpec) Spec describes the desired state of this stream",
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"docker_image_repository": {
							Type:        schema.TypeString,
							Description: "(string) dockerImageRepository is optional, if specified this stream is backed by a Docker repository on this server",
							Optional:    true,
						},
						"lookup_policy": {
							Type:        schema.TypeList,
							Description: "(v1.ImageLookupPolicy) lookupPolicy controls how other resources reference images within this namespace.",
							Optional:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"local": {
										Type:        schema.TypeBool,
										Description: "(boolean) local will change the docker short image references (like 'mysql' or 'php:latest') on objects in this namespace to the image ID whenever they match this image stream, instead of reaching out to a remote registry. The name will be fully qualified to an image ID if found. The tag's referencePolicy is taken into account on the replaced value. Only works within the current namespace.",
										Optional:    true,
									},
								},
							},
						},
						"tag": {
							Type:        schema.TypeList,
							Description: "(array) tags map arbitrary string values to specific image locators",
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"annotations": {
										Type:         schema.TypeMap,
										Description:  "(object) Annotations associated with images using this tag",
										Optional:     true,
										Elem:         &schema.Schema{Type: schema.TypeString},
										ValidateFunc: validateAnnotations,
									},
									"from": {
										Type:        schema.TypeList,
										Description: "(v1.ObjectReference) From is a reference to an image stream tag or image stream this tag should track",
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"api_version": {
													Type:        schema.TypeString,
													Description: "(string) API version of the referent.",
													Optional:    true,
												},
												"field_path": {
													Type:        schema.TypeString,
													Description: "(string) If referring to a piece of an object instead of an entire object, this string should contain a valid JSON/Go field access statement, such as desiredState.manifest.containers[2]. For example, if the object reference is to a container within a pod, this would take on a value like: \"spec.containers{name}\" (where \"name\" refers to the name of the container that triggered the event) or if no container name is specified \"spec.containers[2]\" (container with index 2 in this pod). This syntax is chosen only to have some well-defined way of referencing a part of an object.",
													Optional:    true,
												},
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
												"resource_version": {
													Type:        schema.TypeString,
													Description: "(string) Specific resource_version to which this reference is made, if any. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency",
													Optional:    true,
												},
												"uid": {
													Type:        schema.TypeString,
													Description: "(string) UID of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#uids",
													Optional:    true,
												},
											},
										},
									},
									"generation": {
										Type:        schema.TypeInt,
										Description: "(integer) Generation is the image stream generation that updated this tag - setting it to 0 is an indication that the generation must be updated. Legacy clients will send this as nil, which means the client doesn't know or care.",
										Computed:    true,
									},
									"import_policy": {
										Type:        schema.TypeList,
										Description: "(v1.TagImportPolicy) Import is information that controls how images may be imported by the server.",
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"insecure": {
													Type:        schema.TypeBool,
													Description: "(boolean) Insecure is true if the server may bypass certificate verification or connect directly over HTTP during image import.",
													Optional:    true,
												},
												"scheduled": {
													Type:        schema.TypeBool,
													Description: "(boolean) Scheduled indicates to the server that this tag should be periodically checked to ensure it is up to date, and imported",
													Optional:    true,
												},
											},
										},
									},
									"name": {
										Type:         schema.TypeString,
										Description:  "(string) Name of the tag",
										Optional:     true,
										ForceNew:     true,
										Computed:     true,
										ValidateFunc: validateName,
									},
									"reference": {
										Type:        schema.TypeBool,
										Description: "(boolean) Reference states if the tag will be imported. Default value is false, which means the tag will be imported.",
										Default:     false,
										Optional:    true,
									},
									"reference_policy": {
										Type:        schema.TypeList,
										Description: "(v1.TagReferencePolicy) ReferencePolicy defines how other components should consume the image",
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"type": {
													Type:        schema.TypeString,
													Description: "(string) Type determines how the image pull spec should be transformed when the image stream tag is used in deployment config triggers or new builds. The default value is `Source`, indicating the original location of the image should be used (if imported). The user may also specify `Local`, indicating that the pull spec should point to the integrated Docker registry and leverage the registry's ability to proxy the pull to an upstream registry. `Local` allows the credentials used to pull this image to be managed from the image stream's namespace, so others on the platform can access a remote image but have no access to the remote secret. It also allows the image layers to be mirrored into the local registry which the images can still be pulled even if the upstream registry is unavailable.",
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
	}
}

func resourceOpenshiftImageStreamCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := client_v1.NewForConfig(meta.(*rest.Config))
	if err != nil {
		return diag.FromErr(err)
	}

	metadata := expandMetadata(d.Get("metadata").([]interface{}))
	spec, err := expandImageStreamSpec(d.Get("spec").([]interface{}))
	if err != nil {
		return diag.FromErr(err)
	}

	imageStream := api.ImageStream{
		ObjectMeta: metadata,
		Spec:       spec,
	}

	log.Printf("[INFO] Creating new imagestream: %#v", imageStream)
	out, err := client.ImageStreams(metadata.Namespace).Create(ctx, &imageStream, meta_v1.CreateOptions{})
	if err != nil {
		return diag.Errorf("Failed to create imagestream: %s", err)
	}
	log.Printf("[INFO] Submitted new imagestream: %#v", out)
	d.SetId(buildId(out.ObjectMeta))

	return resourceOpenshiftImageStreamRead(ctx, d, meta)
}

func resourceOpenshiftImageStreamRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	exists, err := resourceOpenshiftImageStreamExists(ctx, d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	if !exists {
		return diag.Diagnostics{}
	}
	client, err := client_v1.NewForConfig(meta.(*rest.Config))
	if err != nil {
		return diag.FromErr(err)
	}

	namespace, name, err := idParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Reading imagestream %s", name)
	imageStream, err := client.ImageStreams(namespace).Get(ctx, name, meta_v1.GetOptions{})
	if err != nil {
		log.Printf("[DEBUG] Received error: %#v", err)
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Received imagestream: %#v", imageStream)
	err = d.Set("metadata", flattenMetadata(imageStream.ObjectMeta, d))
	if err != nil {
		return diag.FromErr(err)
	}

	spec, err := flattenImageStreamSpec(imageStream.Spec, d)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("spec", spec)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceOpenshiftImageStreamUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := client_v1.NewForConfig(meta.(*rest.Config))
	if err != nil {
		return diag.FromErr(err)
	}

	namespace, name, err := idParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	ops := patchMetadata("metadata.0.", "/metadata/", d)
	if d.HasChange("spec") {
		spec, err := expandImageStreamSpec(d.Get("spec").([]interface{}))
		if err != nil {
			return diag.FromErr(err)
		}

		ops = append(ops, &ReplaceOperation{
			Path:  "/spec",
			Value: spec,
		})
	}

	data, err := ops.MarshalJSON()
	if err != nil {
		return diag.Errorf("Failed to marshal update operations: %s", err)
	}

	log.Printf("[INFO] Updating imagestream %q: %v", name, string(data))
	out, err := client.ImageStreams(namespace).Patch(ctx, name, pkgApi.JSONPatchType, data, meta_v1.PatchOptions{})
	if err != nil {
		return diag.Errorf("Failed to update imagestream: %s", err)
	}
	log.Printf("[INFO] Submitted updated imagestream: %#v", out)

	d.SetId(buildId(out.ObjectMeta))

	return resourceOpenshiftImageStreamRead(ctx, d, meta)
}

func resourceOpenshiftImageStreamDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := client_v1.NewForConfig(meta.(*rest.Config))
	if err != nil {
		return diag.FromErr(err)
	}

	namespace, name, err := idParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Deleting imagestream: %#v", name)

	err = client.ImageStreams(namespace).Delete(ctx, name, meta_v1.DeleteOptions{})
	if err != nil {
		return diag.Errorf("Failed to delete imagestream: %s", err)
	}

	d.SetId("")
	return nil
}

func resourceOpenshiftImageStreamExists(ctx context.Context, d *schema.ResourceData, meta interface{}) (bool, error) {
	client, err := client_v1.NewForConfig(meta.(*rest.Config))
	if err != nil {
		return false, err
	}

	namespace, name, err := idParts(d.Id())
	if err != nil {
		return false, err
	}

	log.Printf("[INFO] Checking imagestream %s", name)
	_, err = client.ImageStreams(namespace).Get(ctx, name, meta_v1.GetOptions{})
	if err != nil {
		if statusErr, ok := err.(*errors.StatusError); ok && statusErr.ErrStatus.Code == 404 {
			return false, nil
		}
		log.Printf("[DEBUG] Received error: %#v", err)
	}
	return true, err
}
