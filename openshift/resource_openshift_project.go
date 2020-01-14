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

func resourceOpenshiftProject() *schema.Resource {
	return &schema.Resource{
		Create: resourceOpenshiftProjectCreate,
		Read:   resourceOpenshiftProjectRead,
		Update: resourceOpenshiftProjectUpdate,
		Delete: resourceOpenshiftProjectDelete,
		Exists: resourceOpenshiftProjectExists,
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

func resourceOpenshiftProjectCreate(d *schema.ResourceData, meta interface{}) error {
	conn, err := client_v1.NewForConfig(meta.(*rest.Config))

	metadata := expandMetadata(d.Get("metadata").([]interface{}))
	project := api.Project{
		ObjectMeta: metadata,
	}

	log.Printf("[INFO] Creating new project: %#v", project)
	out, err := conn.Projects().Create(&project)
	if err != nil {
		return err
	}
	log.Printf("[INFO] Submitted new project: %#v", out)
	d.SetId(out.Name)

	return resourceOpenshiftProjectRead(d, meta)
}

func resourceOpenshiftProjectRead(d *schema.ResourceData, meta interface{}) error {
	conn, err := client_v1.NewForConfig(meta.(*rest.Config))

	name := d.Id()
	log.Printf("[INFO] Reading project %s", name)
	project, err := conn.Projects().Get(name, meta_v1.GetOptions{})
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

func resourceOpenshiftProjectUpdate(d *schema.ResourceData, meta interface{}) error {
	conn, err := client_v1.NewForConfig(meta.(*rest.Config))

	ops := patchMetadata("metadata.0.", "/metadata/", d)
	metadata, err := ops.MarshalJSON()
	if err != nil {
		return fmt.Errorf("Failed to marshal update operations: %s", err)
	}

	log.Printf("[INFO] Updating project: %s", ops)
	out, err := conn.Projects().Patch(d.Id(), pkgApi.JSONPatchType, metadata)
	if err != nil {
		return err
	}
	log.Printf("[INFO] Submitted updated project: %#v", out)
	d.SetId(out.Name)

	return resourceOpenshiftProjectRead(d, meta)
}

func resourceOpenshiftProjectDelete(d *schema.ResourceData, meta interface{}) error {
	conn, err := client_v1.NewForConfig(meta.(*rest.Config))

	name := d.Id()
	log.Printf("[INFO] Deleting project: %#v", name)
	err = conn.Projects().Delete(name, &meta_v1.DeleteOptions{})
	if err != nil {
		return err
	}

	stateConf := &resource.StateChangeConf{
		Target:  []string{},
		Pending: []string{"Terminating"},
		Timeout: d.Timeout(schema.TimeoutDelete),
		Refresh: func() (interface{}, string, error) {
			out, err := conn.Projects().Get(name, meta_v1.GetOptions{})
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

func resourceOpenshiftProjectExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	conn, err := client_v1.NewForConfig(meta.(*rest.Config))

	name := d.Id()
	log.Printf("[INFO] Checking project %s", name)
	_, err = conn.Projects().Get(name, meta_v1.GetOptions{})
	if err != nil {
		if statusErr, ok := err.(*errors.StatusError); ok && statusErr.ErrStatus.Code == 404 {
			return false, nil
		}
		log.Printf("[DEBUG] Received error: %#v", err)
	}
	log.Printf("[INFO] Project %s exists", name)
	return true, err
}
