'use strict';

/**
 * @ngdoc function
 * @name openshiftConsole.controller:ServicesController
 * @description
 * # UpdateModalController
 * Controller of the openshiftConsole
 */
angular.module('openshiftConsole')
  .controller('UpdateModalController', function ($scope, $uibModalInstance, AlertMessageService) {
    $scope.update = function() {
      $uibModalInstance.close('update');
    };

    $scope.cancel = function() {
      $uibModalInstance.dismiss('cancel');
    };
  });
