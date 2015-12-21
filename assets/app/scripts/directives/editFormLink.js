'use strict';

angular.module("openshiftConsole")
  .directive("editFormLink", function ($modal, Logger) {
    return {
      restrict: "E",
      scope: {
        resource: "=",
        kind: "@",
        alerts: "=?"
      },
      templateUrl: "views/directives/edit-form-link.html",
      replace: true,
      link: function(scope) {
        scope.openEditModal = function() {
          // Clear any previous edit success message to avoid confusion if the edit is cancelled
          if (scope.alerts) {
            delete scope.alerts['edit-yaml'];
          }

          var modalInstance = $modal.open({
            animation: true,
            templateUrl: 'views/modals/edit-resource.html',
            controller: 'EditModalController',
            scope: scope,
            size: 'lg',
            backdrop: 'static' // don't close modal and lose edits when clicking backdrop
          });
          modalInstance.result.then(function(result) {
            if (scope.alerts) {
              switch (result) {
              case 'no-changes':
                scope.alerts['edit-yaml'] = {
                  type: "warning",
                  message: "There were no changes to " + scope.resource.metadata.name + " to save. Edit cancelled."
                };
                break;
              case 'save':
                scope.alerts['edit-yaml'] = {
                  type: "success",
                  message: scope.resource.metadata.name + " was updated."
                };
                break;
              default:
                Logger.warn('Unknown edit modal result: ' + result);
              }
            }
          });
        };
      }
    };
  });

