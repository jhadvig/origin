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

            $scope.options = {
              pickedBuildFromType: buildFrom.kind,
              pickedBuildFromNamespace: buildFrom.namespace,
              pickedBuildFromImageStream: buildFrom.name.split(":")[0],
              pickedBuildFromImageStreamTag: buildFrom.name.split(":")[1],
              pickedProjectImageStreams: {},
              pickedBuildFromDockerImage: "",
              // imageStreamIndex: 0,
              // imageStreamTagIndex: 0,
            };


            $scope.available.projects = ["openshift"];
            DataService.list("projects", $scope, function(projects) {
              var projects = projects.by("metadata.name");
              for (var name in projects) $scope.available.projects.push(name);
              $scope.updateImageStreams($scope.options.pickedBuildFromNamespace, $scope.options.pickedBuildFromImageStream);
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

    $scope.updateImageStreams = function(ns, pickedImageStream) {
      DataService.list("imagestreams", {namespace: ns}, function(imageStreams) {
        $scope.available.imageStreams = [];
        $scope.available.tags = [];

        var projectImageStreams = imageStreams.by("metadata.name");
        angular.forEach(projectImageStreams, function(value, key) {
          $scope.available.imageStreams.push(key);
          var tagList = [];
          value.status.tags.forEach(function(item){ 
            tagList.push(item["tag"]); 
          });
          $scope.available.tags[key] = tagList;
          console.log("ads");
        });

        // for(var imageStream in projectImageStreams) {

        //   $scope.available.projects[ns] = {imageStream: {}};
        // }
        console.log("ads");
        // projectImageStreams[$scope.options.pickedBuildFromImageStream]
      });
    }


  });