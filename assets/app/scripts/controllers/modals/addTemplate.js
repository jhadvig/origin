'use strict';
/* jshint unused: false */

/**
 * @ngdoc function
 * @name openshiftConsole.controller:AddTemplateController
 * @description
 * # AddTemplateController
 * Controller of the openshiftConsole
 */
angular.module('openshiftConsole')
  .controller('AddTemplateController', function ($scope, $filter, $uibModalInstance, DataService) {

    $scope.aceLoaded = function(editor) {
      var session = editor.getSession();
      session.setOption('tabSize', 2);
      session.setOption('useSoftTabs', true);

      // Resize the editor based on window height.
      var updateEditorHeight = function() {
        var headerHeight = $('.modal-template-add .modal-header').outerHeight();
        var footerHeight = $('.modal-template-add .modal-footer').outerHeight();
        var availableHeight = window.innerHeight - headerHeight - footerHeight;

        // Use 50% of available height. min-height set in CSS.
        var editorHeight = Math.floor(availableHeight * 0.50);

        // Animate the change so it's not janky.
        $('.modal-template-add .editor').animate({
          height: editorHeight + 'px'
        }, 30, function() {
          editor.resize();
        });
      };

      setTimeout(updateEditorHeight, 10);

      var onResize = _.debounce(updateEditorHeight, 200);
      $(window).resize(onResize);
      $scope.$on('$destroy', function() {
        // Stop listening for resize events.
        $(window).off('resize', onResize);
      });
    };

    $scope.create = function() {
      console.log($scope.newTemplate);

      var newResources;
      try {
        newResources = YAML.parse($scope.newTemplate);
      } catch (e) {
        $scope.error = e;
        return;
      }

      if (_.isEqual(newResources, null)) {
        $uibModalInstance.close('no-template');
        return;
      }

      DataService.create('templates', 'kar', newResources, {
        namespace: $scope.projectName
      }).then(
        // success
        function() {
          $uibModalInstance.close('save');
        },
        // failure
        function(result) {
          $scope.error = {
            message: $filter('getErrorDetails')(result)
          };
        }
      );
    };

    $scope.cancel = function() {
      $uibModalInstance.dismiss('cancel');
    };
  });
