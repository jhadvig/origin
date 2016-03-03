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
          var file;
          if (scope.file) {
            file = scope.file;
          } else {
            file = $('input[type=file]', this)[0].files[0];
          }
          var reader = new FileReader();
          reader.onloadend = function(){
            scope.$apply(function(){
              scope.fileName = file.name;
              scope.model = reader.result;
            });
          };
          reader.onerror = function(e){
            scope.supportsFileUpload = false;
            scope.uploadError = true;
            Logger.error("Could not read file", e);
          };
          reader.readAsText(file);
        });
        $('.btn-file').on('click', function() {
          delete scope.file;
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
              scope.file = e.originalEvent.dataTransfer.files[0];
              $(element).trigger('change');
            }
          }
          return false;
        });
      }
    };
  });
