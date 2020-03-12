package openshift

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"fmt"
	"log"

	api "github.com/openshift/api/build/v1"
	client_v1 "github.com/openshift/client-go/build/clientset/versioned/typed/build/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	pkgApi "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

func resourceOpenshiftBuildConfig() *schema.Resource {
	return &schema.Resource{
		Create: resourceOpenshiftBuildConfigCreate,
		Read:   resourceOpenshiftBuildConfigRead,
		Update: resourceOpenshiftBuildConfigUpdate,
		Delete: resourceOpenshiftBuildConfigDelete,
		Exists: resourceOpenshiftBuildConfigExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"metadata": namespacedMetadataSchema("build config", true),
			"spec": {
				Type:        schema.TypeList,
				Description: "(v1.BuildConfigSpec) spec holds all the input necessary to produce a new build, and the conditions when to trigger them.",
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"completion_deadline_seconds": {
							Type:        schema.TypeInt,
							Description: "(integer) completionDeadlineSeconds is an optional duration in seconds, counted from the time when a build pod gets scheduled in the system, that the build may be active on a node before the system actively tries to terminate the build; value must be positive integer",
							Optional:    true,
							Default:     600,
						},
						"failed_builds_history_limit": {
							Type:        schema.TypeInt,
							Description: "(integer) failedBuildsHistoryLimit is the number of old failed builds to retain. If not specified, all failed builds are retained.",
							Optional:    true,
							Default:     3,
						},
						"node_selector": {
							Type:        schema.TypeMap,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Optional:    true,
							Computed:    true,
							Description: "(object) nodeSelector is a selector which must be true for the build pod to fit on a node If nil, it can be overridden by default build nodeselector values for the cluster. If set to an empty map or a map with any values, default build nodeselector values are ignored.",
						},
						"output": {
							Type:        schema.TypeList,
							Description: "(v1.BuildOutput) output describes the Docker image the Strategy should produce.",
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"image_label": {
										Type:        schema.TypeList,
										Description: "(array) imageLabels define a list of labels that are applied to the resulting image. If there are multiple labels with the same name then the last one in the list is used.",
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:        schema.TypeString,
													Description: "(string) name defines the name of the label. It must have non-zero length.",
													Optional:    true,
												},
												"value": {
													Type:        schema.TypeString,
													Description: "(string) value defines the literal value of the label.",
													Optional:    true,
												},
											},
										},
									},
									"push_secret": {
										Type:        schema.TypeList,
										Description: "(v1.LocalObjectReference) PushSecret is the name of a Secret that would be used for setting up the authentication for executing the Docker push to authentication enabled Docker Registry (or Docker Hub).",
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:        schema.TypeString,
													Description: "(string) Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names",
													Optional:    true,
												},
											},
										},
									},
									"to": {
										Type:        schema.TypeList,
										Description: "(v1.ObjectReference) to defines an optional location to push the output of this build to. Kind must be one of 'ImageStreamTag' or 'DockerImage'. This value will be used to look up a Docker image repository to push to. In the case of an ImageStreamTag, the ImageStreamTag will be looked for in the namespace of the build unless Namespace is specified.",
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
								},
							},
						},
						"post_commit": {
							Type:        schema.TypeList,
							Description: "(v1.BuildPostCommitSpec) postCommit is a build hook executed after the build output image is committed, before it is pushed to a registry.",
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"args": {
										Type:        schema.TypeSet,
										Description: "(array) args is a list of arguments that are provided to either Command, Script or the Docker image's default entrypoint. The arguments are placed immediately after the command to be run.",
										Optional:    true,
										Elem:        &schema.Schema{Type: schema.TypeString},
										Set:         schema.HashString,
									},
									"command": {
										Type:        schema.TypeSet,
										Description: "(array) command is the command to run. It may not be specified with Script. This might be needed if the image doesn't have `/bin/sh`, or if you do not want to use a shell. In all other cases, using Script might be more convenient.",
										Optional:    true,
										Elem:        &schema.Schema{Type: schema.TypeString},
										Set:         schema.HashString,
									},
									"script": {
										Type:        schema.TypeString,
										Description: "(string) script is a shell script to be run with `/bin/sh -ic`. It may not be specified with Command. Use Script when a shell script is appropriate to execute the post build hook, for example for running unit tests with `rake test`. If you need control over the image entrypoint, or if the image does not have `/bin/sh`, use Command and/or Args. The `-i` flag is needed to support CentOS and RHEL images that use Software Collections (SCL), in order to have the appropriate collections enabled in the shell. E.g., in the Ruby image, this is necessary to make `ruby`, `bundle` and other binaries available in the PATH.",
										Optional:    true,
									},
								},
							},
						},
						"resources": {
							Type:        schema.TypeList,
							Description: "(v1.ResourceRequirements) resources computes resource requirements to execute the build.",
							Optional:    true,
							MaxItems:    1,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: resourcesField(),
							},
						},
						"run_policy": {
							Type:        schema.TypeString,
							Description: "(string) RunPolicy describes how the new build created from this build configuration will be scheduled for execution. This is optional, if not specified we default to \"Serial\".",
							Optional:    true,
							Computed:    true,
						},
						"service_account": {
							Type:        schema.TypeString,
							Description: "(string) serviceAccount is the name of the ServiceAccount to use to run the pod created by this build. The pod will be allowed to use secrets referenced by the ServiceAccount",
							Optional:    true,
						},
						"source": {
							Type:        schema.TypeList,
							Description: "(v1.BuildSource) source describes the SCM in use.",
							Optional:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"binary": {
										Type:        schema.TypeList,
										Description: "(v1.BinaryBuildSource) binary builds accept a binary as their input. The binary is generally assumed to be a tar, gzipped tar, or zip file depending on the strategy. For Docker builds, this is the build context and an optional Dockerfile may be specified to override any Dockerfile in the build context. For Source builds, this is assumed to be an archive as described above. For Source and Docker builds, if binary.asFile is set the build will receive a directory with a single file. contextDir may be used when an archive is provided. Custom builds will receive this binary as input on STDIN.",
										Optional:    true,
										MaxItems:    1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"as_file": {
													Type:        schema.TypeString,
													Description: "(string) asFile indicates that the provided binary input should be considered a single file within the build input. For example, specifying \"webapp.war\" would place the provided binary as `/webapp.war` for the builder. If left empty, the Docker and Source build strategies assume this file is a zip, tar, or tar.gz file and extract it as the source. The custom strategy receives this binary as standard input. This filename may not contain slashes or be '..' or '.'.",
													Optional:    true,
												},
											},
										},
									},
									"context_dir": {
										Type:        schema.TypeString,
										Description: "(string) contextDir specifies the sub-directory where the source code for the application exists. This allows to have buildable sources in directory other than root of repository.",
										Optional:    true,
									},
									"dockerfile": {
										Type:        schema.TypeString,
										Description: "(string) dockerfile is the raw contents of a Dockerfile which should be built. When this option is specified, the FROM may be modified based on your strategy base image and additional ENV stanzas from your strategy environment will be added after the FROM, but before the rest of your Dockerfile stanzas. The Dockerfile source type may be used with other options like git - in those cases the Git repo will have any innate Dockerfile replaced in the context dir.",
										Optional:    true,
									},
									"git": {
										Type:        schema.TypeList,
										Description: "(v1.GitBuildSource) git contains optional information about git build source",
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												//"http_proxy": {
												//	Type:        schema.TypeString,
												//	Description: "(string) httpProxy is a proxy used to reach the git repository over http",
												//	Optional:    true,
												//},
												//"https_proxy": {
												//	Type:        schema.TypeString,
												//	Description: "(string) httpsProxy is a proxy used to reach the git repository over https",
												//	Optional:    true,
												//},
												//"no_proxy": {
												//	Type:        schema.TypeString,
												//	Description: "(string) noProxy is the list of domains for which the proxy should not be used",
												//	Optional:    true,
												//},
												"ref": {
													Type:        schema.TypeString,
													Description: "(string) ref is the branch/tag/ref to build.",
													Optional:    true,
												},
												"uri": {
													Type:        schema.TypeString,
													Description: "(string) uri points to the source that will be built. The structure of the source will depend on the type of build to run",
													Optional:    true,
												},
											},
										},
									},
									"image": {
										Type:        schema.TypeList,
										Description: "(array) images describes a set of images to be used to provide source for the build",
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"from": {
													Type:        schema.TypeList,
													Description: "(v1.ObjectReference) from is a reference to an ImageStreamTag, ImageStreamImage, or DockerImage to copy source from.",
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
												"paths": {
													Type:        schema.TypeList,
													Description: "(array) paths is a list of source and destination paths to copy from the image.",
													Optional:    true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"destination_dir": {
																Type:        schema.TypeString,
																Description: "(string) destinationDir is the relative directory within the build directory where files copied from the image are placed.",
																Optional:    true,
															},
															"source_path": {
																Type:        schema.TypeString,
																Description: "(string) sourcePath is the absolute path of the file or directory inside the image to copy to the build directory.  If the source path ends in /. then the content of the directory will be copied, but the directory itself will not be created at the destination.",
																Optional:    true,
															},
														},
													},
												},
												"pull_secret": {
													Type:        schema.TypeList,
													Description: "(v1.LocalObjectReference) pullSecret is a reference to a secret to be used to pull the image from a registry If the image is pulled from the OpenShift registry, this field does not need to be set.",
													Optional:    true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"name": {
																Type:        schema.TypeString,
																Description: "(string) Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names",
																Optional:    true,
															},
														},
													},
												},
											},
										},
									},
									"secret": {
										Type:        schema.TypeList,
										Description: "(array) secrets represents a list of secrets and their destinations that will be used only for the build.",
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"destination_dir": {
													Type:        schema.TypeString,
													Description: "(string) destinationDir is the directory where the files from the secret should be available for the build time. For the Source build strategy, these will be injected into a container where the assemble script runs. Later, when the script finishes, all files injected will be truncated to zero length. For the Docker build strategy, these will be copied into the build directory, where the Dockerfile is located, so users can ADD or COPY them during docker build.",
													Optional:    true,
												},
												"secret": {
													Type:        schema.TypeList,
													Description: "(v1.LocalObjectReference) secret is a reference to an existing secret that you want to use in your build.",
													Optional:    true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"name": {
																Type:        schema.TypeString,
																Description: "(string) Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names",
																Optional:    true,
															},
														},
													},
												},
											},
										},
									},
									"source_secret": {
										Type:        schema.TypeList,
										Description: "(v1.LocalObjectReference) sourceSecret is the name of a Secret that would be used for setting up the authentication for cloning private repository. The secret contains valid credentials for remote repository, where the data's key represent the authentication method to be used and value is the base64 encoded credentials. Supported auth methods are: ssh-privatekey.",
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:        schema.TypeString,
													Description: "(string) Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names",
													Optional:    true,
												},
											},
										},
									},
									"type": {
										Type:         schema.TypeString,
										Description:  "(string) type of build input to accept",
										Optional:     true,
										Default:      "None",
										ValidateFunc: validation.StringInSlice([]string{"Git", "Dockerfile", "Binary", "Image", "None"}, false),
									},
								},
							},
						},
						"strategy": {
							Type:        schema.TypeList,
							Description: "(v1.BuildStrategy) strategy defines how to perform a build.",
							Optional:    true,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									//"custom_strategy": {
									//	Type:        schema.TypeList,
									//	Description: "(v1.CustomBuildStrategy) customStrategy holds the parameters to the Custom build strategy",
									//	Optional:    true,
									//	Elem: &schema.Resource{
									//		Schema: map[string]*schema.Schema{},
									//	},
									//},
									"docker_strategy": {
										Type:        schema.TypeList,
										Description: "(v1.DockerBuildStrategy) dockerStrategy holds the parameters to the Docker build strategy.",
										Optional:    true,
										Computed:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"build_args": {
													Type:        schema.TypeList,
													Description: "(array) buildArgs contains build arguments that will be resolved in the Dockerfile.  See https://docs.docker.com/engine/reference/builder/#/arg for more details.",
													Optional:    true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"name": {
																Type:        schema.TypeString,
																Description: "(string) Name of the environment variable. Must be a C_IDENTIFIER.",
																Optional:    true,
															},
															"value": {
																Type:        schema.TypeString,
																Description: "(string) Variable references $(VAR_NAME) are expanded using the previous defined environment variables in the container and any service environment variables. If a variable cannot be resolved, the reference in the input string will be unchanged. The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not. Defaults to \"\".",
																Optional:    true,
															},
															"value_from": {
																Type:        schema.TypeList,
																Optional:    true,
																MaxItems:    1,
																Description: "Source for the environment variable's value",
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"config_map_key_ref": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			MaxItems:    1,
																			Description: "Selects a key of a ConfigMap.",
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"key": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "The key to select.",
																					},
																					"name": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "Name of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#names",
																					},
																				},
																			},
																		},
																		"field_ref": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			MaxItems:    1,
																			Description: "Selects a field of the pod: supports metadata.name, metadata.namespace, metadata.labels, metadata.annotations, spec.nodeName, spec.serviceAccountName, status.podIP..",
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"api_version": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Default:     "v1",
																						Description: `Version of the schema the FieldPath is written in terms of, defaults to "v1".`,
																					},
																					"field_path": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "Path of the field to select in the specified API version",
																					},
																				},
																			},
																		},
																		"resource_field_ref": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			MaxItems:    1,
																			Description: "Selects a field of the pod: supports metadata.name, metadata.namespace, metadata.labels, metadata.annotations, spec.nodeName, spec.serviceAccountName, status.podIP..",
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"container_name": {
																						Type:     schema.TypeString,
																						Optional: true,
																					},
																					"resource": {
																						Type:        schema.TypeString,
																						Required:    true,
																						Description: "Resource to select",
																					},
																				},
																			},
																		},
																		"secret_key_ref": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			MaxItems:    1,
																			Description: "Selects a field of the pod: supports metadata.name, metadata.namespace, metadata.labels, metadata.annotations, spec.nodeName, spec.serviceAccountName, status.podIP..",
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"key": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "The key of the secret to select from. Must be a valid secret key.",
																					},
																					"name": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "Name of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#names",
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
												"dockerfile_path": {
													Type:        schema.TypeString,
													Description: "(string) dockerfilePath is the path of the Dockerfile that will be used to build the Docker image, relative to the root of the context (contextDir).",
													Optional:    true,
												},
												"env": {
													Type:        schema.TypeList,
													Description: "(array) env contains additional environment variables you want to pass into a builder container.",
													Optional:    true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"name": {
																Type:        schema.TypeString,
																Description: "(string) Name of the environment variable. Must be a C_IDENTIFIER.",
																Optional:    true,
															},
															"value": {
																Type:        schema.TypeString,
																Description: "(string) Variable references $(VAR_NAME) are expanded using the previous defined environment variables in the container and any service environment variables. If a variable cannot be resolved, the reference in the input string will be unchanged. The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not. Defaults to \"\".",
																Optional:    true,
															},
															"value_from": {
																Type:        schema.TypeList,
																Optional:    true,
																MaxItems:    1,
																Description: "Source for the environment variable's value",
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"config_map_key_ref": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			MaxItems:    1,
																			Description: "Selects a key of a ConfigMap.",
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"key": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "The key to select.",
																					},
																					"name": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "Name of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#names",
																					},
																				},
																			},
																		},
																		"field_ref": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			MaxItems:    1,
																			Description: "Selects a field of the pod: supports metadata.name, metadata.namespace, metadata.labels, metadata.annotations, spec.nodeName, spec.serviceAccountName, status.podIP..",
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"api_version": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Default:     "v1",
																						Description: `Version of the schema the FieldPath is written in terms of, defaults to "v1".`,
																					},
																					"field_path": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "Path of the field to select in the specified API version",
																					},
																				},
																			},
																		},
																		"resource_field_ref": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			MaxItems:    1,
																			Description: "Selects a field of the pod: supports metadata.name, metadata.namespace, metadata.labels, metadata.annotations, spec.nodeName, spec.serviceAccountName, status.podIP..",
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"container_name": {
																						Type:     schema.TypeString,
																						Optional: true,
																					},
																					"resource": {
																						Type:        schema.TypeString,
																						Required:    true,
																						Description: "Resource to select",
																					},
																				},
																			},
																		},
																		"secret_key_ref": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			MaxItems:    1,
																			Description: "Selects a field of the pod: supports metadata.name, metadata.namespace, metadata.labels, metadata.annotations, spec.nodeName, spec.serviceAccountName, status.podIP..",
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"key": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "The key of the secret to select from. Must be a valid secret key.",
																					},
																					"name": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "Name of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#names",
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
												"force_pull": {
													Type:        schema.TypeBool,
													Description: "(boolean) forcePull describes if the builder should pull the images from registry prior to building.",
													Optional:    true,
												},
												"from": {
													Type:        schema.TypeList,
													Description: "(v1.ObjectReference) from is a reference to an ImageStreamTag, ImageStreamImage, or DockerImage to copy source from.",
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
												"image_optimization_policy": {
													Type:        schema.TypeString,
													Description: "(string) imageOptimizationPolicy describes what optimizations the system can use when building images to reduce the final size or time spent building the image. The default policy is 'None' which means the final build image will be equivalent to an image created by the Docker build API. The experimental policy 'SkipLayers' will avoid committing new layers in between each image step, and will fail if the Dockerfile cannot provide compatibility with the 'None' policy. An additional experimental policy 'SkipLayersAndWarn' is the same as 'SkipLayers' but simply warns if compatibility cannot be preserved.",
													Optional:    true,
												},
												"no_cache": {
													Type:        schema.TypeBool,
													Description: "(boolean) noCache if set to true indicates that the docker build must be executed with the --no-cache=true flag",
													Optional:    true,
												},
												"pull_secret": {
													Type:        schema.TypeList,
													Description: "(v1.LocalObjectReference) pullSecret is a reference to a secret to be used to pull the image from a registry If the image is pulled from the OpenShift registry, this field does not need to be set.",
													Optional:    true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"name": {
																Type:        schema.TypeString,
																Description: "(string) Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names",
																Optional:    true,
															},
														},
													},
												},
											},
										},
									},
									"jenkins_pipeline_strategy": {
										Type:        schema.TypeList,
										Description: "(v1.JenkinsPipelineBuildStrategy) JenkinsPipelineStrategy holds the parameters to the Jenkins Pipeline build strategy. This strategy is in tech preview.",
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"env": {
													Type:        schema.TypeList,
													Description: "(array) env contains additional environment variables you want to pass into a builder container.",
													Optional:    true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"name": {
																Type:        schema.TypeString,
																Description: "(string) Name of the environment variable. Must be a C_IDENTIFIER.",
																Optional:    true,
															},
															"value": {
																Type:        schema.TypeString,
																Description: "(string) Variable references $(VAR_NAME) are expanded using the previous defined environment variables in the container and any service environment variables. If a variable cannot be resolved, the reference in the input string will be unchanged. The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not. Defaults to \"\".",
																Optional:    true,
															},
															"value_from": {
																Type:        schema.TypeList,
																Optional:    true,
																MaxItems:    1,
																Description: "Source for the environment variable's value",
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"config_map_key_ref": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			MaxItems:    1,
																			Description: "Selects a key of a ConfigMap.",
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"key": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "The key to select.",
																					},
																					"name": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "Name of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#names",
																					},
																				},
																			},
																		},
																		"field_ref": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			MaxItems:    1,
																			Description: "Selects a field of the pod: supports metadata.name, metadata.namespace, metadata.labels, metadata.annotations, spec.nodeName, spec.serviceAccountName, status.podIP..",
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"api_version": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Default:     "v1",
																						Description: `Version of the schema the FieldPath is written in terms of, defaults to "v1".`,
																					},
																					"field_path": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "Path of the field to select in the specified API version",
																					},
																				},
																			},
																		},
																		"resource_field_ref": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			MaxItems:    1,
																			Description: "Selects a field of the pod: supports metadata.name, metadata.namespace, metadata.labels, metadata.annotations, spec.nodeName, spec.serviceAccountName, status.podIP..",
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"container_name": {
																						Type:     schema.TypeString,
																						Optional: true,
																					},
																					"resource": {
																						Type:        schema.TypeString,
																						Required:    true,
																						Description: "Resource to select",
																					},
																				},
																			},
																		},
																		"secret_key_ref": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			MaxItems:    1,
																			Description: "Selects a field of the pod: supports metadata.name, metadata.namespace, metadata.labels, metadata.annotations, spec.nodeName, spec.serviceAccountName, status.podIP..",
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"key": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "The key of the secret to select from. Must be a valid secret key.",
																					},
																					"name": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "Name of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#names",
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
												"jenkinsfile": {
													Type:        schema.TypeString,
													Description: "(string) Jenkinsfile defines the optional raw contents of a Jenkinsfile which defines a Jenkins pipeline build.",
													Optional:    true,
												},
												"jenkinsfile_path": {
													Type:        schema.TypeString,
													Description: "(string) JenkinsfilePath is the optional path of the Jenkinsfile that will be used to configure the pipeline relative to the root of the context (contextDir). If both JenkinsfilePath & Jenkinsfile are both not specified, this defaults to Jenkinsfile in the root of the specified contextDir.",
													Optional:    true,
												},
											},
										},
									},
									"source_strategy": {
										Type:        schema.TypeList,
										Description: "(v1.SourceBuildStrategy) sourceStrategy holds the parameters to the Source build strategy.",
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"env": {
													Type:        schema.TypeList,
													Description: "(array) env contains additional environment variables you want to pass into a builder container.",
													Optional:    true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"name": {
																Type:        schema.TypeString,
																Description: "(string) Name of the environment variable. Must be a C_IDENTIFIER.",
																Optional:    true,
															},
															"value": {
																Type:        schema.TypeString,
																Description: "(string) Variable references $(VAR_NAME) are expanded using the previous defined environment variables in the container and any service environment variables. If a variable cannot be resolved, the reference in the input string will be unchanged. The $(VAR_NAME) syntax can be escaped with a double $$, ie: $$(VAR_NAME). Escaped references will never be expanded, regardless of whether the variable exists or not. Defaults to \"\".",
																Optional:    true,
															},
															"value_from": {
																Type:        schema.TypeList,
																Optional:    true,
																MaxItems:    1,
																Description: "Source for the environment variable's value",
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"config_map_key_ref": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			MaxItems:    1,
																			Description: "Selects a key of a ConfigMap.",
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"key": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "The key to select.",
																					},
																					"name": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "Name of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#names",
																					},
																				},
																			},
																		},
																		"field_ref": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			MaxItems:    1,
																			Description: "Selects a field of the pod: supports metadata.name, metadata.namespace, metadata.labels, metadata.annotations, spec.nodeName, spec.serviceAccountName, status.podIP..",
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"api_version": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Default:     "v1",
																						Description: `Version of the schema the FieldPath is written in terms of, defaults to "v1".`,
																					},
																					"field_path": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "Path of the field to select in the specified API version",
																					},
																				},
																			},
																		},
																		"resource_field_ref": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			MaxItems:    1,
																			Description: "Selects a field of the pod: supports metadata.name, metadata.namespace, metadata.labels, metadata.annotations, spec.nodeName, spec.serviceAccountName, status.podIP..",
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"container_name": {
																						Type:     schema.TypeString,
																						Optional: true,
																					},
																					"resource": {
																						Type:        schema.TypeString,
																						Required:    true,
																						Description: "Resource to select",
																					},
																				},
																			},
																		},
																		"secret_key_ref": {
																			Type:        schema.TypeList,
																			Optional:    true,
																			MaxItems:    1,
																			Description: "Selects a field of the pod: supports metadata.name, metadata.namespace, metadata.labels, metadata.annotations, spec.nodeName, spec.serviceAccountName, status.podIP..",
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"key": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "The key of the secret to select from. Must be a valid secret key.",
																					},
																					"name": {
																						Type:        schema.TypeString,
																						Optional:    true,
																						Description: "Name of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#names",
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
												"force_pull": {
													Type:        schema.TypeBool,
													Description: "(boolean) forcePull describes if the builder should pull the images from registry prior to building.",
													Optional:    true,
												},
												"from": {
													Type:        schema.TypeList,
													Description: "(v1.ObjectReference) from is a reference to an ImageStreamTag, ImageStreamImage, or DockerImage to copy source from.",
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
												"incremental": {
													Type:        schema.TypeBool,
													Description: "(boolean) incremental flag forces the Source build to do incremental builds if true.",
													Optional:    true,
												},
												"pull_secret": {
													Type:        schema.TypeList,
													Description: "(v1.LocalObjectReference) pullSecret is a reference to a secret to be used to pull the image from a registry If the image is pulled from the OpenShift registry, this field does not need to be set.",
													Optional:    true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"name": {
																Type:        schema.TypeString,
																Description: "(string) Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names",
																Optional:    true,
															},
														},
													},
												},
												"runtime_image": {
													Type:        schema.TypeList,
													Description: "(v1.ObjectReference) runtimeImage is an optional image that is used to run an application without unneeded dependencies installed. The building of the application is still done in the builder image but, post build, you can copy the needed artifacts in the runtime image for use. Deprecated: This feature will be removed in a future release. Use ImageSource to copy binary artifacts created from one build into a separate runtime image.",
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
												"runtime_artifacts": {
													Type:        schema.TypeList,
													Description: "(array) runtimeArtifacts specifies a list of source/destination pairs that will be copied from the builder to the runtime image. sourcePath can be a file or directory. destinationDir must be a directory. destinationDir can also be empty or equal to \".\", in this case it just refers to the root of WORKDIR. Deprecated: This feature will be removed in a future release. Use ImageSource to copy binary artifacts created from one build into a separate runtime image.",
													Optional:    true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"destination_dir": {
																Type:        schema.TypeString,
																Description: "(string) destinationDir is the relative directory within the build directory where files copied from the image are placed.",
																Optional:    true,
															},
															"source_path": {
																Type:        schema.TypeString,
																Description: "(string) sourcePath is the absolute path of the file or directory inside the image to copy to the build directory.  If the source path ends in /. then the content of the directory will be copied, but the directory itself will not be created at the destination.",
																Optional:    true,
															},
														},
													},
												},
												"scripts": {
													Type:        schema.TypeString,
													Description: "(string) scripts is the location of Source scripts",
													Optional:    true,
												},
											},
										},
									},
									"type": {
										Type:         schema.TypeString,
										Description:  "(string) type is the kind of build strategy.",
										Optional:     true,
										ValidateFunc: validation.StringInSlice([]string{"Docker", "Source", "Custom", "JenkinsPipeline"}, false),
									},
								},
							},
						},
						"successful_builds_history_limit": {
							Type:        schema.TypeInt,
							Description: "(integer) successfulBuildsHistoryLimit is the number of old successful builds to retain. If not specified, all successful builds are retained.",
							Optional:    true,
							Default:     5,
						},
						"trigger": {
							Type:        schema.TypeList,
							Description: "(array) triggers determine how new Builds can be launched from a BuildConfig. If no triggers are defined, a new build can only occur as a result of an explicit client build creation.",
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:         schema.TypeString,
										Description:  "(string) type is the type of build trigger",
										Optional:     true,
										ValidateFunc: validation.StringInSlice([]string{"GitHub", "Generic", "GitLab", "Bitbucket", "ConfigChange"}, false),
									},
									"bitbucket": {
										Type:        schema.TypeList,
										Description: "(v1.WebHookTrigger) BitbucketWebHook contains the parameters for a Bitbucket webhook type of trigger",
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"allow_env": {
													Type:        schema.TypeBool,
													Description: "(boolean) allowEnv determines whether the webhook can set environment variables; can only be set to true for GenericWebHook.",
													Optional:    true,
												},
												"secret": {
													Type:        schema.TypeString,
													Description: "(string) secret used to validate requests.",
													Optional:    true,
												},
											},
										},
									},
									"generic": {
										Type:        schema.TypeList,
										Description: "(v1.WebHookTrigger) generic contains the parameters for a Generic webhook type of trigger",
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"allow_env": {
													Type:        schema.TypeBool,
													Description: "(boolean) allowEnv determines whether the webhook can set environment variables; can only be set to true for GenericWebHook.",
													Optional:    true,
												},
												"secret": {
													Type:        schema.TypeString,
													Description: "(string) secret used to validate requests.",
													Optional:    true,
												},
											},
										},
									},
									"github": {
										Type:        schema.TypeList,
										Description: "(v1.WebHookTrigger) github contains the parameters for a GitHub webhook type of trigger",
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"allow_env": {
													Type:        schema.TypeBool,
													Description: "(boolean) allowEnv determines whether the webhook can set environment variables; can only be set to true for GenericWebHook.",
													Optional:    true,
												},
												"secret": {
													Type:        schema.TypeString,
													Description: "(string) secret used to validate requests.",
													Optional:    true,
												},
											},
										},
									},
									"gitlab": {
										Type:        schema.TypeList,
										Description: "(v1.WebHookTrigger) GitLabWebHook contains the parameters for a GitLab webhook type of trigger",
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"allow_env": {
													Type:        schema.TypeBool,
													Description: "(boolean) allowEnv determines whether the webhook can set environment variables; can only be set to true for GenericWebHook.",
													Optional:    true,
												},
												"secret": {
													Type:        schema.TypeString,
													Description: "(string) secret used to validate requests.",
													Optional:    true,
												},
											},
										},
									},
									"image_change": {
										Type:        schema.TypeList,
										Description: "(v1.WebHookTrigger) GitLabWebHook contains the parameters for a GitLab webhook type of trigger",
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"from": {
													Type:        schema.TypeList,
													Description: "(v1.ObjectReference) from is a reference to an ImageStreamTag, ImageStreamImage, or DockerImage to copy source from.",
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
												"last_triggered_image_id": {
													Type:        schema.TypeString,
													Description: "(string) lastTriggeredImageID is used internally by the ImageChangeController to save last used image ID for build",
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

func resourceOpenshiftBuildConfigCreate(d *schema.ResourceData, meta interface{}) error {
	client, err := client_v1.NewForConfig(meta.(*rest.Config))
	if err != nil {
		return err
	}

	metadata := expandMetadata(d.Get("metadata").([]interface{}))
	spec, err := expandBuildConfigSpec(d.Get("spec").([]interface{}))
	if err != nil {
		return err
	}

	buildConfig := api.BuildConfig{
		ObjectMeta: metadata,
		Spec:       spec,
	}

	log.Printf("[INFO] Creating new build config: %#v", buildConfig)
	out, err := client.BuildConfigs(metadata.Namespace).Create(&buildConfig)
	if err != nil {
		return fmt.Errorf("Failed to create build config: %s", err)
	}
	log.Printf("[INFO] Submitted new build config: %#v", out)
	d.SetId(buildId(out.ObjectMeta))

	return resourceOpenshiftBuildConfigRead(d, meta)
}

func resourceOpenshiftBuildConfigRead(d *schema.ResourceData, meta interface{}) error {
	client, err := client_v1.NewForConfig(meta.(*rest.Config))
	if err != nil {
		return err
	}

	namespace, name, err := idParts(d.Id())
	if err != nil {
		return err
	}

	log.Printf("[INFO] Reading build config %s", name)
	buildConfig, err := client.BuildConfigs(namespace).Get(name, meta_v1.GetOptions{})
	if err != nil {
		log.Printf("[DEBUG] Received error: %#v", err)
		return err
	}

	log.Printf("[INFO] Received build config: %#v", buildConfig)
	err = d.Set("metadata", flattenMetadata(buildConfig.ObjectMeta, d))
	if err != nil {
		return err
	}

	spec, err := flattenBuildConfigSpec(buildConfig.Spec, d)
	if err != nil {
		return err
	}

	err = d.Set("spec", spec)
	if err != nil {
		return err
	}

	return nil
}

func resourceOpenshiftBuildConfigUpdate(d *schema.ResourceData, meta interface{}) error {
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
		spec, err := expandBuildConfigSpec(d.Get("spec").([]interface{}))
		if err != nil {
			return err
		}

		ops = append(ops, &ReplaceOperation{
			Path:  "/spec",
			Value: spec,
		})
	}

	data, err := ops.MarshalJSON()
	if err != nil {
		return fmt.Errorf("Failed to marshal update operations: %s", err)
	}

	log.Printf("[INFO] Updating build config %q: %v", name, string(data))
	out, err := client.BuildConfigs(namespace).Patch(name, pkgApi.JSONPatchType, data)
	if err != nil {
		return fmt.Errorf("Failed to update build config: %s", err)
	}
	log.Printf("[INFO] Submitted updated build config: %#v", out)

	d.SetId(buildId(out.ObjectMeta))

	return resourceOpenshiftBuildConfigRead(d, meta)
}

func resourceOpenshiftBuildConfigDelete(d *schema.ResourceData, meta interface{}) error {
	client, err := client_v1.NewForConfig(meta.(*rest.Config))
	if err != nil {
		return err
	}

	namespace, name, err := idParts(d.Id())
	if err != nil {
		return err
	}

	log.Printf("[INFO] Deleting build config: %#v", name)

	err = client.BuildConfigs(namespace).Delete(name, &meta_v1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("Failed to delete build config: %s", err)
	}

	d.SetId("")
	return nil
}

func resourceOpenshiftBuildConfigExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client, err := client_v1.NewForConfig(meta.(*rest.Config))
	if err != nil {
		return true, err
	}

	namespace, name, err := idParts(d.Id())
	if err != nil {
		return false, err
	}

	log.Printf("[INFO] Checking build config %s", name)
	_, err = client.BuildConfigs(namespace).Get(name, meta_v1.GetOptions{})
	if err != nil {
		if statusErr, ok := err.(*errors.StatusError); ok && statusErr.ErrStatus.Code == 404 {
			return false, nil
		}
		log.Printf("[DEBUG] Received error: %#v", err)
	}
	return true, err
}
