'use strict';

/**
 * @ngdoc function
 * @name openshiftConsole.controller:EditBuildConfigController
 * @description
 * Controller of the openshiftConsole
 */
angular.module('openshiftConsole')
  .controller('EditBuildConfigController', function ($scope, $routeParams, DataService, ProjectsService, $filter, ApplicationGenerator, Navigate, $location, AlertMessageService) {

    $scope.projectName = $routeParams.project;
    $scope.buildConfig = null;
    $scope.alerts = {};
    $scope.emptyMessage = "Loading...";
    $scope.sourceURLPattern = ApplicationGenerator.sourceURLPattern;
    $scope.options = {};

    $scope.buildFromTypes = [
      {
        "id": "ImageStreamTag",
        "title": "Image Stream Tag"
      },
      {
        "id": "DockerImage",
        "title": "Docker Image Link"
      },
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
    }
    $scope.pushTo = {
      projects: [],
      imageStreams: [],
      tags: {},
    }

    $scope.triggers = {
      "webhook": false,
      "imageChange": false,
      "configChange": false
    };

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
            $scope.envVars = $filter('envVarsPair')($scope.buildStrategy.env);

            angular.forEach($scope.buildConfig.spec.triggers, function(value, key) {
              switch (value.type) {
                case "Generic":
                  break;
                case "GitHub":
                  $scope.triggers["webhook"] = true;
                  break;
                case "ImageChange":
                  $scope.triggers["imageChange"] = true;
                  break;
                case "ConfigChange":
                  $scope.triggers["configChange"] = true;
                  break;
              }
            });
            $scope.pickedTriggers = $scope.triggers;
            

            var buildFrom = $scope.buildStrategy.from;
            var testFrom = $filter('imageObjectRef')(buildFrom, buildConfig.metadata.namespace)
            var pushTo = $filter('imageObjectRef')(buildConfig.spec.output.to, buildConfig.metadata.namespace)

            $scope.options = {
              pickedBuildFromType: buildFrom.kind,
              pickedBuildFromNamespace: buildFrom.namespace || $scope.projectName,
              pickedBuildFromImageStream: buildFrom.name.split(":")[0],
              pickedBuildFromImageStreamTag: buildFrom.name.split(":")[1],

              pickedPushToNamespace: buildConfig.spec.output.to.namespace || buildConfig.metadata.namespace,        
              pickedPushToImageStream: buildConfig.spec.output.to.name.split(":")[0],
              pickedPushToImageStreamTag: buildConfig.spec.output.to.name.split(":")[1],

              pickedBuildFromDockerImage: $filter('buildStrategy')($scope.updatedBuildConfig).from.name || "",

              forcePull: !!$scope.buildStrategy.forcePull,

            };

            $scope.outputImageStream = {
              namespace: $scope.options.pickedPushToNamespace,
              imageStream: $scope.options.pickedPushToImageStream,
              tag: $scope.options.pickedPushToImageStreamTag,
            };

            if ($scope.buildConfig.spec.strategy.type === "Docker") {
              $scope.options.noCache = !!$scope.buildConfig.spec.strategy.dockerStrategy.noCache;
            }

            $scope.buildFrom.projects = ["openshift"];
            DataService.list("projects", $scope, function(projects) {
              var projects = projects.by("metadata.name");
              for (var name in projects) {
                $scope.buildFrom.projects.push(name);
                $scope.pushTo.projects.push(name);
              }
              $scope.updateBuilderImageStreams($scope.options.pickedBuildFromNamespace, false);
              $scope.updateOutputImageStreams($scope.options.pickedPushToNamespace, false);
            });
            $scope.loaded = true;
            // If we found the item successfully, watch for changes on it
            watches.push(DataService.watchObject("buildconfigs", $routeParams.buildconfig, context, function(buildConfig, action) {
              if (action === "DELETED") {
                $scope.alerts["deleted"] = {
                  type: "warning",
                  message: "This build configuration has been deleted."
                };
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

    // updateBuilderImageStreams creates/updates the list of imageStreams and imageStreamTags for the builder image from picked namespace.
    // As parameter takes the picked namespace, selectFirstOption as a boolean that indicates whether the imageStreams and imageStreamTags
    // selectboxes should be select the first option.
    $scope.updateBuilderImageStreams = function(projectName, selectFirstOption) {
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
          });
          if (selectFirstOption) {
            $scope.options.pickedBuildFromImageStream = $scope.buildFrom.imageStreams[0];
            $scope.clearSelectedBuilderTag();
          }
        } else {
          $scope.options.pickedBuildFromImageStream = "";
          $scope.options.pickedBuildFromImageStreamTag = "";
        }
      });
    }

    $scope.clearSelectedBuilderTag = function() {
      var tags = $scope.buildFrom.tags[$scope.options.pickedBuildFromImageStream];
      $scope.options.pickedBuildFromImageStreamTag = _.find(tags, function(tag) { return tag == "latest" }) || tags[0] || "";
    }

    // updateOutputImageStreams creates/updates the list of imageStreams and imageStreamTags for the output image from picked namespace.
    // As parameter takes the picked namespace, selectFirstOption as a boolean that indicates whether the imageStreams and imageStreamTags
    // selectboxes should be select the first option.
    $scope.updateOutputImageStreams = function(projectName, selectFirstOption) {
      var outputImageStreamListEmpty = true;
      DataService.list("imagestreams", {namespace: projectName}, function(imageStreams) {
        outputImageStreamListEmpty = false;
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
            // If the namespace and the imageStream are matching with the picked one, add the unavailable tag to the tagList.  
            // Prevents losing of data, while build is still taking place and the tag for the output imageStream is not available. 
            if (projectName === $scope.outputImageStream.namespace && imageStream.metadata.name === $scope.outputImageStream.imageStream && _.indexOf(tagList, $scope.outputImageStream.tag) === -1) {
              tagList.push($scope.outputImageStream.tag);
            }
            $scope.pushTo.tags[name] = tagList;
          });
          if (selectFirstOption) {
            $scope.options.pickedPushToImageStream = $scope.pushTo.imageStreams[0];
            $scope.clearSelectedOutputTag();
          } 
        } else {
          $scope.options.pickedPushToImageStream = "";
          $scope.options.pickedPushToImageStreamTag = "";
        }
      });
    }

    $scope.clearSelectedOutputTag = function() {
      var tags = $scope.pushTo.tags[$scope.options.pickedPushToImageStream];
      $scope.options.pickedPushToImageStreamTag = _.find(tags, function(tag) { return tag == "latest" }) || tags[0] || "";
    }

    $scope.save = function() {
      $scope.disableInputs = true;
      // Update Configuration
      $filter('buildStrategy')($scope.updatedBuildConfig).forcePull = $scope.options.forcePull;
      if ($scope.updatedBuildConfig.spec.strategy.type === "Docker") {
        $filter('buildStrategy')($scope.updatedBuildConfig).noCache = $scope.options.noCache;
      }

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
      $filter('buildStrategy')($scope.updatedBuildConfig).from = from;

      var to = {
        kind: "ImageStreamTag",
        namespace: $scope.options.pickedPushToNamespace,
        name: $scope.options.pickedPushToImageStream + ":" + $scope.options.pickedPushToImageStreamTag
      };
      $scope.updatedBuildConfig.spec.output.to = to;

      // Update envVars
      var updateEnvVars = [];
      angular.forEach($scope.envVars, function(v, k) {
        var env = {
          name: k,
          value: v
        };
        updateEnvVars.push(env);
      });
      $filter('buildStrategy')($scope.updatedBuildConfig).env = updateEnvVars

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
          $location.path(Navigate.resourceURL($scope.updatedBuildConfig, "BuildConfig", $scope.updatedBuildConfig.metadata.namespace, "browse"));
        },
        function(result) {
          $scope.disableInputs = false;
          AlertMessageService.addAlert({
            name: $scope.updatedBuildConfig.metadata.name,
            data: {
              type: "error",
              message: "An error occurred updating the build " + $scope.updatedBuildConfig.metadata.name + "Build Config",
              details: $filter('getErrorDetails')(result)
            }
          });
        }
      );
    };
  });
