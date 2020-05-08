package openshift

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	api "github.com/openshift/api/project/v1"
	client_v1 "github.com/openshift/client-go/project/clientset/versioned/typed/project/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	pkgApi "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

func resourceOpenshiftProjectRequest() *schema.Resource {
	return &schema.Resource{
		Create: resourceOpenshiftProjectRequestCreate,
		Read:   resourceOpenshiftProjectRequestRead,
		Update: resourceOpenshiftProjectRequestUpdate,
		Delete: resourceOpenshiftProjectRequestDelete,
		Exists: resourceOpenshiftProjectRequestExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"metadata": metadataSchema("project", true),
		},
		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

func resourceOpenshiftProjectRequestCreate(d *schema.ResourceData, meta interface{}) error {
	client, err := client_v1.NewForConfig(meta.(*rest.Config))
	if err != nil {
		return err
	}

	metadata := expandMetadata(d.Get("metadata").([]interface{}))
	project := api.ProjectRequest{
		ObjectMeta: metadata,
	}

	log.Printf("[INFO] Creating new project request: %#v", project)
	out, err := client.ProjectRequests().Create(&project)
	if err != nil {
		return err
	}
	log.Printf("[INFO] Submitted new project request: %#v", out)
	d.SetId(out.Name)

	return resourceOpenshiftProjectRequestRead(d, meta)
}

func resourceOpenshiftProjectRequestRead(d *schema.ResourceData, meta interface{}) error {
	client, err := client_v1.NewForConfig(meta.(*rest.Config))
	if err != nil {
		return err
	}

	name := d.Id()
	log.Printf("[INFO] Reading project %s", name)
	project, err := client.Projects().Get(name, meta_v1.GetOptions{})
	if err != nil {
		log.Printf("[DEBUG] Received error: %#v", err)
		return err
	}
	log.Printf("[INFO] Received project: %#v", project)
	err = d.Set("metadata", flattenMetadata(project.ObjectMeta, d))
	if err != nil {
		return err
	}

	return nil
}

func resourceOpenshiftProjectRequestUpdate(d *schema.ResourceData, meta interface{}) error {
	client, err := client_v1.NewForConfig(meta.(*rest.Config))
	if err != nil {
		return err
	}

	ops := patchMetadata("metadata.0.", "/metadata/", d)
	metadata, err := ops.MarshalJSON()
	if err != nil {
		return fmt.Errorf("Failed to marshal update operations: %s", err)
	}

	log.Printf("[INFO] Updating project: %s", ops)
	out, err := client.Projects().Patch(d.Id(), pkgApi.JSONPatchType, metadata)
	if err != nil {
		return err
	}
	log.Printf("[INFO] Submitted updated project: %#v", out)
	d.SetId(out.Name)

	return resourceOpenshiftProjectRequestRead(d, meta)
}

func resourceOpenshiftProjectRequestDelete(d *schema.ResourceData, meta interface{}) error {
	client, err := client_v1.NewForConfig(meta.(*rest.Config))
	if err != nil {
		return err
	}

	name := d.Id()
	log.Printf("[INFO] Deleting project: %#v", name)
	err = client.Projects().Delete(name, &meta_v1.DeleteOptions{})
	if err != nil {
		return err
	}

	stateConf := &resource.StateChangeConf{
		Target:  []string{},
		Pending: []string{"Terminating"},
		Timeout: d.Timeout(schema.TimeoutDelete),
		Refresh: func() (interface{}, string, error) {
			out, err := client.Projects().Get(name, meta_v1.GetOptions{})
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
		return err
	}
	log.Printf("[INFO] Project %s deleted", name)

	d.SetId("")

	return nil
}

func resourceOpenshiftProjectRequestExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client, err := client_v1.NewForConfig(meta.(*rest.Config))
	if err != nil {
		return false, err
	}

	name := d.Id()
	log.Printf("[INFO] Checking project %s", name)
	_, err = client.Projects().Get(name, meta_v1.GetOptions{})
	if err != nil {
		if statusErr, ok := err.(*errors.StatusError); ok && statusErr.ErrStatus.Code == 404 {
			return false, nil
		}
		log.Printf("[DEBUG] Received error: %#v", err)
	}
	log.Printf("[INFO] Project %s exists", name)
	return true, err
}
