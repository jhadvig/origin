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

        var dropArea = scope.dropArea || element,
        isDisabled = scope.disabled || false,
        highlightDropArea = false,
        showDropArea = false,
        timeout = -1,
        highlightTimeout = -1;

        scope.$watch('disabled', function() {
          isDisabled = scope.disabled;
        }, true);

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
        $(document).on('dragenter', function(e) {
          e.preventDefault();
          e.stopPropagation();
          if (isDisabled) {
            return false;
          }
          $(dropArea).addClass('show-drag-and-drop-area');
          showDropArea = true;
        });
        $(document).on('dragover', function(e) {
          e.preventDefault();
          e.stopPropagation();
          if (isDisabled) {
            return false;
          }
          showDropArea = true;
        });
        $(document).on('dragleave', function(e) {
          e.preventDefault();
          e.stopPropagation();
          if (isDisabled) {
            return false;
          }
          showDropArea = false;
          clearTimeout( timeout );
          timeout = setTimeout( function(){
              if( !showDropArea ){ $(dropArea).removeClass('show-drag-and-drop-area'); }
          }, 200 );
        });
        $(document).on('drop', function(e) {
          e.preventDefault();
          e.stopPropagation();
          return false;
        });
        $(dropArea).on('dragenter', function() {
          $(dropArea).addClass('highlight-drag-and-drop-area');
          highlightDropArea = true;
        });
        $(dropArea).on('dragover', function() {
          highlightDropArea = true;
        });
        $(dropArea).on('dragleave', function() {
          highlightDropArea = false;
          clearTimeout( timeout );
          highlightTimeout = setTimeout( function(){
              if( !highlightDropArea ){ $(dropArea).removeClass('highlight-drag-and-drop-area'); }
          }, 100 );
        });
        $(dropArea).on('drop', function(e) {
          e.preventDefault();
          e.stopPropagation();
          if (isDisabled) {
            return false;
          }
          var files = _.get(e, 'originalEvent.dataTransfer.files', []);
          if (files.length > 0 ) {
            scope.file = e.originalEvent.dataTransfer.files[0];
            $(element).trigger('change');
          }
          angular.forEach($('.show-drag-and-drop-area'), function(el) {
            el.classList.remove('show-drag-and-drop-area');
            el.classList.remove('highlight-drag-and-drop-area');
          });
          return false;
        });
      }
    };
  });
