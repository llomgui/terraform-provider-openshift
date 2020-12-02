package openshift

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	api "github.com/openshift/api/project/v1"
	client_v1 "github.com/openshift/client-go/project/clientset/versioned/typed/project/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	pkgApi "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func resourceOpenshiftProjectRequest() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpenshiftProjectRequestCreate,
		ReadContext:   resourceOpenshiftProjectRequestRead,
		UpdateContext: resourceOpenshiftProjectRequestUpdate,
		DeleteContext: resourceOpenshiftProjectRequestDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"metadata": metadataSchema("project", true),
		},
		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

func resourceOpenshiftProjectRequestCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := client_v1.NewForConfig(meta.(*rest.Config))
	if err != nil {
		diag.FromErr(err)
	}

	metadata := expandMetadata(d.Get("metadata").([]interface{}))
	project := api.ProjectRequest{
		ObjectMeta: metadata,
	}

	log.Printf("[INFO] Creating new project request: %#v", project)
	out, err := client.ProjectRequests().Create(ctx, &project, meta_v1.CreateOptions{})
	if err != nil {
		diag.FromErr(err)
	}
	log.Printf("[INFO] Submitted new project request: %#v", out)
	d.SetId(out.Name)

	return resourceOpenshiftProjectRequestUpdate(ctx, d, meta)
}

func resourceOpenshiftProjectRequestRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	exists, err := resourceOpenshiftProjectRequestExists(ctx, d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	if !exists {
		return diag.Diagnostics{}
	}
	client, err := client_v1.NewForConfig(meta.(*rest.Config))
	if err != nil {
		diag.FromErr(err)
	}

	name := d.Id()
	log.Printf("[INFO] Reading project %s", name)
	project, err := client.Projects().Get(ctx, name, meta_v1.GetOptions{})
	if err != nil {
		log.Printf("[DEBUG] Received error: %#v", err)
		diag.FromErr(err)
	}
	log.Printf("[INFO] Received project: %#v", project)
	err = d.Set("metadata", flattenMetadata(project.ObjectMeta, d))
	if err != nil {
		diag.FromErr(err)
	}

	return nil
}

func resourceOpenshiftProjectRequestUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := kubernetes.NewForConfig(meta.(*rest.Config))
	if err != nil {
		diag.FromErr(err)
	}

	ops := patchMetadata("metadata.0.", "/metadata/", d)
	metadata, err := ops.MarshalJSON()
	if err != nil {
		return diag.Errorf("Failed to marshal update operations: %s", err)
	}

	log.Printf("[INFO] Updating project: %s", ops)
	out, err := client.CoreV1().Namespaces().Patch(ctx, d.Id(), pkgApi.JSONPatchType, metadata, meta_v1.PatchOptions{})
	if err != nil {
		diag.FromErr(err)
	}
	log.Printf("[INFO] Submitted updated project: %#v", out)
	d.SetId(out.Name)

	return resourceOpenshiftProjectRequestRead(ctx, d, meta)
}

func resourceOpenshiftProjectRequestDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := client_v1.NewForConfig(meta.(*rest.Config))
	if err != nil {
		diag.FromErr(err)
	}

	name := d.Id()
	log.Printf("[INFO] Deleting project: %#v", name)
	err = client.Projects().Delete(ctx, name, meta_v1.DeleteOptions{})
	if err != nil {
		diag.FromErr(err)
	}

	stateConf := &resource.StateChangeConf{
		Target:  []string{},
		Pending: []string{"Terminating"},
		Timeout: d.Timeout(schema.TimeoutDelete),
		Refresh: func() (interface{}, string, error) {
			out, err := client.Projects().Get(ctx, name, meta_v1.GetOptions{})
			if err != nil {
				if statusErr, ok := err.(*errors.StatusError); ok && statusErr.ErrStatus.Code == 404 {
					return nil, "", nil
				}
				log.Printf("[ERROR] Received error: %#v", err)
				return out, "Error", err
			}

			statusPhase := fmt.Sprintf("%v", out.Status.Phase)
			log.Printf("[DEBUG] Project %s status received: %#v", out.Name, statusPhase)
			return out, statusPhase, nil
		},
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		diag.FromErr(err)
	}
	log.Printf("[INFO] Project %s deleted", name)

	d.SetId("")

	return nil
}

func resourceOpenshiftProjectRequestExists(ctx context.Context, d *schema.ResourceData, meta interface{}) (bool, error) {
	client, err := client_v1.NewForConfig(meta.(*rest.Config))
	if err != nil {
		return false, err
	}

	name := d.Id()
	log.Printf("[INFO] Checking project %s", name)
	_, err = client.Projects().Get(ctx, name, meta_v1.GetOptions{})
	if err != nil {
		if statusErr, ok := err.(*errors.StatusError); ok && statusErr.ErrStatus.Code == 404 {
			return false, nil
		}
		log.Printf("[DEBUG] Received error: %#v", err)
	}
	log.Printf("[INFO] Project %s exists", name)
	return true, err
}
