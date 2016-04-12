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
        var dropArea = scope.dropArea || element;
        // var body = $('body');
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
        $(dropArea).on('dragleave',function(e) {
          console.log("leave");
          e.preventDefault();
          e.stopPropagation();
          $(this).removeClass('show-drag-and-drop-area');
        });
        $(dropArea).on('dragover', function(e) {
          e.preventDefault();
          e.stopPropagation();
          $(this).addClass('show-drag-and-drop-area');
        });
        $(dropArea).on('dragenter', function(e) {
          e.preventDefault();
          e.stopPropagation();
        });
        $(dropArea).on('drop', function(e) {
          e.preventDefault();
          e.stopPropagation();
          var files = _.get(e, 'originalEvent.dataTransfer.files', []);
          if (files.length > 0 ) {
            scope.file = e.originalEvent.dataTransfer.files[0];
            $(element).trigger('change');
          }
          $(this).removeClass('show-drag-and-drop-area');
          return false;
        });
      }
    };
  });
