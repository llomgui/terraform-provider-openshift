package openshift

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	pkgApi "k8s.io/apimachinery/pkg/types"
	kubernetes "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func resourceOpenshiftSecret() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpenshiftSecretCreate,
		ReadContext:   resourceOpenshiftSecretRead,
		UpdateContext: resourceOpenshiftSecretUpdate,
		DeleteContext: resourceOpenshiftSecretDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"metadata": namespacedMetadataSchema("secret", true),
			"data": {
				Type:        schema.TypeMap,
				Description: "A map of the secret data.",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Sensitive:   true,
			},
			"base64data": {
				Type:        schema.TypeMap,
				Description: "A map of the base64-encoded secret data.",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Sensitive:   true,
			},
			"type": {
				Type:        schema.TypeString,
				Description: "Type of secret",
				Default:     "Opaque",
				Optional:    true,
				ForceNew:    true,
			},
		},
	}
}

func decodeBase64Value(value interface{}) ([]byte, error) {
	enc, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("base64data cannot decode type %T", value)
	}
	return base64.StdEncoding.DecodeString(enc)
}

func resourceOpenshiftSecretCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn, err := kubernetes.NewForConfig(meta.(*rest.Config))
	if err != nil {
		return diag.FromErr(err)
	}

	// Merge data and base64-encoded data into a single data map
	dataMap := d.Get("data").(map[string]interface{})
	for key, value := range d.Get("base64data").(map[string]interface{}) {
		// Decode Terraform's base64 representation to avoid double-encoding in Kubernetes.
		decodedValue, err := decodeBase64Value(value)
		if err != nil {
			return diag.FromErr(err)
		}
		dataMap[key] = string(decodedValue)
	}

	metadata := expandMetadata(d.Get("metadata").([]interface{}))
	secret := api.Secret{
		ObjectMeta: metadata,
		Data:       expandStringMapToByteMap(dataMap),
	}

	if v, ok := d.GetOk("type"); ok {
		secret.Type = api.SecretType(v.(string))
	}

	log.Printf("[INFO] Creating new secret: %#v", secret)
	out, err := conn.CoreV1().Secrets(metadata.Namespace).Create(ctx, &secret, meta_v1.CreateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Submitting new secret: %#v", out)
	d.SetId(buildId(out.ObjectMeta))

	return resourceOpenshiftSecretRead(ctx, d, meta)
}

func resourceOpenshiftSecretRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	exists, err := resourceOpenshiftSecretExists(ctx, d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	if !exists {
		return diag.Diagnostics{}
	}
	conn, err := kubernetes.NewForConfig(meta.(*rest.Config))
	if err != nil {
		return diag.FromErr(err)
	}

	namespace, name, err := idParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Reading secret %s", name)
	secret, err := conn.CoreV1().Secrets(namespace).Get(ctx, name, meta_v1.GetOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Received secret: %#v", secret)
	err = d.Set("metadata", flattenMetadata(secret.ObjectMeta, d))
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("type", secret.Type)
	if err != nil {
		return diag.FromErr(err)
	}

	secretData := flattenByteMapToStringMap(secret.Data)
	// Remove base64data keys from the payload before setting the data key on the resource. If
	// these keys are not removed, they will always show in the diff at update.
	for key, value := range d.Get("base64data").(map[string]interface{}) {
		if _, err := decodeBase64Value(value); err != nil {
			continue
		}
		delete(secretData, key)
	}

	err = d.Set("data", secretData)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceOpenshiftSecretUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn, err := kubernetes.NewForConfig(meta.(*rest.Config))
	if err != nil {
		return diag.FromErr(err)
	}

	namespace, name, err := idParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	ops := patchMetadata("metadata.0.", "/metadata/", d)
	if d.HasChanges([]string{"data", "base64data"}...) {
		oldV, newV := d.GetChange("data")
		oldV = base64EncodeStringMap(oldV.(map[string]interface{}))
		newV = base64EncodeStringMap(newV.(map[string]interface{}))

		oldVB64, newVB64 := d.GetChange("base64data")
		for key, value := range oldVB64.(map[string]interface{}) {
			oldV.(map[string]interface{})[key] = value
		}
		for key, value := range newVB64.(map[string]interface{}) {
			newV.(map[string]interface{})[key] = value
		}

		diffOps := diffStringMap("/data/", oldV.(map[string]interface{}), newV.(map[string]interface{}))

		ops = append(ops, diffOps...)
	}

	data, err := ops.MarshalJSON()
	if err != nil {
		return diag.Errorf("Failed to marshal update operations: %s", err)
	}

	log.Printf("[INFO] Updating secret %q: %v", name, data)
	out, err := conn.CoreV1().Secrets(namespace).Patch(ctx, name, pkgApi.JSONPatchType, data, meta_v1.PatchOptions{})
	if err != nil {
		return diag.Errorf("Failed to update secret: %s", err)
	}

	log.Printf("[INFO] Submitting updated secret: %#v", out)
	d.SetId(buildId(out.ObjectMeta))

	return resourceOpenshiftSecretRead(ctx, d, meta)
}

func resourceOpenshiftSecretDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn, err := kubernetes.NewForConfig(meta.(*rest.Config))
	if err != nil {
		return diag.FromErr(err)
	}

	namespace, name, err := idParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Deleting secret: %q", name)
	err = conn.CoreV1().Secrets(namespace).Delete(ctx, name, meta_v1.DeleteOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Secret %s deleted", name)

	d.SetId("")

	return nil
}

func resourceOpenshiftSecretExists(ctx context.Context, d *schema.ResourceData, meta interface{}) (bool, error) {
	conn, err := kubernetes.NewForConfig(meta.(*rest.Config))
	if err != nil {
		return false, err
	}

	namespace, name, err := idParts(d.Id())
	if err != nil {
		return false, err
	}

	log.Printf("[INFO] Checking secret %s", name)
	_, err = conn.CoreV1().Secrets(namespace).Get(ctx, name, meta_v1.GetOptions{})
	if err != nil {
		if statusErr, ok := err.(*errors.StatusError); ok && statusErr.ErrStatus.Code == 404 {
			return false, nil
		}
		log.Printf("[DEBUG] Received error: %#v", err)
	}

	return true, err
}
