'use strict';

angular.module('openshiftConsole')
  .directive('oscFileInput', function(Logger) {
    return {
      restrict: 'E',
      scope: {
        model: "=",
        required: "=",
        disabled: "=ngDisabled",
        helpText: "@?",
        fileExtension: "=?",
        dropArea: "@?"
      },
      templateUrl: 'views/directives/osc-file-input.html',
      link: function(scope, element){
        scope.helpID = _.uniqueId('help-');
        scope.supportsFileUpload = (window.File && window.FileReader && window.FileList && window.Blob);
        scope.uploadError = false;
        var dropArea = element;
        if (scope.dropArea) {
          dropArea = scope.dropArea;
        }
        $(element).change(function(){
          var file = $('input[type=file]', this)[0].files[0];
          var reader = new FileReader();
          reader.onloadend = function(){
            scope.$apply(function(){
              scope.fileName = file.name;
              scope.model = reader.result;
              var splittedName = file.name.split(".");
              scope.fileExtension = splittedName[splittedName.length-1];
            });
          };
          reader.onerror = function(e){
            scope.supportsFileUpload = false;
            scope.uploadError = true;
            Logger.error("Could not read file", e);
          };
          reader.readAsText(file);
        });
        $(dropArea).on('dragover', function(e) {
          e.preventDefault();
          e.stopPropagation();
        });
        $(dropArea).on('dragenter', function(e) {
            e.preventDefault();
            e.stopPropagation();
        });
        $(dropArea).on('drop', function(e) {
          e.preventDefault();
          e.stopPropagation();
          if (e.originalEvent.dataTransfer){
            if (e.originalEvent.dataTransfer.files.length > 0) {
              $('input[type=file]', this)[0].files = e.originalEvent.dataTransfer.files;
            }
          }
          return false;
        });
      }
    };
  });
