'use strict';

angular.module("openshiftConsole")
  .factory("BuildConfigsService", function(){
    function BuildConfigsService() {}

    BuildConfigsService.prototype.updateEnvVars = function(envVars) {
      var updatedEnvVars = [];
      angular.forEach(envVars, function(v, k) {
        var env = {
          name: k,
          value: v
        };
        updatedEnvVars.push(env);
      });
      return updatedEnvVars;
    };

    BuildConfigsService.prototype.getTriggerMap = function(triggerMap, triggers) {
      triggers.forEach(function(value) {
        switch (value.type) {
          case "Generic":
            break;
          case "GitHub":
            triggerMap.webhook = true;
            break;
          case "ImageChange":
            triggerMap.imageChange = true;
            break;
          case "ConfigChange":
            triggerMap.configChange = true;
            break;
        }
      });
      return triggerMap;
    };

    BuildConfigsService.prototype.getSourceMap = function(sourceMap, sources) {
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

    BuildConfigsService.prototype.setBuildFromVariables = function(optionsModel, type, ns, is, ist, isi, di) {
      optionsModel.pickedBuildFromType = type;
      optionsModel.pickedBuildFromNamespace = ns;
      optionsModel.pickedBuildFromImageStream = is;
      optionsModel.pickedBuildFromImageStreamTag = ist;
      optionsModel.pickedBuildFromImageStreamImage = isi;
      optionsModel.pickedBuildFromDockerImage = di;
      return optionsModel;
    };

    BuildConfigsService.prototype.setPushToVariables = function(optionsModel, type, ns, is, ist, di) {
      optionsModel.pickedPushToType = type;
      optionsModel.pickedPushToNamespace = ns;
      optionsModel.pickedPushToImageStream = is;
      optionsModel.pickedPushToImageStreamTag = ist;
      optionsModel.pickedPushToDockerImage = di;
      return optionsModel;
    };

    return new BuildConfigsService();
  });