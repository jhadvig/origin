'use strict';

/**
 * @ngdoc function
 * @name openshiftConsole.controller:EditBuildConfigController
 * @description
 * Controller of the openshiftConsole
 */
angular.module('openshiftConsole')
  .controller('EditBuildConfigController', function ($scope, $routeParams, DataService, ProjectsService, $filter) {

    $scope.projectName = $routeParams.project;
    $scope.buildConfig = null;
    $scope.alerts = {};
    $scope.emptyMessage = "Loading...";
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

    $scope.available = {
      projects: [],
      imageStreams: [],
      tags: {},
    }

    var watches = [];

    ProjectsService
      .get($routeParams.project)
      .then(_.spread(function(project, context) {
        $scope.project = project;
        DataService.get("buildconfigs", $routeParams.buildconfig, context).then(
          // success
          function(buildConfig) {
            $scope.buildConfig = buildConfig;
            $scope.bcEdit = angular.copy($scope.buildConfig);

            var buildFrom = $filter('buildStrategy')($scope.bcEdit).from;

            var pushTo = $filter('imageObjectRef')(buildConfig.spec.output.to, buildConfig.metadata.namespace)

            $scope.options = {
              pickedBuildFromType: buildFrom.kind,
              pickedBuildFromNamespace: buildFrom.namespace,
              pickedBuildFromImageStream: buildFrom.name.split(":")[0],
              pickedBuildFromImageStreamTag: buildFrom.name.split(":")[1],

              pickedPushToNamespace: buildConfig.spec.output.to.namespace || buildConfig.metadata.namespace,              
              pickedPushToImageStream: buildConfig.spec.output.to.name.split(":")[0],
              pickedPushToImageStreamTag: buildConfig.spec.output.to.name.split(":")[1],

              // selectedBuildFromType: buildFrom.kind,
              // selectedBuildFromNamespace: buildFrom.namespace,
              // selectedBuildFromImageStream: buildFrom.name.split(":")[0],
              // selectedBuildFromImageStreamTag: buildFrom.name.split(":")[1],

              pickedBuildFromDockerImage: "",
            };


            $scope.available.projects = ["openshift"];
            DataService.list("projects", $scope, function(projects) {
              var projects = projects.by("metadata.name");
              for (var name in projects) $scope.available.projects.push(name);
              $scope.updateImageStreams($scope.options.pickedBuildFromNamespace, $scope.options.pickedBuildFromImageStream, $scope.options.pickedBuildFromImageStreamTag);
              console.log("asd")
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

    $scope.updateImageStreams = function(project, imageStream, tag) {
      $scope.options.pickedBuildFromImageStreamTag = tag;
      $scope.options.pickedBuildFromImageStream = imageStream;
      DataService.list("imagestreams", {namespace: project}, function(imageStreams) {
        $scope.available.imageStreams = [];
        $scope.available.tags = {};

        var projectImageStreams = imageStreams.by("metadata.name");
        angular.forEach(projectImageStreams, function(value, key) {
          $scope.available.imageStreams.push(key);
          var tagList = [];
          value.status.tags.forEach(function(item){ 
            tagList.push(item["tag"]); 
          });
          $scope.available.tags[key] = tagList;
        });
      });
      console.log($scope.options.pickedBuildFromNamespace);
      console.log($scope.options.pickedBuildFromImageStream);
      console.log($scope.options.pickedBuildFromImageStreamTag);
      console.log("-----------");
    }

    $scope.updateTags = function(imageStream) {
      $scope.options.pickedBuildFromImageStreamTag = "";
      console.log($scope.options.pickedBuildFromNamespace);
      console.log($scope.options.pickedBuildFromImageStream);
      console.log($scope.options.pickedBuildFromImageStreamTag);
      console.log("-----------");
    }


  });