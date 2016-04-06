'use strict';

/**
 * @ngdoc function
 * @name openshiftConsole.controller:ServicesController
 * @description
 * # ProcessTemplateModalController
 * Controller of the openshiftConsole
 */
angular.module('openshiftConsole')
  .controller('ProcessTemplateModalController', function ($scope, $uibModalInstance) {
    $scope.create = function() {
      $uibModalInstance.close('create');
    };

    $scope.cancel = function() {
      $uibModalInstance.dismiss('cancel');
    };
  });
