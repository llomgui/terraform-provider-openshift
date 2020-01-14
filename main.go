package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/llomgui/terraform-provider-openshift/openshift"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: openshift.Provider})
}
