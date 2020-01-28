package openshift

import (
	v1 "k8s.io/api/core/v1"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// Flatteners

func flattenAWSElasticBlockStoreVolumeSource(in *v1.AWSElasticBlockStoreVolumeSource) []interface{} {
	att := make(map[string]interface{})
	att["volume_id"] = in.VolumeID
	if in.FSType != "" {
		att["fs_type"] = in.FSType
	}
	if in.Partition != 0 {
		att["partition"] = in.Partition
	}
	if in.ReadOnly {
		att["read_only"] = in.ReadOnly
	}
	return []interface{}{att}
}

func flattenAzureDiskVolumeSource(in *v1.AzureDiskVolumeSource) []interface{} {
	att := make(map[string]interface{})
	att["disk_name"] = in.DiskName
	att["data_disk_uri"] = in.DataDiskURI
	att["caching_mode"] = string(*in.CachingMode)
	if in.FSType != nil {
		att["fs_type"] = *in.FSType
	}
	if in.ReadOnly != nil {
		att["read_only"] = *in.ReadOnly
	}
	return []interface{}{att}
}

func flattenAzureFileVolumeSource(in *v1.AzureFileVolumeSource) []interface{} {
	att := make(map[string]interface{})
	att["secret_name"] = in.SecretName
	att["share_name"] = in.ShareName
	if in.ReadOnly {
		att["read_only"] = in.ReadOnly
	}
	return []interface{}{att}
}

func flattenCephFSVolumeSource(in *v1.CephFSVolumeSource) []interface{} {
	att := make(map[string]interface{})
	att["monitors"] = newStringSet(schema.HashString, in.Monitors)
	if in.Path != "" {
		att["path"] = in.Path
	}
	if in.User != "" {
		att["user"] = in.User
	}
	if in.SecretFile != "" {
		att["secret_file"] = in.SecretFile
	}
	if in.SecretRef != nil {
		att["secret_ref"] = flattenLocalObjectReference(in.SecretRef)
	}
	if in.ReadOnly {
		att["read_only"] = in.ReadOnly
	}
	return []interface{}{att}
}

func flattenCinderVolumeSource(in *v1.CinderVolumeSource) []interface{} {
	att := make(map[string]interface{})
	att["volume_id"] = in.VolumeID
	if in.FSType != "" {
		att["fs_type"] = in.FSType
	}
	if in.ReadOnly {
		att["read_only"] = in.ReadOnly
	}
	return []interface{}{att}
}

func flattenFCVolumeSource(in *v1.FCVolumeSource) []interface{} {
	att := make(map[string]interface{})
	att["target_ww_ns"] = newStringSet(schema.HashString, in.TargetWWNs)
	att["lun"] = *in.Lun
	if in.FSType != "" {
		att["fs_type"] = in.FSType
	}
	if in.ReadOnly {
		att["read_only"] = in.ReadOnly
	}
	return []interface{}{att}
}

func flattenFlexVolumeSource(in *v1.FlexVolumeSource) []interface{} {
	att := make(map[string]interface{})
	att["driver"] = in.Driver
	if in.FSType != "" {
		att["fs_type"] = in.FSType
	}
	if in.SecretRef != nil {
		att["secret_ref"] = flattenLocalObjectReference(in.SecretRef)
	}
	if in.ReadOnly {
		att["read_only"] = in.ReadOnly
	}
	if len(in.Options) > 0 {
		att["options"] = in.Options
	}
	return []interface{}{att}
}

func flattenFlockerVolumeSource(in *v1.FlockerVolumeSource) []interface{} {
	att := make(map[string]interface{})
	att["dataset_name"] = in.DatasetName
	att["dataset_uuid"] = in.DatasetUUID
	return []interface{}{att}
}

func flattenGCEPersistentDiskVolumeSource(in *v1.GCEPersistentDiskVolumeSource) []interface{} {
	att := make(map[string]interface{})
	att["pd_name"] = in.PDName
	if in.FSType != "" {
		att["fs_type"] = in.FSType
	}
	if in.Partition != 0 {
		att["partition"] = in.Partition
	}
	if in.ReadOnly {
		att["read_only"] = in.ReadOnly
	}
	return []interface{}{att}
}

func flattenGlusterfsVolumeSource(in *v1.GlusterfsVolumeSource) []interface{} {
	att := make(map[string]interface{})
	att["endpoints_name"] = in.EndpointsName
	att["path"] = in.Path
	if in.ReadOnly {
		att["read_only"] = in.ReadOnly
	}
	return []interface{}{att}
}

func flattenHostPathVolumeSource(in *v1.HostPathVolumeSource) []interface{} {
	att := make(map[string]interface{})
	att["path"] = in.Path
	return []interface{}{att}
}

func flattenISCSIVolumeSource(in *v1.ISCSIVolumeSource) []interface{} {
	att := make(map[string]interface{})
	if in.TargetPortal != "" {
		att["target_portal"] = in.TargetPortal
	}
	if in.IQN != "" {
		att["iqn"] = in.IQN
	}
	if in.Lun != 0 {
		att["lun"] = in.Lun
	}
	if in.ISCSIInterface != "" {
		att["iscsi_interface"] = in.ISCSIInterface
	}
	if in.FSType != "" {
		att["fs_type"] = in.FSType
	}
	if in.ReadOnly {
		att["read_only"] = in.ReadOnly
	}
	return []interface{}{att}
}

func flattenLocalObjectReference(in *v1.LocalObjectReference) []interface{} {
	att := make(map[string]interface{})
	if in.Name != "" {
		att["name"] = in.Name
	}
	return []interface{}{att}
}

func flattenNFSVolumeSource(in *v1.NFSVolumeSource) []interface{} {
	att := make(map[string]interface{})
	att["server"] = in.Server
	att["path"] = in.Path
	if in.ReadOnly {
		att["read_only"] = in.ReadOnly
	}
	return []interface{}{att}
}

func flattenPhotonPersistentDiskVolumeSource(in *v1.PhotonPersistentDiskVolumeSource) []interface{} {
	att := make(map[string]interface{})
	att["pd_id"] = in.PdID
	if in.FSType != "" {
		att["fs_type"] = in.FSType
	}
	return []interface{}{att}
}

func flattenQuobyteVolumeSource(in *v1.QuobyteVolumeSource) []interface{} {
	att := make(map[string]interface{})
	att["registry"] = in.Registry
	att["volume"] = in.Volume
	if in.ReadOnly {
		att["read_only"] = in.ReadOnly
	}
	if in.User != "" {
		att["user"] = in.User
	}
	if in.Group != "" {
		att["group"] = in.Group
	}
	return []interface{}{att}
}

func flattenRBDVolumeSource(in *v1.RBDVolumeSource) []interface{} {
	att := make(map[string]interface{})
	att["ceph_monitors"] = newStringSet(schema.HashString, in.CephMonitors)
	att["rbd_image"] = in.RBDImage
	if in.FSType != "" {
		att["fs_type"] = in.FSType
	}
	if in.RBDPool != "" {
		att["rbd_pool"] = in.RBDPool
	}
	if in.RadosUser != "" {
		att["rados_user"] = in.RadosUser
	}
	if in.Keyring != "" {
		att["keyring"] = in.Keyring
	}
	if in.SecretRef != nil {
		att["secret_ref"] = flattenLocalObjectReference(in.SecretRef)
	}
	if in.ReadOnly {
		att["read_only"] = in.ReadOnly
	}
	return []interface{}{att}
}

func flattenVsphereVirtualDiskVolumeSource(in *v1.VsphereVirtualDiskVolumeSource) []interface{} {
	att := make(map[string]interface{})
	att["volume_path"] = in.VolumePath
	if in.FSType != "" {
		att["fs_type"] = in.FSType
	}
	return []interface{}{att}
}

// Expanders

func expandAWSElasticBlockStoreVolumeSource(l []interface{}) *v1.AWSElasticBlockStoreVolumeSource {
	if len(l) == 0 || l[0] == nil {
		return &v1.AWSElasticBlockStoreVolumeSource{}
	}
	in := l[0].(map[string]interface{})
	obj := &v1.AWSElasticBlockStoreVolumeSource{
		VolumeID: in["volume_id"].(string),
	}
	if v, ok := in["fs_type"].(string); ok {
		obj.FSType = v
	}
	if v, ok := in["partition"].(int); ok {
		obj.Partition = int32(v)
	}
	if v, ok := in["read_only"].(bool); ok {
		obj.ReadOnly = v
	}
	return obj
}

func expandAzureDiskVolumeSource(l []interface{}) *v1.AzureDiskVolumeSource {
	if len(l) == 0 || l[0] == nil {
		return &v1.AzureDiskVolumeSource{}
	}
	in := l[0].(map[string]interface{})
	cachingMode := v1.AzureDataDiskCachingMode(in["caching_mode"].(string))
	obj := &v1.AzureDiskVolumeSource{
		CachingMode: &cachingMode,
		DiskName:    in["disk_name"].(string),
		DataDiskURI: in["data_disk_uri"].(string),
	}
	if v, ok := in["fs_type"].(string); ok {
		obj.FSType = ptrToString(v)
	}
	if v, ok := in["read_only"].(bool); ok {
		obj.ReadOnly = ptrToBool(v)
	}
	return obj
}

func expandAzureFileVolumeSource(l []interface{}) *v1.AzureFileVolumeSource {
	if len(l) == 0 || l[0] == nil {
		return &v1.AzureFileVolumeSource{}
	}
	in := l[0].(map[string]interface{})
	obj := &v1.AzureFileVolumeSource{
		SecretName: in["secret_name"].(string),
		ShareName:  in["share_name"].(string),
	}
	if v, ok := in["read_only"].(bool); ok {
		obj.ReadOnly = v
	}
	return obj
}

func expandCephFSVolumeSource(l []interface{}) *v1.CephFSVolumeSource {
	if len(l) == 0 || l[0] == nil {
		return &v1.CephFSVolumeSource{}
	}
	in := l[0].(map[string]interface{})
	obj := &v1.CephFSVolumeSource{
		Monitors: sliceOfString(in["monitors"].(*schema.Set).List()),
	}
	if v, ok := in["path"].(string); ok {
		obj.Path = v
	}
	if v, ok := in["user"].(string); ok {
		obj.User = v
	}
	if v, ok := in["secret_file"].(string); ok {
		obj.SecretFile = v
	}
	if v, ok := in["secret_ref"].([]interface{}); ok && len(v) > 0 {
		obj.SecretRef = expandLocalObjectReference(v)
	}
	if v, ok := in["read_only"].(bool); ok {
		obj.ReadOnly = v
	}
	return obj
}

func expandCinderVolumeSource(l []interface{}) *v1.CinderVolumeSource {
	if len(l) == 0 || l[0] == nil {
		return &v1.CinderVolumeSource{}
	}
	in := l[0].(map[string]interface{})
	obj := &v1.CinderVolumeSource{
		VolumeID: in["volume_id"].(string),
	}
	if v, ok := in["fs_type"].(string); ok {
		obj.FSType = v
	}
	if v, ok := in["read_only"].(bool); ok {
		obj.ReadOnly = v
	}
	return obj
}

func expandFCVolumeSource(l []interface{}) *v1.FCVolumeSource {
	if len(l) == 0 || l[0] == nil {
		return &v1.FCVolumeSource{}
	}
	in := l[0].(map[string]interface{})
	obj := &v1.FCVolumeSource{
		TargetWWNs: sliceOfString(in["target_ww_ns"].(*schema.Set).List()),
		Lun:        ptrToInt32(int32(in["lun"].(int))),
	}
	if v, ok := in["fs_type"].(string); ok {
		obj.FSType = v
	}
	if v, ok := in["read_only"].(bool); ok {
		obj.ReadOnly = v
	}
	return obj
}

func expandFlexVolumeSource(l []interface{}) *v1.FlexVolumeSource {
	if len(l) == 0 || l[0] == nil {
		return &v1.FlexVolumeSource{}
	}
	in := l[0].(map[string]interface{})
	obj := &v1.FlexVolumeSource{
		Driver: in["driver"].(string),
	}
	if v, ok := in["fs_type"].(string); ok {
		obj.FSType = v
	}
	if v, ok := in["secret_ref"].([]interface{}); ok && len(v) > 0 {
		obj.SecretRef = expandLocalObjectReference(v)
	}
	if v, ok := in["read_only"].(bool); ok {
		obj.ReadOnly = v
	}
	if v, ok := in["options"].(map[string]interface{}); ok && len(v) > 0 {
		obj.Options = expandStringMap(v)
	}
	return obj
}

func expandFlockerVolumeSource(l []interface{}) *v1.FlockerVolumeSource {
	if len(l) == 0 || l[0] == nil {
		return &v1.FlockerVolumeSource{}
	}
	in := l[0].(map[string]interface{})
	obj := &v1.FlockerVolumeSource{
		DatasetName: in["dataset_name"].(string),
		DatasetUUID: in["dataset_uuid"].(string),
	}
	return obj
}

func expandGCEPersistentDiskVolumeSource(l []interface{}) *v1.GCEPersistentDiskVolumeSource {
	if len(l) == 0 || l[0] == nil {
		return &v1.GCEPersistentDiskVolumeSource{}
	}
	in := l[0].(map[string]interface{})
	obj := &v1.GCEPersistentDiskVolumeSource{
		PDName: in["pd_name"].(string),
	}
	if v, ok := in["fs_type"].(string); ok {
		obj.FSType = v
	}
	if v, ok := in["partition"].(int); ok {
		obj.Partition = int32(v)
	}
	if v, ok := in["read_only"].(bool); ok {
		obj.ReadOnly = v
	}
	return obj
}

func expandGlusterfsVolumeSource(l []interface{}) *v1.GlusterfsVolumeSource {
	if len(l) == 0 || l[0] == nil {
		return &v1.GlusterfsVolumeSource{}
	}
	in := l[0].(map[string]interface{})
	obj := &v1.GlusterfsVolumeSource{
		EndpointsName: in["endpoints_name"].(string),
		Path:          in["path"].(string),
	}
	if v, ok := in["read_only"].(bool); ok {
		obj.ReadOnly = v
	}
	return obj
}

func expandHostPathVolumeSource(l []interface{}) *v1.HostPathVolumeSource {
	if len(l) == 0 || l[0] == nil {
		return &v1.HostPathVolumeSource{}
	}
	in := l[0].(map[string]interface{})
	obj := &v1.HostPathVolumeSource{
		Path: in["path"].(string),
	}
	return obj
}

func expandISCSIVolumeSource(l []interface{}) *v1.ISCSIVolumeSource {
	if len(l) == 0 || l[0] == nil {
		return &v1.ISCSIVolumeSource{}
	}
	in := l[0].(map[string]interface{})
	obj := &v1.ISCSIVolumeSource{
		TargetPortal: in["target_portal"].(string),
		IQN:          in["iqn"].(string),
	}
	if v, ok := in["lun"].(int); ok {
		obj.Lun = int32(v)
	}
	if v, ok := in["iscsi_interface"].(string); ok {
		obj.ISCSIInterface = v
	}
	if v, ok := in["fs_type"].(string); ok {
		obj.FSType = v
	}
	if v, ok := in["read_only"].(bool); ok {
		obj.ReadOnly = v
	}
	return obj
}

func expandLocalObjectReference(l []interface{}) *v1.LocalObjectReference {
	if len(l) == 0 || l[0] == nil {
		return &v1.LocalObjectReference{}
	}
	in := l[0].(map[string]interface{})
	obj := &v1.LocalObjectReference{}
	if v, ok := in["name"].(string); ok {
		obj.Name = v
	}
	return obj
}

func expandNFSVolumeSource(l []interface{}) *v1.NFSVolumeSource {
	if len(l) == 0 || l[0] == nil {
		return &v1.NFSVolumeSource{}
	}
	in := l[0].(map[string]interface{})
	obj := &v1.NFSVolumeSource{
		Server: in["server"].(string),
		Path:   in["path"].(string),
	}
	if v, ok := in["read_only"].(bool); ok {
		obj.ReadOnly = v
	}
	return obj
}

func expandPhotonPersistentDiskVolumeSource(l []interface{}) *v1.PhotonPersistentDiskVolumeSource {
	if len(l) == 0 || l[0] == nil {
		return &v1.PhotonPersistentDiskVolumeSource{}
	}
	in := l[0].(map[string]interface{})
	obj := &v1.PhotonPersistentDiskVolumeSource{
		PdID: in["pd_id"].(string),
	}
	if v, ok := in["fs_type"].(string); ok {
		obj.FSType = v
	}
	return obj
}

func expandQuobyteVolumeSource(l []interface{}) *v1.QuobyteVolumeSource {
	if len(l) == 0 || l[0] == nil {
		return &v1.QuobyteVolumeSource{}
	}
	in := l[0].(map[string]interface{})
	obj := &v1.QuobyteVolumeSource{
		Registry: in["registry"].(string),
		Volume:   in["volume"].(string),
	}
	if v, ok := in["read_only"].(bool); ok {
		obj.ReadOnly = v
	}
	if v, ok := in["user"].(string); ok {
		obj.User = v
	}
	if v, ok := in["group"].(string); ok {
		obj.Group = v
	}
	return obj
}

func expandRBDVolumeSource(l []interface{}) *v1.RBDVolumeSource {
	if len(l) == 0 || l[0] == nil {
		return &v1.RBDVolumeSource{}
	}
	in := l[0].(map[string]interface{})
	obj := &v1.RBDVolumeSource{
		CephMonitors: expandStringSlice(in["ceph_monitors"].(*schema.Set).List()),
		RBDImage:     in["rbd_image"].(string),
	}
	if v, ok := in["fs_type"].(string); ok {
		obj.FSType = v
	}
	if v, ok := in["rbd_pool"].(string); ok {
		obj.RBDPool = v
	}
	if v, ok := in["rados_user"].(string); ok {
		obj.RadosUser = v
	}
	if v, ok := in["keyring"].(string); ok {
		obj.Keyring = v
	}
	if v, ok := in["secret_ref"].([]interface{}); ok && len(v) > 0 {
		obj.SecretRef = expandLocalObjectReference(v)
	}
	if v, ok := in["read_only"].(bool); ok {
		obj.ReadOnly = v
	}
	return obj
}

func expandVsphereVirtualDiskVolumeSource(l []interface{}) *v1.VsphereVirtualDiskVolumeSource {
	if len(l) == 0 || l[0] == nil {
		return &v1.VsphereVirtualDiskVolumeSource{}
	}
	in := l[0].(map[string]interface{})
	obj := &v1.VsphereVirtualDiskVolumeSource{
		VolumePath: in["volume_path"].(string),
	}
	if v, ok := in["fs_type"].(string); ok {
		obj.FSType = v
	}
	return obj
}
