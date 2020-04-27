package openshift

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	api "github.com/openshift/api/network/v1"
	client_v1 "github.com/openshift/client-go/network/clientset/versioned/typed/network/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	pkgApi "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

func resourceOpenshiftNetNamespace() *schema.Resource {
	return &schema.Resource{
		Create: resourceOpenshiftNetNamespaceCreate,
		Read:   resourceOpenshiftNetNamespaceRead,
		Update: resourceOpenshiftNetNamespaceUpdate,
		Delete: resourceOpenshiftNetNamespaceDelete,
		Exists: resourceOpenshiftNetNamespaceExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"metadata": namespacedMetadataSchemaIsTemplate("netnamespace", true, true),
			"netid": {
				Type:        schema.TypeString,
				Description: "(integer) NetID is the network identifier of the network namespace assigned to each overlay network packet. This can be manipulated with the \"oc adm pod-network\" commands.",
				Computed:    true,
			},
			"netname": {
				Type:        schema.TypeString,
				Description: "(string) NetName is the name of the network namespace. (This is the same as the object's name, but both fields must be set.)",
				Required:    true,
			},
			"egress_ips": {
				Type:        schema.TypeSet,
				Description: "EgressIPs is a list of reserved IPs that will be used as the source for external traffic coming from pods in this namespace. (If empty, external traffic will be masqueraded to Node IPs.)",
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},
		},
	}
}

func resourceOpenshiftNetNamespaceCreate(d *schema.ResourceData, meta interface{}) error {
	client, err := client_v1.NewForConfig(meta.(*rest.Config))
	if err != nil {
		return err
	}

	netNamespace := api.NetNamespace{}

	netNamespace.ObjectMeta = expandMetadata(d.Get("metadata").([]interface{}))
	netNamespace.NetID = d.Get("netid").(uint32)
	netNamespace.NetName = d.Get("netname").(string)

	if v, ok := d.GetOk("egress_ips"); ok && v.(*schema.Set).Len() > 0 {
		netNamespace.EgressIPs = sliceOfString(v.(*schema.Set).List())
	}

	log.Printf("[INFO] Creating new netnamespace: %#v", netNamespace)
	out, err := client.NetNamespaces().Create(&netNamespace)
	if err != nil {
		return fmt.Errorf("Failed to create netnamespace: %s", err)
	}
	log.Printf("[INFO] Submitted new netnamespace: %#v", out)
	d.SetId(out.NetName)

	return resourceOpenshiftNetNamespaceRead(d, meta)
}

func resourceOpenshiftNetNamespaceRead(d *schema.ResourceData, meta interface{}) error {
	client, err := client_v1.NewForConfig(meta.(*rest.Config))
	if err != nil {
		return err
	}

	name := d.Id()

	log.Printf("[INFO] Reading netnamespace %s", name)
	netNamespace, err := client.NetNamespaces().Get(name, meta_v1.GetOptions{})
	if err != nil {
		log.Printf("[DEBUG] Received error: %#v", err)
		return err
	}
	log.Printf("[INFO] Received netnamespace: %#v", netNamespace)

	err = d.Set("metadata", flattenMetadata(netNamespace.ObjectMeta, d))
	if err != nil {
		return err
	}

	err = d.Set("netname", netNamespace.NetName)
	if err != nil {
		return err
	}

	err = d.Set("netid", strconv.FormatUint(uint64(netNamespace.NetID), 10))
	if err != nil {
		return err
	}

	if len(netNamespace.EgressIPs) > 0 {
		if err := d.Set("egress_ips", newStringSet(schema.HashString, netNamespace.EgressIPs)); err != nil {
			return fmt.Errorf("error setting egress_ips: %s", err)
		}
	}

	return nil
}

func resourceOpenshiftNetNamespaceUpdate(d *schema.ResourceData, meta interface{}) error {
	client, err := client_v1.NewForConfig(meta.(*rest.Config))
	if err != nil {
		return err
	}

	name := d.Id()

	ops := patchMetadata("metadata.0.", "/metadata/", d)
	if d.HasChange("netid") {
		netid := d.Get("netid").(uint32)
		ops = append(ops, &ReplaceOperation{
			Path:  "/netid",
			Value: netid,
		})
	}

	if d.HasChange("netname") {
		netname := d.Get("netname").(string)
		ops = append(ops, &ReplaceOperation{
			Path:  "/netname",
			Value: netname,
		})
	}

	if d.HasChange("egress_ips") {
		ops = append(ops, &ReplaceOperation{
			Path:  "/egressIPs",
			Value: d.Get("egress_ips").(*schema.Set).List(),
		})
	}

	data, err := ops.MarshalJSON()
	if err != nil {
		return fmt.Errorf("Failed to marshal update operations: %s", err)
	}

	log.Printf("[INFO] Updating netnamespace %q: %v", name, string(data))
	out, err := client.NetNamespaces().Patch(name, pkgApi.JSONPatchType, data)
	if err != nil {
		return fmt.Errorf("Failed to update netnamespace: %s", err)
	}
	log.Printf("[INFO] Submitted updated netnamespace: %#v", out)

	d.SetId(out.NetName)

	return resourceOpenshiftNetNamespaceRead(d, meta)
}

func resourceOpenshiftNetNamespaceDelete(d *schema.ResourceData, meta interface{}) error {
	client, err := client_v1.NewForConfig(meta.(*rest.Config))
	if err != nil {
		return err
	}

	name := d.Id()

	log.Printf("[INFO] Deleting netnamespace: %#v", name)

	err = client.NetNamespaces().Delete(name, &meta_v1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("Failed to delete netnamespace: %s", err)
	}

	d.SetId("")
	return nil
}

func resourceOpenshiftNetNamespaceExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client, err := client_v1.NewForConfig(meta.(*rest.Config))
	if err != nil {
		return false, err
	}

	name := d.Id()

	log.Printf("[INFO] Checking netnamespace %s", name)
	_, err = client.NetNamespaces().Get(name, meta_v1.GetOptions{})
	if err != nil {
		if statusErr, ok := err.(*errors.StatusError); ok && statusErr.ErrStatus.Code == 404 {
			return false, nil
		}
		log.Printf("[DEBUG] Received error: %#v", err)
	}
	return true, err
}
