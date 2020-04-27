package openshift

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	api "github.com/openshift/api/route/v1"
	client_v1 "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	pkgApi "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

func resourceOpenshiftRoute() *schema.Resource {
	return &schema.Resource{
		Create: resourceOpenshiftRouteCreate,
		Read:   resourceOpenshiftRouteRead,
		Update: resourceOpenshiftRouteUpdate,
		Delete: resourceOpenshiftRouteDelete,
		Exists: resourceOpenshiftRouteExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"metadata": namespacedMetadataSchema("route", true),
			"spec": {
				Type:        schema.TypeList,
				Description: "Spec defines the behavior of a route.",
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host": {
							Type:        schema.TypeString,
							Description: "(string) host is an alias/DNS that points to the service. Optional. If not specified a route name will typically be automatically chosen. Must follow DNS952 subdomain conventions.",
							Computed:    true,
							Optional:    true,
						},
						"path": {
							Type:        schema.TypeString,
							Description: "(string) Path that the router watches for, to route traffic for to the service. Optional",
							Optional:    true,
						},
						"port": {
							Type:        schema.TypeList,
							Description: "(v1.RoutePort) If specified, the port to be used by the router. Most routers will use all endpoints exposed by the service by default - set this value to instruct routers which port to use.",
							MaxItems:    1,
							Optional:    true,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"target_port": {
										Type:        schema.TypeString,
										Description: "(intstr.IntOrString) The target port on pods selected by the service this route points to. If this is a string, it will be looked up as a named port in the target endpoints port list. Required",
										Required:    true,
									},
								},
							},
						},
						"tls": {
							Type:        schema.TypeList,
							Description: "(v1.TLSConfig) The tls field provides the ability to configure certificates and termination for the route.",
							MaxItems:    1,
							Optional:    true,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"termination": {
										Type:         schema.TypeString,
										Description:  "(string) termination indicates termination type.",
										Computed:     true,
										Optional:     true,
										ValidateFunc: validation.StringInSlice([]string{"edge", "passthrough", "reencrypt"}, false),
									},
									"certificate": {
										Type:        schema.TypeString,
										Description: "(string) certificate provides certificate contents",
										Optional:    true,
									},
									"key": {
										Type:        schema.TypeString,
										Description: "(string) key provides key file contents",
										Optional:    true,
									},
									"ca_certificate": {
										Type:        schema.TypeString,
										Description: "(string) caCertificate provides the cert authority certificate contents",
										Optional:    true,
									},
									"destination_ca_certificate": {
										Type:        schema.TypeString,
										Description: "(string) destinationCACertificate provides the contents of the ca certificate of the final destination.  When using reencrypt termination this file should be provided in order to have routers use it for health checks on the secure connection. If this field is not specified, the router may provide its own destination CA and perform hostname validation using the short service name (service.namespace.svc), which allows infrastructure generated certificates to automatically verify.",
										Optional:    true,
									},
									"insecure_edge_termination_policy": {
										Type:         schema.TypeString,
										Description:  "(string) insecureEdgeTerminationPolicy indicates the desired behavior for insecure connections to a route. While each router may make its own decisions on which ports to expose, this is normally port 80.",
										Computed:     true,
										Optional:     true,
										ValidateFunc: validation.StringInSlice([]string{"None", "Allow", "Redirect"}, false),
									},
								},
							},
						},
						"to": {
							Type:        schema.TypeList,
							Description: "(v1.RouteTargetReference) to is an object the route should use as the primary backend. Only the Service kind is allowed, and it will be defaulted to Service. If the weight field (0-256 default 1) is set to zero, no traffic will be sent to this backend.",
							Required:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"kind": {
										Type:        schema.TypeString,
										Description: "(string) The kind of target that the route is referring to. Currently, only 'Service' is allowed",
										Required:    true,
									},
									"name": {
										Type:        schema.TypeString,
										Description: "(string) name of the service/target that is being referred to. e.g. name of the service",
										Required:    true,
									},
									"weight": {
										Type:        schema.TypeInt,
										Description: "(integer) weight as an integer between 0 and 256, default 1, that specifies the target's relative weight against other target reference objects. 0 suppresses requests to this backend.",
										Computed:    true,
										Optional:    true,
									},
								},
							},
						},
						"wildcard_policy": {
							Type:         schema.TypeString,
							Description:  "(string) Wildcard policy if any for the route. Currently only 'Subdomain' or 'None' is allowed.",
							Optional:     true,
							Default:      "None",
							ValidateFunc: validation.StringInSlice([]string{"Subdomain", "None"}, false),
						},
					},
				},
			},
		},
	}
}

