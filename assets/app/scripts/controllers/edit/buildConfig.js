'use strict';

/**
 * @ngdoc function
 * @name openshiftConsole.controller:EditBuildConfigController
 * @description
 * Controller of the openshiftConsole
 */
angular.module('openshiftConsole')
  .controller('EditBuildConfigController', function ($scope, $routeParams, DataService, ProjectsService, BuildConfigsService, $filter, ApplicationGenerator, Navigate, $location, AlertMessageService, SOURCE_URL_PATTERN) {

    $scope.projectName = $routeParams.project;
    $scope.buildConfig = null;
    $scope.alerts = {};
    $scope.emptyMessage = "Loading...";
    $scope.sourceURLPattern = SOURCE_URL_PATTERN;
    $scope.options = {};
    $scope.openshiftBuilderImages = ["ruby","python","perl","php","nodejs","wildfly"]
    $scope.buildFromTypes = [
      {
        "id": "ImageStreamTag",
        "title": "Image Stream Tag"
      },
      {
        "id": "ImageStreamImage",
        "title": "Image Stream Image"  
      },
      {
        "id": "DockerImage",
        "title": "Docker Image Link"
      }
    ];
    $scope.pushToTypes = [
      {
        "id": "ImageStreamTag",
        "title": "Image Stream Tag"
      },
      {
        "id": "DockerImage",
        "title": "Docker Image Link"
      },
      {
        "id": "None",
        "title": "--- None ---"
      }
    ];
    $scope.breadcrumbs = [
      {
        title: "Builds",
        link: "project/" + $routeParams.project + "/browse/builds"
      },
      {
        title: $routeParams.buildconfig,
        link: "project/" + $routeParams.project + "/browse/builds/" + $routeParams.buildconfig
      },
      {
        title: "Edit"
      }
    ];
    $scope.buildFrom = {
      projects: [],
      imageStreams: [],
      tags: {},
    };
    $scope.pushTo = {
      projects: [],
      imageStreams: [],
      tags: {},
    };
    $scope.sources = {
      "binary": false,
      "dockerfile": false,
      "git": false,
      "image": false,
      "contextDir": false
    };
    $scope.triggers = {
      "webhook": false,
      "imageChange": false,
      "configChange": false
    };
    $scope.availableProjects = [];

    AlertMessageService.getAlerts().forEach(function(alert) {
      $scope.alerts[alert.name] = alert.data;
    });
    AlertMessageService.clearAlerts();
    var watches = [];

    ProjectsService
      .get($routeParams.project)
      .then(_.spread(function(project, context) {
        $scope.project = project;
        DataService.get("buildconfigs", $routeParams.buildconfig, context).then(
          // success
          function(buildConfig) {
            $scope.buildConfig = buildConfig;
            $scope.updatedBuildConfig = angular.copy($scope.buildConfig);
            $scope.buildStrategy = $filter('buildStrategy')($scope.updatedBuildConfig);
            $scope.strategyType = $scope.buildConfig.spec.strategy.type;
            $scope.envVars = $filter('envVarsPair')($scope.buildStrategy.env);
            $scope.triggers = BuildConfigsService.getTriggerMap($scope.triggers, $scope.buildConfig.spec.triggers)
            $scope.pickedTriggers = $scope.triggers;
            $scope.sources = $scope.getSourceMap($scope.sources, $scope.buildConfig.spec.source);

            if ($scope.buildStrategy.from) {
              var buildFrom = $scope.buildStrategy.from;
              BuildConfigsService.setBuildFromVariables(
                $scope.options,
                buildFrom.kind,
                buildFrom.namespace || buildConfig.metadata.namespace, 
                buildFrom.name.split(":")[0],
                buildFrom.name.split(":")[1],
                (buildFrom.kind === "ImageStreamImage") ? buildFrom.name : "",
                (buildFrom.kind === "ImageStreamTag") ? buildConfig.metadata.namespace + "/" + buildFrom.name : buildFrom.name);
            } else {
              BuildConfigsService.setBuildFromVariables($scope.options, "None", buildConfig.metadata.namespace, "", "", "");
            }

            if ($scope.updatedBuildConfig.spec.output.to) {
              var pushTo = $scope.updatedBuildConfig.spec.output.to
              BuildConfigsService.setPushToVariables(
                $scope.options,
                pushTo.kind,
                pushTo.namespace || buildConfig.metadata.namespace,
                pushTo.name.split(":")[0],
                pushTo.name.split(":")[1],
                (pushTo.kind === "ImageStreamTag") ? buildConfig.metadata.namespace + "/" + pushTo.name : pushTo.name);
            } else {
              BuildConfigsService.setPushToVariables($scope.options, "None", buildConfig.metadata.namespace, "", "", "");
            }

            $scope.builderImageStream = {
              namespace: $scope.options.pickedBuildFromNamespace,
              imageStream: $scope.options.pickedBuildFromImageStream,
              tag: $scope.options.pickedBuildFromImageStreamTag,
            };

            $scope.outputImageStream = {
              namespace: $scope.options.pickedPushToNamespace,
              imageStream: $scope.options.pickedPushToImageStream,
              tag: $scope.options.pickedPushToImageStreamTag,
            };

            $scope.options.forcePull = !!$scope.buildStrategy.forcePull;

            if ($scope.sources.image) {
              // Initialize structure in the same way builder and output image.
              $scope.imageSourceBuildFrom = {
                projects: [],
                imageStreams: [],
                tags: {},
              };

              $scope.imageSourcePaths = $filter('destinationSourcePair')($scope.buildConfig.spec.source.image.paths);
              $scope.imageSourceTypes = angular.copy($scope.buildFromTypes);

              var imageSourceFrom = $scope.buildConfig.spec.source.image.from;
              $scope.options.pickedImageSourceType = imageSourceFrom.kind;
              $scope.options.pickedImageSourceNamespace = imageSourceFrom.namespace || buildConfig.metadata.namespace;
              $scope.options.pickedImageSourceImageStream = imageSourceFrom.name.split(":")[0];
              $scope.options.pickedImageSourceImageStreamTag = imageSourceFrom.name.split(":")[1];
              $scope.options.pickedImageSourceImageStreamImage = (imageSourceFrom.kind === "ImageStreamImage") ? imageSourceFrom.name : "";
              $scope.options.pickedImageSourceDockerImage = (imageSourceFrom.kind === "ImageStreamTag") ? buildConfig.metadata.namespace + "/" + imageSourceFrom.name : imageSourceFrom.name;

              // Save loaded value in case namespace, imageStream or tag are not available, to prevent data loss.
              $scope.imageSourceImageStream = {
                namespace: $scope.options.pickedImageSourceNamespace,
                imageStream: $scope.options.pickedImageSourceImageStream,
                tag: $scope.options.pickedImageSourceImageStreamTag,
              };
            }

            if ($scope.sources.binary) {
              $scope.options.binaryAsFile = ($scope.buildConfig.spec.source.binary.asFile) ? $scope.buildConfig.spec.source.binary.asFile : "";
            }

            if ($scope.strategyType === "Docker") {
              $scope.options.noCache = !!$scope.buildConfig.spec.strategy.dockerStrategy.noCache;
              // Only DockerStrategy can have empty Strategy object and therefore it's from object
              $scope.buildFromTypes.push({"id": "None", "title": "--- None ---"});
            }

            $scope.buildFrom.projects = ["openshift"];
            $scope.pushTo.projects = ["openshift"];
            DataService.list("projects", $scope, function(projects) {
              var projects = projects.by("metadata.name");
              for (var name in projects) {
                $scope.buildFrom.projects.push(name);
                $scope.pushTo.projects.push(name);
              }
              $scope.availableProjects = angular.copy($scope.buildFrom.projects);

              // If builder or output image namespace is not part of users available namespaces, add it to 
              // the namespace array anyway together with and call that checks the availability of the namespace.
              if (!$scope.buildFrom.projects.contains($scope.options.pickedBuildFromNamespace)) {
                $scope.checkNamespaceAvailability($scope.options.pickedBuildFromNamespace, "builder");
                $scope.buildFrom.projects.push($scope.options.pickedBuildFromNamespace);
              }
              if (!$scope.pushTo.projects.contains($scope.options.pickedPushToNamespace)) {
                $scope.checkNamespaceAvailability($scope.options.pickedPushToNamespace, "output");
                $scope.pushTo.projects.push($scope.options.pickedPushToNamespace);
              }

              // If builder or output image reference kind is DockerImage select the first imageSteam and imageStreamTag
              // in the picker, so when the user changes the reference kind to ImageStreamTag the picker is filled with
              // default(first) value.
              var builderSelectFirstOption = $scope.options.pickedBuildFromType === "DockerImage";
              $scope.updateBuilderImageStreams($scope.options.pickedBuildFromNamespace, builderSelectFirstOption);

              var outputSelectFirstOption = $scope.options.pickedPushToType === "DockerImage";
              $scope.updateOutputImageStreams($scope.options.pickedPushToNamespace, outputSelectFirstOption);

              if ($scope.sources.image) {
                $scope.imageSourceBuildFrom.projects = angular.copy($scope.buildFrom.projects);

                if (!$scope.imageSourceBuildFrom.projects.contains($scope.options.pickedImageSourceNamespace)) {
                  $scope.checkNamespaceAvailability($scope.options.pickedImageSourceNamespace, "imageSource");
                  $scope.imageSourceBuildFrom.projects.push($scope.options.pickedImageSourceNamespace);
                }
                var imageSourceSelectFirstOption = $scope.options.pickedImageSourceType === "DockerImage";
                $scope.updateImageSourceImageStreams($scope.options.pickedImageSourceNamespace, imageSourceSelectFirstOption);
              }
            });
            $scope.loaded = true;
            // If we found the item successfully, watch for changes on it
            watches.push(DataService.watchObject("buildconfigs", $routeParams.buildconfig, context, function(buildConfig, action) {
              if (action === "DELETED") {
                $scope.alerts["deleted"] = {
                  type: "warning",
                  message: "This build configuration has been deleted."
                };
                $scope.disableInputs = true;
              }
              $scope.buildConfig = buildConfig;
            }));
          },
          // failure
          function(e) {
            $scope.loaded = true;
            $scope.alerts["load"] = {
              type: "error",
              message: "The build configuration details could not be loaded.",
              details: "Reason: " + $filter('getErrorDetails')(e)
            };
          }
        );
      })
    );

    // When BuildFrom/PushTo/ImageBuildFrom type is change, appeared fields need to be filled with proper values.
    $scope.assambleInputType = function(type, pickedType) {
      switch (type) {
        case "builder":
          if ( pickedType === "DockerImage") {
            $scope.options.pickedBuildFromDockerImage = $scope.options.pickedBuildFromNamespace + "/" + $scope.options.pickedBuildFromImageStream + ":" + $scope.options.pickedBuildFromImageStreamTag;
          } else if ( pickedType === "ImageStreamTag") {
            $scope.updateBuilderImageStreams($scope.options.pickedBuildFromNamespace, true)
          }
          break;
        case "output":
          if ( pickedType === "DockerImage") {
            $scope.options.pickedPushToDockerImage = $scope.options.pickedPushToNamespace + "/" + $scope.options.pickedPushToImageStream + ":" + $scope.options.pickedPushToImageStreamTag;
          } else if ( pickedType === "ImageStreamTag") {
            $scope.updateOutputImageStreams($scope.options.pickedPushToNamespace, true)
          }
          break;
        case "imageSource":
          if ( pickedType === "DockerImage") {
            $scope.options.pickedImageSourceDockerImage = $scope.options.pickedImageSourceNamespace + "/" + $scope.options.pickedImageSourceImageStream + ":" + $scope.options.pickedImageSourceImageStreamTag;
          } else if ( pickedType === "ImageStreamTag") {
            $scope.updateImageSourceImageStreams($scope.options.pickedImageSourceNamespace, true)
          }
          break;
      }
    }

    $scope.aceLoaded = function(editor) {
      var session = editor.getSession();
      session.setOption('tabSize', 2);
      session.setOption('useSoftTabs', true);
      editor.$blockScrolling = Infinity;
    };

    // updateImageSourceImageStreams creates/updates the list of imageStreams and imageStreamTags for the builder image from picked namespace.
    // As parameter takes the picked namespace, selectFirstOption as a boolean that indicates whether the imageStreams and imageStreamTags
    // selectboxes should be select the first option.
    $scope.updateImageSourceImageStreams = function(projectName, selectFirstOption) {
      if (!$scope.availableProjects.contains(projectName)) {
        $scope.imageSourceBuildFrom.imageStreams = [];
        $scope.imageSourceBuildFrom.tags = {};
        $scope.imageSourceBuildFrom.imageStreams.push($scope.imageSourceImageStream.imageStream);
        $scope.options.pickedImageStreamImageStream = $scope.imageSourceImageStream.imageStream;
        $scope.imageSourceBuildFrom.tags[$scope.imageSourceImageStream.imageStream] = [$scope.imageSourceImageStream.tag];
        $scope.options.pickedImageStreamImageStreamTag = $scope.imageSourceImageStream.tag;
      } else {
        DataService.list("imagestreams", {namespace: projectName}, function(imageStreams) {
          $scope.imageSourceBuildFrom.imageStreams = [];
          $scope.imageSourceBuildFrom.tags = {};
          var projectImageStreams = imageStreams.by("metadata.name");
          if (!_.isEmpty(projectImageStreams)) {
            if (!Object.keys(projectImageStreams).contains($scope.imageSourceBuildFrom.imageStream) && projectName === $scope.imageSourceBuildFrom.namespace) {
              $scope.imageSourceBuildFrom.imageStreams.push($scope.imageSourceImageStream.imageStream);
              $scope.options.pickedImageStreamImageStream = $scope.imageSourceImageStream.imageStream;
              $scope.imageSourceBuildFrom.tags[$scope.imageSourceImageStream.imageStream] = [$scope.imageSourceImageStream.tag];
              $scope.options.pickedImageStreamImageStreamTag = $scope.imageSourceImageStream.tag;
            }
            angular.forEach(projectImageStreams, function(imageStream, name) {
              // List only OpenShift builder images
              if (isOpenshift && !$scope.openshiftBuilderImages.contains(name)) {return;}
              var tagList = [];
              if (imageStream.status.tags) {
                imageStream.status.tags.forEach(function(item){
                  tagList.push(item["tag"]);
                });
              }
              $scope.imageSourceBuildFrom.imageStreams.push(name);
              // If the namespace and the imageStream are matching with the picked one, add the unavailable tag to the tagList.
              // Prevents losing of data, while build is still taking place and the tag for the imageSource imageStream is not available.
              if (projectName === $scope.imageSourceBuildFrom.namespace && imageStream.metadata.name === $scope.imageSourceBuildFrom.imageStream && _.indexOf(tagList, $scope.imageSourceBuildFrom.tag) === -1) {
                tagList.push($scope.imageSourceBuildFrom.tag);
              }
              $scope.imageSourceBuildFrom.tags[name] = tagList;
              // If ImageStream doesn't have any tags, set tag to empty string, so the user has to pick from a existing Tags.
              if (name === $scope.options.pickedImageSourceImageStream && _.isEmpty(tagList)) {
                $scope.options.pickedImageSourceImageStreamTag = "";
              }
            });
            // If defined Image Stream is not present in the Image Streams of picked Namespace set tag to empty string,
            // so the user has to pick from a existing Image Streams.
            if (!$scope.imageSourceBuildFrom.imageStreams.contains($scope.options.pickedImageSourceImageStream)) {
              $scope.options.pickedImageSourceImageStreamTag = "";
            }
            if (selectFirstOption) {
              $scope.options.pickedImageSourceImageStream = $scope.imageSourceBuildFrom.imageStreams[0];
              $scope.clearSelectedTag($scope.imageSourceBuildFrom.tags, $scope.options.pickedImageSourceImageStream);
            }
          } else if (projectName === $scope.outputImageStream.namespace) {
            $scope.imageSourceBuildFrom.imageStreams.push($scope.imageSourceImageStream.imageStream);
            $scope.options.pickedImageStreamImageStream = $scope.imageSourceImageStream.imageStream;
            $scope.imageSourceBuildFrom.tags[$scope.imageSourceImageStream.imageStream] = [$scope.imageSourceImageStream.tag];
            $scope.options.pickedImageStreamImageStreamTag = $scope.imageSourceImageStream.tag;
          } else {
            $scope.options.pickedImageSourceImageStream = "";
            $scope.options.pickedImageSourceImageStreamTag = "";
          }
        });
      }
    }

    // updateBuilderImageStreams creates/updates the list of imageStreams and imageStreamTags for the builder image from picked namespace.
    // As parameter takes the picked namespace, selectFirstOption as a boolean that indicates whether the imageStreams and imageStreamTags
    // selectboxes should be select the first option.
    $scope.updateBuilderImageStreams = function(projectName, selectFirstOption) {
      if (!$scope.availableProjects.contains(projectName)) {
        $scope.buildFrom.imageStreams = [];
        $scope.buildFrom.tags = {};
        $scope.buildFrom.imageStreams.push($scope.builderImageStream.imageStream);
        $scope.options.pickedBuildFromImageStream = $scope.builderImageStream.imageStream;
        $scope.buildFrom.tags[$scope.builderImageStream.imageStream] = [$scope.builderImageStream.tag];
        $scope.options.pickedBuildFromImageStreamTag = $scope.builderImageStream.tag;
      } else {
        DataService.list("imagestreams", {namespace: projectName}, function(imageStreams) {
          var isOpenshift = (projectName === "openshift") ? true : false
          $scope.buildFrom.imageStreams = [];
          $scope.buildFrom.tags = {};
          var projectImageStreams = imageStreams.by("metadata.name");
          if (!_.isEmpty(projectImageStreams)) {
            if (!Object.keys(projectImageStreams).contains($scope.builderImageStream.imageStream) && projectName === $scope.builderImageStream.namespace) {
              $scope.buildFrom.imageStreams.push($scope.builderImageStream.imageStream);
              $scope.options.pickedBuildFromImageStream = $scope.builderImageStream.imageStream;
              $scope.buildFrom.tags[$scope.builderImageStream.imageStream] = [$scope.builderImageStream.tag];
              $scope.options.pickedBuildFromImageStreamTag = $scope.builderImageStream.tag;
            }
            angular.forEach(projectImageStreams, function(imageStream, name) {
              // List only OpenShift builder images
              if (isOpenshift && !$scope.openshiftBuilderImages.contains(name)) {return;}
              var tagList = [];
              if (imageStream.status.tags) {
                imageStream.status.tags.forEach(function(item){
                  tagList.push(item["tag"]);
                });
              }
              $scope.buildFrom.imageStreams.push(name);
              // If the namespace and the imageStream are matching with the picked one, add the unavailable tag to the tagList.
              // Prevents losing of data, while build is still taking place and the tag for the output imageStream is not available.
              if (projectName === $scope.builderImageStream.namespace && imageStream.metadata.name === $scope.builderImageStream.imageStream && _.indexOf(tagList, $scope.builderImageStream.tag) === -1) {
                tagList.push($scope.builderImageStream.tag);
              }

              $scope.buildFrom.tags[name] = tagList;
              // If ImageStream doesn't have any tags, set tag to empty string, so the user has to pick from a existing Tags.
              if (name === $scope.options.pickedBuildFromImageStream &&_.isEmpty(tagList)) {
                $scope.options.pickedBuildFromImageStreamTag = "";
              }
            });
            // If defined Image Stream is not present in the Image Streams of picked Namespace set tag to empty string,
            // so the user has to pick from a existing Image Streams.
            if (!$scope.buildFrom.imageStreams.contains($scope.options.pickedBuildFromImageStream)) {
              $scope.options.pickedBuildFromImageStreamTag = "";
            }
            if (selectFirstOption) {
              $scope.options.pickedBuildFromImageStream = $scope.buildFrom.imageStreams[0];
              $scope.clearSelectedTag($scope.buildFrom.tags, $scope.options.pickedBuildFromImageStream);
            }
          } else if (projectName === $scope.builderImageStream.namespace) {
            $scope.buildFrom.imageStreams.push($scope.builderImageStream.imageStream);
            $scope.options.pickedBuildFromImageStream = $scope.builderImageStream.imageStream;
            $scope.buildFrom.tags[$scope.builderImageStream.imageStream] = [$scope.builderImageStream.tag];
            $scope.options.pickedBuildFromImageStreamTag = $scope.builderImageStream.tag;
          } else {
            $scope.options.pickedBuildFromImageStream = "";
            $scope.options.pickedBuildFromImageStreamTag = "";
          }
        });
      }
    }

    // updateOutputImageStreams creates/updates the list of imageStreams and imageStreamTags for the output image from picked namespace.
    // As parameter takes the picked namespace, selectFirstOption as a boolean that indicates whether the imageStreams and imageStreamTags
    // selectboxes should be select the first option.
    $scope.updateOutputImageStreams = function(projectName, selectFirstOption) {
      if (!$scope.availableProjects.contains(projectName)) {
        $scope.pushTo.imageStreams = [];
        $scope.pushTo.tags = {};
        $scope.pushTo.imageStreams.push($scope.outputImageStream.imageStream);
        $scope.options.pickedPushToImageStream = $scope.outputImageStream.imageStream;
        $scope.options.pickedPushToImageStreamTag = $scope.outputImageStream.tag;
      } else {
        DataService.list("imagestreams", {namespace: projectName}, function(imageStreams) {
          $scope.pushTo.imageStreams = [];
          $scope.pushTo.tags = {};
          var projectImageStreams = imageStreams.by("metadata.name");
          if (!_.isEmpty(projectImageStreams)) {
            if (!Object.keys(projectImageStreams).contains($scope.outputImageStream.imageStream) && projectName === $scope.outputImageStream.namespace) {
              $scope.pushTo.imageStreams.push($scope.outputImageStream.imageStream);
              $scope.options.pickedPushToImageStream = $scope.outputImageStream.imageStream;
              $scope.options.pickedPushToImageStreamTag = $scope.outputImageStream.tag;
            }
            angular.forEach(projectImageStreams, function(imageStream, name) {
              var tagList = [];
              if (imageStream.status.tags) {
                imageStream.status.tags.forEach(function(item){
                  tagList.push(item["tag"]);
                });
              }
              $scope.pushTo.imageStreams.push(name);
              $scope.pushTo.tags[name] = tagList;
            });

            if (selectFirstOption) {
              $scope.options.pickedPushToImageStream = $scope.pushTo.imageStreams[0];
              $scope.clearSelectedTag($scope.pushTo.tags, $scope.options.pickedPushToImageStream);
              // If defined Image Stream is not present in the Image Streams of picked Namespace set tag to empty string,
              // so the user has to pick from a existing Image Streams.
            } else if (!$scope.pushTo.imageStreams.contains($scope.options.pickedPushToImageStream)) {
              $scope.options.pickedPushToImageStreamTag = "";
            }
          } else if (projectName === $scope.outputImageStream.namespace) {
            $scope.pushTo.imageStreams.push($scope.outputImageStream.imageStream);
            $scope.options.pickedPushToImageStream = $scope.outputImageStream.imageStream;
            $scope.options.pickedPushToImageStreamTag = $scope.outputImageStream.tag;
          } else {
            $scope.options.pickedPushToImageStream = "";
            $scope.options.pickedPushToImageStreamTag = "";
          }
        });
      }
    }

    $scope.clearSelectedTag = function(tagHash, pickedImageStream) {
      var tags = tagHash[pickedImageStream];
      if (tags.length > 0) {
        return _.find(tags, function(tag) { return tag == "latest" }) || tags[0]
      } else {
        return  "latest";
      }
    }

    // Check if the namespace is available. If so add him to available namespaces and remove him from unavailable 
    $scope.checkNamespaceAvailability = function(ns, type) {
      DataService.get("projects", ns, {}, { errorNotification: false})
      .then(function() {
        $scope.availableProjects.push(ns);
      }, function(result) {
        if (result.status === 403) {
          $scope.alerts["load-" + ns] = {
            type: "error",
            message: "Project '" + ns + "' is not available for your account.",
            details: "Reason: " + $filter('getErrorDetails')(result)
          };
        } else {
          $scope.alerts["load-" + ns] = {
            type: "error",
            message: "An error occurred loading the " + ns + " project.",
            details: "Reason: " + $filter('getErrorDetails')(result)
          };
        }
      });
    }

    $scope.save = function() {
      $scope.disableInputs = true;
      // Update Configuration
      $filter('buildStrategy')($scope.updatedBuildConfig).forcePull = $scope.options.forcePull;
      if ($scope.strategyType === "Docker") {
        $filter('buildStrategy')($scope.updatedBuildConfig).noCache = $scope.options.noCache;
      }

      // If binarySource check if the AsFile string is set and construct the object accordingly.
      if ($scope.options.binaryAsFile) {
        if ($scope.options.binaryAsFile !== "") {
          $scope.updatedBuildConfig.spec.source.binary.asFile = $scope.options.binaryAsFile;
        } else {
          $scope.updatedBuildConfig.spec.source.binary = {};
        }
      }

      // If imageSource is present update From and Paths.
      if ($scope.sources.image) {
        var updatedImageSourcePath = [];
        angular.forEach($scope.imageSourcePaths, function(v, k) {
          var env = {
            sourcePath: k,
            destinationDir: v
          };
          updatedImageSourcePath.push(env);
        });
        $scope.updatedBuildConfig.spec.source.image.paths  = updatedImageSourcePath;

        // Construct updated imageSource builder image object based on it's kind
        var from = {};
        if ($scope.options.pickedImageSourceType === "ImageStreamTag") {
          from = {
            kind: $scope.options.pickedImageSourceType,
            namespace: $scope.options.pickedImageSourceNamespace,
            name: $scope.options.pickedImageSourceImageStream + ":" + $scope.options.pickedImageSourceImageStreamTag
          };
        } else if ($scope.options.pickedImageSourceType === "DockerImage") {
          from = {
            kind: $scope.options.pickedImageSourceType,
            name: $scope.options.pickedImageSourceDockerImage
          };
        } else if($scope.options.pickedImageSourceType === "ImageStreamImage") {
          from = {
            kind: $scope.options.pickedImageSourceType,
            namespace: $scope.options.pickedImageSourceNamespace,
            name: $scope.options.pickedImageSourceImageStreamImage
          }
        }
        $scope.updatedBuildConfig.spec.source.image.from = from;
      }

      // Construct updated builder image object based on it's kind
      var from = {};
      if ($scope.options.pickedBuildFromType === "ImageStreamTag") {
        from = {
          kind: $scope.options.pickedBuildFromType,
          namespace: $scope.options.pickedBuildFromNamespace,
          name: $scope.options.pickedBuildFromImageStream + ":" + $scope.options.pickedBuildFromImageStreamTag
        };
      } else if ($scope.options.pickedBuildFromType === "DockerImage") {
        from = {
          kind: $scope.options.pickedBuildFromType,
          name: $scope.options.pickedBuildFromDockerImage
        };
      } else if ($scope.options.pickedBuildFromType === "ImageStreamImage") {
        from = {
          kind: $scope.options.pickedBuildFromType,
          namespace: $scope.options.pickedBuildFromNamespace,
          name: $scope.options.pickedBuildFromImageStreamImage
        };
      }
      if ($scope.options.pickedBuildFromType === "None") {
        delete $filter('buildStrategy')($scope.updatedBuildConfig).from
      } else {
        $filter('buildStrategy')($scope.updatedBuildConfig).from = from;
      }

      // Construct updated output image object based on it's kind
      var to = {};
      if ($scope.options.pickedPushToType === "ImageStreamTag") {
        to = {
          kind: $scope.options.pickedPushToType,
          namespace: $scope.options.pickedPushToNamespace,
          name: $scope.options.pickedPushToImageStream + ":" + $scope.options.pickedPushToImageStreamTag
        };
      } else if ($scope.options.pickedPushToType === "DockerImage") {
        to = {
          kind: $scope.options.pickedPushToType,
          name: $scope.options.pickedPushToDockerImage
        };
      }
      if ($scope.options.pickedPushToType === "None") {
        // If user will change the output reference to 'None' shall the potential PushSecret be deleted as well?
        // This case won't delete them.
        delete($scope.updatedBuildConfig.spec.output.to)
      } else {
        $scope.updatedBuildConfig.spec.output.to = to;
      }

      // Update envVars
      $filter('buildStrategy')($scope.updatedBuildConfig).env = BuildConfigsService.updateEnvVars($scope.envVars);

      // Update triggers
      var triggers = [];
      if ($scope.triggers.webhook) {
        var webhooks = _.filter($scope.buildConfig.spec.triggers, function(obj) { return obj.type == "GitHub" })
        if (webhooks.length === 0) {
          webhooks.push({
            github: {
              secret: ApplicationGenerator._generateSecret()
            },
            type: "GitHub"
          });
        }
        triggers = triggers.concat(webhooks);
      }
      triggers = triggers.concat(_.filter($scope.buildConfig.spec.triggers, function(obj) { return obj.type == "Generic" }));

      if ($scope.triggers.imageChange) {
        var imageChangeTriggers = _.filter($scope.buildConfig.spec.triggers, function(obj) { return obj.type == "ImageChange" });
        if (imageChangeTriggers.length === 0) {
          imageChangeTriggers.push({
            imageChange: {},
            type: "ImageChange"
          });
        }
        triggers = triggers.concat(imageChangeTriggers);
      }
      if ($scope.triggers.configChange) {
        triggers.push({
          type: "ConfigChange"
        });
      }
      $scope.updatedBuildConfig.spec.triggers = triggers;

      DataService.update("buildconfigs", $scope.updatedBuildConfig.metadata.name, $scope.updatedBuildConfig, {
        namespace: $scope.updatedBuildConfig.metadata.namespace
      }).then(
        function() {
          AlertMessageService.addAlert({
            name: $scope.updatedBuildConfig.metadata.name,
            data: {
              type: "success",
              message: "Build Config " + $scope.updatedBuildConfig.metadata.name + " was sucessfully updated."
            }
          });
          $location.path(Navigate.resourceURL($scope.updatedBuildConfig, "BuildConfig", $scope.updatedBuildConfig.metadata.namespace));
        },
        function(result) {
          $scope.disableInputs = false;

          $scope.alerts["save"] = {
            type: "error",
            message: "An error occurred updating the build " + $scope.updatedBuildConfig.metadata.name + "Build Config",
            details: $filter('getErrorDetails')(result)
          }
        }
      );
    };

    $scope.isNamespaceAvailable = function(namespace) {
      if ($scope.availableProjects.contains(namespace)) {
        return true;
      }
      return false;
    };

    $scope.inspectNamespace = function(imageStreams, pickedImageStream) {
      if (imageStreams.length === 0) {
        return "empty";
      } 
      if (imageStreams.length !== 0 && !imageStreams.contains(pickedImageStream)) {
        return "noMatch";
      }
      return "";
    };

    $scope.inspectTags = function(tagHash, pickedImageStream, pickedTag) {
      if (tagHash[pickedImageStream]) {
        if (tagHash[pickedImageStream].length === 0) {
          return "empty";
        } 
        if (tagHash[pickedImageStream].length !== 0 && !tagHash[pickedImageStream].contains(pickedTag)) {
          return "noMatch";
        }
      }
      return "";
    };

    $scope.showOutputTagWarning = function(form) {
      if (form.outputNamespace.$dirty || form.outputImageStream.$dirty || form.outputTag.$dirty) {
        if ($scope.pushTo.tags[$scope.options.pickedPushToImageStream] && $scope.pushTo.tags[$scope.options.pickedPushToImageStream].contains($scope.options.pickedPushToImageStreamTag)) {
          return true;
        }
      }
      return false;
    };

    $scope.getSourceMap = function(sourceMap, sources) {
      angular.forEach(sources, function(value, key) {
        switch (key) {
          case "binary":
            sourceMap.binary = true;
            break;
          case "dockerfile":
            sourceMap.dockerfile = true;
            break;
          case "git":
            sourceMap.git = true;
            break;
          case "image":
            sourceMap.image = true;
            break;
          case "contextDir":
            sourceMap.contextDir = true;
            break;
        }
      });
      return sourceMap;
    };

  });
