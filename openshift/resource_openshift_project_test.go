package openshift

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	api "github.com/openshift/api/project/v1"
	client_v1 "github.com/openshift/client-go/project/clientset/versioned/typed/project/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

func TestAccOpenshiftProject_basic(t *testing.T) {
	var conf api.Project
	nsName := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "openshift_project.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckOpenshiftProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOpenshiftProjectConfig_basic(nsName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOpenshiftProjectExists("openshift_project.test", &conf),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.annotations.%", "0"),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.labels.%", "0"),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.name", nsName),
					resource.TestCheckResourceAttrSet("openshift_project.test", "metadata.0.generation"),
					resource.TestCheckResourceAttrSet("openshift_project.test", "metadata.0.resource_version"),
					resource.TestCheckResourceAttrSet("openshift_project.test", "metadata.0.self_link"),
					resource.TestCheckResourceAttrSet("openshift_project.test", "metadata.0.uid"),
				),
			},
			{
				Config: testAccOpenshiftProjectConfig_addAnnotations(nsName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOpenshiftProjectExists("openshift_project.test", &conf),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.annotations.%", "2"),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.annotations.TestAnnotationOne", "one"),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.annotations.TestAnnotationTwo", "two"),
					testAccCheckMetaAnnotations(&conf.ObjectMeta, map[string]string{"TestAnnotationOne": "one", "TestAnnotationTwo": "two"}),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.labels.%", "0"),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.name", nsName),
					resource.TestCheckResourceAttrSet("openshift_project.test", "metadata.0.generation"),
					resource.TestCheckResourceAttrSet("openshift_project.test", "metadata.0.resource_version"),
					resource.TestCheckResourceAttrSet("openshift_project.test", "metadata.0.self_link"),
					resource.TestCheckResourceAttrSet("openshift_project.test", "metadata.0.uid"),
				),
			},
			{
				Config: testAccOpenshiftProjectConfig_addLabels(nsName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOpenshiftProjectExists("openshift_project.test", &conf),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.annotations.%", "2"),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.annotations.TestAnnotationOne", "one"),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.annotations.TestAnnotationTwo", "two"),
					testAccCheckMetaAnnotations(&conf.ObjectMeta, map[string]string{"TestAnnotationOne": "one", "TestAnnotationTwo": "two"}),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.labels.%", "3"),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.labels.TestLabelOne", "one"),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.labels.TestLabelTwo", "two"),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.labels.TestLabelThree", "three"),
					testAccCheckMetaLabels(&conf.ObjectMeta, map[string]string{"TestLabelOne": "one", "TestLabelTwo": "two", "TestLabelThree": "three"}),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.name", nsName),
					resource.TestCheckResourceAttrSet("openshift_project.test", "metadata.0.generation"),
					resource.TestCheckResourceAttrSet("openshift_project.test", "metadata.0.resource_version"),
					resource.TestCheckResourceAttrSet("openshift_project.test", "metadata.0.self_link"),
					resource.TestCheckResourceAttrSet("openshift_project.test", "metadata.0.uid"),
				),
			},
			{
				Config: testAccOpenshiftProjectConfig_smallerLists(nsName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOpenshiftProjectExists("openshift_project.test", &conf),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.annotations.%", "2"),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.annotations.TestAnnotationOne", "one"),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.annotations.Different", "1234"),
					testAccCheckMetaAnnotations(&conf.ObjectMeta, map[string]string{"TestAnnotationOne": "one", "Different": "1234"}),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.labels.%", "2"),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.labels.TestLabelOne", "one"),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.labels.TestLabelThree", "three"),
					testAccCheckMetaLabels(&conf.ObjectMeta, map[string]string{"TestLabelOne": "one", "TestLabelThree": "three"}),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.name", nsName),
					resource.TestCheckResourceAttrSet("openshift_project.test", "metadata.0.generation"),
					resource.TestCheckResourceAttrSet("openshift_project.test", "metadata.0.resource_version"),
					resource.TestCheckResourceAttrSet("openshift_project.test", "metadata.0.self_link"),
					resource.TestCheckResourceAttrSet("openshift_project.test", "metadata.0.uid"),
				),
			},
			{
				Config: testAccOpenshiftProjectConfig_noLists(nsName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOpenshiftProjectExists("openshift_project.test", &conf),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.annotations.%", "0"),
					testAccCheckMetaAnnotations(&conf.ObjectMeta, map[string]string{}),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.labels.%", "0"),
					testAccCheckMetaLabels(&conf.ObjectMeta, map[string]string{}),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.name", nsName),
					resource.TestCheckResourceAttrSet("openshift_project.test", "metadata.0.generation"),
					resource.TestCheckResourceAttrSet("openshift_project.test", "metadata.0.resource_version"),
					resource.TestCheckResourceAttrSet("openshift_project.test", "metadata.0.self_link"),
					resource.TestCheckResourceAttrSet("openshift_project.test", "metadata.0.uid"),
				),
			},
		},
	})
}

