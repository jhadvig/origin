'use strict';

angular.module("openshiftConsole")
  .factory('BuildConfigsService', function(DataService, $filter){
    function BuildConfigsService() {}

    BuildConfigsService.prototype.canBuild = function($scope) {
      if (!$scope.buildConfig) {
        return false;
      }

      if ($scope.buildConfig.metadata.deletionTimestamp) {
        return false;
      }

      if ($scope.buildConfigBuildsInProgress &&
          $filter('hashSize')($scope.buildConfigBuildsInProgress[$scope.buildConfig.metadata.name]) > 0) {
        return false;
      }

      if ($scope.buildConfig.metadata.annotations["openshift.io/build-config.paused"] === "true") {
        return false;
      }

      return true;
    };

    return new BuildConfigsService();
  });
