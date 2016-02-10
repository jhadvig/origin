'use strict';

angular.module('openshiftConsole')
  .directive('events', function($routeParams, $filter, DataService, ProjectsService, Logger) {
    return {
      restrict: 'E',
      scope: {
        resourceKind: "@",
        resourceName: "@"
      },
      templateUrl: 'views/directives/events.html',
      controller: function($scope){

        var filterEvent = function(event) {
          var resourceName = $scope.resourceName;
          var resourceKind = $scope.resourceKind;
          // For Build resource watch events from the builder Pod
          if (resourceKind === "Build") {
            resourceKind = "Pod";
            resourceName = $scope.$parent.build.metadata.annotations["openshift.io/build.pod-name"];
          } else if (resourceKind === "Deployment") {
            // For Deployment resource watch ReplicationController events
            resourceKind = "ReplicationController";
          }
          return (event.involvedObject.kind === resourceKind) && (event.involvedObject.name === resourceName);
        };

        var watches = [];
        watches.push(DataService.watch("events", $scope.$parent.logContext, function(events) {
          $scope.emptyMessage = "No events to show";
          $scope.eventsArray = $filter('toArray')(events.by("metadata.name"));
          $scope.filteredEvents = _.filter($scope.eventsArray, function(event) { 
            return filterEvent(event);
          });
          Logger.log("events (subscribe)", $scope.events);
        }));

        $scope.$on('$destroy', function(){
          DataService.unwatchAll(watches);
        });

      },
    };
  });