func TestAccOpenshiftProject_importBasic(t *testing.T) {
	resourceName := "openshift_project.test"
	nsName := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckOpenshiftProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOpenshiftProjectConfig_basic(nsName),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metadata.0.resource_version"},
			},
		},
	})
}

func TestAccOpenshiftProject_generatedName(t *testing.T) {
	var conf api.Project
	prefix := "tf-acc-test-gen-"

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "openshift_project.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckOpenshiftProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOpenshiftProjectConfig_generatedName(prefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOpenshiftProjectExists("openshift_project.test", &conf),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.annotations.%", "0"),
					testAccCheckMetaAnnotations(&conf.ObjectMeta, map[string]string{}),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.labels.%", "0"),
					testAccCheckMetaLabels(&conf.ObjectMeta, map[string]string{}),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.generate_name", prefix),
					resource.TestMatchResourceAttr("openshift_project.test", "metadata.0.name", regexp.MustCompile("^"+prefix)),
					resource.TestCheckResourceAttrSet("openshift_project.test", "metadata.0.generation"),
					resource.TestCheckResourceAttrSet("openshift_project.test", "metadata.0.resource_version"),
					resource.TestCheckResourceAttrSet("openshift_project.test", "metadata.0.self_link"),
					resource.TestCheckResourceAttrSet("openshift_project.test", "metadata.0.uid"),
				),
			},
		},
	})
}

func TestAccOpenshiftProject_withSpecialCharacters(t *testing.T) {
	var conf api.Project
	nsName := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "openshift_project.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckOpenshiftProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOpenshiftProjectConfig_specialCharacters(nsName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOpenshiftProjectExists("openshift_project.test", &conf),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.annotations.%", "2"),
					testAccCheckMetaAnnotations(&conf.ObjectMeta, map[string]string{
						"myhost.co.uk/any-path": "one",
						"Different":             "1234",
					}),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.annotations.myhost.co.uk/any-path", "one"),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.annotations.Different", "1234"),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.labels.%", "2"),
					testAccCheckMetaLabels(&conf.ObjectMeta, map[string]string{
						"myhost.co.uk/any-path": "one",
						"TestLabelThree":        "three",
					}),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.labels.myhost.co.uk/any-path", "one"),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.labels.TestLabelThree", "three"),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.name", nsName),
					resource.TestCheckResourceAttrSet("openshift_project.test", "metadata.0.generation"),
					resource.TestCheckResourceAttrSet("openshift_project.test", "metadata.0.resource_version"),
					resource.TestCheckResourceAttrSet("openshift_project.test", "metadata.0.self_link"),
					resource.TestCheckResourceAttrSet("openshift_project.test", "metadata.0.uid"),
				),
			},
		},
	})
}

func TestAccOpenshiftProject_importGeneratedName(t *testing.T) {
	resourceName := "openshift_project.test"
	prefix := "tf-acc-test-gen-import-"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckOpenshiftProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOpenshiftProjectConfig_generatedName(prefix),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metadata.0.resource_version"},
			},
		},
	})
}

func TestAccOpenshiftProject_deleteTimeout(t *testing.T) {
	var conf api.Project
	nsName := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "openshift_project.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckOpenshiftProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOpenshiftProjectConfig_deleteTimeout(nsName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOpenshiftProjectExists("openshift_project.test", &conf),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.annotations.%", "0"),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.labels.%", "0"),
					resource.TestCheckResourceAttr("openshift_project.test", "metadata.0.name", nsName),
					resource.TestCheckResourceAttrSet("openshift_project.test", "metadata.0.generation"),
					resource.TestCheckResourceAttrSet("openshift_project.test", "metadata.0.resource_version"),
					resource.TestCheckResourceAttrSet("openshift_project.test", "metadata.0.self_link"),
					resource.TestCheckResourceAttrSet("openshift_project.test", "metadata.0.uid"),
				),
			},
		},
	})
}

