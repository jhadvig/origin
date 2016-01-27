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
    $scope.buildFromTypes = [
      {
        "id": "ImageStreamTag",
        "title": "Image Stream Tag"
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
    $scope.buildFrom1 = {
      projects: [],
      imageStreams: [],
      tags: {},
    };
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

    $scope.myOptions = [];

    $scope.buildFromTest = [];

    $scope.namespaceConfig = {
      create: true,
      valueField: 'title',
      labelField: 'title',
      delimiter: '|',
      placeholder: 'Pick Namespace',
      onChange: function(item){
        console.log("ns");
        console.log(item);
      },
      maxItems: 1
    };

    $scope.imageStreamConfig = {
      create: true,
      valueField: 'title',
      labelField: 'title',
      delimiter: '|',
      placeholder: 'Pick Image Stream',
      onChange: function(item){
        console.log(item);
      },
      maxItems: 1
    };

    $scope.tagConfig = {
      create: true,
      valueField: 'title',
      labelField: 'title',
      delimiter: '|',
      placeholder: 'Pick Tag',
      onChange: function(item){
        console.log(item);
      },
      maxItems: 1
    };

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

            angular.forEach($scope.buildConfig.spec.triggers, function(value, key) {
              switch (value.type) {
                case "Generic":
                  break;
                case "GitHub":
                  $scope.triggers.webhook = true;
                  break;
                case "ImageChange":
                  $scope.triggers.imageChange = true;
                  break;
                case "ConfigChange":
                  $scope.triggers.configChange = true;
                  break;
              }
            });
            $scope.pickedTriggers = $scope.triggers;

            angular.forEach($scope.buildConfig.spec.source, function(value, key) {
              switch (key) {
                case "binary":
                  $scope.sources.binary = true;
                  break
                case "dockerfile":
                  $scope.sources.dockerfile = true;
                  break
                case "git":
                  $scope.sources.git = true;
                  break
                case "image":
                  $scope.sources.image = true;
                  break
                case "contextDir":
                  $scope.sources.contextDir = true;
                  break
              }
            });

            var setBuildFromVariables = function (type, ns, is, ist, di) {
              $scope.options.pickedBuildFromType = type;
              $scope.options.pickedBuildFromNamespace = ns;
              $scope.options.pickedBuildFromImageStream = is;
              $scope.options.pickedBuildFromImageStreamTag = ist;
              $scope.options.pickedBuildFromDockerImage = di;
            }
            var setPushToVariables = function (type, ns, is, ist, di) {
              $scope.options.pickedPushToType = type;
              $scope.options.pickedPushToNamespace = ns;
              $scope.options.pickedPushToImageStream = is;
              $scope.options.pickedPushToImageStreamTag = ist;
              $scope.options.pickedPushToDockerImage = di;
            }

            if ($scope.buildStrategy.from) {
              var buildFrom = $scope.buildStrategy.from;
              setBuildFromVariables(buildFrom.kind,
                                    buildFrom.namespace || buildConfig.metadata.namespace, 
                                    buildFrom.name.split(":")[0], 
                                    buildFrom.name.split(":")[1], 
                                    (buildFrom.kind === "ImageStreamTag") ? buildConfig.metadata.namespace + "/" + buildFrom.name : buildFrom.name);
            } else {
              setBuildFromVariables("None", buildConfig.metadata.namespace, "", "", "");
            }

            if ($scope.updatedBuildConfig.spec.output.to) {
              var pushTo = $scope.updatedBuildConfig.spec.output.to
              setPushToVariables(pushTo.kind,
                                pushTo.namespace || buildConfig.metadata.namespace,
                                pushTo.name.split(":")[0],
                                pushTo.name.split(":")[1],
                                (pushTo.kind === "ImageStreamTag") ? buildConfig.metadata.namespace + "/" + pushTo.name : pushTo.name);
            } else {
              setPushToVariables("None", buildConfig.metadata.namespace, "", "", "");
            }

            $scope.options.forcePull = !!$scope.buildStrategy.forcePull,

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

            if ($scope.sources.image) {
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
              $scope.options.pickedImageSourceDockerImage = (imageSourceFrom.kind === "ImageStreamTag") ? buildConfig.metadata.namespace + "/" + imageSourceFrom.name : imageSourceFrom.name;

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

            // $scope.buildFrom.projects = ["openshift"];
            $scope.buildFrom1.projects = [{title: "openshift"}];

            
            DataService.list("projects", $scope, function(projects) {
              var projects = projects.by("metadata.name");
              for (var name in projects) {
                $scope.buildFrom.projects.push(name);
                $scope.pushTo.projects.push(name);

                $scope.buildFrom1.projects.push({title: name});
              }

              $scope.pushTo.projects.forEach(function(project, index) {
                $scope.buildFromTest.push({
                  id: index,
                  title: project
                });
              });
              $scope.myModelTest = 1;

              // If builder or output image reference kind is DockerImage select the first imageSteam and imageStreamTag
              // in the picker, so when the user changes the reference kind to ImageStreamTag the picker is filled with
              // default(first) value.
              var builderSelectFirstOption = $scope.options.pickedBuildFromType === "DockerImage";
              var outputSelectFirstOption = $scope.options.pickedPushToType === "DockerImage";
              $scope.updateBuilderImageStreams($scope.options.pickedBuildFromNamespace, builderSelectFirstOption);
              $scope.updateOutputImageStreams($scope.options.pickedPushToNamespace, outputSelectFirstOption);

              if ($scope.sources.image) {
                $scope.imageSourceBuildFrom.projects = angular.copy($scope.buildFrom.projects);
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
      }));

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
      DataService.list("imagestreams", {namespace: projectName}, function(imageStreams) {
        $scope.imageSourceBuildFrom.imageStreams = [];
        $scope.imageSourceBuildFrom.tags = {};
        var projectImageStreams = imageStreams.by("metadata.name");
        if (!_.isEmpty(projectImageStreams)) {
          angular.forEach(projectImageStreams, function(imageStream, name) {
            var tagList = [];
            if (imageStream.status.tags) {
              imageStream.status.tags.forEach(function(item){
                tagList.push(item["tag"]);
              });
            }
            $scope.imageSourceBuildFrom.imageStreams.push(name);
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
            $scope.clearSelectedImageSourceTag();
          }
        } else {
          $scope.options.pickedImageSourceImageStream = "";
          $scope.options.pickedImageSourceImageStreamTag = "";
        }
      });
    }

    $scope.clearSelectedImageSourceTag = function() {
      var tags = $scope.imageSourceBuildFrom.tags[$scope.options.pickedImageSourceImageStream];
      $scope.options.pickedImageSourceImageStreamTag = _.find(tags, function(tag) { return tag == "latest" }) || tags[0] || "";
    }

    // updateBuilderImageStreams creates/updates the list of imageStreams and imageStreamTags for the builder image from picked namespace.
    // As parameter takes the picked namespace, selectFirstOption as a boolean that indicates whether the imageStreams and imageStreamTags
    // selectboxes should be select the first option.
    $scope.updateBuilderImageStreams = function(projectName, selectFirstOption) {
      if ($scope.options.pickedBuildFromType === 'None') {
        $scope.pickedTriggers['imageChange'] = false;
      } else {
        DataService.list("imagestreams", {namespace: projectName}, function(imageStreams) {
          $scope.buildFrom.imageStreams = [];
          $scope.buildFrom.tags = {};
          var projectImageStreams = imageStreams.by("metadata.name");
          if (!_.isEmpty(projectImageStreams)) {
            angular.forEach(projectImageStreams, function(imageStream, name) {
              var tagList = [];
              if (imageStream.status.tags) {
                imageStream.status.tags.forEach(function(item){
                  tagList.push(item["tag"]);
                });
              }
              

              $scope.buildFrom.imageStreams.push(name);
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
              $scope.clearSelectedBuilderTag();
            }

            $scope.myOptions = BuildConfigsService.arrayToOptions($scope.buildFrom.imageStreams);
            $scope.myModel = $scope.options.pickedPushToImageStream;
          } else {
            $scope.options.pickedBuildFromImageStream = "";
            $scope.options.pickedBuildFromImageStreamTag = "";
          }
        });
      }
    }

    $scope.clearSelectedBuilderTag = function() {
      var tags = $scope.buildFrom.tags[$scope.options.pickedBuildFromImageStream];
      $scope.options.pickedBuildFromImageStreamTag = _.find(tags, function(tag) { return tag == "latest" }) || tags[0] || "";
    }

    // updateOutputImageStreams creates/updates the list of imageStreams and imageStreamTags for the output image from picked namespace.
    // As parameter takes the picked namespace, selectFirstOption as a boolean that indicates whether the imageStreams and imageStreamTags
    // selectboxes should be select the first option.
    $scope.updateOutputImageStreams = function(projectName, selectFirstOption) {
      DataService.list("imagestreams", {namespace: projectName}, function(imageStreams) {
        $scope.pushTo.imageStreams = [];
        $scope.pushTo.tags = {};
        var projectImageStreams = imageStreams.by("metadata.name");
        if (!_.isEmpty(projectImageStreams)) {
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
            $scope.clearSelectedOutputTag();
            // If defined Image Stream is not present in the Image Streams of picked Namespace set tag to empty string,
            // so the user has to pick from a existing Image Streams.
          } else if (!$scope.pushTo.imageStreams.contains($scope.options.pickedPushToImageStream)) {
            $scope.options.pickedPushToImageStreamTag = "";
          }
        } else {
          $scope.options.pickedPushToImageStream = "";
          $scope.options.pickedPushToImageStreamTag = "";
        }
      });
    }

    $scope.clearSelectedOutputTag = function() {
      var tags = $scope.pushTo.tags[$scope.options.pickedPushToImageStream];
      $scope.options.pickedPushToImageStreamTag = _.find(tags, function(tag) { return tag == "latest" }) || tags[0] || "latest";
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
      var updateEnvVars = [];
      angular.forEach($scope.envVars, function(v, k) {
        var env = {
          name: k,
          value: v
        };
        updateEnvVars.push(env);
      });
      $filter('buildStrategy')($scope.updatedBuildConfig).env = updateEnvVars;

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
  });
