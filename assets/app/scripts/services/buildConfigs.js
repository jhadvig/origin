'use strict';

angular.module("openshiftConsole")
  .factory("BuildConfigsService", function(){
    function BuildConfigsService() {}

    BuildConfigsService.prototype.arrayToOptions = function(array) {
    	var optionsArray = [];
    	array.forEach(function(item){
    		optionsArray.push({
    			title: item
    		});
    	});
      return optionsArray;
    };

    BuildConfigsService.prototype.initializeOptions = function(buildConfig, optionType, $scope) {
      switch (optionType) {
        case "buildeImager":
          break;
        case "outputImage":
          var pushTo = $scope.updatedBuildConfig.spec.output.to;

          $scope.options.loadedPushToType = pushTo.kind;
          $scope.options.loadedPushToNamespace = pushTo.namespace || buildConfig.metadata.namespace;
          $scope.options.loadedPushToImageStream = pushTo.name.split(":")[0];
          $scope.options.loadedPushToImageStreamTag = pushTo.name.split(":")[1];
          $scope.options.loadedPushToDockerImage = (pushTo.kind === "ImageStreamTag") ? buildConfig.metadata.namespace + "/" + pushTo.name : pushTo.name;
          break;
        case "ImageSource":
          break;
      }
    };

    return new BuildConfigsService();
  });