func testAccCheckMetaAnnotations(om *meta_v1.ObjectMeta, expected map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(expected) == 0 && len(om.Annotations) == 0 {
			return nil
		}

		// Remove any internal k8s annotations unless we expect them
		annotations := om.Annotations
		for key := range annotations {
			_, isExpected := expected[key]
			if isInternalKey(key) && !isExpected {
				delete(annotations, key)
			}
		}

		if !reflect.DeepEqual(annotations, expected) {
			return fmt.Errorf("%s annotations don't match.\nExpected: %q\nGiven: %q",
				om.Name, expected, om.Annotations)
		}
		return nil
	}
}

func testAccCheckMetaLabels(om *meta_v1.ObjectMeta, expected map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(expected) == 0 && len(om.Labels) == 0 {
			return nil
		}

		// Remove any internal k8s labels unless we expect them
		labels := om.Labels
		for key := range labels {
			_, isExpected := expected[key]
			if isInternalKey(key) && !isExpected {
				delete(labels, key)
			}
		}

		if !reflect.DeepEqual(labels, expected) {
			return fmt.Errorf("%s labels don't match.\nExpected: %q\nGiven: %q",
				om.Name, expected, om.Labels)
		}
		return nil
	}
}

func testAccCheckOpenshiftProjectDestroy(s *terraform.State) error {
	client, err := client_v1.NewForConfig(testAccProvider.Meta().(*rest.Config))
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "openshift_project" {
			continue
		}

		resp, err := client.Projects().Get(rs.Primary.ID, meta_v1.GetOptions{})
		if err == nil {
			if resp.Name == rs.Primary.ID {
				return fmt.Errorf("Project still exists: %s", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckOpenshiftProjectExists(n string, obj *api.Project) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		client, err := client_v1.NewForConfig(testAccProvider.Meta().(*rest.Config))
		if err != nil {
			return err
		}
		out, err := client.Projects().Get(rs.Primary.ID, meta_v1.GetOptions{})
		if err != nil {
			return err
		}

		*obj = *out
		return nil
	}
}

func testAccOpenshiftProjectConfig_basic(nsName string) string {
	return fmt.Sprintf(`
resource "openshift_project" "test" {
  metadata {
    name = "%s"
  }
}
`, nsName)
}

func testAccOpenshiftProjectConfig_addAnnotations(nsName string) string {
	return fmt.Sprintf(`
resource "openshift_project" "test" {
  metadata {
		annotations = {
      TestAnnotationOne = "one"
      TestAnnotationTwo = "two"
    }
    name = "%s"
  }
}
`, nsName)
}
func testAccOpenshiftProjectConfig_addLabels(nsName string) string {
	return fmt.Sprintf(`
resource "openshift_project" "test" {
  metadata {
    annotations = {
      TestAnnotationOne = "one"
      TestAnnotationTwo = "two"
    }

    labels = {
      TestLabelOne   = "one"
      TestLabelTwo   = "two"
      TestLabelThree = "three"
    }

    name = "%s"
  }
}
`, nsName)
}

func testAccOpenshiftProjectConfig_smallerLists(nsName string) string {
	return fmt.Sprintf(`
resource "openshift_project" "test" {
  metadata {
    annotations = {
      TestAnnotationOne = "one"
      Different         = "1234"
    }

    labels = {
      TestLabelOne   = "one"
      TestLabelThree = "three"
    }

    name = "%s"
  }
}
`, nsName)
}

func testAccOpenshiftProjectConfig_noLists(nsName string) string {
	return fmt.Sprintf(`
resource "openshift_project" "test" {
  metadata {
    name = "%s"
  }
}
`, nsName)
}

func testAccOpenshiftProjectConfig_generatedName(prefix string) string {
	return fmt.Sprintf(`
resource "openshift_project" "test" {
  metadata {
    generate_name = "%s"
  }
}
`, prefix)
}

func testAccOpenshiftProjectConfig_specialCharacters(nsName string) string {
	return fmt.Sprintf(`
resource "openshift_project" "test" {
  metadata {
    annotations = {
      "myhost.co.uk/any-path" = "one"
      "Different"             = "1234"
    }

    labels = {
      "myhost.co.uk/any-path" = "one"
      "TestLabelThree"        = "three"
    }

    name = "%s"
  }
}
`, nsName)
}

func testAccOpenshiftProjectConfig_deleteTimeout(nsName string) string {
	return fmt.Sprintf(`
resource "openshift_project" "test" {
  metadata {
    name = "%s"
  }
  timeouts{
    delete = "30m"
  }
}
`, nsName)
}