func resourceOpenshiftRouteCreate(d *schema.ResourceData, meta interface{}) error {
	client, err := client_v1.NewForConfig(meta.(*rest.Config))
	if err != nil {
		return err
	}

	metadata := expandMetadata(d.Get("metadata").([]interface{}))
	route := api.Route{
		ObjectMeta: metadata,
		Spec:       expandRouteSpec(d.Get("spec").([]interface{})),
	}

	log.Printf("[INFO] Creating new route: %#v", route)
	out, err := client.Routes(metadata.Namespace).Create(&route)
	if err != nil {
		return fmt.Errorf("Failed to create route: %s", err)
	}
	log.Printf("[INFO] Submitted new route: %#v", out)
	d.SetId(buildId(out.ObjectMeta))

	return resourceOpenshiftRouteRead(d, meta)
}

func resourceOpenshiftRouteRead(d *schema.ResourceData, meta interface{}) error {
	client, err := client_v1.NewForConfig(meta.(*rest.Config))
	if err != nil {
		return err
	}

	namespace, name, err := idParts(d.Id())
	if err != nil {
		return err
	}

	log.Printf("[INFO] Reading route %s", name)
	route, err := client.Routes(namespace).Get(name, meta_v1.GetOptions{})
	if err != nil {
		log.Printf("[DEBUG] Received error: %#v", err)
		return err
	}

	log.Printf("[INFO] Received route: %#v", route)
	err = d.Set("metadata", flattenMetadata(route.ObjectMeta, d))
	if err != nil {
		return err
	}

	spec, err := flattenRouteSpec(route.Spec)
	if err != nil {
		return err
	}

	err = d.Set("spec", spec)
	if err != nil {
		return err
	}

	return nil
}

func resourceOpenshiftRouteUpdate(d *schema.ResourceData, meta interface{}) error {
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
		diffOps, err := patchRouteSpec("spec.0.", "/spec/", d)
		if err != nil {
			return err
		}
		ops = append(ops, diffOps...)
	}

	data, err := ops.MarshalJSON()
	if err != nil {
		return fmt.Errorf("Failed to marshal update operations: %s", err)
	}

	log.Printf("[INFO] Updating route %q: %v", name, string(data))
	out, err := client.Routes(namespace).Patch(name, pkgApi.JSONPatchType, data)
	if err != nil {
		return fmt.Errorf("Failed to update route: %s", err)
	}
	log.Printf("[INFO] Submitted updated route: %#v", out)

	d.SetId(buildId(out.ObjectMeta))

	return resourceOpenshiftRouteRead(d, meta)
}

func resourceOpenshiftRouteDelete(d *schema.ResourceData, meta interface{}) error {
	client, err := client_v1.NewForConfig(meta.(*rest.Config))
	if err != nil {
		return err
	}

	namespace, name, err := idParts(d.Id())
	if err != nil {
		return err
	}

	log.Printf("[INFO] Deleting route: %#v", name)

	err = client.Routes(namespace).Delete(name, &meta_v1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("Failed to delete route: %s", err)
	}

	d.SetId("")
	return nil
}

func resourceOpenshiftRouteExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client, err := client_v1.NewForConfig(meta.(*rest.Config))
	if err != nil {
		return false, err
	}

	namespace, name, err := idParts(d.Id())
	if err != nil {
		return false, err
	}

	log.Printf("[INFO] Checking route %s", name)
	_, err = client.Routes(namespace).Get(name, meta_v1.GetOptions{})
	if err != nil {
		if statusErr, ok := err.(*errors.StatusError); ok && statusErr.ErrStatus.Code == 404 {
			return false, nil
		}
		log.Printf("[DEBUG] Received error: %#v", err)
	}
	return true, err
}
