'use strict';

angular.module("openshiftConsole")
  .factory("BuildConfigsService", function(ProjectsService){
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
    }

    BuildConfigsService.prototype.getTriggerMap = function(triggerMap, triggers) {
      angular.forEach(triggers, function(value, key) {
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
    }

    BuildConfigsService.prototype.getSourceMap = function(sourceMap, sources) {
      angular.forEach(sources, function(value, key) {
        switch (key) {
          case "binary":
            sourceMap.binary = true;
            break
          case "dockerfile":
            sourceMap.dockerfile = true;
            break
          case "git":
            sourceMap.git = true;
            break
          case "image":
            sourceMap.image = true;
            break
          case "contextDir":
            sourceMap.contextDir = true;
            break
        }
      });
      return sourceMap;
    }

    BuildConfigsService.prototype.setBuildFromVariables = function(optionModel, type, ns, is, ist, di) {
      optionModel.pickedBuildFromType = type;
      optionModel.pickedBuildFromNamespace = ns;
      optionModel.pickedBuildFromImageStream = is;
      optionModel.pickedBuildFromImageStreamTag = ist;
      optionModel.pickedBuildFromDockerImage = di;
      return optionModel;
    }

    BuildConfigsService.prototype.setPushToVariables = function(optionModel, type, ns, is, ist, di) {
      optionModel.pickedPushToType = type;
      optionModel.pickedPushToNamespace = ns;
      optionModel.pickedPushToImageStream = is;
      optionModel.pickedPushToImageStreamTag = ist;
      optionModel.pickedPushToDockerImage = di;
      return optionModel;
    }

    return new BuildConfigsService();
  });