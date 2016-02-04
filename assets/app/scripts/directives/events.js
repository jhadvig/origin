'use strict';

angular.module('openshiftConsole')
  .directive('eventIcon', function() {
    return {
      restrict: "E",
      scope: false,
      template: '<span class="pficon {{type | eventIcon}}" aria-hidden="true" data-toggle="tooltip" data-placement="right"' +
                'data-original-title="{{type}}"></span>',
      replace: true,
      link: function(scope, element, attrs) {
        scope.type = attrs.type;
      }
    };
  })
  .directive('events', function($routeParams, $filter, DataService, ProjectsService, Logger) {
    return {
      restrict: 'E',
      scope: {
        kind: "@",
      },
      templateUrl: 'views/directives/events.html',
      controller: function($scope){

        var filterEvent = function(resourceType, resource, event) {
          var involvedObjectKind;
          if (resourceType === "pod") {
            involvedObjectKind = "Pod";
          } else if (resourceType === "service") {
            involvedObjectKind = "Service";
          } else if (resourceType === "deploymentconfig") {
            involvedObjectKind = "DeploymentConfig";
          } else if (resourceType === "deployment") {
            involvedObjectKind = "ReplicationController";
          } else if (resourceType === "build") {
            involvedObjectKind = "Pod";
            return (event.involvedObject.kind === involvedObjectKind) && (event.involvedObject.name === resource + "-build");
          }
          return (event.involvedObject.kind === involvedObjectKind) && (event.involvedObject.name === resource);
        };

        var watches = [];
        ProjectsService
          .get($routeParams.project)
          .then(_.spread(function(project, context) {
            $scope.project = project;
            watches.push(DataService.watch("events", context, function(events) {
              $scope.emptyMessage = "No events to show";
              $scope.eventsArray = $filter('toArray')(events.by("metadata.name"));
              $scope.filteredEvents = _.filter($scope.eventsArray, function(event) { 
                return filterEvent($scope.kind, $routeParams[$scope.kind], event);
              });
              Logger.log("events (subscribe)", $scope.events);
            }));

            $scope.$on('$destroy', function(){
              DataService.unwatchAll(watches);
            });
          }));
      },
    };
  });
